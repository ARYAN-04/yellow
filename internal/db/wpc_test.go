package db

import (
	"path/filepath"
	"testing"
	"yellow/internal/models"
)

func TestTeamAndSpeakerTrajectory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "trajectory_test.db")
	sqlDB, err := InitTournamentDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init tournament db: %v", err)
	}
	defer sqlDB.Close()
	store := NewSQLTournamentStore(sqlDB)

	// Setup institutions, teams, speakers, adjudicators, rounds
	if err := store.CreateInstitution("inst-1", "Harvard", "HAR"); err != nil {
		t.Fatalf("failed to create institution: %v", err)
	}
	if err := store.CreateInstitution("inst-2", "Yale", "YAL"); err != nil {
		t.Fatalf("failed to create institution: %v", err)
	}

	inst1 := "inst-1"
	inst2 := "inst-2"
	if err := store.CreateTeam("team-1", "Harvard A", "HAR-A", inst1, []models.SpeakerRequest{
		{Name: "Alice"},
		{Name: "Bob"},
	}, "tok-team-1"); err != nil {
		t.Fatalf("failed to create team 1: %v", err)
	}

	if err := store.CreateTeam("team-2", "Yale A", "YAL-A", inst2, []models.SpeakerRequest{
		{Name: "Charlie"},
		{Name: "Dave"},
	}, "tok-team-2"); err != nil {
		t.Fatalf("failed to create team 2: %v", err)
	}

	// Fetch speaker IDs created for team-1 and team-2
	team1Data, err := store.GetTeamTrajectory("team-1", true)
	if err != nil || len(team1Data.Team.Speakers) < 2 {
		t.Fatalf("failed to load team-1 speakers: %v", err)
	}
	sp1ID := team1Data.Team.Speakers[0].ID
	sp2ID := team1Data.Team.Speakers[1].ID

	team2Data, err := store.GetTeamTrajectory("team-2", true)
	if err != nil || len(team2Data.Team.Speakers) < 2 {
		t.Fatalf("failed to load team-2 speakers: %v", err)
	}
	sp3ID := team2Data.Team.Speakers[0].ID
	sp4ID := team2Data.Team.Speakers[1].ID

	if err := store.CreateAdjudicator("adj-1", "Judge Judy", inst1, 8.5, "tok-adj-1"); err != nil {
		t.Fatalf("failed to create adj 1: %v", err)
	}
	if err := store.CreateAdjudicator("adj-2", "Judge Dredd", inst2, 7.0, "tok-adj-2"); err != nil {
		t.Fatalf("failed to create adj 2: %v", err)
	}

	// Round 1 (Public & released)
	if err := store.CreateRound("r-1", 1, "Round 1", "preliminary"); err != nil {
		t.Fatalf("failed to create round 1: %v", err)
	}
	rel := true
	if err := store.UpdateRound("r-1", nil, &rel, &rel); err != nil {
		t.Fatalf("failed to update round 1: %v", err)
	}

	// Round 2 (Silent round with draw released but results unreleased / silent)
	if err := store.CreateRound("r-2", 2, "Round 2 (Silent)", "preliminary"); err != nil {
		t.Fatalf("failed to create round 2: %v", err)
	}
	silent := true
	unrel := false
	if err := store.UpdateRound("r-2", &silent, &rel, &unrel); err != nil {
		t.Fatalf("failed to update round 2: %v", err)
	}

	// Draw for Round 1
	r1DebateInput := []models.DebateDrawInput{
		{
			DebateID: "deb-1",
			Venue:    "Room 101",
			Teams: []models.TeamAssignment{
				{TeamID: "team-1", Side: "OG"},
				{TeamID: "team-2", Side: "OO"},
			},
			Adjudicators: []models.AdjudicatorAssignment{
				{AdjudicatorID: "adj-1", Role: "chair"},
				{AdjudicatorID: "adj-2", Role: "panel"},
			},
		},
	}
	if err := store.SaveDraw("r-1", r1DebateInput); err != nil {
		t.Fatalf("failed to save round 1 draw: %v", err)
	}

	// Draw for Round 2
	r2DebateInput := []models.DebateDrawInput{
		{
			DebateID: "deb-2",
			Venue:    "Room 102",
			Teams: []models.TeamAssignment{
				{TeamID: "team-1", Side: "OO"},
				{TeamID: "team-2", Side: "OG"},
			},
			Adjudicators: []models.AdjudicatorAssignment{
				{AdjudicatorID: "adj-1", Role: "chair"},
			},
		},
	}
	if err := store.SaveDraw("r-2", r2DebateInput); err != nil {
		t.Fatalf("failed to save round 2 draw: %v", err)
	}

	// Submit and Confirm Ballot for Round 1
	sp1Score := 78.0
	sp2Score := 76.0
	r1Results := []models.TeamBallotResult{
		{
			TeamID:        "team-1",
			Points:        3,
			SpeakerPoints: 154.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: sp1ID, Score: sp1Score, SpeechOrder: 1, IsReply: false},
				{SpeakerID: sp2ID, Score: sp2Score, SpeechOrder: 2, IsReply: false},
			},
		},
		{
			TeamID:        "team-2",
			Points:        0,
			SpeakerPoints: 148.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: sp3ID, Score: 75.0, SpeechOrder: 1, IsReply: false},
				{SpeakerID: sp4ID, Score: 73.0, SpeechOrder: 2, IsReply: false},
			},
		},
	}
	if err := store.SubmitBallot("deb-1", "ballot-1", "adjudicator", "adj-1", "confirmed", false, "", r1Results); err != nil {
		t.Fatalf("failed to submit ballot 1: %v", err)
	}

	// Submit and Confirm Ballot for Round 2 (Silent)
	r2Results := []models.TeamBallotResult{
		{
			TeamID:        "team-1",
			Points:        2,
			SpeakerPoints: 152.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: sp1ID, Score: 76.0, SpeechOrder: 1, IsReply: false},
				{SpeakerID: sp2ID, Score: 76.0, SpeechOrder: 2, IsReply: false},
			},
		},
		{
			TeamID:        "team-2",
			Points:        1,
			SpeakerPoints: 150.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: sp3ID, Score: 75.0, SpeechOrder: 1, IsReply: false},
				{SpeakerID: sp4ID, Score: 75.0, SpeechOrder: 2, IsReply: false},
			},
		},
	}
	if err := store.SubmitBallot("deb-2", "ballot-2", "adjudicator", "adj-1", "confirmed", false, "", r2Results); err != nil {
		t.Fatalf("failed to submit ballot 2: %v", err)
	}

	// Test Team Trajectory - Admin View (Should see both round results)
	adminTraj, err := store.GetTeamTrajectory("team-1", true)
	if err != nil {
		t.Fatalf("failed to get team trajectory (admin): %v", err)
	}
	if adminTraj.Team.Name != "Harvard A" {
		t.Errorf("expected team name Harvard A, got %s", adminTraj.Team.Name)
	}
	if len(adminTraj.Debates) != 2 {
		t.Fatalf("expected 2 debates, got %d", len(adminTraj.Debates))
	}
	if adminTraj.Debates[0].Points == nil || *adminTraj.Debates[0].Points != 3 {
		t.Errorf("expected round 1 points 3, got %v", adminTraj.Debates[0].Points)
	}
	if adminTraj.Debates[1].Points == nil || *adminTraj.Debates[1].Points != 2 {
		t.Errorf("expected round 2 (silent) points 2 in admin view, got %v", adminTraj.Debates[1].Points)
	}
	if len(adminTraj.Debates[0].Opponents) != 1 || adminTraj.Debates[0].Opponents[0].TeamName != "Yale A" {
		t.Errorf("expected opponent Yale A, got %+v", adminTraj.Debates[0].Opponents)
	}
	if len(adminTraj.Debates[0].SpeakerScores) != 2 {
		t.Errorf("expected 2 speaker scores in round 1, got %d", len(adminTraj.Debates[0].SpeakerScores))
	}

	// Test Team Trajectory - Public/Non-Admin View (Silent round results must be masked)
	publicTraj, err := store.GetTeamTrajectory("team-1", false)
	if err != nil {
		t.Fatalf("failed to get team trajectory (public): %v", err)
	}
	if len(publicTraj.Debates) != 2 {
		t.Fatalf("expected 2 debates, got %d", len(publicTraj.Debates))
	}
	// Round 1 results visible
	if publicTraj.Debates[0].Points == nil || *publicTraj.Debates[0].Points != 3 {
		t.Errorf("expected round 1 points 3 in public view, got %v", publicTraj.Debates[0].Points)
	}
	// Round 2 results MASKED
	if publicTraj.Debates[1].Points != nil {
		t.Errorf("expected round 2 points to be masked/nil for silent round in public view, got %v", *publicTraj.Debates[1].Points)
	}
	if len(publicTraj.Debates[1].SpeakerScores) != 0 {
		t.Errorf("expected 0 speaker scores in masked round 2, got %d", len(publicTraj.Debates[1].SpeakerScores))
	}

	// Test Speaker Trajectory - Admin View for sp-1 (Alice)
	adminSpTraj, err := store.GetSpeakerTrajectory(sp1ID, true)
	if err != nil {
		t.Fatalf("failed to get speaker trajectory: %v", err)
	}
	if adminSpTraj.Speaker.Name != "Alice" {
		t.Errorf("expected speaker Alice, got %s", adminSpTraj.Speaker.Name)
	}
	if len(adminSpTraj.Speeches) != 2 {
		t.Fatalf("expected 2 speeches for Alice, got %d", len(adminSpTraj.Speeches))
	}
	if adminSpTraj.Speeches[0].Score == nil || *adminSpTraj.Speeches[0].Score != 78.0 {
		t.Errorf("expected round 1 score 78.0, got %v", adminSpTraj.Speeches[0].Score)
	}
	if adminSpTraj.Speeches[1].Score == nil || *adminSpTraj.Speeches[1].Score != 76.0 {
		t.Errorf("expected round 2 score 76.0 in admin view, got %v", adminSpTraj.Speeches[1].Score)
	}

	// Test Speaker Trajectory - Public View (Silent round score masked)
	publicSpTraj, err := store.GetSpeakerTrajectory(sp1ID, false)
	if err != nil {
		t.Fatalf("failed to get speaker trajectory (public): %v", err)
	}
	if publicSpTraj.Speeches[0].Score == nil || *publicSpTraj.Speeches[0].Score != 78.0 {
		t.Errorf("expected round 1 score 78.0, got %v", publicSpTraj.Speeches[0].Score)
	}
	if publicSpTraj.Speeches[1].Score != nil {
		t.Errorf("expected round 2 score masked in public view, got %v", *publicSpTraj.Speeches[1].Score)
	}
}
