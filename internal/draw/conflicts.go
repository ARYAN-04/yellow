package draw

import (
	"yellow/internal/models"
)

// Conflict penalty weights for pairing costs.
const (
	penaltyHard = 100
	penaltySoft = 1
)

// ConflictIndex provides O(1) lookups of declared conflicts for the draw.
type ConflictIndex struct {
	pairHard map[string]bool
	pairSoft map[string]bool
	adjHard  map[string]map[string]bool
	adjSoft  map[string]map[string]bool
}

func pairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func addAdjTarget(m map[string]map[string]bool, adjID, targetID string) {
	if m[adjID] == nil {
		m[adjID] = make(map[string]bool)
	}
	m[adjID][targetID] = true
}

// BuildConflictIndex indexes conflicts for team pairing and adjudicator checks.
func BuildConflictIndex(conflicts []models.Conflict) *ConflictIndex {
	ci := &ConflictIndex{
		pairHard: make(map[string]bool),
		pairSoft: make(map[string]bool),
		adjHard:  make(map[string]map[string]bool),
		adjSoft:  make(map[string]map[string]bool),
	}

	for _, c := range conflicts {
		hard := c.Weight == "hard"
		switch {
		case c.SubjectType == "team" && c.TargetType == "team":
			key := pairKey(c.SubjectID, c.TargetID)
			if hard {
				ci.pairHard[key] = true
			} else {
				ci.pairSoft[key] = true
			}
		case c.SubjectType == "adjudicator" && c.TargetType != "adjudicator":
			if hard {
				addAdjTarget(ci.adjHard, c.SubjectID, c.TargetID)
			} else {
				addAdjTarget(ci.adjSoft, c.SubjectID, c.TargetID)
			}
		}
	}
	return ci
}

// teamPairPenalty returns the conflict penalty between two teams: hard=100, soft=1, none=0.
func (ci *ConflictIndex) teamPairPenalty(a, b string) int {
	key := pairKey(a, b)
	switch {
	case ci.pairHard[key]:
		return penaltyHard
	case ci.pairSoft[key]:
		return penaltySoft
	}
	return 0
}

// debatePenalty sums pairwise team penalties within one debate's pairing.
func (ci *ConflictIndex) debatePenalty(teams []models.TeamDrawInfo) int {
	total := 0
	for i := range teams {
		for j := i + 1; j < len(teams); j++ {
			total += ci.teamPairPenalty(teams[i].ID, teams[j].ID)
		}
	}
	return total
}

// adjHasHard reports whether an adjudicator has a hard conflict against any
// team in the debate, either declared or via matching institutions.
func (ci *ConflictIndex) adjHasHard(adj models.AdjDrawInfo, teamInst map[string]string, teams []models.TeamAssignment) bool {
	targets := ci.adjHard[adj.ID]
	for _, t := range teams {
		if targets != nil && targets[t.TeamID] {
			return true
		}
		inst := teamInst[t.TeamID]
		if inst != "" && (targets[inst] || (adj.InstitutionID != "" && adj.InstitutionID == inst)) {
			return true
		}
	}
	return false
}

// adjSoftCount counts soft conflicts between an adjudicator and the debate's teams.
func (ci *ConflictIndex) adjSoftCount(adj models.AdjDrawInfo, teamInst map[string]string, teams []models.TeamAssignment) int {
	count := 0
	targets := ci.adjSoft[adj.ID]
	for _, t := range teams {
		if targets != nil && targets[t.TeamID] {
			count++
			continue
		}
		if inst := teamInst[t.TeamID]; inst != "" && targets[inst] {
			count++
		}
	}
	return count
}

// trySwap evaluates swapping team si of debate di with team sj of debate dj,
// returning true if the swap strictly reduces total conflict penalty.
func trySwap(pairings [][]models.TeamDrawInfo, ci *ConflictIndex, di, si, dj, sj int) bool {
	before := ci.debatePenalty(pairings[di]) + ci.debatePenalty(pairings[dj])
	pairings[di][si], pairings[dj][sj] = pairings[dj][sj], pairings[di][si]
	after := ci.debatePenalty(pairings[di]) + ci.debatePenalty(pairings[dj])
	if after < before {
		return true
	}
	pairings[di][si], pairings[dj][sj] = pairings[dj][sj], pairings[di][si]
	return false
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// resolveTeamPairConflicts attempts to remove team-level conflicts by swapping
// teams between adjacent debates. Hard conflicts are prioritized implicitly by cost:
// each accepted swap must reduce combined penalty. Unresolvable conflicts remain.
func resolveTeamPairConflicts(pairings [][]models.TeamDrawInfo, ci *ConflictIndex, passes int) {
	for p := 0; p < passes; p++ {
		swappedAny := false
		for di := range pairings {
			for dj := range pairings {
				if di == dj || absInt(di-dj) > 1 {
					continue
				}
				for si := range pairings[di] {
					for sj := range pairings[dj] {
						if trySwap(pairings, ci, di, si, dj, sj) {
							swappedAny = true
						}
					}
				}
			}
		}
		if !swappedAny {
			break
		}
	}
}
