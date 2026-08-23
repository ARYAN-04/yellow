package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"yellow/internal/models"
)

// GetBreak computes the live qualifier list for a break category.
func (api *API) GetBreak(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	categoryID := r.PathValue("category_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	res, err := tdb.ComputeBreak(categoryID)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSONResponse(w, res, http.StatusOK)
}

// PublishBreak persists a break snapshot so brackets survive later data changes.
func (api *API) PublishBreak(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	categoryID := r.PathValue("category_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	res, err := tdb.ComputeBreak(categoryID)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := tdb.SaveBreakSnapshot(categoryID, res.Qualifiers); err != nil {
		JSONError(w, "failed to persist break: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, res, http.StatusOK)
}

type GenerateBracketRequest struct {
	CategoryID string `json:"category_id"`
}

// GenerateBracket seeds an elimination round 1vN from the published break.
func (api *API) GenerateBracket(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req GenerateBracketRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	categoryID := strings.TrimSpace(req.CategoryID)
	if categoryID == "" {
		categoryID = "open"
	}

	if err := tdb.GenerateBracket(roundID, categoryID); err != nil {
		JSONError(w, "bracket generation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	JSONResponse(w, map[string]string{"status": "success", "category_id": categoryID}, http.StatusOK)
}

// AdvanceRound creates the next elimination round from the confirmed winners of this one.
func (api *API) AdvanceRound(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	newID, err := tdb.AdvanceEliminationRound(roundID)
	if err != nil {
		JSONError(w, "advance failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	round, err := tdb.GetRound(newID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, round, http.StatusCreated)
}

// GetBracket returns all elimination rounds for the visualizer.
func (api *API) GetBracket(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	rounds, err := tdb.GetBracket()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if rounds == nil {
		rounds = []models.BracketRound{}
	}

	JSONResponse(w, rounds, http.StatusOK)
}
