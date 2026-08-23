package api

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"yellow/internal/db"

	"github.com/google/uuid"
)

// UploadArchive handles POST /api/archive/upload
func (api *API) UploadArchive(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB memory limit)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		JSONError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.ToLower(strings.TrimSpace(r.FormValue("slug")))

	if name == "" || slug == "" {
		JSONError(w, "name and slug are required", http.StatusBadRequest)
		return
	}

	// Match slug format validation
	if !slugRegex.MatchString(slug) {
		JSONError(w, "slug can only contain lowercase letters, numbers, hyphens, and underscores", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		JSONError(w, "file field is required in form-data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Ensure tournaments directory exists
	if err := os.MkdirAll("tournaments", 0755); err != nil {
		JSONError(w, "failed to create tournaments directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Write upload to a temp file
	tempPath := filepath.Join("tournaments", fmt.Sprintf("%s_temp_%s.db", slug, uuid.New().String()))
	tempFile, err := os.Create(tempPath)
	if err != nil {
		JSONError(w, "failed to create temporary file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		JSONError(w, "failed to write uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tempFile.Close()

	// Validate SQLite database schema and extract stats
	valDB, err := sql.Open("sqlite", tempPath)
	if err != nil {
		JSONError(w, "uploaded file is not a valid SQLite database: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer valDB.Close()

	store := db.NewSQLTournamentStore(valDB)
	teamCount, roundCount, err := store.GetStats()
	if err != nil {
		JSONError(w, "database validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	valDB.Close()

	// Move to final path location
	destPath := filepath.Join("tournaments", slug+".db")
	_ = os.Remove(destPath)

	err = os.Rename(tempPath, destPath)
	if err != nil {
		// Copy fallback if rename fails
		err = copyFileHelper(tempPath, destPath)
		if err != nil {
			JSONError(w, "failed to copy database to destination: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Register archived entry in the global directory database
	id := slug
	_, err = api.GlobalDB.Exec(`
		INSERT INTO tournaments (id, name, slug, db_path, is_archived, team_count, round_count)
		VALUES (?, ?, ?, ?, 1, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			name = excluded.name,
			db_path = excluded.db_path,
			is_archived = 1,
			team_count = excluded.team_count,
			round_count = excluded.round_count
	`, id, name, slug, destPath, teamCount, roundCount)
	if err != nil {
		JSONError(w, "failed to record tournament in global registry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Evict connection cache in manager
	api.DBMgr.CloseAll()

	JSONResponse(w, map[string]interface{}{
		"status":      "success",
		"name":        name,
		"slug":        slug,
		"team_count":  teamCount,
		"round_count": roundCount,
	}, http.StatusOK)
}

func copyFileHelper(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
