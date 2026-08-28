package draw

import (
	"database/sql"
	"testing"

	"yellow/internal/db"
	"yellow/internal/models"
)

func TestDrawVenueAllocation(t *testing.T) {
	tdb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer tdb.Close()

	if _, err := tdb.Exec(db.TournamentSchema); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}

	store := db.NewSQLTournamentStore(tdb)

	// Create 8 teams (2 debates)
	for i := 1; i <= 8; i++ {
		_, err := tdb.Exec("INSERT INTO teams (id, name) VALUES (?, ?)", string(rune('A'+i-1)), string(rune('A'+i-1)))
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create venues: Premier Hall (priority 100), Room B (priority 20)
	if err := store.CreateVenue(models.Venue{
		ID:           "v1",
		Name:         "Premier Hall",
		Priority:     100,
		IsAccessible: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateVenue(models.Venue{
		ID:           "v2",
		Name:         "Room B",
		Priority:     20,
		IsAccessible: false,
	}); err != nil {
		t.Fatal(err)
	}

	// Round 1
	roundID := "r1"
	_, err = tdb.Exec("INSERT INTO rounds (id, seq, name, stage) VALUES (?, 1, 'Round 1', 'preliminary')", roundID)
	if err != nil {
		t.Fatal(err)
	}

	if err := GenerateDraw(store, roundID); err != nil {
		t.Fatalf("failed to generate draw: %v", err)
	}

	drawResult, err := store.GetRoundDraw(roundID)
	if err != nil {
		t.Fatalf("failed to get round draw: %v", err)
	}
	if len(drawResult) != 2 {
		t.Fatalf("expected 2 debates, got %d", len(drawResult))
	}

	// Debate 1 should have Premier Hall (accessible) and Debate 2 should have Room B
	venueNames := []string{drawResult[0].Venue, drawResult[1].Venue}
	hasPremier := false
	hasRoomB := false
	for _, d := range drawResult {
		if d.Venue == "Premier Hall" {
			hasPremier = true
			if !d.VenueAccessible {
				t.Errorf("expected Premier Hall to be accessible")
			}
		}
		if d.Venue == "Room B" {
			hasRoomB = true
			if d.VenueAccessible {
				t.Errorf("expected Room B not to be accessible")
			}
		}
	}
	if !hasPremier || !hasRoomB {
		t.Errorf("expected both Premier Hall and Room B assigned, got venues: %v", venueNames)
	}
}
