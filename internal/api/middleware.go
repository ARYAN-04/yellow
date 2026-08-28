package api

import (
	"net/http"

	"yellow/internal/db"
	"yellow/internal/models"
)

// UserRole returns the authenticated role ("admin", "assistant", or "").
func (api *API) UserRole(r *http.Request) string {
	cookie, err := r.Cookie("admin_session")
	if err == nil {
		if cookie.Value == "authenticated" || cookie.Value == "admin" {
			return "admin"
		}
		if cookie.Value == "assistant" {
			return "assistant"
		}
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "Bearer admin" || authHeader == "Bearer authenticated" {
		return "admin"
	}
	if authHeader == "Bearer assistant" {
		return "assistant"
	}

	return ""
}

// IsAdmin checks if the incoming request is authorized (either full admin or assistant operator).
func (api *API) IsAdmin(r *http.Request) bool {
	return api.UserRole(r) != ""
}

// IsFullAdmin checks if the incoming request is authorized with full admin privileges.
func (api *API) IsFullAdmin(r *http.Request) bool {
	return api.UserRole(r) == "admin"
}

// RequireAdmin ensures the requester is an authorized organizer/admin or assistant before executing next.
// If the tournament is archived, read-only (GET) access is allowed for public records.
func (api *API) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if api.IsAdmin(r) {
			next(w, r)
			return
		}

		slug := r.PathValue("slug")
		if slug != "" && r.Method == http.MethodGet {
			var isArchived bool
			err := api.GlobalDB.QueryRow("SELECT is_archived FROM tournaments WHERE slug = ?", slug).Scan(&isArchived)
			if err == nil && isArchived {
				next(w, r)
				return
			}
		}

		JSONError(w, "unauthorized: admin session required", http.StatusUnauthorized)
	}
}

// RequireFullAdmin ensures the requester is an authorized full admin (not assistant).
// If unauthenticated, returns 401 Unauthorized. If authenticated as assistant, returns 403 Forbidden.
func (api *API) RequireFullAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := api.UserRole(r)
		if role == "admin" {
			next(w, r)
			return
		}
		if role == "assistant" {
			JSONError(w, "forbidden: action requires full admin privileges", http.StatusForbidden)
			return
		}

		slug := r.PathValue("slug")
		if slug != "" && r.Method == http.MethodGet {
			var isArchived bool
			err := api.GlobalDB.QueryRow("SELECT is_archived FROM tournaments WHERE slug = ?", slug).Scan(&isArchived)
			if err == nil && isArchived {
				next(w, r)
				return
			}
		}

		JSONError(w, "unauthorized: admin session required", http.StatusUnauthorized)
	}
}

// CheckAuth handles GET /api/auth/me to verify current admin session status and role.
func (api *API) CheckAuth(w http.ResponseWriter, r *http.Request) {
	role := api.UserRole(r)
	if role != "" {
		JSONResponse(w, map[string]interface{}{
			"authenticated": true,
			"user":          role,
			"role":          role,
		}, http.StatusOK)
		return
	}
	JSONError(w, "unauthorized", http.StatusUnauthorized)
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
		if slug == "" {
			token := r.PathValue("token")
			if token != "" {
				if info, err := api.resolveTokenHelper(token); err == nil && info != nil {
					slug = info.Slug
				}
			}
		}

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
