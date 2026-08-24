package draw

import (
	"yellow/internal/models"
)

// traineeDivisor is the fraction of remaining (non-chair) adjudicators assigned
// as trainees: the lowest-scored quartile.
const traineeDivisor = 4

// snakeSlot returns the slot visited at position i of a snake rotation across
// n slots (0,1,...,n-1,n-1,...,1,0,...).
func snakeSlot(i, n int) int {
	col := i % n
	if (i/n)%2 == 1 {
		col = n - 1 - col
	}
	return col
}

// distributeSnake places ids into slots following a snake rotation; an id that
// is disallowed in its rotation slot tries subsequent positions and is dropped
// if it fits nowhere. Returns the per-slot id lists.
func distributeSnake(ids []string, n int, allowed func(slot int, id string) bool) [][]string {
	out := make([][]string, n)
	pos := 0
	for _, id := range ids {
		for k := 0; k < n; k++ {
			slot := snakeSlot(pos+k, n)
			if allowed == nil || allowed(slot, id) {
				out[slot] = append(out[slot], id)
				pos++
				break
			}
		}
	}
	return out
}

func adjIDs(adjs []models.AdjDrawInfo) []string {
	ids := make([]string, len(adjs))
	for i, a := range adjs {
		ids[i] = a.ID
	}
	return ids
}

// AllocatePanels assigns chairs, panelists, and trainees to debates. Chairs are
// taken best-first in the given debate importance order (skipping
// hard-conflicted adjudicators); remaining adjudicators split into a
// lowest-quartile trainee pool and a panelist pool, each snake-distributed so
// panel strength tracks bracket importance.
func AllocatePanels(order []int, debates []models.DebateDrawInput, adjs []models.AdjDrawInfo, cIdx *ConflictIndex, teamInstMap map[string]string) {
	n := len(debates)
	if n == 0 {
		return
	}

	used := make(map[string]bool)
	for _, di := range order {
		chairID := selectChair(adjs, used, cIdx, teamInstMap, debates[di].Teams)
		if chairID == "" {
			continue
		}
		used[chairID] = true
		debates[di].Adjudicators = append(debates[di].Adjudicators,
			models.AdjudicatorAssignment{AdjudicatorID: chairID, Role: "chair"})
	}

	var rest []models.AdjDrawInfo
	for _, a := range adjs {
		if !used[a.ID] {
			rest = append(rest, a)
		}
	}
	byID := make(map[string]models.AdjDrawInfo, len(adjs))
	for _, a := range adjs {
		byID[a.ID] = a
	}

	split := len(rest) - len(rest)/traineeDivisor
	panelSlots := distributeSnake(adjIDs(rest[:split]), n, func(slot int, id string) bool {
		di := order[slot]
		return !cIdx.adjHasHard(byID[id], teamInstMap, debates[di].Teams)
	})
	traineeSlots := distributeSnake(adjIDs(rest[split:]), n, func(slot int, id string) bool {
		di := order[slot]
		return !cIdx.adjHasHard(byID[id], teamInstMap, debates[di].Teams)
	})

	for slot := 0; slot < n; slot++ {
		di := order[slot]
		for _, id := range panelSlots[slot] {
			debates[di].Adjudicators = append(debates[di].Adjudicators,
				models.AdjudicatorAssignment{AdjudicatorID: id, Role: "panel"})
		}
		for _, id := range traineeSlots[slot] {
			debates[di].Adjudicators = append(debates[di].Adjudicators,
				models.AdjudicatorAssignment{AdjudicatorID: id, Role: "trainee"})
		}
	}
}
