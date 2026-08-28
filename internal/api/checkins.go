package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"yellow/internal/models"
)

// ListCheckins handles GET /api/t/{slug}/checkins
func (api *API) ListCheckins(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	checkins, err := tdb.ListCheckins()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, checkins, http.StatusOK)
}

// SetCheckin handles POST /api/t/{slug}/checkins/{entity_type}/{entity_id}
func (api *API) SetCheckin(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req struct {
		CheckedIn bool `json:"checked_in"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	entityType := r.PathValue("entity_type")
	entityID := r.PathValue("entity_id")
	switch entityType {
	case "team", "adjudicator", "venue":
	default:
		JSONError(w, "invalid entity type", http.StatusBadRequest)
		return
	}

	if err := tdb.SetCheckedIn(entityType, entityID, req.CheckedIn); err != nil {
		JSONError(w, "checkin update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]bool{"checked_in": req.CheckedIn}, http.StatusOK)
}

// resolveCheckinTokenHelper searches all tournament databases for a check-in token.
func (api *API) resolveCheckinTokenHelper(token string) (*models.CheckinTokenInfo, string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, "", os.ErrInvalid
	}

	rows, err := api.GlobalDB.Query("SELECT slug FROM tournaments")
	if err != nil {
		return nil, "", err
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
		info, err := tdb.ResolveCheckinToken(token)
		if err == nil {
			return info, slug, nil
		}
	}
	return nil, "", os.ErrNotExist
}

// ResolveCheckinToken handles GET /api/checkin/{token}
func (api *API) ResolveCheckinTokenHandler(w http.ResponseWriter, r *http.Request) {
	info, _, err := api.resolveCheckinTokenHelper(r.PathValue("token"))
	if err != nil {
		JSONError(w, "check-in link not recognized", http.StatusNotFound)
		return
	}
	JSONResponse(w, info, http.StatusOK)
}

// SelfCheckin handles POST /api/checkin/{token} — idempotent self check-in.
func (api *API) SelfCheckin(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	info, slug, err := api.resolveCheckinTokenHelper(token)
	if err != nil {
		JSONError(w, "check-in link not recognized", http.StatusNotFound)
		return
	}

	if !info.CheckedIn {
		tdb, err := api.DBMgr.Get(slug)
		if err != nil {
			JSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		if err := tdb.SetCheckedIn(info.EntityType, info.EntityID, true); err != nil {
			JSONError(w, "check-in failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		info.CheckedIn = true
	}

	JSONResponse(w, info, http.StatusOK)
}

// GetRoundAvailability handles GET /api/t/{slug}/rounds/{round_id}/availability
func (api *API) GetRoundAvailability(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	overrides, err := tdb.GetRoundAvailability(roundID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	overrideMap := make(map[string]models.AvailabilityOverride)
	for _, o := range overrides {
		overrideMap[o.EntityType+":"+o.EntityID] = o
	}

	checkins, err := tdb.ListCheckins()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	checkedMap := make(map[string]bool)
	for _, c := range checkins {
		checkedMap[c.EntityType+":"+c.EntityID] = c.CheckedIn
	}

	var entries []models.AvailabilityEntry
	teams, err := tdb.ListTeams()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, t := range teams {
		entries = append(entries, buildAvailabilityEntry("team", t.ID, t.Name, overrideMap, checkedMap))
	}
	adjudicators, err := tdb.ListAdjudicators()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, a := range adjudicators {
		entries = append(entries, buildAvailabilityEntry("adjudicator", a.ID, a.Name, overrideMap, checkedMap))
	}

	JSONResponse(w, entries, http.StatusOK)
}

func buildAvailabilityEntry(entityType, entityID, name string, overrideMap map[string]models.AvailabilityOverride, checkedMap map[string]bool) models.AvailabilityEntry {
	entry := models.AvailabilityEntry{
		EntityType: entityType,
		EntityID:   entityID,
		Name:       name,
		CheckedIn:  checkedMap[entityType+":"+entityID],
	}
	if o, ok := overrideMap[entityType+":"+entityID]; ok {
		v := o.IsAvailable
		entry.IsAvailable = &v
	}
	return entry
}

// PutRoundAvailability handles PUT /api/t/{slug}/rounds/{round_id}/availability
func (api *API) PutRoundAvailability(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var updates []models.AvailabilityOverride
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	for _, u := range updates {
		switch u.EntityType {
		case "team", "adjudicator", "venue":
		default:
			JSONError(w, "invalid entity type: "+u.EntityType, http.StatusBadRequest)
			return
		}
		if err := tdb.SetRoundAvailability(roundID, u.EntityType, u.EntityID, u.IsAvailable); err != nil {
			JSONError(w, "update failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

// SyncRoundAvailability handles POST /api/t/{slug}/rounds/{round_id}/availability/sync
func (api *API) SyncRoundAvailability(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	roundID := r.PathValue("round_id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := tdb.SyncAvailabilityFromCheckins(roundID); err != nil {
		JSONError(w, "sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}
