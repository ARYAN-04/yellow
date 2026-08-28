package db

import (
	"path/filepath"
	"testing"

	"yellow/internal/models"

	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) (*SQLTournamentStore, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test_tournament.db")
	dbConn, err := InitTournamentDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	store := NewSQLTournamentStore(dbConn)
	cleanup := func() {
		_ = store.Close()
	}
	return store, cleanup
}

func TestSpeakerStandingsCalculation(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// 1. Create institutions
	inst1ID := uuid.New().String()
	inst2ID := uuid.New().String()
	if err := store.CreateInstitution(inst1ID, "Harvard", "HAR"); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateInstitution(inst2ID, "Yale", "YAL"); err != nil {
		t.Fatal(err)
	}

	// 2. Create teams & speakers
	team1ID := uuid.New().String()
	team2ID := uuid.New().String()
	team3ID := uuid.New().String()

	aliceID := uuid.New().String()
	bobID := uuid.New().String()
	charlieID := uuid.New().String()
	davidID := uuid.New().String()
	eveID := uuid.New().String()
	frankID := uuid.New().String()

	if err := store.CreateTeam(team1ID, "Harvard A", "HA", inst1ID, []models.SpeakerRequest{{Name: "Alice"}, {Name: "Bob"}}, "tok1"); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateTeam(team2ID, "Harvard B", "HB", inst1ID, []models.SpeakerRequest{{Name: "Charlie"}, {Name: "David"}}, "tok2"); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateTeam(team3ID, "Yale A", "YA", inst2ID, []models.SpeakerRequest{{Name: "Eve"}, {Name: "Frank"}}, "tok3"); err != nil {
		t.Fatal(err)
	}

	// Get generated speaker IDs and set flags
	teams, err := store.ListTeams()
	if err != nil {
		t.Fatal(err)
	}
	var spAlice, spBob, spCharlie, spEve models.Speaker
	for _, team := range teams {
		for _, sp := range team.Speakers {
			switch sp.Name {
			case "Alice":
				sp.IsNovice = true
				_ = store.UpsertSpeaker(team.ID, sp)
				spAlice = sp
			case "Bob":
				spBob = sp
			case "Charlie":
				spCharlie = sp
			case "Eve":
				sp.IsEsl = true
				_ = store.UpsertSpeaker(team.ID, sp)
				spEve = sp
			}
		}
	}
	_ = aliceID
	_ = bobID
	_ = charlieID
	_ = davidID
	_ = eveID
	_ = frankID

	// 3. Create rounds
	r1ID := uuid.New().String()
	r2ID := uuid.New().String()
	r3ID := uuid.New().String()
	if err := store.CreateRound(r1ID, 1, "Round 1", "preliminary"); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateRound(r2ID, 2, "Round 2", "preliminary"); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateRound(r3ID, 3, "Round 3 (Silent)", "preliminary"); err != nil {
		t.Fatal(err)
	}
	silentTrue := true
	if err := store.UpdateRound(r3ID, &silentTrue, nil, nil); err != nil {
		t.Fatal(err)
	}

	// 4. Create debates
	d1ID := uuid.New().String()
	d2ID := uuid.New().String()
	d3ID := uuid.New().String()
	if err := store.SaveDraw(r1ID, []models.DebateDrawInput{{
		DebateID: d1ID,
		Venue:    "Room 1",
		Teams:    []models.TeamAssignment{{TeamID: team1ID, Side: "OG"}, {TeamID: team2ID, Side: "OO"}},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveDraw(r2ID, []models.DebateDrawInput{{
		DebateID: d2ID,
		Venue:    "Room 1",
		Teams:    []models.TeamAssignment{{TeamID: team1ID, Side: "OG"}, {TeamID: team3ID, Side: "OO"}},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveDraw(r3ID, []models.DebateDrawInput{{
		DebateID: d3ID,
		Venue:    "Room 1",
		Teams:    []models.TeamAssignment{{TeamID: team1ID, Side: "OG"}, {TeamID: team2ID, Side: "OO"}},
	}}); err != nil {
		t.Fatal(err)
	}

	// 5. Submit confirmed ballots with speaker scores
	b1ID := uuid.New().String()
	err = store.SubmitBallot(d1ID, b1ID, "adjudicator", "adj1", "confirmed", false, "", []models.TeamBallotResult{
		{
			TeamID:        team1ID,
			Points:        3,
			SpeakerPoints: 153.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: spAlice.ID, Score: 78.0, IsReply: false, SpeechOrder: 1},
				{SpeakerID: spBob.ID, Score: 75.0, IsReply: false, SpeechOrder: 2},
				{SpeakerID: spAlice.ID, Score: 39.0, IsReply: true, SpeechOrder: 3}, // reply speech
			},
		},
		{
			TeamID:        team2ID,
			Points:        2,
			SpeakerPoints: 150.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: spCharlie.ID, Score: 80.0, IsReply: false, SpeechOrder: 1},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	b2ID := uuid.New().String()
	err = store.SubmitBallot(d2ID, b2ID, "adjudicator", "adj1", "confirmed", false, "", []models.TeamBallotResult{
		{
			TeamID:        team1ID,
			Points:        3,
			SpeakerPoints: 159.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: spAlice.ID, Score: 82.0, IsReply: false, SpeechOrder: 1},
				{SpeakerID: spBob.ID, Score: 77.0, IsReply: false, SpeechOrder: 2},
			},
		},
		{
			TeamID:        team3ID,
			Points:        1,
			SpeakerPoints: 160.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: spEve.ID, Score: 81.0, IsReply: false, SpeechOrder: 1},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	b3ID := uuid.New().String()
	err = store.SubmitBallot(d3ID, b3ID, "adjudicator", "adj1", "confirmed", false, "", []models.TeamBallotResult{
		{
			TeamID:        team1ID,
			Points:        3,
			SpeakerPoints: 164.0,
			SpeakerScores: []models.SpeakerScoreInput{
				{SpeakerID: spAlice.ID, Score: 85.0, IsReply: false, SpeechOrder: 1},
				{SpeakerID: spBob.ID, Score: 79.0, IsReply: false, SpeechOrder: 2},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// 6. Test Non-Admin (includeSilent=false): Alice has 2 substantive speeches (78, 82)
	standingsNonAdmin, err := store.GetSpeakerStandings("", false, false)
	if err != nil {
		t.Fatal(err)
	}
	var aliceNonAdmin *models.SpeakerStanding
	for i := range standingsNonAdmin {
		if standingsNonAdmin[i].SpeakerID == spAlice.ID {
			aliceNonAdmin = &standingsNonAdmin[i]
			break
		}
	}
	if aliceNonAdmin == nil {
		t.Fatal("Alice not found in speaker standings")
	}
	if aliceNonAdmin.SpeechCount != 2 {
		t.Fatalf("Alice speech count non-admin = %d, want 2", aliceNonAdmin.SpeechCount)
	}
	if aliceNonAdmin.TotalScore != 160.0 {
		t.Fatalf("Alice total score non-admin = %f, want 160.0", aliceNonAdmin.TotalScore)
	}
	if aliceNonAdmin.AverageScore != 80.0 {
		t.Fatalf("Alice avg score non-admin = %f, want 80.0", aliceNonAdmin.AverageScore)
	}

	// 7. Test Admin (includeSilent=true): Alice has 3 speeches (78, 82, 85)
	standingsAdmin, err := store.GetSpeakerStandings("", false, true)
	if err != nil {
		t.Fatal(err)
	}
	var aliceAdmin *models.SpeakerStanding
	for i := range standingsAdmin {
		if standingsAdmin[i].SpeakerID == spAlice.ID {
			aliceAdmin = &standingsAdmin[i]
			break
		}
	}
	if aliceAdmin == nil {
		t.Fatal("Alice not found in admin speaker standings")
	}
	if aliceAdmin.SpeechCount != 3 {
		t.Fatalf("Alice speech count admin = %d, want 3", aliceAdmin.SpeechCount)
	}
	if aliceAdmin.TotalScore != 245.0 {
		t.Fatalf("Alice total score admin = %f, want 245.0", aliceAdmin.TotalScore)
	}
	// Trimmed score drops 78 and 85 -> remaining is 82.0
	if aliceAdmin.TrimmedScore != 82.0 {
		t.Fatalf("Alice trimmed score admin = %f, want 82.0", aliceAdmin.TrimmedScore)
	}

	// 8. Test Category Filtering
	noviceStandings, err := store.GetSpeakerStandings("novice", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(noviceStandings) != 1 || noviceStandings[0].SpeakerID != spAlice.ID {
		t.Fatalf("novice speaker standings len=%d, want 1 (Alice)", len(noviceStandings))
	}

	eslStandings, err := store.GetSpeakerStandings("esl", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(eslStandings) != 1 || eslStandings[0].SpeakerID != spEve.ID {
		t.Fatalf("esl speaker standings len=%d, want 1 (Eve)", len(eslStandings))
	}
}

func TestFormatPresetApplication(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	presets := []struct {
		name       string
		wantSides  string
		wantSpPerT string
		wantReply  string
	}{
		{"bp", "OG,OO,CG,CO", "2", "false"},
		{"australs", "Aff,Neg", "3", "true"},
		{"asians", "Gov,Opp", "3", "true"},
		{"wsdc", "Prop,Opp", "3", "true"},
		{"apda", "Gov,Opp", "2", "false"},
	}

	for _, tc := range presets {
		if err := store.ApplyFormatPreset(tc.name); err != nil {
			t.Fatalf("ApplyFormatPreset(%q) failed: %v", tc.name, err)
		}
		sides, _ := store.GetConfig("sides")
		if sides != tc.wantSides {
			t.Errorf("preset %s sides = %q, want %q", tc.name, sides, tc.wantSides)
		}
		spPerT, _ := store.GetConfig("speakers_per_team")
		if spPerT != tc.wantSpPerT {
			t.Errorf("preset %s speakers_per_team = %q, want %q", tc.name, spPerT, tc.wantSpPerT)
		}
		hasReply, _ := store.GetConfig("has_reply_speeches")
		if hasReply != tc.wantReply {
			t.Errorf("preset %s has_reply_speeches = %q, want %q", tc.name, hasReply, tc.wantReply)
		}
	}

	if err := store.ApplyFormatPreset("invalid_preset"); err == nil {
		t.Error("ApplyFormatPreset(invalid) expected error, got nil")
	}
}

func TestInstitutionalBreakCaps(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Institutions
	instHarvard := uuid.New().String()
	instYale := uuid.New().String()
	instOxford := uuid.New().String()
	_ = store.CreateInstitution(instHarvard, "Harvard", "HAR")
	_ = store.CreateInstitution(instYale, "Yale", "YAL")
	_ = store.CreateInstitution(instOxford, "Oxford", "OXF")

	// Teams
	tHA := uuid.New().String()
	tHB := uuid.New().String()
	tHC := uuid.New().String()
	tYA := uuid.New().String()
	tOA := uuid.New().String()

	_ = store.CreateTeam(tHA, "Harvard A", "HA", instHarvard, nil, "tok1")
	_ = store.CreateTeam(tHB, "Harvard B", "HB", instHarvard, nil, "tok2")
	_ = store.CreateTeam(tHC, "Harvard C", "HC", instHarvard, nil, "tok3")
	_ = store.CreateTeam(tYA, "Yale A", "YA", instYale, nil, "tok4")
	_ = store.CreateTeam(tOA, "Oxford A", "OA", instOxford, nil, "tok5")

	// Create rounds & debates with points:
	// Harvard A: 9 pts
	// Harvard B: 8 pts
	// Harvard C: 7 pts
	// Yale A: 7 pts
	// Oxford A: 6 pts
	rID := uuid.New().String()
	_ = store.CreateRound(rID, 1, "R1", "preliminary")
	dID := uuid.New().String()
	_ = store.SaveDraw(rID, []models.DebateDrawInput{{DebateID: dID, Venue: "Room 1"}})

	bID := uuid.New().String()
	_ = store.SubmitBallot(dID, bID, "admin", "admin", "confirmed", false, "", []models.TeamBallotResult{
		{TeamID: tHA, Points: 9, SpeakerPoints: 300},
		{TeamID: tHB, Points: 8, SpeakerPoints: 290},
		{TeamID: tHC, Points: 7, SpeakerPoints: 285},
		{TeamID: tYA, Points: 7, SpeakerPoints: 280},
		{TeamID: tOA, Points: 6, SpeakerPoints: 270},
	})

	// Create BreakCategory: Size=4, MaxTeamsPerInstitution=2
	size := 4
	maxInst := 2
	catID := uuid.New().String()
	err := store.CreateBreakCategory(models.BreakCategory{
		ID:                     catID,
		Name:                   "Open Break",
		Seq:                    1,
		Size:                   &size,
		MaxTeamsPerInstitution: &maxInst,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := store.ComputeBreak(catID)
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 4 qualifiers: Harvard A (1), Harvard B (2), Yale A (3), Oxford A (4).
	// Harvard C should be excluded because Harvard already reached cap of 2.
	if len(res.Qualifiers) != 4 {
		t.Fatalf("got %d qualifiers, want 4. Qualifiers: %+v", len(res.Qualifiers), res.Qualifiers)
	}

	wantTeamIDs := []string{tHA, tHB, tYA, tOA}
	for i, q := range res.Qualifiers {
		if q.TeamID != wantTeamIDs[i] {
			t.Errorf("qualifier %d = %s, want %s", i+1, q.TeamName, wantTeamIDs[i])
		}
		if q.Rank != i+1 {
			t.Errorf("qualifier %d rank = %d, want %d", i+1, q.Rank, i+1)
		}
	}
}
