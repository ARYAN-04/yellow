package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"yellow/internal/models"

	"github.com/google/uuid"
)

// --- Configuration Handlers ---

var configDefaultPrecedence = []string{"points", "speaker_points", "margin"}

var configValidPrecedence = map[string]bool{
	"points":         true,
	"speaker_points": true,
	"margin":         true,
}

type ConfigResponse struct {
	Sides             string          `json:"sides"`
	ScoreMin          float64         `json:"score_min"`
	ScoreMax          float64         `json:"score_max"`
	RankingPrecedence []string        `json:"ranking_precedence"`
	PublicFeatures    map[string]bool `json:"public_features"`
}

type ConfigUpdateRequest struct {
	Sides             *string          `json:"sides"`
	ScoreMin          *float64         `json:"score_min"`
	ScoreMax          *float64         `json:"score_max"`
	RankingPrecedence []string         `json:"ranking_precedence"`
	PublicFeatures    *map[string]bool `json:"public_features"`
}

func (api *API) GetConfig(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	cfg := ConfigResponse{
		Sides:             "",
		ScoreMin:          0,
		ScoreMax:          100,
		RankingPrecedence: configDefaultPrecedence,
		PublicFeatures:    map[string]bool{"results_public": true},
	}

	if v, err := tdb.GetConfig("sides"); err == nil {
		cfg.Sides = v
	}
	if v, err := tdb.GetConfig("score_min"); err == nil {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil {
			cfg.ScoreMin = f
		}
	}
	if v, err := tdb.GetConfig("score_max"); err == nil {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil {
			cfg.ScoreMax = f
		}
	}
	if v, err := tdb.GetConfig("ranking_precedence"); err == nil {
		var prec []string
		if uerr := json.Unmarshal([]byte(v), &prec); uerr == nil && len(prec) > 0 {
			cfg.RankingPrecedence = prec
		}
	}
	if v, err := tdb.GetConfig("public_features"); err == nil {
		var features map[string]bool
		if uerr := json.Unmarshal([]byte(v), &features); uerr == nil {
			cfg.PublicFeatures = features
		}
	}

	JSONResponse(w, cfg, http.StatusOK)
}

func (api *API) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Sides != nil {
		if err := tdb.SetConfig("sides", strings.TrimSpace(*req.Sides)); err != nil {
			JSONError(w, "failed to save sides: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if req.ScoreMin != nil {
		if err := tdb.SetConfig("score_min", strconv.FormatFloat(*req.ScoreMin, 'f', -1, 64)); err != nil {
			JSONError(w, "failed to save score_min: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if req.ScoreMax != nil {
		if err := tdb.SetConfig("score_max", strconv.FormatFloat(*req.ScoreMax, 'f', -1, 64)); err != nil {
			JSONError(w, "failed to save score_max: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if req.RankingPrecedence != nil {
		for _, k := range req.RankingPrecedence {
			if !configValidPrecedence[k] {
				JSONError(w, "unknown ranking rule: "+k, http.StatusBadRequest)
				return
			}
		}
		b, err := json.Marshal(req.RankingPrecedence)
		if err != nil {
			JSONError(w, "failed to encode ranking_precedence: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tdb.SetConfig("ranking_precedence", string(b)); err != nil {
			JSONError(w, "failed to save ranking_precedence: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if req.PublicFeatures != nil {
		b, err := json.Marshal(*req.PublicFeatures)
		if err != nil {
			JSONError(w, "failed to encode public_features: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tdb.SetConfig("public_features", string(b)); err != nil {
			JSONError(w, "failed to save public_features: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	api.GetConfig(w, r)
}

// --- Break Category Handlers ---

type BreakCategoryRequest struct {
	Name       string `json:"name"`
	Seq        int    `json:"seq"`
	Size       *int   `json:"size"`
	BasePoints *int   `json:"base_points"`
}

func (api *API) ListBreakCategories(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	results, err := tdb.ListBreakCategories()
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

func (api *API) CreateBreakCategory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req BreakCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		JSONError(w, "name is required", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	c := models.BreakCategory{ID: id, Name: req.Name, Seq: req.Seq, Size: req.Size, BasePoints: req.BasePoints}
	err = tdb.CreateBreakCategory(c)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "break category name already exists", http.StatusConflict)
		} else {
			JSONError(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, c, http.StatusCreated)
}

func (api *API) UpdateBreakCategory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req BreakCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		JSONError(w, "name is required", http.StatusBadRequest)
		return
	}

	c := models.BreakCategory{ID: id, Name: req.Name, Seq: req.Seq, Size: req.Size, BasePoints: req.BasePoints}
	err = tdb.UpdateBreakCategory(c)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "break category not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "break category name already exists", http.StatusConflict)
		} else {
			JSONError(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, c, http.StatusOK)
}

func (api *API) DeleteBreakCategory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	err = tdb.DeleteBreakCategory(id)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "break category not found", http.StatusNotFound)
		} else {
			JSONError(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Team Update Handler ---

type TeamUpdateRequest struct {
	Name          *string          `json:"name"`
	Code          *string          `json:"code"`
	InstitutionID *string          `json:"institution_id"`
	Speakers      []models.Speaker `json:"speakers"`
	IsNovice      *bool            `json:"is_novice"`
	IsEsl         *bool            `json:"is_esl"`
	IsEfl         *bool            `json:"is_efl"`
	IsStandby     *bool            `json:"is_standby"`
}

func (api *API) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id := r.PathValue("id")
	tdb, err := api.DBMgr.Get(slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req TeamUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = tdb.UpdateTeam(id, req.Name, req.Code, req.InstitutionID, req.IsNovice, req.IsEsl, req.IsEfl, req.IsStandby)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "team not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "UNIQUE") {
			JSONError(w, "team name already exists", http.StatusConflict)
		} else {
			JSONError(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	for _, sp := range req.Speakers {
		if err := tdb.UpsertSpeaker(id, sp); err != nil {
			JSONError(w, "speaker update failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}
