package api

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"yellow/internal/models"
)

// ListVenues returns all venues configured in the tournament.
func (api *API) ListVenues(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	venues, err := tdb.ListVenues()
	if err != nil {
		JSONError(w, "failed to list venues: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, venues, http.StatusOK)
}

// CreateVenue creates a new venue with name, priority, and accessibility flag.
func (api *API) CreateVenue(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var v models.Venue
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		JSONError(w, "venue name is required", http.StatusBadRequest)
		return
	}

	if err := tdb.CreateVenue(v); err != nil {
		JSONError(w, "failed to create venue: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, v, http.StatusCreated)
}

// UpdateVenue updates an existing venue's properties.
func (api *API) UpdateVenue(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var v models.Venue
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	v.ID = id
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		JSONError(w, "venue name is required", http.StatusBadRequest)
		return
	}

	if err := tdb.UpdateVenue(v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			JSONError(w, "venue not found", http.StatusNotFound)
			return
		}
		JSONError(w, "failed to update venue: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, v, http.StatusOK)
}

// DeleteVenue deletes a venue by ID.
func (api *API) DeleteVenue(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := tdb.DeleteVenue(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			JSONError(w, "venue not found", http.StatusNotFound)
			return
		}
		JSONError(w, "failed to delete venue: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]string{"status": "deleted"}, http.StatusOK)
}

// ImportVenues handles bulk CSV uploads to insert or update venues.
func (api *API) ImportVenues(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		JSONError(w, "file field is required in form-data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		JSONError(w, "failed to read CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) < 2 {
		JSONError(w, "CSV is empty or missing headers", http.StatusBadRequest)
		return
	}

	header := records[0]
	nameIdx, priorityIdx, catIdx, accessIdx := -1, -1, -1, -1

	for idx, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "name" || h == "venue" || h == "room" {
			nameIdx = idx
		} else if h == "priority" {
			priorityIdx = idx
		} else if h == "category" || h == "category_id" {
			catIdx = idx
		} else if h == "is_accessible" || h == "accessible" || h == "wheelchair" {
			accessIdx = idx
		}
	}

	if nameIdx == -1 {
		JSONError(w, "missing required header: 'name' must be present", http.StatusBadRequest)
		return
	}

	var venues []models.Venue
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= nameIdx {
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			continue
		}

		priority := 0
		if priorityIdx != -1 && len(row) > priorityIdx && row[priorityIdx] != "" {
			if p, err := strconv.Atoi(strings.TrimSpace(row[priorityIdx])); err == nil {
				priority = p
			}
		}

		var catID *string
		if catIdx != -1 && len(row) > catIdx && strings.TrimSpace(row[catIdx]) != "" {
			c := strings.TrimSpace(row[catIdx])
			catID = &c
		}

		isAccessible := false
		if accessIdx != -1 && len(row) > accessIdx && row[accessIdx] != "" {
			val := strings.ToLower(strings.TrimSpace(row[accessIdx]))
			isAccessible = val == "1" || val == "true" || val == "yes" || val == "y"
		}

		venues = append(venues, models.Venue{
			Name:         name,
			Priority:     priority,
			CategoryID:   catID,
			IsAccessible: isAccessible,
		})
	}

	inserted, err := tdb.ImportVenues(venues)
	if err != nil {
		JSONError(w, "failed to import venues: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]interface{}{"status": "success", "imported": inserted}, http.StatusOK)
}
