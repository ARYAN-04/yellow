package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"GoTabs/internal/models"

	"github.com/google/uuid"
)

// resolveTokenHelper searches all databases in the tournaments directory to resolve the token.
func (api *API) resolveTokenHelper(token string) (*models.TokenInfo, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	files, err := os.ReadDir("tournaments")
	if err != nil {
		return nil, fmt.Errorf("failed to read tournaments directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".db") || file.Name() == "global.db" {
			continue
		}

		slug := strings.TrimSuffix(file.Name(), ".db")
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
