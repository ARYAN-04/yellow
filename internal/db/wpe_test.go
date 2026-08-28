package db

import (
	"path/filepath"
	"testing"
)

func TestWPE_AdjudicatorAllocationsAndStandings(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "wpe-test.db")
	db, err := InitTournamentDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init tourney db: %v", err)
	}
	defer db.Close()

	store := NewSQLTournamentStore(db)

	// Seed institution, teams, adjudicators, round, debates
	_, err = db.Exec("INSERT INTO institutions (id, name, code) VALUES ('inst-1', 'Oxford University', 'OXF')")
	if err != nil {
		t.Fatalf("failed to insert inst: %v", err)
	}
	_, err = db.Exec("INSERT INTO teams (id, name, code, institution_id) VALUES ('t-1', 'Oxford A', 'OXA', 'inst-1'), ('t-2', 'Oxford B', 'OXB', 'inst-1')")
	if err != nil {
		t.Fatalf("failed to insert teams: %v", err)
	}
	_, err = db.Exec("INSERT INTO adjudicators (id, name, institution_id, test_score, rating) VALUES ('adj-1', 'Judge Alice', 'inst-1', 90.0, 92.5), ('adj-2', 'Judge Bob', NULL, 80.0, NULL), ('adj-3', 'Judge Charlie', NULL, 70.0, 68.0)")
	if err != nil {
		t.Fatalf("failed to insert adjudicators: %v", err)
	}
	_, err = db.Exec("INSERT INTO rounds (id, seq, name, stage, draw_released, results_released) VALUES ('r-1', 1, 'Round 1', 'preliminary', 1, 1)")
	if err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}
	_, err = db.Exec("INSERT INTO debates (id, round_id, venue) VALUES ('d-1', 'r-1', 'Room 101'), ('d-2', 'r-1', 'Room 102')")
	if err != nil {
		t.Fatalf("failed to insert debates: %v", err)
	}
	_, err = db.Exec("INSERT INTO debate_teams (id, debate_id, team_id, side) VALUES ('dt-1', 'd-1', 't-1', 'OG'), ('dt-2', 'd-1', 't-2', 'OO')")
	if err != nil {
		t.Fatalf("failed to insert debate teams: %v", err)
	}
	_, err = db.Exec("INSERT INTO debate_adjudicators (id, debate_id, adjudicator_id, role) VALUES ('da-1', 'd-1', 'adj-1', 'chair')")
	if err != nil {
		t.Fatalf("failed to insert debate adj: %v", err)
	}

	// 1. Test AddAdjudicatorToDebate
	roundID, err := store.AddAdjudicatorToDebate("d-1", "adj-2", "panel")
	if err != nil {
		t.Fatalf("AddAdjudicatorToDebate failed: %v", err)
	}
	if roundID != "r-1" {
		t.Errorf("expected roundID r-1, got %s", roundID)
	}

	// Test duplicate addition in same round fails
	_, err = store.AddAdjudicatorToDebate("d-2", "adj-2", "chair")
	if err == nil {
		t.Errorf("expected error adding already assigned adjudicator to another debate in same round, got nil")
	}

	// 2. Test MoveSwapAdjudicatorAssignment with role update
	var da2ID string
	err = db.QueryRow("SELECT id FROM debate_adjudicators WHERE debate_id = 'd-1' AND adjudicator_id = 'adj-2'").Scan(&da2ID)
	if err != nil {
		t.Fatalf("failed to find da2 id: %v", err)
	}

	roundID, err = store.MoveSwapAdjudicatorAssignment(da2ID, "d-1", "trainee")
	if err != nil {
		t.Fatalf("MoveSwapAdjudicatorAssignment (role update) failed: %v", err)
	}
	var updatedRole string
	_ = db.QueryRow("SELECT role FROM debate_adjudicators WHERE id = ?", da2ID).Scan(&updatedRole)
	if updatedRole != "trainee" {
		t.Errorf("expected role 'trainee', got %s", updatedRole)
	}

	// 3. Test RemoveAdjudicatorFromDebate
	roundID, err = store.RemoveAdjudicatorFromDebate(da2ID)
	if err != nil {
		t.Fatalf("RemoveAdjudicatorFromDebate failed: %v", err)
	}
	if roundID != "r-1" {
		t.Errorf("expected roundID r-1, got %s", roundID)
	}
	var cnt int
	_ = db.QueryRow("SELECT COUNT(*) FROM debate_adjudicators WHERE id = ?", da2ID).Scan(&cnt)
	if cnt != 0 {
		t.Errorf("expected da2 assignment to be deleted, count = %d", cnt)
	}

	// Re-add adj-2 as panelist to test standings
	_, err = store.AddAdjudicatorToDebate("d-1", "adj-2", "panel")
	if err != nil {
		t.Fatalf("failed to re-add adj-2: %v", err)
	}

	// Seed feedback submissions
	var qID string
	err = db.QueryRow("SELECT id FROM feedback_questions LIMIT 1").Scan(&qID)
	if err != nil {
		t.Fatalf("failed to query feedback question id: %v", err)
	}

	score95 := 9.5
	score80 := 8.0
	err = store.SubmitFeedback("d-1", "team", "t-1", "adj-1", &score95, map[string]string{qID: "9.5"})
	if err != nil {
		t.Fatalf("SubmitFeedback failed: %v", err)
	}
	err = store.SubmitFeedback("d-1", "team", "t-2", "adj-1", &score80, map[string]string{qID: "8.0"})
	if err != nil {
		t.Fatalf("SubmitFeedback 2 failed: %v", err)
	}

	// 4. Test GetAdjudicatorStandings
	standings, err := store.GetAdjudicatorStandings(true)
	if err != nil {
		t.Fatalf("GetAdjudicatorStandings failed: %v", err)
	}
	if len(standings) != 3 {
		t.Fatalf("expected 3 adjudicator standings, got %d", len(standings))
	}

	// adj-1 should be Rank 1 (rating 92.5)
	if standings[0].ID != "adj-1" {
		t.Errorf("expected rank 1 to be adj-1, got %s", standings[0].ID)
	}
	if standings[0].ChairsCount != 1 || standings[0].DebatesCount != 1 {
		t.Errorf("expected adj-1 to have 1 chair debate, got chairs=%d total=%d", standings[0].ChairsCount, standings[0].DebatesCount)
	}
	if standings[0].FeedbackCount != 2 {
		t.Errorf("expected adj-1 to have 2 feedback submissions, got %d", standings[0].FeedbackCount)
	}
	if standings[0].AverageFeedbackScore == nil || *standings[0].AverageFeedbackScore != 8.75 {
		t.Errorf("expected avg feedback 8.75, got %v", standings[0].AverageFeedbackScore)
	}
	if standings[0].InstitutionCode == nil || *standings[0].InstitutionCode != "OXF" {
		t.Errorf("expected institution code OXF, got %v", standings[0].InstitutionCode)
	}

	// adj-2 should be Rank 2 (test_score 80.0, panel count 1)
	if standings[1].ID != "adj-2" {
		t.Errorf("expected rank 2 to be adj-2, got %s", standings[1].ID)
	}
	if standings[1].PanelsCount != 1 {
		t.Errorf("expected adj-2 to have 1 panel debate, got %d", standings[1].PanelsCount)
	}

	// 5. Test GetAdjudicatorTrajectory
	traj, err := store.GetAdjudicatorTrajectory("adj-1", true)
	if err != nil {
		t.Fatalf("GetAdjudicatorTrajectory failed: %v", err)
	}
	if traj.Adjudicator.Name != "Judge Alice" {
		t.Errorf("expected name Judge Alice, got %s", traj.Adjudicator.Name)
	}
	if len(traj.Debates) != 1 {
		t.Fatalf("expected 1 debate in trajectory, got %d", len(traj.Debates))
	}
	if traj.Debates[0].Role != "chair" {
		t.Errorf("expected role chair, got %s", traj.Debates[0].Role)
	}
	if len(traj.Debates[0].Teams) != 2 {
		t.Errorf("expected 2 teams in debate, got %d", len(traj.Debates[0].Teams))
	}
	if len(traj.Debates[0].CoAdjudicators) != 1 {
		t.Errorf("expected 1 co-adjudicator (adj-2), got %d", len(traj.Debates[0].CoAdjudicators))
	}
}
