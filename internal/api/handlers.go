package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"GoTabs/internal/db"
	"GoTabs/internal/models"
)

// API encapsulates the dependencies for HTTP JSON API endpoints.
type API struct {
	GlobalDB *sql.DB
	DBMgr    *db.ConnectionManager
}

// NewAPI returns an initialized API instance.
func NewAPI(globalDB *sql.DB, dbMgr *db.ConnectionManager) *API {
	return &API{
		GlobalDB: globalDB,
		DBMgr:    dbMgr,
	}
}

// JSONError writes a structured JSON error response to the client.
func JSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// JSONResponse writes a structured JSON success response to the client.
func JSONResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

// CreateTournamentRequest represents the expected payload to create a tournament.
type CreateTournamentRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9-_]+$`)

// CreateTournament registers a tournament in the global database and initializes its SQLite file.
func (api *API) CreateTournament(w http.ResponseWriter, r *http.Request) {
	var req CreateTournamentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))

	if req.Name == "" || req.Slug == "" {
		JSONError(w, "name and slug are required", http.StatusBadRequest)
		return
	}

	if !slugRegex.MatchString(req.Slug) {
		JSONError(w, "slug can only contain lowercase letters, numbers, hyphens, and underscores", http.StatusBadRequest)
		return
	}

	// Insert into global database
	id := req.Slug // Let's use the slug or uuid as ID
	dbPath := filepath.Join("tournaments", req.Slug+".db")

	_, err := api.GlobalDB.Exec(
		"INSERT INTO tournaments (id, name, slug, db_path) VALUES (?, ?, ?, ?)",
		id, req.Name, req.Slug, dbPath,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "a tournament with this slug already exists", http.StatusConflict)
		} else {
			JSONError(w, "failed to record tournament: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Initialize the tournament DB schema dynamically by triggering Get()
	_, err = api.DBMgr.Get(req.Slug)
	if err != nil {
		// Rollback global entry if initialization failed
		_, _ = api.GlobalDB.Exec("DELETE FROM tournaments WHERE slug = ?", req.Slug)
		JSONError(w, "failed to initialize tournament database schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]string{"id": id, "name": req.Name, "slug": req.Slug}, http.StatusCreated)
}

// ListTournaments lists all registered tournaments in the global database.
func (api *API) ListTournaments(w http.ResponseWriter, r *http.Request) {
	rows, err := api.GlobalDB.Query("SELECT id, name, slug, db_path, created_at FROM tournaments ORDER BY created_at DESC")
	if err != nil {
		JSONError(w, "failed to query tournaments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := make([]models.Tournament, 0)
	for rows.Next() {
		var t models.Tournament
		var createdAtStr string
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.DBPath, &createdAtStr); err != nil {
			JSONError(w, "failed to parse query results: "+err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, t)
	}

	JSONResponse(w, list, http.StatusOK)
}
