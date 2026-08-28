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

func setupWPDTestAPI(t *testing.T) (*API, *http.ServeMux, string, string, string, func()) {
	t.Helper()
	dir := t.TempDir()
	globalDBPath := filepath.Join(dir, "global.db")
	globalDB, err := db.InitGlobalDB(globalDBPath)
	if err != nil {
		t.Fatalf("failed to init global db: %v", err)
	}

	dbMgr := db.NewConnectionManager(globalDB, dir)
	api := NewAPI(globalDB, dbMgr)

	slug := "wpd-tourney"
	tID := uuid.New().String()
	tDBPath := filepath.Join(dir, "wpd-tourney.db")
	tDB, err := db.InitTournamentDB(tDBPath)
	if err != nil {
		t.Fatalf("failed to init tourney db: %v", err)
	}

	// Insert round, teams, speakers, adjudicators, debate
	rID := "r-1"
	if _, err := tDB.Exec("INSERT INTO rounds (id, seq, name, stage, draw_released, results_released) VALUES (?, 1, 'Round 1', 'preliminary', 1, 1)", rID); err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO teams (id, name, code) VALUES ('team-1', 'Team One', 'T1'), ('team-2', 'Team Two', 'T2')"); err != nil {
		t.Fatalf("failed to insert teams: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO speakers (id, name, team_id) VALUES ('sp-1', 'Alice', 'team-1'), ('sp-2', 'Bob', 'team-1'), ('sp-3', 'Charlie', 'team-2'), ('sp-4', 'Dave', 'team-2')"); err != nil {
		t.Fatalf("failed to insert speakers: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO adjudicators (id, name, test_score) VALUES ('adj-1', 'Judge Judy', 85.0)"); err != nil {
		t.Fatalf("failed to insert adjudicator: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO debates (id, round_id, venue) VALUES ('deb-1', 'r-1', 'Room 1')"); err != nil {
		t.Fatalf("failed to insert debate: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO debate_teams (id, debate_id, team_id, side) VALUES ('dt-1', 'deb-1', 'team-1', 'OG'), ('dt-2', 'deb-1', 'team-2', 'OO')"); err != nil {
		t.Fatalf("failed to insert debate teams: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO debate_adjudicators (id, debate_id, adjudicator_id, role) VALUES ('da-1', 'deb-1', 'adj-1', 'chair')"); err != nil {
		t.Fatalf("failed to insert debate adjs: %v", err)
	}
	if _, err := tDB.Exec("INSERT INTO access_tokens (token, type, owner_id) VALUES ('tok-team', 'team', 'team-1'), ('tok-adj', 'adjudicator', 'adj-1')"); err != nil {
		t.Fatalf("failed to insert tokens: %v", err)
	}
	tDB.Close()

	if _, err := globalDB.Exec("INSERT INTO tournaments (id, name, slug, db_path) VALUES (?, 'WPD Tournament', ?, ?)", tID, slug, tDBPath); err != nil {
		t.Fatalf("failed to register tournament: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/token/{token}", api.ResolveToken)
	mux.HandleFunc("GET /api/token/{token}/debates", api.GetTokenDebates)
	mux.HandleFunc("POST /api/token/{token}/debates/{debate_id}/ballots", api.WriteBlocker(api.SubmitTokenBallot))
	mux.HandleFunc("GET /api/token/{token}/checkin", api.GetTokenCheckin)
	mux.HandleFunc("POST /api/token/{token}/checkin", api.WriteBlocker(api.SetTokenCheckin))

	cleanup := func() {
		globalDB.Close()
	}

	return api, mux, slug, "tok-team", "tok-adj", cleanup
}

func TestWPD_TokenCheckinAndBallotSubmission(t *testing.T) {
	_, mux, _, teamTok, adjTok, cleanup := setupWPDTestAPI(t)
	defer cleanup()

	// 1. GET checkin for team token (initially false)
	req := httptest.NewRequest("GET", "/api/token/"+teamTok+"/checkin", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET checkin, got %d: %s", w.Code, w.Body.String())
	}
	var checkinResp struct {
		CheckedIn  bool   `json:"checked_in"`
		EntityType string `json:"entity_type"`
		EntityName string `json:"entity_name"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &checkinResp); err != nil {
		t.Fatalf("failed to decode checkin response: %v", err)
	}
	if checkinResp.CheckedIn {
		t.Errorf("expected initially not checked in, got true")
	}
	if checkinResp.EntityType != "team" || checkinResp.EntityName != "Team One" {
		t.Errorf("unexpected checkin entity info: %+v", checkinResp)
	}

	// 2. POST checkin to toggle to checked_in = true
	req = httptest.NewRequest("POST", "/api/token/"+teamTok+"/checkin", bytes.NewReader([]byte(`{"checked_in": true}`)))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for POST checkin, got %d: %s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &checkinResp); err != nil {
		t.Fatalf("failed to decode POST checkin response: %v", err)
	}
	if !checkinResp.CheckedIn {
		t.Errorf("expected checked_in to be true after POST")
	}

	// 3. GET checkin again to verify persisted
	req = httptest.NewRequest("GET", "/api/token/"+teamTok+"/checkin", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if err := json.Unmarshal(w.Body.Bytes(), &checkinResp); err != nil {
		t.Fatalf("failed to decode GET checkin response: %v", err)
	}
	if !checkinResp.CheckedIn {
		t.Errorf("expected persisted checked_in to be true")
	}

	// 4. Submit Ballot with Speaker Scores and Roles via Adjudicator token
	ballotBody, _ := json.Marshal(SubmitBallotRequest{
		SubmitterType: "adjudicator",
		SubmitterID:   "adj-1",
		Results: []models.TeamBallotResult{
			{
				TeamID:        "team-1",
				Points:        3,
				SpeakerPoints: 153.0,
				SpeakerScores: []models.SpeakerScoreInput{
					{SpeakerID: "sp-1", Score: 77.0, IsReply: false, SpeechOrder: 1, Role: "PM"},
					{SpeakerID: "sp-2", Score: 76.0, IsReply: false, SpeechOrder: 2, Role: "DPM"},
				},
			},
			{
				TeamID:        "team-2",
				Points:        2,
				SpeakerPoints: 150.0,
				SpeakerScores: []models.SpeakerScoreInput{
					{SpeakerID: "sp-3", Score: 75.0, IsReply: false, SpeechOrder: 1, Role: "LO"},
					{SpeakerID: "sp-4", Score: 75.0, IsReply: false, SpeechOrder: 2, Role: "DLO"},
				},
			},
		},
	})

	req = httptest.NewRequest("POST", "/api/token/"+adjTok+"/debates/deb-1/ballots", bytes.NewReader(ballotBody))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for submit token ballot, got %d: %s", w.Code, w.Body.String())
	}
}
