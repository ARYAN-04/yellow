package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"yellow/internal/db"
	"yellow/internal/models"

	"github.com/google/uuid"
)

func setupWPETestAPI(t *testing.T) (*API, *http.ServeMux, string, string, func()) {
	t.Helper()
	dir := t.TempDir()
	globalDBPath := filepath.Join(dir, "global.db")
	globalDB, err := db.InitGlobalDB(globalDBPath)
	if err != nil {
		t.Fatalf("failed to init global db: %v", err)
	}

	dbMgr := db.NewConnectionManager(globalDB, dir)
	api := NewAPI(globalDB, dbMgr)

	slug := "wpe-tourney"
	tID := uuid.New().String()
	tDBPath := filepath.Join(dir, "wpe-tourney.db")
	tDB, err := db.InitTournamentDB(tDBPath)
	if err != nil {
		t.Fatalf("failed to init tourney db: %v", err)
	}

	// Insert into global DB
	_, err = globalDB.Exec("INSERT INTO tournaments (id, name, slug, db_path) VALUES (?, 'WPE Tournament', ?, ?)", tID, slug, tDBPath)
	if err != nil {
		t.Fatalf("failed to register tournament: %v", err)
	}

	// Insert test data in tournament DB
	rID := "r-1"
	if _, err := tDB.Exec("INSERT INTO rounds (id, seq, name, stage, draw_released, results_released) VALUES (?, 1, 'Round 1', 'preliminary', 1, 1)", rID); err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO teams (id, name, code) VALUES ('team-1', 'Team One', 'T1'), ('team-2', 'Team Two', 'T2')"); err != nil {
		t.Fatalf("failed to insert teams: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO adjudicators (id, name, test_score, rating) VALUES ('adj-1', 'Judge Judy', 85.0, 90.0), ('adj-2', 'Judge Joe', 75.0, NULL)"); err != nil {
		t.Fatalf("failed to insert adjudicators: %v", err)
	}
	debID := "deb-1"
	if _, err := tDB.Exec("INSERT INTO debates (id, round_id, venue) VALUES (?, 'r-1', 'Room 1')", debID); err != nil {
		t.Fatalf("failed to insert debate: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO debate_teams (id, debate_id, team_id, side) VALUES ('dt-1', 'deb-1', 'team-1', 'OG'), ('dt-2', 'deb-1', 'team-2', 'OO')"); err != nil {
		t.Fatalf("failed to insert debate teams: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO debate_adjudicators (id, debate_id, adjudicator_id, role) VALUES ('da-1', 'deb-1', 'adj-1', 'chair')"); err != nil {
		t.Fatalf("failed to insert debate adj: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/t/{slug}/rounds", api.ListRounds)
	mux.HandleFunc("GET /api/t/{slug}/rounds/{round_id}/draw", api.GetRoundDraw)
	mux.HandleFunc("POST /api/t/{slug}/debates/{debate_id}/adjudicators", api.AddDebateAdjudicator)
	mux.HandleFunc("DELETE /api/t/{slug}/debates/{debate_id}/adjudicators/{adj_assignment_id}", api.DeleteDebateAdjudicator)
	mux.HandleFunc("PUT /api/t/{slug}/debates/{debate_id}/adjudicators/{adj_assignment_id}", api.MoveAdjudicatorAssignment)
	mux.HandleFunc("GET /api/t/{slug}/standings/adjudicators", api.GetAdjudicatorStandings)
	mux.HandleFunc("GET /api/t/{slug}/adjudicators/{id}/trajectory", api.GetAdjudicatorTrajectory)

	cleanup := func() {
		dbMgr.CloseAll()
		globalDB.Close()
	}

	return api, mux, slug, debID, cleanup
}

func TestWPE_APIAdjudicatorAllocationsAndStandings(t *testing.T) {
	_, mux, slug, debID, cleanup := setupWPETestAPI(t)
	defer cleanup()

	// 1. Test POST /api/t/{slug}/debates/{debate_id}/adjudicators
	addBody, _ := json.Marshal(map[string]string{
		"adjudicator_id": "adj-2",
		"role":           "panel",
	})
	req := httptest.NewRequest("POST", "/api/t/"+slug+"/debates/"+debID+"/adjudicators", bytes.NewReader(addBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("POST add adjudicator failed with status %d: %s", rec.Code, rec.Body.String())
	}

	var draw []models.DebateDraw
	if err := json.Unmarshal(rec.Body.Bytes(), &draw); err != nil {
		t.Fatalf("failed to decode draw: %v", err)
	}
	if len(draw) != 1 || len(draw[0].Adjudicators) != 2 {
		t.Fatalf("expected 2 adjudicators in debate draw, got %d", len(draw[0].Adjudicators))
	}

	var da2ID string
	for _, a := range draw[0].Adjudicators {
		if a.AdjudicatorID == "adj-2" {
			da2ID = a.ID
			if a.Role != "panel" {
				t.Errorf("expected role panel, got %s", a.Role)
			}
		}
	}
	if da2ID == "" {
		t.Fatalf("adj-2 assignment ID not found in draw")
	}

	// 2. Test PUT /api/t/{slug}/debates/{debate_id}/adjudicators/{adj_assignment_id} (role change)
	roleBody, _ := json.Marshal(map[string]string{
		"role": "trainee",
	})
	req = httptest.NewRequest("PUT", "/api/t/"+slug+"/debates/"+debID+"/adjudicators/"+da2ID, bytes.NewReader(roleBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT role update failed with status %d: %s", rec.Code, rec.Body.String())
	}

	// 3. Test DELETE /api/t/{slug}/debates/{debate_id}/adjudicators/{adj_assignment_id}
	req = httptest.NewRequest("DELETE", "/api/t/"+slug+"/debates/"+debID+"/adjudicators/"+da2ID, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE adjudicator failed with status %d: %s", rec.Code, rec.Body.String())
	}

	var drawAfterDelete []models.DebateDraw
	_ = json.Unmarshal(rec.Body.Bytes(), &drawAfterDelete)
	if len(drawAfterDelete[0].Adjudicators) != 1 {
		t.Errorf("expected 1 adjudicator remaining after delete, got %d", len(drawAfterDelete[0].Adjudicators))
	}

	// 4. Test GET /api/t/{slug}/standings/adjudicators
	req = httptest.NewRequest("GET", "/api/t/"+slug+"/standings/adjudicators", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET adjudicator standings failed with status %d: %s", rec.Code, rec.Body.String())
	}
	var standings []models.AdjudicatorStanding
	if err := json.Unmarshal(rec.Body.Bytes(), &standings); err != nil {
		t.Fatalf("failed to decode standings: %v", err)
	}
	if len(standings) != 2 {
		t.Fatalf("expected 2 adjudicators in standings, got %d", len(standings))
	}
	if standings[0].ID != "adj-1" {
		t.Errorf("expected rank 1 to be adj-1 (rating 90.0), got %s", standings[0].ID)
	}

	// 5. Test GET /api/t/{slug}/adjudicators/{id}/trajectory
	req = httptest.NewRequest("GET", "/api/t/"+slug+"/adjudicators/adj-1/trajectory", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET adjudicator trajectory failed with status %d: %s", rec.Code, rec.Body.String())
	}
	var traj models.AdjudicatorTrajectory
	if err := json.Unmarshal(rec.Body.Bytes(), &traj); err != nil {
		t.Fatalf("failed to decode adjudicator trajectory: %v", err)
	}
	if traj.Adjudicator.Name != "Judge Judy" {
		t.Errorf("expected name Judge Judy, got %s", traj.Adjudicator.Name)
	}
	if len(traj.Debates) != 1 {
		t.Errorf("expected 1 debate in trajectory, got %d", len(traj.Debates))
	}
}
