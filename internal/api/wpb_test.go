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
	_ "modernc.org/sqlite"
)

func setupTestAPI(t *testing.T) (*API, *http.ServeMux, string, func()) {
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
	if _, err := tDB.Exec("INSERT INTO rounds (id, seq, name, stage) VALUES (?, 1, 'Round 1', 'preliminary')", rID); err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}
	tDB.Close()

	_, err = globalDB.Exec("INSERT INTO tournaments (id, name, slug, db_path) VALUES (?, 'Test Tournament', ?, ?)", tID, slug, tDBPath)
	if err != nil {
		t.Fatalf("failed to register tournament: %v", err)
	}

	mux := http.NewServeMux()
	// Motions & Vetoes
	mux.HandleFunc("GET /api/t/{slug}/rounds/{round_id}/motions", api.ListRoundMotions)
	mux.HandleFunc("POST /api/t/{slug}/rounds/{round_id}/motions", api.RequireAdmin(api.WriteBlocker(api.CreateRoundMotion)))
	mux.HandleFunc("PUT /api/t/{slug}/motions/{id}", api.RequireAdmin(api.WriteBlocker(api.UpdateMotion)))
	mux.HandleFunc("DELETE /api/t/{slug}/motions/{id}", api.RequireAdmin(api.WriteBlocker(api.DeleteMotion)))
	mux.HandleFunc("POST /api/t/{slug}/rounds/{round_id}/motions/release", api.RequireAdmin(api.WriteBlocker(api.ReleaseRoundMotions)))
	mux.HandleFunc("GET /api/t/{slug}/debates/{debate_id}/vetoes", api.GetDebateVetoes)
	mux.HandleFunc("POST /api/t/{slug}/debates/{debate_id}/vetoes", api.RequireAdmin(api.WriteBlocker(api.RecordDebateVetoes)))
	mux.HandleFunc("GET /api/t/{slug}/motions/statistics", api.GetMotionStatistics)

	// Venues
	mux.HandleFunc("GET /api/t/{slug}/venues", api.RequireAdmin(api.ListVenues))
	mux.HandleFunc("POST /api/t/{slug}/venues", api.RequireAdmin(api.WriteBlocker(api.CreateVenue)))
	mux.HandleFunc("PUT /api/t/{slug}/venues/{id}", api.RequireAdmin(api.WriteBlocker(api.UpdateVenue)))
	mux.HandleFunc("DELETE /api/t/{slug}/venues/{id}", api.RequireAdmin(api.WriteBlocker(api.DeleteVenue)))

	cleanup := func() {
		dbMgr.CloseAll()
		globalDB.Close()
	}
	return api, mux, slug, cleanup
}

func adminRequest(req *http.Request) *http.Request {
	req.AddCookie(&http.Cookie{Name: "admin_session", Value: "authenticated"})
	return req
}

func TestAPIMotionsAndVenues(t *testing.T) {
	_, mux, slug, cleanup := setupTestAPI(t)
	defer cleanup()

	rID := "round-1"

	// Insert round into tourney DB directly via store
	// We'll test POST /api/t/{slug}/rounds/{round_id}/motions
	// Let's create the round directly
	// Get DB
	// Let's do a venue creation first
	venuePayload, _ := json.Marshal(models.Venue{
		Name:         "Hall 101",
		Priority:     50,
		IsAccessible: true,
	})
	req := httptest.NewRequest("POST", "/api/t/"+slug+"/venues", bytes.NewReader(venuePayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, adminRequest(req))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for venue, got %d: %s", w.Code, w.Body.String())
	}

	var createdVenue models.Venue
	_ = json.NewDecoder(w.Body).Decode(&createdVenue)

	// List venues
	req = httptest.NewRequest("GET", "/api/t/"+slug+"/venues", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, adminRequest(req))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var venues []models.Venue
	_ = json.NewDecoder(w.Body).Decode(&venues)
	if len(venues) != 1 || venues[0].Name != "Hall 101" {
		t.Fatalf("expected 1 venue 'Hall 101', got %+v", venues)
	}

	// Create a motion for round
	motionPayload, _ := json.Marshal(models.Motion{
		Text:      "This House would tax sugary drinks.",
		Reference: "THW Tax",
		InfoSlide: "Context on sugar consumption.",
	})
	req = httptest.NewRequest("POST", "/api/t/"+slug+"/rounds/"+rID+"/motions", bytes.NewReader(motionPayload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, adminRequest(req))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for motion, got %d: %s", w.Code, w.Body.String())
	}

	var createdMotion models.Motion
	_ = json.NewDecoder(w.Body).Decode(&createdMotion)

	// List motions as admin
	req = httptest.NewRequest("GET", "/api/t/"+slug+"/rounds/"+rID+"/motions", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, adminRequest(req))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	var motions []models.Motion
	_ = json.NewDecoder(w.Body).Decode(&motions)
	if len(motions) != 1 {
		t.Fatalf("expected 1 motion as admin, got %d", len(motions))
	}

	// List motions as non-admin (unreleased -> should be empty)
	req = httptest.NewRequest("GET", "/api/t/"+slug+"/rounds/"+rID+"/motions", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	var pubMotions []models.Motion
	_ = json.NewDecoder(w.Body).Decode(&pubMotions)
	if len(pubMotions) != 0 {
		t.Fatalf("expected 0 public motions before release, got %d", len(pubMotions))
	}

	// Release motions
	relPayload, _ := json.Marshal(map[string]bool{"release": true})
	req = httptest.NewRequest("POST", "/api/t/"+slug+"/rounds/"+rID+"/motions/release", bytes.NewReader(relPayload))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, adminRequest(req))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on release, got %d", w.Code)
	}

	// Now non-admin should see the motion
	req = httptest.NewRequest("GET", "/api/t/"+slug+"/rounds/"+rID+"/motions", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	_ = json.NewDecoder(w.Body).Decode(&pubMotions)
	if len(pubMotions) != 1 {
		t.Fatalf("expected 1 public motion after release, got %d", len(pubMotions))
	}
}
