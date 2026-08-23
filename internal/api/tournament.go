package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"yellow/internal/db"
	"yellow/internal/draw"
	"yellow/internal/models"

	"github.com/google/uuid"
)

// --- Institutions Handlers ---

type InstitutionRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func (api *API) ListInstitutions(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	results, err := tdb.ListInstitutions()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

func (api *API) CreateInstitution(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req InstitutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))

	if req.Name == "" || req.Code == "" {
		JSONError(w, "name and code are required", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	err = tdb.CreateInstitution(id, req.Name, req.Code)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "institution name or code already exists", http.StatusConflict)
		} else {
			JSONError(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"id": id, "name": req.Name, "code": req.Code}, http.StatusCreated)
}

func (api *API) UpdateInstitution(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req InstitutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))

	if req.Name == "" || req.Code == "" {
		JSONError(w, "name and code are required", http.StatusBadRequest)
		return
	}

	err = tdb.UpdateInstitution(id, req.Name, req.Code)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "institution not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "institution name or code already exists", http.StatusConflict)
		} else {
			JSONError(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"id": id, "name": req.Name, "code": req.Code}, http.StatusOK)
}

func (api *API) DeleteInstitution(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	err = tdb.DeleteInstitution(id)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "institution not found", http.StatusNotFound)
		} else {
			JSONError(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Teams & Speakers Handlers ---

type TeamRequest struct {
	Name          string                  `json:"name"`
	Code          string                  `json:"code"`
	InstitutionID string                  `json:"institution_id"`
	Speakers      []models.SpeakerRequest `json:"speakers"`
}

func (api *API) ListTeams(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	teamList, err := tdb.ListTeams()
	if err != nil {
		JSONError(w, "query teams failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, teamList, http.StatusOK)
}

func (api *API) CreateTeam(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Code = strings.TrimSpace(req.Code)

	if req.Name == "" {
		JSONError(w, "team name is required", http.StatusBadRequest)
		return
	}

	teamID := uuid.New().String()
	token := uuid.New().String()

	err = tdb.CreateTeam(teamID, req.Name, req.Code, req.InstitutionID, req.Speakers, token)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "team name already exists", http.StatusConflict)
		} else {
			JSONError(w, "team insertion failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	type spResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var createdSpeakers []spResp
	for _, s := range req.Speakers {
		sName := strings.TrimSpace(s.Name)
		if sName == "" {
			continue
		}
		createdSpeakers = append(createdSpeakers, spResp{ID: "", Name: sName})
	}

	var instID interface{} = nil
	if req.InstitutionID != "" {
		instID = req.InstitutionID
	}

	JSONResponse(w, map[string]interface{}{
		"id":             teamID,
		"name":           req.Name,
		"code":           req.Code,
		"institution_id": instID,
		"speakers":       createdSpeakers,
		"token":          token,
	}, http.StatusCreated)
}

func (api *API) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	err = tdb.DeleteTeam(id)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "team not found", http.StatusNotFound)
		} else {
			JSONError(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Adjudicators Handlers ---

type AdjudicatorRequest struct {
	Name          string  `json:"name"`
	InstitutionID string  `json:"institution_id"`
	TestScore     float64 `json:"test_score"`
}

func (api *API) ListAdjudicators(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	results, err := tdb.ListAdjudicators()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

func (api *API) CreateAdjudicator(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req AdjudicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		JSONError(w, "name is required", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	token := uuid.New().String()

	err = tdb.CreateAdjudicator(id, req.Name, req.InstitutionID, req.TestScore, token)
	if err != nil {
		JSONError(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var instID interface{} = nil
	if req.InstitutionID != "" {
		instID = req.InstitutionID
	}

	JSONResponse(w, map[string]interface{}{
		"id":             id,
		"name":           req.Name,
		"institution_id": instID,
		"test_score":     req.TestScore,
		"token":          token,
	}, http.StatusCreated)
}

func (api *API) DeleteAdjudicator(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	err = tdb.DeleteAdjudicator(id)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "adjudicator not found", http.StatusNotFound)
		} else {
			JSONError(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Rounds Handlers ---

type RoundRequest struct {
	Name  string `json:"name"`
	Seq   int    `json:"seq"`
	Stage string `json:"stage"` // 'preliminary', 'elimination'
}

func (api *API) ListRounds(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	results, err := tdb.ListRounds()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

func (api *API) CreateRound(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req RoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Stage = strings.TrimSpace(req.Stage)
	if req.Name == "" || req.Stage == "" {
		JSONError(w, "name and stage are required", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	err = tdb.CreateRound(id, req.Seq, req.Name, req.Stage)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "a round with this sequence number already exists", http.StatusConflict)
		} else {
			JSONError(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]interface{}{
		"id":               id,
		"seq":              req.Seq,
		"name":             req.Name,
		"stage":            req.Stage,
		"silent":           false,
		"draw_released":    false,
		"results_released": false,
	}, http.StatusCreated)
}

type UpdateRoundRequest struct {
	Silent          *bool `json:"silent"`
	DrawReleased    *bool `json:"draw_released"`
	ResultsReleased *bool `json:"results_released"`
}

func (api *API) UpdateRound(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req UpdateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = tdb.UpdateRound(roundID, req.Silent, req.DrawReleased, req.ResultsReleased)
	if err != nil {
		JSONError(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

// GenerateRoundDraw handles the generation of pairings and allocations for the round.
func (api *API) GenerateRoundDraw(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	err = draw.GenerateDraw(tdb, roundID)
	if err != nil {
		JSONError(w, "draw generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

// GetRoundDraw retrieves all debates and their side assignments/panels for the round.
func (api *API) GetRoundDraw(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	debates, err := tdb.GetRoundDraw(roundID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, debates, http.StatusOK)
}

// SubmitBallotRequest represents the payload for ballot submission.
type SubmitBallotRequest struct {
	SubmitterType string                    `json:"submitter_type"`
	SubmitterID   string                    `json:"submitter_id"`
	IsSplit       bool                      `json:"is_split"`
	EntryGroup    string                    `json:"entry_group"`
	Results       []models.TeamBallotResult `json:"results"`
}

// validateBallotRequest enforces score scale and split/consensus rules on a ballot payload.
func validateBallotRequest(req *SubmitBallotRequest, scoreMin, scoreMax float64) error {
	if len(req.Results) == 0 {
		return errors.New("results are required")
	}
	seen := make(map[string]bool)
	for _, res := range req.Results {
		if res.Points < 0 {
			return errors.New("points must be >= 0")
		}
		if res.SpeakerPoints < scoreMin || res.SpeakerPoints > scoreMax {
			return fmt.Errorf("speaker score %s out of range [%s,%s]",
				strconv.FormatFloat(res.SpeakerPoints, 'f', -1, 64),
				strconv.FormatFloat(scoreMin, 'f', -1, 64),
				strconv.FormatFloat(scoreMax, 'f', -1, 64))
		}
		hasAdj := res.AdjudicatorID != nil && *res.AdjudicatorID != ""
		if !req.IsSplit && hasAdj {
			return errors.New("consensus ballots must not specify adjudicator_id on results")
		}
		if req.IsSplit && !hasAdj {
			return errors.New("split ballots require adjudicator_id on every result")
		}
		adj := ""
		if hasAdj {
			adj = *res.AdjudicatorID
		}
		key := res.TeamID + "|" + adj
		if seen[key] {
			return errors.New("each adjudicator may appear at most once per team")
		}
		seen[key] = true
	}
	return nil
}

func ballotScoreBounds(tdb db.TournamentStore) (float64, float64) {
	scoreMin, scoreMax := 0.0, 100.0
	if v, err := tdb.GetConfig("score_min"); err == nil {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil {
			scoreMin = f
		}
	}
	if v, err := tdb.GetConfig("score_max"); err == nil {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil {
			scoreMax = f
		}
	}
	return scoreMin, scoreMax
}

// SubmitBallot records ballot results for a debate.
func (api *API) SubmitBallot(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	debateID := r.PathValue("debate_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req SubmitBallotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	scoreMin, scoreMax := ballotScoreBounds(tdb)
	if verr := validateBallotRequest(&req, scoreMin, scoreMax); verr != nil {
		JSONError(w, verr.Error(), http.StatusBadRequest)
		return
	}

	ballotID := uuid.New().String()
	err = tdb.SubmitBallot(debateID, ballotID, req.SubmitterType, req.SubmitterID, "submitted", req.IsSplit, req.EntryGroup, req.Results)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "debate not found", http.StatusNotFound)
		} else {
			JSONError(w, "failed to submit ballot: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"id": ballotID, "status": "submitted"}, http.StatusCreated)
}

// ConfirmBallot transitions a ballot's status to 'confirmed', incorporating it into standings.
// For double-entry ballots the sibling draft is compared first: a match confirms both,
// a mismatch flags both as 'discrepancy' and returns the diffs.
func (api *API) ConfirmBallot(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	ballotID := r.PathValue("ballot_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	if !api.IsAdmin(r) {
		JSONError(w, "unauthorized: organizer admin required", http.StatusUnauthorized)
		return
	}

	err = confirmWithDoubleEntry(tdb, ballotID)
	if err != nil {
		if dispErr, ok := err.(*DiscrepancyError); ok {
			JSONResponse(w, map[string]interface{}{"error": "discrepancy", "diffs": dispErr.Diffs}, http.StatusConflict)
			return
		}
		if err == sql.ErrNoRows {
			JSONError(w, "ballot not found", http.StatusNotFound)
		} else {
			JSONError(w, "failed to confirm ballot: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"id": ballotID, "status": "confirmed"}, http.StatusOK)
}

// DiscrepancyError signals mismatched double-entry ballots along with their field diffs.
type DiscrepancyError struct {
	Diffs []models.BallotDiff
}

func (e *DiscrepancyError) Error() string { return "double-entry ballots do not match" }

// confirmWithDoubleEntry runs the double-entry comparison flow for a ballot confirmation.
func confirmWithDoubleEntry(tdb db.TournamentStore, ballotID string) error {
	ballot, err := tdb.GetBallotByID(ballotID)
	if err != nil {
		return err
	}

	if ballot.EntryGroup == nil || *ballot.EntryGroup == "" {
		return tdb.ConfirmBallot(ballotID)
	}

	pending, _, _, err := tdb.CompareEntryGroup(*ballot.EntryGroup)
	if err != nil {
		return err
	}
	if len(pending) < 2 {
		return tdb.ConfirmBallot(ballotID)
	}

	diffs := db.CompareBallotSummaries(pending[0], pending[1])
	if len(diffs) > 0 {
		for _, b := range pending {
			if serr := tdb.SetBallotStatus(b.ID, "discrepancy"); serr != nil {
				return serr
			}
		}
		return &DiscrepancyError{Diffs: diffs}
	}

	for _, b := range pending {
		if serr := tdb.SetBallotStatus(b.ID, "confirmed"); serr != nil {
			return serr
		}
	}
	return nil
}

// GetRoundBallots returns the ballot registry for a round.
func (api *API) GetRoundBallots(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	results, err := tdb.GetBallotsForRound(roundID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

// GetStandings dynamically computes team rankings based on confirmed ballots.
func (api *API) GetStandings(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var precedence []string
	if p := strings.TrimSpace(r.URL.Query().Get("precedence")); p != "" {
		precedence = strings.Split(p, ",")
	} else if v, cerr := tdb.GetConfig("ranking_precedence"); cerr == nil {
		_ = json.Unmarshal([]byte(v), &precedence)
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))

	// Non-admin viewers must not infer results of silent, unreleased rounds.
	list, err := tdb.GetStandingsWithPrecedenceEx(precedence, category, api.IsAdmin(r))
	if err != nil {
		JSONError(w, "failed to compute standings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, list, http.StatusOK)
}
