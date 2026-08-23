package draw

import (
	"reflect"
	"testing"

	"yellow/internal/models"
)

func TestSelectActiveTeamsAdmitsStandby(t *testing.T) {
	sorted := []models.TeamDrawInfo{
		{ID: "a"}, {ID: "b"}, {ID: "x"}, {ID: "c", Standby: true},
	}
	got := SelectActiveTeams(sorted, 2)
	if len(got) != 4 {
		t.Fatalf("expected 4 active teams, got %d: %v", len(got), got)
	}
	foundStandby := false
	for _, g := range got {
		if g.ID == "c" {
			foundStandby = true
		}
	}
	if !foundStandby {
		t.Errorf("expected standby team c to be admitted, got %v", got)
	}
}

func TestSelectActiveTeamsDropsSurplus(t *testing.T) {
	sorted := []models.TeamDrawInfo{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}, {ID: "f"}, {ID: "g"},
	}
	got := SelectActiveTeams(sorted, 2)
	if len(got) != 6 {
		t.Fatalf("expected 6 active teams, got %d", len(got))
	}
	if got[len(got)-1].ID == "g" {
		t.Errorf("expected lowest-ranked team g to sit out")
	}
}

func TestSelectActiveTeamsNoUnneededStandby(t *testing.T) {
	sorted := []models.TeamDrawInfo{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "s", Standby: true},
	}
	got := SelectActiveTeams(sorted, 2)
	if len(got) != 4 {
		t.Fatalf("expected 4 active teams, got %d", len(got))
	}
	for _, g := range got {
		if g.ID == "s" {
			t.Errorf("standby team should not be admitted when brackets are full")
		}
	}
}

func TestDetectPullUps(t *testing.T) {
	points := map[string]int{"w1": 3, "w2": 3, "l1": 0}
	pairings := [][]models.TeamDrawInfo{
		{{ID: "w1"}, {ID: "w2"}},
		{{ID: "w2"}, {ID: "l1"}},
	}

	flags := DetectPullUps(pairings, points, false)
	if flags[0][0] || flags[0][1] || flags[1][1] {
		t.Errorf("round one should never flag pull-ups: %v", flags)
	}

	flags = DetectPullUps(pairings, points, true)
	want := [][]bool{{false, false}, {false, true}}
	if !reflect.DeepEqual(flags, want) {
		t.Errorf("expected %v, got %v", want, flags)
	}
}

func TestSnakeSlot(t *testing.T) {
	want := []int{0, 1, 2, 2, 1, 0, 0, 1, 2}
	for i, w := range want {
		if got := snakeSlot(i, 3); got != w {
			t.Errorf("snakeSlot(%d, 3) = %d, want %d", i, got, w)
		}
	}
}

func TestDistributeSnake(t *testing.T) {
	out := distributeSnake([]string{"a", "b", "c", "d", "e", "f", "g"}, 3, nil)
	want := [][]string{{"a", "f", "g"}, {"b", "e"}, {"c", "d"}}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("expected %v, got %v", want, out)
	}
}

func TestDistributeSnakeConflictSkip(t *testing.T) {
	allowed := func(slot int, id string) bool {
		return id != "a" || slot == 2
	}
	out := distributeSnake([]string{"a", "b", "c", "d", "e", "f"}, 3, allowed)
	want := [][]string{{"f"}, {"b", "e"}, {"a", "c", "d"}}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("expected %v, got %v", want, out)
	}
}

func TestAllocatePanelsRoles(t *testing.T) {
	adjs := []models.AdjDrawInfo{
		{ID: "j1", Score: 5}, {ID: "j2", Score: 4}, {ID: "j3", Score: 3}, {ID: "j4", Score: 2},
		{ID: "j5", Score: 1.5}, {ID: "j6", Score: 1}, {ID: "j7", Score: 0.5}, {ID: "j8", Score: 0},
	}
	debates := []models.DebateDrawInput{
		{DebateID: "d1", Teams: []models.TeamAssignment{{TeamID: "t1"}}},
		{DebateID: "d2", Teams: []models.TeamAssignment{{TeamID: "t2"}}},
	}
	order := debateImportanceOrder(debates, map[string]int{"t1": 6, "t2": 0}, false, nil)

	AllocatePanels(order, debates, adjs, BuildConflictIndex(nil), map[string]string{})

	var roles []string
	for _, d := range debates {
		chairs, panels, trainees := 0, 0, 0
		for _, a := range d.Adjudicators {
			switch a.Role {
			case "chair":
				chairs++
			case "panel":
				panels++
			case "trainee":
				trainees++
			default:
				t.Errorf("unexpected role %q", a.Role)
			}
			roles = append(roles, a.Role)
		}
		if chairs != 1 || panels < 1 {
			t.Errorf("debate %s expected >=1 chair and >=1 panel, got chairs=%d panels=%d trainees=%d", d.DebateID, chairs, panels, trainees)
		}
	}

	total := 0
	for _, d := range debates {
		total += len(d.Adjudicators)
	}
	if total != len(adjs) {
		t.Errorf("expected all %d adjudicators allocated, got %d", len(adjs), total)
	}
	if countRole(debates, "chair") != 2 || countRole(debates, "trainee") != 1 {
		t.Errorf("expected 2 chairs and 1 trainee (lowest quartile of 6 remaining)")
	}
}

func countRole(debates []models.DebateDrawInput, role string) int {
	n := 0
	for _, d := range debates {
		for _, a := range d.Adjudicators {
			if a.Role == role {
				n++
			}
		}
	}
	return n
}

func TestDebateImportanceOrder(t *testing.T) {
	points := map[string]int{"hi": 9, "lo": 0}
	debates := []models.DebateDrawInput{
		{DebateID: "weak", Teams: []models.TeamAssignment{{TeamID: "lo"}}},
		{DebateID: "strong", Teams: []models.TeamAssignment{{TeamID: "hi"}}},
	}
	order := debateImportanceOrder(debates, points, false, nil)
	if debates[order[0]].DebateID != "strong" {
		t.Errorf("expected strongest debate first, got order %v", order)
	}
}
