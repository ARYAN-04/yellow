package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

var conflictSubjectTypes = map[string]bool{"adjudicator": true, "team": true}
var conflictTargetTypes = map[string]bool{"team": true, "speaker": true, "adjudicator": true, "institution": true}
var conflictWeights = map[string]bool{"hard": true, "soft": true}

type ConflictRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"`
	TargetType  string `json:"target_type"`
	TargetID    string `json:"target_id"`
	Weight      string `json:"weight"`
}

func (api *API) ListConflicts(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	results, err := tdb.ListConflicts()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

func (api *API) CreateConflict(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req ConflictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.SubjectType = strings.TrimSpace(req.SubjectType)
	req.TargetType = strings.TrimSpace(req.TargetType)
	req.Weight = strings.TrimSpace(req.Weight)
	if req.Weight == "" {
		req.Weight = "soft"
	}

	if !conflictSubjectTypes[req.SubjectType] || !conflictTargetTypes[req.TargetType] {
		JSONError(w, "invalid subject or target type", http.StatusBadRequest)
		return
	}
	if !conflictWeights[req.Weight] {
		JSONError(w, "weight must be 'hard' or 'soft'", http.StatusBadRequest)
		return
	}
	if req.SubjectID == "" || req.TargetID == "" {
		JSONError(w, "subject_id and target_id are required", http.StatusBadRequest)
		return
	}
	if req.SubjectID == req.TargetID && req.SubjectType == req.TargetType {
		JSONError(w, "an entity cannot conflict with itself", http.StatusBadRequest)
		return
	}

	err = tdb.CreateConflict(req.SubjectType, req.SubjectID, req.TargetType, req.TargetID, req.Weight)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "this conflict already exists", http.StatusConflict)
		} else {
			JSONError(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"status": "created"}, http.StatusCreated)
}

func (api *API) DeleteConflict(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	err = tdb.DeleteConflict(id)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "conflict not found", http.StatusNotFound)
		} else {
			JSONError(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDebateConflicts returns hard/soft conflict descriptions for a saved debate.
func (api *API) GetDebateConflicts(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	debateID := r.PathValue("debate_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	hard, soft, err := tdb.GetDebateConflicts(debateID)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "debate not found", http.StatusNotFound)
		} else {
			JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if hard == nil {
		hard = []string{}
	}
	if soft == nil {
		soft = []string{}
	}

	JSONResponse(w, map[string][]string{"hard": hard, "soft": soft}, http.StatusOK)
}
