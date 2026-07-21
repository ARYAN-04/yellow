package draw

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"GoTabs/internal/db"
	"GoTabs/internal/models"

	"github.com/google/uuid"
)

// GenerateDraw runs the bracket power pairing, side-balancing, and adjudicator allocation for a round.
func GenerateDraw(store db.TournamentStore, roundID string) error {
	// 1. Get round info
	round, err := store.GetRound(roundID)
	if err != nil {
		return fmt.Errorf("round not found: %w", err)
	}
	seq := round.Seq

	// 2. Resolve side configuration from config (Sides as Data constraint)
	sidesStr, err := store.GetSidesConfig()
	if err != nil {
		if err == sql.ErrNoRows {
			// Default to 4-team BP style
			sidesStr = "OG,OO,CG,CO"
		} else {
			return fmt.Errorf("failed to read sides configuration: %w", err)
		}
	}
	sides := strings.Split(sidesStr, ",")
	numSides := len(sides)

	// 3. Fetch all teams
	teamsDraw, err := store.GetTeamsForDraw()
	if err != nil {
		return fmt.Errorf("failed to query teams: %w", err)
	}

	if len(teamsDraw) == 0 {
		return fmt.Errorf("no teams registered in the tournament")
	}
	if len(teamsDraw)%numSides != 0 {
		return fmt.Errorf("number of teams (%d) is not a multiple of number of sides (%d)", len(teamsDraw), numSides)
	}

	teamInstMap := make(map[string]string)
	for _, t := range teamsDraw {
		if t.InstitutionID != "" {
			teamInstMap[t.ID] = t.InstitutionID
		}
	}

	numDebates := len(teamsDraw) / numSides

	// 4. Group teams into pairings
	var pairings [][]models.TeamDrawInfo
	rSource := rand.New(rand.NewSource(time.Now().UnixNano()))

	if seq == 1 {
		// Round 1: Randomize and pair
		shuffled := make([]models.TeamDrawInfo, len(teamsDraw))
		copy(shuffled, teamsDraw)
		rSource.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		for i := 0; i < numDebates; i++ {
			pairings = append(pairings, shuffled[i*numSides:(i+1)*numSides])
		}
	} else {
		// Power Pairing: Group teams by cumulative wins/points
		teamPoints, err := store.GetConfirmedPoints()
		if err != nil {
			return fmt.Errorf("failed to query team points: %w", err)
		}

		// Fill in zero points for teams without any ballot results yet
		for _, t := range teamsDraw {
			if _, ok := teamPoints[t.ID]; !ok {
				teamPoints[t.ID] = 0
			}
		}

		// Sort teams by points
		sort.Slice(teamsDraw, func(i, j int) bool {
			if teamPoints[teamsDraw[i].ID] == teamPoints[teamsDraw[j].ID] {
				return rSource.Float64() < 0.5 // Random tie break
			}
			return teamPoints[teamsDraw[i].ID] > teamPoints[teamsDraw[j].ID]
		})

		for i := 0; i < numDebates; i++ {
			pairings = append(pairings, teamsDraw[i*numSides:(i+1)*numSides])
		}
	}

	// 5. Assign sides within each pairing using the Hungarian Algorithm
	history, err := store.GetSideHistory(seq)
	if err != nil {
		return fmt.Errorf("failed to query side history: %w", err)
	}

	var debatesToSave []models.DebateDrawInput

	for debateIdx, pair := range pairings {
		// Construct cost matrix (N x N where N = numSides)
		costMatrix := make([][]int, numSides)
		for i := 0; i < numSides; i++ {
			costMatrix[i] = make([]int, numSides)
			for j := 0; j < numSides; j++ {
				// Cost is the number of times this team has played this side
				costMatrix[i][j] = history[models.SideHistKey{TeamID: pair[i].ID, Side: sides[j]}]
			}
		}

		// Solve assignment problem
		assignment := SolveHungarian(costMatrix)

		// Save debate record
		debateID := uuid.New().String()
		venue := fmt.Sprintf("Room %d", debateIdx+1)

		var debateTeams []models.TeamAssignment
		// Save debate team assignments
		for teamIdx, sideIdx := range assignment {
			debateTeams = append(debateTeams, models.TeamAssignment{
				TeamID: pair[teamIdx].ID,
				Side:   sides[sideIdx],
			})
		}

		debatesToSave = append(debatesToSave, models.DebateDrawInput{
			DebateID:     debateID,
			Venue:        venue,
			Teams:        debateTeams,
			Adjudicators: []models.AdjudicatorAssignment{},
		})
	}

	// 6. Adjudicator Allocation
	adjudicators, err := store.GetAdjudicatorsForDraw()
	if err != nil {
		return fmt.Errorf("failed to query adjudicators: %w", err)
	}

	// Sort adjudicators by score descending (greedy panels placement)
	sort.Slice(adjudicators, func(i, j int) bool {
		return adjudicators[i].Score > adjudicators[j].Score
	})

	// Allocate chairs first, keeping clash avoidance in mind
	usedAdjs := make(map[string]bool)

	for idx := range debatesToSave {
		d := &debatesToSave[idx]

		// Allocate a conflict-free chair
		var chairID string
		for _, adj := range adjudicators {
			if usedAdjs[adj.ID] {
				continue
			}

			// Conflict Check: Adjudicator institution matches any team's institution in this debate
			hasConflict := false
			for _, t := range d.Teams {
				teamInst := teamInstMap[t.TeamID]
				if adj.InstitutionID != "" && teamInst != "" && adj.InstitutionID == teamInst {
					hasConflict = true
					break
				}
			}

			if !hasConflict {
				chairID = adj.ID
				usedAdjs[adj.ID] = true
				break
			}
		}

		if chairID != "" {
			d.Adjudicators = append(d.Adjudicators, models.AdjudicatorAssignment{
				AdjudicatorID: chairID,
				Role:          "chair",
			})
		}
	}

	// 7. Save draw in store
	err = store.SaveDraw(roundID, debatesToSave)
	if err != nil {
		return fmt.Errorf("failed to save draw: %w", err)
	}

	return nil
}
