package draw

import (
	"math/rand"
	"sort"

	"yellow/internal/models"
)

// splitByStandby partitions a points-sorted team list into regular and standby
// teams, preserving relative order within each partition.
func splitByStandby(sorted []models.TeamDrawInfo) (regular, standby []models.TeamDrawInfo) {
	for _, t := range sorted {
		if t.Standby {
			standby = append(standby, t)
		} else {
			regular = append(regular, t)
		}
	}
	return regular, standby
}

// SelectActiveTeams trims a points-sorted team list down to a multiple of
// numSides: standby teams are admitted highest-ranked-first only when needed to
// complete a bracket; any lowest-ranked surplus sits out the round.
func SelectActiveTeams(sorted []models.TeamDrawInfo, numSides int) []models.TeamDrawInfo {
	regular, standby := splitByStandby(sorted)
	if short := numSides - len(regular)%numSides; short < numSides && len(standby) > 0 {
		admit := min(short, len(standby))
		regular = append(regular, standby[:admit]...)
	}
	return regular[:len(regular)/numSides*numSides]
}

// DetectPullUps flags teams whose confirmed points sit below their own debate's
// top seed — i.e., pulled up from a lower points bracket by sequential chunking.
// Round 1 pairings (powerPaired=false) never produce pull-ups.
func DetectPullUps(pairings [][]models.TeamDrawInfo, points map[string]int, powerPaired bool) [][]bool {
	flags := make([][]bool, len(pairings))
	for di, pair := range pairings {
		flags[di] = make([]bool, len(pair))
		if !powerPaired {
			continue
		}
		top := 0
		for _, t := range pair {
			if p := points[t.ID]; p > top {
				top = p
			}
		}
		for i, t := range pair {
			flags[di][i] = points[t.ID] < top
		}
	}
	return flags
}

// sortByPoints orders teams by confirmed points descending with random tiebreak.
func sortByPoints(teams []models.TeamDrawInfo, points map[string]int, r *rand.Rand) {
	r.Shuffle(len(teams), func(i, j int) {
		teams[i], teams[j] = teams[j], teams[i]
	})
	sort.SliceStable(teams, func(i, j int) bool {
		return points[teams[i].ID] > points[teams[j].ID]
	})
}

// buildPairings chunks the ordered team list into debates; round one shuffles
// first while later rounds rely on points-descending order (power pairing).
func buildPairings(teams []models.TeamDrawInfo, numSides int, roundOne bool, r *rand.Rand) [][]models.TeamDrawInfo {
	pool := make([]models.TeamDrawInfo, len(teams))
	copy(pool, teams)
	if roundOne {
		r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	}
	numDebates := len(pool) / numSides
	out := make([][]models.TeamDrawInfo, numDebates)
	for i := 0; i < numDebates; i++ {
		out[i] = append(out[i], pool[i*numSides:(i+1)*numSides]...)
	}
	return out
}

// debateAvgPoints returns the average confirmed points of a debate's teams.
func debateAvgPoints(d models.DebateDrawInput, points map[string]int) float64 {
	total := 0
	for _, t := range d.Teams {
		total += points[t.TeamID]
	}
	if len(d.Teams) == 0 {
		return 0
	}
	return float64(total) / float64(len(d.Teams))
}

// debateImportanceOrder returns debate indices sorted by average confirmed
// points descending; round one uses random order instead.
func debateImportanceOrder(debates []models.DebateDrawInput, points map[string]int, roundOne bool, r *rand.Rand) []int {
	order := make([]int, len(debates))
	for i := range order {
		order[i] = i
	}
	if roundOne {
		r.Shuffle(len(order), func(a, b int) { order[a], order[b] = order[b], order[a] })
		return order
	}
	sort.SliceStable(order, func(a, b int) bool {
		return debateAvgPoints(debates[order[a]], points) > debateAvgPoints(debates[order[b]], points)
	})
	return order
}
