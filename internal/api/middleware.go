package api

import (
	"net/http"

	"yellow/internal/db"
	"yellow/internal/models"
)

// IsAdmin checks if the incoming request is authorized as organizer/admin.
// For Phase 1, we use a simple cookie check ("admin_session=authenticated")
// or an Authorization header check ("Bearer admin").
func (api *API) IsAdmin(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err == nil && cookie.Value == "authenticated" {
		return true
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "Bearer admin" {
		return true
	}

	return false
}

// ValidateToken looks up the token inside the tournament database to resolve ownership.
func (api *API) ValidateToken(tdb db.TournamentStore, token string) (*models.TokenOwner, error) {
	return tdb.ValidateToken(token)
}

// CanViewRound checks visibility permissions for a specific round based on whether the round is silent
// or if its draw/results are released yet.
func (api *API) CanViewRound(tdb db.TournamentStore, roundID string, isAdmin bool) (bool, error) {
	return tdb.CanViewRound(roundID, isAdmin)
}

// WriteBlocker blocks modifying requests (POST, PUT, DELETE) if the tournament is archived.
func (api *API) WriteBlocker(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug != "" {
			var isArchived bool
			err := api.GlobalDB.QueryRow("SELECT is_archived FROM tournaments WHERE slug = ?", slug).Scan(&isArchived)
			if err == nil && isArchived {
				if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
					JSONError(w, "this tournament is archived and read-only", http.StatusForbidden)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	}
}
