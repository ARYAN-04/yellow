package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"yellow/internal/models"
)

// ListRoundMotions returns the motions configured for a given round.
// If the requester is not an admin, only released motions are returned.
func (api *API) ListRoundMotions(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	motions, err := tdb.ListMotions(roundID)
	if err != nil {
		JSONError(w, "failed to list motions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	isAdmin := api.IsAdmin(r)
	if !isAdmin {
		// Filter to only released motions
		released := make([]models.Motion, 0, len(motions))
		for _, m := range motions {
			if m.ReleasedAt != nil && *m.ReleasedAt != "" {
				released = append(released, m)
			}
		}
		motions = released
	}

	JSONResponse(w, motions, http.StatusOK)
}

// CreateRoundMotion creates a new motion under the specified round.
func (api *API) CreateRoundMotion(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var m models.Motion
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	m.RoundID = roundID
	if strings.TrimSpace(m.Text) == "" {
		JSONError(w, "motion text is required", http.StatusBadRequest)
		return
	}

	if err := tdb.CreateMotion(m); err != nil {
		JSONError(w, "failed to create motion: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, m, http.StatusCreated)
}

// UpdateMotion updates an existing motion's text, reference, seq, or info slide.
func (api *API) UpdateMotion(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var m models.Motion
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	m.ID = id
	if strings.TrimSpace(m.Text) == "" {
		JSONError(w, "motion text is required", http.StatusBadRequest)
		return
	}

	if err := tdb.UpdateMotion(m); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			JSONError(w, "motion not found", http.StatusNotFound)
			return
		}
		JSONError(w, "failed to update motion: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, m, http.StatusOK)
}

// DeleteMotion deletes a motion by ID.
func (api *API) DeleteMotion(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := tdb.DeleteMotion(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			JSONError(w, "motion not found", http.StatusNotFound)
			return
		}
		JSONError(w, "failed to delete motion: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}

// ReleaseRoundMotions toggles the released status on all motions in a round.
func (api *API) ReleaseRoundMotions(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req struct {
		Release *bool `json:"release"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	shouldRelease := true
	if req.Release != nil {
		shouldRelease = *req.Release
	}

	if err := tdb.ReleaseMotions(roundID, shouldRelease); err != nil {
		JSONError(w, "failed to update motion release status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]interface{}{"status": "success", "released": shouldRelease}, http.StatusOK)
}

// GetDebateVetoes returns the recorded veto rankings for a debate.
func (api *API) GetDebateVetoes(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	debateID := r.PathValue("debate_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	vetoes, err := tdb.GetDebateVetoes(debateID)
	if err != nil {
		JSONError(w, "failed to get debate vetoes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, vetoes, http.StatusOK)
}

// RecordDebateVetoes records team preference rankings/vetoes for motions in a debate.
func (api *API) RecordDebateVetoes(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	debateID := r.PathValue("debate_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	type VetoItem struct {
		TeamID     string `json:"team_id"`
		MotionID   string `json:"motion_id"`
		Preference int    `json:"preference"`
	}

	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var items []VetoItem
	// Try parsing as array
	if err := json.Unmarshal(raw, &items); err != nil {
		// Try parsing as wrapper object { vetoes: [] }
		var wrapper struct {
			Vetoes []VetoItem `json:"vetoes"`
		}
		if err := json.Unmarshal(raw, &wrapper); err == nil && len(wrapper.Vetoes) > 0 {
			items = wrapper.Vetoes
		} else {
			// Try parsing as single object
			var single VetoItem
			if err := json.Unmarshal(raw, &single); err == nil && single.TeamID != "" && single.MotionID != "" {
				items = []VetoItem{single}
			}
		}
	}

	if len(items) == 0 {
		JSONError(w, "at least one veto entry is required", http.StatusBadRequest)
		return
	}

	for _, item := range items {
		if item.TeamID == "" || item.MotionID == "" {
			JSONError(w, "team_id and motion_id are required for each veto", http.StatusBadRequest)
			return
		}
		if err := tdb.RecordMotionVeto(debateID, item.TeamID, item.MotionID, item.Preference); err != nil {
			JSONError(w, "failed to record veto: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

// GetMotionStatistics returns side win rates and position distributions for all motions.
func (api *API) GetMotionStatistics(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	stats, err := tdb.GetMotionStatistics()
	if err != nil {
		JSONError(w, "failed to calculate motion statistics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, stats, http.StatusOK)
}
