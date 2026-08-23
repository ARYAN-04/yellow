package draw

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"yellow/internal/db"
	"yellow/internal/models"

	"github.com/google/uuid"
)

// selectChair picks the best unused adjudicator for a debate: adjudicators with
// hard conflicts (declared or institutional) are skipped; remaining candidates
// are ranked by fewest soft conflicts, then by score descending.
func selectChair(adjudicators []models.AdjDrawInfo, usedAdjs map[string]bool, cIdx *ConflictIndex, teamInstMap map[string]string, teams []models.TeamAssignment) string {
	var candidates []models.AdjDrawInfo
	for _, adj := range adjudicators {
		if usedAdjs[adj.ID] {
			continue
		}
		if cIdx.adjHasHard(adj, teamInstMap, teams) {
			continue
		}
		candidates = append(candidates, adj)
	}
	if len(candidates) == 0 {
		return ""
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return cIdx.adjSoftCount(candidates[i], teamInstMap, teams) < cIdx.adjSoftCount(candidates[j], teamInstMap, teams)
	})
	return candidates[0].ID
}

// resolveSides reads the side configuration from tournament config (Sides as
// Data constraint), defaulting to BP-style sides when unset.
func resolveSides(store db.TournamentStore) ([]string, int, error) {
	sidesStr, err := store.GetSidesConfig()
	if err == sql.ErrNoRows {
		sidesStr = "OG,OO,CG,CO"
	} else if err != nil {
		return nil, 0, fmt.Errorf("failed to read sides configuration: %w", err)
	}
	sides := strings.Split(sidesStr, ",")
	return sides, len(sides), nil
}

// assignSides assigns the sides of one pairing via the Hungarian algorithm,
// minimizing repeated side history across prior rounds.
func assignSides(pair []models.TeamDrawInfo, sides []string, history map[models.SideHistKey]int) []models.TeamAssignment {
	costMatrix := make([][]int, len(pair))
	for i := range costMatrix {
		costMatrix[i] = make([]int, len(sides))
		for j := range sides {
			costMatrix[i][j] = history[models.SideHistKey{TeamID: pair[i].ID, Side: sides[j]}]
		}
	}

	assignment := SolveHungarian(costMatrix)
	teams := make([]models.TeamAssignment, 0, len(pair))
	for teamIdx, sideIdx := range assignment {
		teams = append(teams, models.TeamAssignment{TeamID: pair[teamIdx].ID, Side: sides[sideIdx]})
	}
	return teams
}

// buildDebates converts resolved pairings into savable debate inputs with
// Hungarian-assigned sides and pull-up flags applied per team.
func buildDebates(pairings [][]models.TeamDrawInfo, pullFlags [][]bool, sides []string, history map[models.SideHistKey]int) []models.DebateDrawInput {
	out := make([]models.DebateDrawInput, 0, len(pairings))
	for debateIdx, pair := range pairings {
		teams := assignSides(pair, sides, history)
		for i := range teams {
			teams[i].PullUp = pullFlags[debateIdx][i]
		}
		out = append(out, models.DebateDrawInput{
			DebateID: uuid.New().String(),
			Venue:    fmt.Sprintf("Room %d", debateIdx+1),
			Teams:    teams,
		})
	}
	return out
}

// drawContext bundles everything the draw pipeline needs for one round.
type drawContext struct {
	seq             int
	roundOne        bool
	sides           []string
	numSides        int
	active          []models.TeamDrawInfo
	points          map[string]int
	cIdx            *ConflictIndex
	teamInstMap     map[string]string
	unavailableAdjs map[string]bool
	rnd             *rand.Rand
}

// loadUnavailableSets reads the round's availability overrides and returns the
// IDs of unavailable teams and adjudicators. An empty override list means no
// availability tracking is in effect for the round.
func loadUnavailableSets(store db.TournamentStore, roundID string) (teams map[string]bool, adjs map[string]bool, err error) {
	overrides, err := store.GetRoundAvailability(roundID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query round availability: %w", err)
	}
	teams = make(map[string]bool)
	adjs = make(map[string]bool)
	for _, o := range overrides {
		if o.IsAvailable {
			continue
		}
		switch o.EntityType {
		case "team":
			teams[o.EntityID] = true
		case "adjudicator":
			adjs[o.EntityID] = true
		}
	}
	return teams, adjs, nil
}

// prepareDrawContext loads the round, resolves sides and points, applies the
// standby admission rules, and indexes conflicts for pairing.
func prepareDrawContext(store db.TournamentStore, roundID string) (*drawContext, error) {
	round, err := store.GetRound(roundID)
	if err != nil {
		return nil, fmt.Errorf("round not found: %w", err)
	}
	sides, numSides, err := resolveSides(store)
	if err != nil {
		return nil, err
	}

	teams, err := store.GetTeamsForDraw()
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams registered in the tournament")
	}

	unavailableTeams, unavailableAdjs, err := loadUnavailableSets(store, roundID)
	if err != nil {
		return nil, err
	}
	if len(unavailableTeams) > 0 {
		available := make([]models.TeamDrawInfo, 0, len(teams))
		for _, t := range teams {
			if !unavailableTeams[t.ID] {
				available = append(available, t)
			}
		}
		teams = available
	}

	points, err := store.GetConfirmedPoints()
	if err != nil {
		return nil, fmt.Errorf("failed to query team points: %w", err)
	}
	for _, t := range teams {
		if _, ok := points[t.ID]; !ok {
			points[t.ID] = 0
		}
	}

	conflicts, err := store.GetConflictsForDraw()
	if err != nil {
		return nil, fmt.Errorf("failed to query conflicts: %w", err)
	}

	teamInstMap := make(map[string]string)
	for _, t := range teams {
		if t.InstitutionID != "" {
			teamInstMap[t.ID] = t.InstitutionID
		}
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	sortByPoints(teams, points, rnd)

	return &drawContext{
		seq:             round.Seq,
		roundOne:        round.Seq == 1,
		sides:           sides,
		numSides:        numSides,
		active:          SelectActiveTeams(teams, numSides),
		points:          points,
		cIdx:            BuildConflictIndex(conflicts),
		teamInstMap:     teamInstMap,
		unavailableAdjs: unavailableAdjs,
		rnd:             rnd,
	}, nil
}

// GenerateDraw runs bracket formation (with standby admission and surplus
// sit-outs), conflict resolution, side balancing, pull-up detection, and
// strength-balanced panel allocation for a round.
func GenerateDraw(store db.TournamentStore, roundID string) error {
	ctx, err := prepareDrawContext(store, roundID)
	if err != nil {
		return err
	}

	pairings := buildPairings(ctx.active, ctx.numSides, ctx.roundOne, ctx.rnd)
	resolveTeamPairConflicts(pairings, ctx.cIdx, 3)
	pullFlags := DetectPullUps(pairings, ctx.points, !ctx.roundOne)

	history, err := store.GetSideHistory(ctx.seq)
	if err != nil {
		return fmt.Errorf("failed to query side history: %w", err)
	}
	debatesToSave := buildDebates(pairings, pullFlags, ctx.sides, history)

	adjudicators, err := store.GetAdjudicatorsForDraw()
	if err != nil {
		return fmt.Errorf("failed to query adjudicators: %w", err)
	}
	if len(ctx.unavailableAdjs) > 0 {
		available := make([]models.AdjDrawInfo, 0, len(adjudicators))
		for _, a := range adjudicators {
			if !ctx.unavailableAdjs[a.ID] {
				available = append(available, a)
			}
		}
		adjudicators = available
	}
	sort.Slice(adjudicators, func(i, j int) bool {
		return adjudicators[i].Score > adjudicators[j].Score
	})

	order := debateImportanceOrder(debatesToSave, ctx.points, ctx.roundOne, ctx.rnd)
	AllocatePanels(order, debatesToSave, adjudicators, ctx.cIdx, ctx.teamInstMap)

	if err := store.SaveDraw(roundID, debatesToSave); err != nil {
		return fmt.Errorf("failed to save draw: %w", err)
	}
	return nil
}
