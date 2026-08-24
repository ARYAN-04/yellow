package db

import (
	"path/filepath"
	"testing"
	"yellow/internal/models"

	"github.com/google/uuid"
)

func TestFixes_StoreIntegrityAndCheckins(t *testing.T) {
	dir := t.TempDir()
	store, err := NewSQLiteStore(filepath.Join(dir, "fixes_verify.db"))
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer store.Close()

	// 1. Verify foreign keys are enabled by default
	var fk int
	err = store.db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil || fk != 1 {
		t.Fatalf("expected PRAGMA foreign_keys = 1, got %d, err: %v", fk, err)
	}

	// 2. Verify admin SetCheckedIn upsert for a brand new team
	teamID := uuid.New().String()
	err = store.CreateTeam(teamID, "Test Team", "TT", "", []models.SpeakerRequest{{Name: "Sp1"}, {Name: "Sp2"}}, "tok-t")
	if err != nil {
		t.Fatalf("CreateTeam failed: %v", err)
	}

	// Calling SetCheckedIn without prior ListCheckins must succeed
	if err := store.SetCheckedIn("team", teamID, true); err != nil {
		t.Fatalf("SetCheckedIn failed: %v", err)
	}

	checkins, err := store.ListCheckins()
	if err != nil {
		t.Fatalf("ListCheckins failed: %v", err)
	}
	var found bool
	for _, c := range checkins {
		if c.EntityID == teamID && c.CheckedIn {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected team %s to be checked in", teamID)
	}

	// 3. Verify ListTeams retains all speakers with slice growth (25 teams)
	for i := 0; i < 25; i++ {
		tid := uuid.New().String()
		_ = store.CreateTeam(tid, "Team "+string(rune('A'+i)), "T"+string(rune('A'+i)), "", []models.SpeakerRequest{
			{Name: "Speaker A"},
			{Name: "Speaker B"},
		}, "tok-"+tid)
	}

	teams, err := store.ListTeams()
	if err != nil {
		t.Fatalf("ListTeams failed: %v", err)
	}
	if len(teams) != 26 {
		t.Fatalf("expected 26 teams, got %d", len(teams))
	}
	missingSpeakers := 0
	for _, team := range teams {
		if len(team.Speakers) != 2 {
			missingSpeakers++
		}
	}
	if missingSpeakers > 0 {
		t.Fatalf("ListTeams dropped speakers from %d teams due to pointer aliasing!", missingSpeakers)
	}

	// 4. Verify cascade delete
	delID := teams[0].ID
	if err := store.DeleteTeam(delID); err != nil {
		t.Fatalf("DeleteTeam failed: %v", err)
	}
	var spCount int
	_ = store.db.QueryRow("SELECT COUNT(*) FROM speakers WHERE team_id = ?", delID).Scan(&spCount)
	if spCount != 0 {
		t.Fatalf("expected orphaned speakers to be cascade-deleted, found %d", spCount)
	}
}

func TestFixes_DoubleEntryDraftFlow(t *testing.T) {
	dir := t.TempDir()
	store, err := NewSQLiteStore(filepath.Join(dir, "ballot_verify.db"))
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer store.Close()

	rID := uuid.New().String()
	if err := store.CreateRound(rID, 1, "Round 1", "preliminary"); err != nil {
		t.Fatalf("CreateRound: %v", err)
	}
	t1 := uuid.New().String()
	t2 := uuid.New().String()
	_ = store.CreateTeam(t1, "Team 1", "T1", "", nil, "tok-1")
	_ = store.CreateTeam(t2, "Team 2", "T2", "", nil, "tok-2")

	dID := uuid.New().String()
	err = store.SaveDraw(rID, []models.DebateDrawInput{
		{
			DebateID: dID,
			Venue:    "Room 1",
			Teams: []models.TeamAssignment{
				{TeamID: t1, Side: "OG"},
				{TeamID: t2, Side: "OO"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveDraw: %v", err)
	}

	// Submit first double-entry draft in group "eg-1"
	b1ID := uuid.New().String()
	err = store.SubmitBallot(dID, b1ID, "organizer", "admin", "draft", false, "eg-1", []models.TeamBallotResult{
		{TeamID: t1, Points: 3, SpeakerPoints: 75},
		{TeamID: t2, Points: 2, SpeakerPoints: 74},
	})
	if err != nil {
		t.Fatalf("SubmitBallot 1: %v", err)
	}

	// Verify status is draft
	b1, err := store.GetBallotByID(b1ID)
	if err != nil {
		t.Fatalf("GetBallotByID: %v", err)
	}
	if b1.Status != "draft" {
		t.Fatalf("expected status draft, got %s", b1.Status)
	}

	// CompareEntryGroup with only 1 entry should report 1 pending
	pending, _, _, err := store.CompareEntryGroup("eg-1")
	if err != nil {
		t.Fatalf("CompareEntryGroup: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending entry, got %d", len(pending))
	}
}
