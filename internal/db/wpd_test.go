package db

import (
	"path/filepath"
	"testing"
	"yellow/internal/models"

	"github.com/google/uuid"
)

func TestWPD_SpeakerScoresRolesAndTrajectories(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "wpd-test.db")
	db, err := InitTournamentDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init tourney db: %v", err)
	}
	defer db.Close()

	store := NewSQLTournamentStore(db)

	// Seed institutions, teams, speakers, adjudicators, round, debate
	_, err = db.Exec("INSERT INTO institutions (id, name, code) VALUES ('inst-1', 'Test University', 'TU')")
	if err != nil {
		t.Fatalf("failed to insert inst: %v", err)
	}
	_, err = db.Exec("INSERT INTO teams (id, name, code, institution_id) VALUES ('t-1', 'TU Alpha', 'TUA', 'inst-1'), ('t-2', 'TU Beta', 'TUB', 'inst-1')")
	if err != nil {
		t.Fatalf("failed to insert teams: %v", err)
	}
	_, err = db.Exec("INSERT INTO speakers (id, name, team_id, is_novice) VALUES ('sp-1', 'Alice', 't-1', 0), ('sp-2', 'Bob', 't-1', 1), ('sp-3', 'Charlie', 't-2', 0), ('sp-4', 'Dave', 't-2', 0)")
	if err != nil {
		t.Fatalf("failed to insert speakers: %v", err)
	}
	_, err = db.Exec("INSERT INTO adjudicators (id, name, test_score) VALUES ('adj-1', 'Judge Chair', 80.0), ('adj-2', 'Judge Panel', 75.0)")
	if err != nil {
		t.Fatalf("failed to insert adjudicators: %v", err)
	}
	_, err = db.Exec("INSERT INTO rounds (id, seq, name, stage, draw_released, results_released) VALUES ('r-1', 1, 'Round 1', 'preliminary', 1, 1)")
	if err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}
	_, err = db.Exec("INSERT INTO motions (id, round_id, seq, reference, text) VALUES ('m-1', 'r-1', 1, 'R1M1', 'This House would ban artificial intelligence.')")
	if err != nil {
		t.Fatalf("failed to insert motion: %v", err)
	}
	_, err = db.Exec("INSERT INTO debates (id, round_id, venue) VALUES ('d-1', 'r-1', 'Room 101')")
	if err != nil {
		t.Fatalf("failed to insert debate: %v", err)
	}
	_, err = db.Exec("INSERT INTO debate_teams (id, debate_id, team_id, side) VALUES ('dt-1', 'd-1', 't-1', 'OG'), ('dt-2', 'd-1', 't-2', 'OO')")
	if err != nil {
		t.Fatalf("failed to insert debate teams: %v", err)
	}
	_, err = db.Exec("INSERT INTO debate_adjudicators (id, debate_id, adjudicator_id, role) VALUES ('da-1', 'd-1', 'adj-1', 'chair'), ('da-2', 'd-1', 'adj-2', 'panel')")
	if err != nil {
		t.Fatalf("failed to insert debate adjs: %v", err)
	}
	_, err = db.Exec("INSERT INTO access_tokens (token, type, owner_id) VALUES ('token-team-1', 'team', 't-1'), ('token-adj-1', 'adjudicator', 'adj-1')")
	if err != nil {
		t.Fatalf("failed to insert tokens: %v", err)
	}

	// 1. Submit Ballot with Speaker Scores including Role
	ballotID := uuid.New().String()
	results := []models.TeamBallotResult{
		{
			TeamID:        "t-1",
			Points:        3,
			SpeakerPoints: 154.5,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: "sp-1", Score: 77.5, IsReply: false, SpeechOrder: 1, Role: "PM"},
				{SpeakerID: "sp-2", Score: 77.0, IsReply: false, SpeechOrder: 2, Role: "DPM"},
			},
		},
		{
			TeamID:        "t-2",
			Points:        2,
			SpeakerPoints: 150.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: "sp-3", Score: 75.0, IsReply: false, SpeechOrder: 1, Role: "LO"},
				{SpeakerID: "sp-4", Score: 75.0, IsReply: false, SpeechOrder: 2, Role: "DLO"},
			},
		},
	}

	if err := store.SubmitBallot("d-1", ballotID, "adjudicator", "adj-1", "confirmed", false, "", results); err != nil {
		t.Fatalf("SubmitBallot failed: %v", err)
	}

	// 2. Test GetTeamTrajectory includes roles
	traj, err := store.GetTeamTrajectory("t-1", true)
	if err != nil {
		t.Fatalf("GetTeamTrajectory failed: %v", err)
	}
	if len(traj.Debates) != 1 {
		t.Fatalf("expected 1 debate in trajectory, got %d", len(traj.Debates))
	}
	if len(traj.Debates[0].SpeakerScores) != 2 {
		t.Fatalf("expected 2 speaker scores, got %d", len(traj.Debates[0].SpeakerScores))
	}
	if traj.Debates[0].SpeakerScores[0].Role != "PM" || traj.Debates[0].SpeakerScores[1].Role != "DPM" {
		t.Errorf("unexpected roles in team trajectory: %v, %v", traj.Debates[0].SpeakerScores[0].Role, traj.Debates[0].SpeakerScores[1].Role)
	}

	// 3. Test GetSpeakerTrajectory includes role
	spTraj, err := store.GetSpeakerTrajectory("sp-1", true)
	if err != nil {
		t.Fatalf("GetSpeakerTrajectory failed: %v", err)
	}
	if len(spTraj.Speeches) != 1 {
		t.Fatalf("expected 1 speech, got %d", len(spTraj.Speeches))
	}
	if spTraj.Speeches[0].Role != "PM" {
		t.Errorf("expected role 'PM', got '%s'", spTraj.Speeches[0].Role)
	}

	// 4. Test ResolveToken for Team returns populated Speakers
	tokenInfo, err := store.ResolveToken("token-team-1")
	if err != nil {
		t.Fatalf("ResolveToken failed: %v", err)
	}
	if len(tokenInfo.Speakers) != 2 {
		t.Errorf("expected 2 speakers in token info, got %d", len(tokenInfo.Speakers))
	}
	if tokenInfo.Speakers[0].Name != "Alice" || tokenInfo.Speakers[1].Name != "Bob" {
		t.Errorf("unexpected speakers list in token info: %+v", tokenInfo.Speakers)
	}

	// 5. Test GetTokenDebates for Adjudicator returns populated speakers per team and motion
	adjDebates, err := store.GetTokenDebates("adj-1", "adjudicator")
	if err != nil {
		t.Fatalf("GetTokenDebates for adj failed: %v", err)
	}
	if len(adjDebates) != 1 {
		t.Fatalf("expected 1 debate, got %d", len(adjDebates))
	}
	if adjDebates[0].Motion == "" {
		t.Errorf("expected motion to be populated on debate info")
	}
	if len(adjDebates[0].Teams) != 2 {
		t.Fatalf("expected 2 teams in debate, got %d", len(adjDebates[0].Teams))
	}
	if len(adjDebates[0].Teams[0].Speakers) != 2 {
		t.Errorf("expected 2 speakers for team 0, got %d", len(adjDebates[0].Teams[0].Speakers))
	}

	// 6. Test GetTokenDebates for Team returns speaker scores and panel info
	teamDebates, err := store.GetTokenDebates("t-1", "team")
	if err != nil {
		t.Fatalf("GetTokenDebates for team failed: %v", err)
	}
	if len(teamDebates) != 1 {
		t.Fatalf("expected 1 debate, got %d", len(teamDebates))
	}
	if len(teamDebates[0].SpeakerScores) != 2 {
		t.Errorf("expected 2 speaker scores in team debates, got %d", len(teamDebates[0].SpeakerScores))
	}
	if teamDebates[0].Chair != "Judge Chair" {
		t.Errorf("expected chair 'Judge Chair', got '%s'", teamDebates[0].Chair)
	}
}
