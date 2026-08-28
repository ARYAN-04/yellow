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

func setupWPCTestAPI(t *testing.T) (*API, *http.ServeMux, string, func()) {
	t.Helper()
	dir := t.TempDir()
	globalDBPath := filepath.Join(dir, "global.db")
	globalDB, err := db.InitGlobalDB(globalDBPath)
	if err != nil {
		t.Fatalf("failed to init global db: %v", err)
	}

	dbMgr := db.NewConnectionManager(globalDB, dir)
	api := NewAPI(globalDB, dbMgr)

	slug := "test-tourney"
	tID := uuid.New().String()
	tDBPath := filepath.Join(dir, "test-tourney.db")
	tDB, err := db.InitTournamentDB(tDBPath)
	if err != nil {
		t.Fatalf("failed to init tourney db: %v", err)
	}
	rID := "round-1"
	if _, err := tDB.Exec("INSERT INTO rounds (id, seq, name, stage, draw_released, results_released) VALUES (?, 1, 'Round 1', 'preliminary', 1, 1)", rID); err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO teams (id, name, code) VALUES ('team-1', 'Team One', 'T1')"); err != nil {
		t.Fatalf("failed to insert team: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO speakers (id, name, team_id) VALUES ('sp-1', 'Speaker One', 'team-1')"); err != nil {
		t.Fatalf("failed to insert speaker: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO debates (id, round_id, venue) VALUES ('deb-1', 'round-1', 'Room A')"); err != nil {
		t.Fatalf("failed to insert debate: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO debate_teams (id, debate_id, team_id, side) VALUES ('dt-1', 'deb-1', 'team-1', 'OG')"); err != nil {
		t.Fatalf("failed to insert debate team: %v", err)
	}
	tDB.Close()

	_, err = globalDB.Exec("INSERT INTO tournaments (id, name, slug, db_path) VALUES (?, 'Test Tournament', ?, ?)", tID, slug, tDBPath)
	if err != nil {
		t.Fatalf("failed to register tournament: %v", err)
	}

	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("GET /api/auth/me", api.CheckAuth)
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, "invalid request", http.StatusBadRequest)
			return
		}
		if req.Password == "admin" {
			http.SetCookie(w, &http.Cookie{Name: "admin_session", Value: "authenticated", Path: "/"})
			JSONResponse(w, map[string]interface{}{"status": "success", "role": "admin"}, http.StatusOK)
		} else if req.Password == "assistant" {
			http.SetCookie(w, &http.Cookie{Name: "admin_session", Value: "assistant", Path: "/"})
			JSONResponse(w, map[string]interface{}{"status": "success", "role": "assistant"}, http.StatusOK)
		} else {
			JSONError(w, "invalid password", http.StatusUnauthorized)
		}
	})

	// Routes
	mux.HandleFunc("POST /api/tournaments", api.RequireFullAdmin(api.CreateTournament))
	mux.HandleFunc("POST /api/t/{slug}/teams", api.RequireFullAdmin(api.WriteBlocker(api.CreateTeam)))
	mux.HandleFunc("POST /api/t/{slug}/rounds/{round_id}/draw", api.RequireFullAdmin(api.WriteBlocker(api.GenerateRoundDraw)))
	mux.HandleFunc("POST /api/t/{slug}/debates/{debate_id}/ballots", api.RequireAdmin(api.WriteBlocker(api.SubmitBallot)))
	mux.HandleFunc("POST /api/t/{slug}/checkins/{entity_type}/{entity_id}", api.RequireAdmin(api.WriteBlocker(api.SetCheckin)))
	mux.HandleFunc("GET /api/t/{slug}/teams/{id}/trajectory", api.GetTeamTrajectory)
	mux.HandleFunc("GET /api/t/{slug}/speakers/{id}/trajectory", api.GetSpeakerTrajectory)

	cleanup := func() {
		dbMgr.CloseAll()
		globalDB.Close()
	}
	return api, mux, slug, cleanup
}

func assistantRequest(req *http.Request) *http.Request {
	req.AddCookie(&http.Cookie{Name: "admin_session", Value: "assistant"})
	return req
}

func TestAssistantRolePermissions(t *testing.T) {
	_, mux, slug, cleanup := setupWPCTestAPI(t)
	defer cleanup()

	// 1. Check /api/auth/me for admin vs assistant vs anonymous
	// Admin
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, adminRequest(req))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin auth check, got %d", w.Code)
	}
	var adminAuth map[string]interface{}
	json.NewDecoder(w.Body).Decode(&adminAuth)
	if adminAuth["role"] != "admin" {
		t.Errorf("expected role 'admin', got %v", adminAuth["role"])
	}

	// Assistant
	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, assistantRequest(req))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for assistant auth check, got %d", w.Code)
	}
	var assistantAuth map[string]interface{}
	json.NewDecoder(w.Body).Decode(&assistantAuth)
	if assistantAuth["role"] != "assistant" {
		t.Errorf("expected role 'assistant', got %v", assistantAuth["role"])
	}

	// Anonymous
	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous auth check, got %d", w.Code)
	}

	// 2. Assistant CAN perform ballot entry
	ballotBody, _ := json.Marshal(SubmitBallotRequest{
		SubmitterType: "organizer",
		SubmitterID:   "assistant-1",
		Results: []models.TeamBallotResult{
			{
				TeamID:        "team-1",
				Points:        3,
				SpeakerPoints: 75.0,
			},
		},
	})
	req = httptest.NewRequest("POST", "/api/t/"+slug+"/debates/deb-1/ballots", bytes.NewReader(ballotBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, assistantRequest(req))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for assistant ballot entry, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Assistant CAN perform checkin
	req = httptest.NewRequest("POST", "/api/t/"+slug+"/checkins/team/team-1", bytes.NewReader([]byte(`{"checked_in": true}`)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, assistantRequest(req))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for assistant checkin, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Assistant CANNOT generate round draw (RequireFullAdmin -> 403 Forbidden)
	req = httptest.NewRequest("POST", "/api/t/"+slug+"/rounds/round-1/draw", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, assistantRequest(req))
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for assistant generating draw, got %d: %s", w.Code, w.Body.String())
	}

	// 5. Assistant CANNOT create teams (RequireFullAdmin -> 403 Forbidden)
	teamBody, _ := json.Marshal(map[string]interface{}{"name": "Team Two", "code": "T2"})
	req = httptest.NewRequest("POST", "/api/t/"+slug+"/teams", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, assistantRequest(req))
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for assistant creating team, got %d: %s", w.Code, w.Body.String())
	}

	// 6. Anonymous gets 401 for RequireFullAdmin and RequireAdmin
	req = httptest.NewRequest("POST", "/api/t/"+slug+"/teams", bytes.NewReader(teamBody))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for anonymous creating team, got %d", w.Code)
	}

	// 7. Test Trajectory endpoints
	req = httptest.NewRequest("GET", "/api/t/"+slug+"/teams/team-1/trajectory", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for team trajectory, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("GET", "/api/t/"+slug+"/speakers/sp-1/trajectory", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for speaker trajectory, got %d: %s", w.Code, w.Body.String())
	}
}
