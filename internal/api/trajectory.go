package api

import (
	"database/sql"
	"net/http"
)

// GetTeamTrajectory handles GET /api/t/{slug}/teams/{id}/trajectory
func (api *API) GetTeamTrajectory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	teamID := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	trajectory, err := tdb.GetTeamTrajectory(teamID, api.IsAdmin(r))
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "team not found", http.StatusNotFound)
		} else {
			JSONError(w, "failed to get team trajectory: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, trajectory, http.StatusOK)
}

// GetSpeakerTrajectory handles GET /api/t/{slug}/speakers/{id}/trajectory
func (api *API) GetSpeakerTrajectory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	speakerID := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	trajectory, err := tdb.GetSpeakerTrajectory(speakerID, api.IsAdmin(r))
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "speaker not found", http.StatusNotFound)
		} else {
			JSONError(w, "failed to get speaker trajectory: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, trajectory, http.StatusOK)
}
