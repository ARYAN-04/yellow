package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"yellow/internal/db"
)

// MoveAssignmentRequest represents the payload for manual reallocation swaps.
type MoveAssignmentRequest struct {
	TargetDebateID string `json:"target_debate_id"`
	Role           string `json:"role,omitempty"`
}

// writeMoveError maps store-level move errors to HTTP responses; returns true
// when an error was written.
func writeMoveError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, db.ErrAssignmentNotFound):
		JSONError(w, "assignment not found", http.StatusNotFound)
	case errors.Is(err, db.ErrDebateNotFound):
		JSONError(w, "debate not found", http.StatusNotFound)
	case errors.Is(err, db.ErrCrossRoundMove):
		JSONError(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, db.ErrAlreadyInDebate):
		JSONError(w, err.Error(), http.StatusConflict)
	default:
		JSONError(w, "move failed: "+err.Error(), http.StatusInternalServerError)
	}
	return true
}

func decodeMoveRequest(w http.ResponseWriter, r *http.Request) (MoveAssignmentRequest, bool) {
	var req MoveAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return req, false
	}
	if req.TargetDebateID == "" {
		JSONError(w, "target_debate_id is required", http.StatusBadRequest)
		return req, false
	}
	return req, true
}

// MoveTeamAssignment moves a team assignment into the target debate by swapping
// with the assignment holding the same side there.
func (api *API) MoveTeamAssignment(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	req, ok := decodeMoveRequest(w, r)
	if !ok {
		return
	}

	roundID, err := tdb.MoveSwapTeamAssignment(r.PathValue("team_assignment_id"), req.TargetDebateID)
	if writeMoveError(w, err) {
		return
	}

	draw, err := tdb.GetRoundDraw(roundID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, draw, http.StatusOK)
}

// MoveAdjudicatorAssignment moves an adjudicator assignment into the target
// debate using swap-with-same-role semantics, optionally changing its role.
func (api *API) MoveAdjudicatorAssignment(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	req, ok := decodeMoveRequest(w, r)
	if !ok {
		return
	}

	roundID, err := tdb.MoveSwapAdjudicatorAssignment(r.PathValue("adj_assignment_id"), req.TargetDebateID, req.Role)
	if writeMoveError(w, err) {
		return
	}

	draw, err := tdb.GetRoundDraw(roundID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, draw, http.StatusOK)
}
