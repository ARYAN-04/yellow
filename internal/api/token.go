package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"yellow/internal/models"

	"github.com/google/uuid"
)

// resolveTokenHelper searches all databases in the tournaments directory to resolve the token.
func (api *API) resolveTokenHelper(token string) (*models.TokenInfo, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	rows, err := api.GlobalDB.Query("SELECT slug FROM tournaments")
	if err != nil {
		return nil, fmt.Errorf("failed to query tournaments: %w", err)
	}
	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err == nil {
			slugs = append(slugs, slug)
		}
	}
	rows.Close()

	for _, slug := range slugs {
		tdb, err := api.DBMgr.Get(slug)
		if err != nil {
			continue
		}

		info, err := tdb.ResolveToken(token)
		if err == nil {
			info.Slug = slug
			return info, nil
		}
	}

	return nil, fmt.Errorf("token not found in any active tournament")
}

// ResolveToken handles GET /api/token/{token}
func (api *API) ResolveToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	info, err := api.resolveTokenHelper(token)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	JSONResponse(w, info, http.StatusOK)
}

// GetTokenDebates handles GET /api/token/{token}/debates
func (api *API) GetTokenDebates(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	info, err := api.resolveTokenHelper(token)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	tdb, err := api.DBMgr.Get(info.Slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	debates, err := tdb.GetTokenDebates(info.OwnerID, info.Type)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, debates, http.StatusOK)
}

// SubmitTokenBallot handles POST /api/token/{token}/debates/{debate_id}/ballots
func (api *API) SubmitTokenBallot(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	debateID := r.PathValue("debate_id")
	info, err := api.resolveTokenHelper(token)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	if info.Type != "adjudicator" {
		JSONError(w, "only adjudicators can submit ballots", http.StatusForbidden)
		return
	}

	tdb, err := api.DBMgr.Get(info.Slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req SubmitBallotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	scoreMin, scoreMax, replyMin, replyMax := ballotScoreBounds(tdb)
	if err := validateBallotRequest(&req, scoreMin, scoreMax, replyMin, replyMax); err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ballotID := uuid.New().String()
	err = tdb.SubmitTokenBallot(debateID, ballotID, info.OwnerID, req.Results)
	if err != nil {
		if err == models.ErrNotAssigned {
			JSONError(w, err.Error(), http.StatusForbidden)
		} else if err == sql.ErrNoRows {
			JSONError(w, "debate not found", http.StatusNotFound)
		} else {
			JSONError(w, "failed to insert ballot: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"id": ballotID, "status": "submitted"}, http.StatusCreated)
}

// GetTokenCheckin handles GET /api/token/{token}/checkin
func (api *API) GetTokenCheckin(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	info, err := api.resolveTokenHelper(token)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	tdb, err := api.DBMgr.Get(info.Slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	checkins, err := tdb.ListCheckins()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	isCheckedIn := false
	for _, c := range checkins {
		if c.EntityType == info.Type && c.EntityID == info.OwnerID {
			isCheckedIn = c.CheckedIn
			break
		}
	}

	checkedAtStr := ""
	if isCheckedIn {
		checkedAtStr = "Checked In"
	}

	JSONResponse(w, map[string]interface{}{
		"checked_in":    isCheckedIn,
		"checked_in_at": checkedAtStr,
		"entity_type":   info.Type,
		"entity_name":   info.OwnerName,
	}, http.StatusOK)
}

// SetTokenCheckin handles POST /api/token/{token}/checkin
func (api *API) SetTokenCheckin(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	info, err := api.resolveTokenHelper(token)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	tdb, err := api.DBMgr.Get(info.Slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	checkins, err := tdb.ListCheckins()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	currentCheckedIn := false
	for _, c := range checkins {
		if c.EntityType == info.Type && c.EntityID == info.OwnerID {
			currentCheckedIn = c.CheckedIn
			break
		}
	}

	targetCheckedIn := !currentCheckedIn
	var req struct {
		CheckedIn *bool `json:"checked_in"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.CheckedIn != nil {
			targetCheckedIn = *req.CheckedIn
		}
	}

	if err := tdb.SetCheckedIn(info.Type, info.OwnerID, targetCheckedIn); err != nil {
		JSONError(w, "checkin update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	checkedAtStr := ""
	if targetCheckedIn {
		checkedAtStr = "Checked In"
	}

	JSONResponse(w, map[string]interface{}{
		"checked_in":    targetCheckedIn,
		"checked_in_at": checkedAtStr,
		"entity_type":   info.Type,
		"entity_name":   info.OwnerName,
	}, http.StatusOK)
}
