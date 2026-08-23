package draw

import (
	"database/sql"
	"testing"

	"yellow/internal/db"
)

func TestGenerateDrawRound1(t *testing.T) {
	// Initialize in-memory SQLite database
	tdb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer tdb.Close()

	// Apply tournament schema DDL
	if _, err := tdb.Exec(db.TournamentSchema); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}

	// Create 4 teams (enough for 1 BP debate)
	teams := []struct {
		ID   string
		Name string
	}{
		{"team-1", "A-Team"},
		{"team-2", "B-Team"},
		{"team-3", "C-Team"},
		{"team-4", "D-Team"},
	}

	for _, tm := range teams {
		_, err := tdb.Exec("INSERT INTO teams (id, name) VALUES (?, ?)", tm.ID, tm.Name)
		if err != nil {
			t.Fatalf("failed to insert team: %v", err)
		}
	}

	// Create a round
	roundID := "round-1"
	_, err = tdb.Exec("INSERT INTO rounds (id, seq, name, stage) VALUES (?, 1, 'Round 1', 'preliminary')", roundID)
	if err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}

	// Generate draw
	store := db.NewSQLTournamentStore(tdb)
	err = GenerateDraw(store, roundID)
	if err != nil {
		t.Fatalf("failed to generate draw: %v", err)
	}

	// Verify debate was created
	var debateCount int
	err = tdb.QueryRow("SELECT COUNT(*) FROM debates WHERE round_id = ?", roundID).Scan(&debateCount)
	if err != nil {
		t.Fatalf("failed to query debates: %v", err)
	}
	if debateCount != 1 {
		t.Errorf("expected 1 debate, got %d", debateCount)
	}

	// Verify debate teams were assigned with correct sides (OG, OO, CG, CO)
	rows, err := tdb.Query("SELECT team_id, side FROM debate_teams")
	if err != nil {
		t.Fatalf("failed to query debate teams: %v", err)
	}
	defer rows.Close()

	assignedSides := make(map[string]string)
	for rows.Next() {
		var tid, side string
		if err := rows.Scan(&tid, &side); err != nil {
			t.Fatalf("failed to scan debate team: %v", err)
		}
		assignedSides[tid] = side
	}

	if len(assignedSides) != 4 {
		t.Errorf("expected 4 team side assignments, got %d", len(assignedSides))
	}

	// Ensure all four BP side positions are represented
	sidesMap := make(map[string]bool)
	for _, side := range assignedSides {
		sidesMap[side] = true
	}

	for _, side := range []string{"OG", "OO", "CG", "CO"} {
		if !sidesMap[side] {
			t.Errorf("missing expected side position assignment: %s", side)
		}
	}
}
