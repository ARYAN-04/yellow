package draw

import (
	"testing"

	"yellow/internal/models"
)

func testConflicts() []models.Conflict {
	return []models.Conflict{
		{SubjectType: "team", SubjectID: "t1", TargetType: "team", TargetID: "t2", Weight: "hard"},
		{SubjectType: "team", SubjectID: "t3", TargetType: "team", TargetID: "t4", Weight: "soft"},
		{SubjectType: "adjudicator", SubjectID: "a1", TargetType: "team", TargetID: "t1", Weight: "hard"},
		{SubjectType: "adjudicator", SubjectID: "a1", TargetType: "institution", TargetID: "inst-x", Weight: "soft"},
		{SubjectType: "adjudicator", SubjectID: "a2", TargetType: "team", TargetID: "t5", Weight: "soft"},
	}
}

func TestTeamPairPenalty(t *testing.T) {
	ci := BuildConflictIndex(testConflicts())

	if got := ci.teamPairPenalty("t1", "t2"); got != penaltyHard {
		t.Errorf("expected hard penalty %d, got %d", penaltyHard, got)
	}
	if got := ci.teamPairPenalty("t3", "t4"); got != penaltySoft {
		t.Errorf("expected soft penalty %d, got %d", penaltySoft, got)
	}
	if got := ci.teamPairPenalty("t4", "t3"); got != penaltySoft {
		t.Errorf("expected symmetric soft penalty, got %d", got)
	}
	if got := ci.teamPairPenalty("t1", "t3"); got != 0 {
		t.Errorf("expected no penalty, got %d", got)
	}
}

func TestDebatePenalty(t *testing.T) {
	ci := BuildConflictIndex(testConflicts())
	teams := []models.TeamDrawInfo{
		{ID: "t1"}, {ID: "t2"}, {ID: "t3"}, {ID: "t4"},
	}
	if got := ci.debatePenalty(teams); got != penaltyHard+penaltySoft {
		t.Errorf("expected combined penalty %d, got %d", penaltyHard+penaltySoft, got)
	}
}

func TestResolveTeamPairConflicts(t *testing.T) {
	ci := BuildConflictIndex([]models.Conflict{
		{SubjectType: "team", SubjectID: "t1", TargetType: "team", TargetID: "t2", Weight: "hard"},
	})
	pairings := [][]models.TeamDrawInfo{
		{{ID: "t1"}, {ID: "t2"}},
		{{ID: "t3"}, {ID: "t4"}},
	}

	resolveTeamPairConflicts(pairings, ci, 3)

	for _, p := range pairings {
		if pen := ci.debatePenalty(p); pen != 0 {
			t.Fatalf("expected zero penalty after swap resolution, debate still has penalty %d", pen)
		}
	}
}

func TestResolveTeamPairConflictsUnresolvable(t *testing.T) {
	ci := BuildConflictIndex([]models.Conflict{
		{SubjectType: "team", SubjectID: "t1", TargetType: "team", TargetID: "t2", Weight: "hard"},
	})
	pairings := [][]models.TeamDrawInfo{
		{{ID: "t1"}, {ID: "t2"}},
	}

	resolveTeamPairConflicts(pairings, ci, 3)

	if pen := ci.debatePenalty(pairings[0]); pen == 0 {
		t.Error("single-debate hard conflict should remain (unresolvable), but penalty was zero")
	}
}

func adjTestSetup() (*ConflictIndex, map[string]string) {
	ci := BuildConflictIndex(testConflicts())
	return ci, map[string]string{
		"t1": "inst-y",
		"t5": "inst-x",
	}
}

func TestAdjHasHard(t *testing.T) {
	ci, insts := adjTestSetup()

	teamsT1 := []models.TeamAssignment{{TeamID: "t1"}}
	if !ci.adjHasHard(models.AdjDrawInfo{ID: "a1"}, insts, teamsT1) {
		t.Error("a1 should have declared hard conflict with t1")
	}

	teamsInst := []models.TeamAssignment{{TeamID: "t5"}}
	if !ci.adjHasHard(models.AdjDrawInfo{ID: "a9", InstitutionID: "inst-x"}, insts, teamsInst) {
		t.Error("adjudicator from inst-x should have institutional conflict with t5")
	}

	if ci.adjHasHard(models.AdjDrawInfo{ID: "a9"}, insts, teamsT1) {
		t.Error("a9 has no conflict with t1")
	}
}

func TestAdjSoftCount(t *testing.T) {
	ci, insts := adjTestSetup()

	teamsT5 := []models.TeamAssignment{{TeamID: "t5"}}
	if got := ci.adjSoftCount(models.AdjDrawInfo{ID: "a2"}, insts, teamsT5); got != 1 {
		t.Errorf("expected 1 soft conflict for a2 vs t5, got %d", got)
	}

	teamsInstSoft := []models.TeamAssignment{{TeamID: "t5"}}
	if got := ci.adjSoftCount(models.AdjDrawInfo{ID: "a3"}, insts, teamsInstSoft); got != 0 {
		t.Errorf("a3 has no declared conflicts, got %d", got)
	}
}

func TestSelectChair(t *testing.T) {
	ci, insts := adjTestSetup()
	adjs := []models.AdjDrawInfo{
		{ID: "a1", Score: 10}, // hard conflict with t1
		{ID: "a2", Score: 9},  // soft conflict with t5
		{ID: "a3", Score: 8},  // clean
	}
	used := make(map[string]bool)
	teams := []models.TeamAssignment{{TeamID: "t1"}, {TeamID: "t5"}}

	got := selectChair(adjs, used, ci, insts, teams)
	if got != "a3" {
		t.Errorf("expected clean adjudicator a3 to chair, got %q", got)
	}

	used["a3"] = true
	got = selectChair(adjs, used, ci, insts, teams)
	if got != "a2" {
		t.Errorf("expected adjudicator with fewest soft conflicts a2 to chair next, got %q", got)
	}

	used["a2"] = true
	got = selectChair(adjs, used, ci, insts, teams)
	if got != "" {
		t.Errorf("expected no chair when only hard-conflicted adjudicator remains, got %q", got)
	}
}
