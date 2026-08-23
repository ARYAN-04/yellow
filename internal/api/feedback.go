package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"yellow/internal/db"
	"yellow/internal/models"
)

// --- Admin Feedback Question Handlers ---

type FeedbackQuestionRequest struct {
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	Options  []string `json:"options"`
	Required bool     `json:"required"`
	FromType string   `json:"from_type"`
	ToType   string   `json:"to_type"`
}

var validQuestionTypes = map[string]bool{"scale": true, "text": true, "checkbox": true, "select": true}

func validateFeedbackQuestionRequest(req *FeedbackQuestionRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.FromType = strings.TrimSpace(req.FromType)
	req.ToType = strings.TrimSpace(req.ToType)
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !validQuestionTypes[req.Type] {
		return fmt.Errorf("type must be one of scale, text, checkbox, select")
	}
	if req.FromType != "team" && req.FromType != "adjudicator" {
		return fmt.Errorf("from_type must be 'team' or 'adjudicator'")
	}
	if req.ToType != "adjudicator" {
		return fmt.Errorf("to_type must be 'adjudicator'")
	}
	if req.Type == "select" && len(req.Options) < 2 {
		return fmt.Errorf("select questions require at least two options")
	}
	return nil
}

// ListFeedbackQuestions handles GET /api/t/{slug}/feedback/questions
func (api *API) ListFeedbackQuestions(w http.ResponseWriter, r *http.Request) {
	tdb, err := api.tournamentFromSlug(w, r)
	if err != nil {
		return
	}

	results, err := tdb.ListFeedbackQuestions(r.URL.Query().Get("from_type"), r.URL.Query().Get("to_type"))
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

// CreateFeedbackQuestion handles POST /api/t/{slug}/feedback/questions
func (api *API) CreateFeedbackQuestion(w http.ResponseWriter, r *http.Request) {
	tdb, err := api.tournamentFromSlug(w, r)
	if err != nil {
		return
	}

	var req FeedbackQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if verr := validateFeedbackQuestionRequest(&req); verr != nil {
		JSONError(w, verr.Error(), http.StatusBadRequest)
		return
	}

	q, err := tdb.CreateFeedbackQuestion(models.FeedbackQuestion{
		Type: req.Type, Name: req.Name, Options: req.Options,
		Required: req.Required, FromType: req.FromType, ToType: req.ToType,
	})
	if err != nil {
		JSONError(w, "insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, q, http.StatusCreated)
}

// UpdateFeedbackQuestion handles PUT /api/t/{slug}/feedback/questions/{id}
func (api *API) UpdateFeedbackQuestion(w http.ResponseWriter, r *http.Request) {
	tdb, err := api.tournamentFromSlug(w, r)
	if err != nil {
		return
	}

	var req FeedbackQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if verr := validateFeedbackQuestionRequest(&req); verr != nil {
		JSONError(w, verr.Error(), http.StatusBadRequest)
		return
	}

	err = tdb.UpdateFeedbackQuestion(models.FeedbackQuestion{
		ID: r.PathValue("id"), Type: req.Type, Name: req.Name, Options: req.Options,
		Required: req.Required, FromType: req.FromType, ToType: req.ToType,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "question not found", http.StatusNotFound)
		} else {
			JSONError(w, "update failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

// MoveFeedbackQuestion handles POST /api/t/{slug}/feedback/questions/{id}/move
func (api *API) MoveFeedbackQuestion(w http.ResponseWriter, r *http.Request) {
	tdb, err := api.tournamentFromSlug(w, r)
	if err != nil {
		return
	}

	var req struct {
		Direction string `json:"direction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Direction != "up" && req.Direction != "down") {
		JSONError(w, "direction must be 'up' or 'down'", http.StatusBadRequest)
		return
	}

	err = tdb.MoveFeedbackQuestion(r.PathValue("id"), req.Direction)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "question not found", http.StatusNotFound)
		} else {
			JSONError(w, "move failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	JSONResponse(w, map[string]string{"status": "success"}, http.StatusOK)
}

// DeleteFeedbackQuestion handles DELETE /api/t/{slug}/feedback/questions/{id}
func (api *API) DeleteFeedbackQuestion(w http.ResponseWriter, r *http.Request) {
	tdb, err := api.tournamentFromSlug(w, r)
	if err != nil {
		return
	}

	err = tdb.DeleteFeedbackQuestion(r.PathValue("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "question not found", http.StatusNotFound)
		} else {
			JSONError(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListFeedbackSubmissions handles GET /api/t/{slug}/feedback/submissions?round_id=
func (api *API) ListFeedbackSubmissions(w http.ResponseWriter, r *http.Request) {
	tdb, err := api.tournamentFromSlug(w, r)
	if err != nil {
		return
	}

	results, err := tdb.ListFeedbackSubmissions(strings.TrimSpace(r.URL.Query().Get("round_id")))
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, results, http.StatusOK)
}

// --- Token-scoped Feedback Handlers ---

type TokenFeedbackTargetsResponse struct {
	Questions []models.FeedbackQuestion `json:"questions"`
	Targets   []models.FeedbackTarget   `json:"targets"`
}

type SubmitTokenFeedbackRequest struct {
	DebateID          string            `json:"debate_id"`
	TargetAdjudicator string            `json:"target_adjudicator_id"`
	Answers           map[string]string `json:"answers"`
}

// scaleBounds derives the accepted numeric range for a scale question from its
// options: the first two numeric option entries are used as min/max, else 0..10.
func scaleBounds(options []string) (float64, float64) {
	var bounds []float64
	for _, o := range options {
		if f, err := strconv.ParseFloat(strings.TrimSpace(o), 64); err == nil {
			bounds = append(bounds, f)
			if len(bounds) == 2 {
				break
			}
		}
	}
	if len(bounds) == 2 {
		return bounds[0], bounds[1]
	}
	return 0, 10
}

// validateFeedbackAnswers checks answers against the applicable questionnaire and
// returns the derived score (average of scale answers), if any.
func validateFeedbackAnswers(questions []models.FeedbackQuestion, answers map[string]string) (*float64, error) {
	qByID := make(map[string]models.FeedbackQuestion, len(questions))
	for _, q := range questions {
		qByID[q.ID] = q
	}

	var sum float64
	var scaleCount int
	seen := make(map[string]bool, len(answers))
	for qid, val := range answers {
		q, ok := qByID[qid]
		if !ok {
			return nil, fmt.Errorf("unknown question id %s", qid)
		}
		seen[qid] = true
		val = strings.TrimSpace(val)

		switch q.Type {
		case "scale":
			min, max := scaleBounds(q.Options)
			n, err := strconv.ParseFloat(val, 64)
			if err != nil || n < min || n > max {
				return nil, fmt.Errorf("answer to '%s' must be a number between %s and %s",
					q.Name, strconv.FormatFloat(min, 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
			}
			sum += n
			scaleCount++
		case "select":
			found := false
			for _, o := range q.Options {
				if o == val {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("answer to '%s' is not one of the allowed options", q.Name)
			}
		case "checkbox":
			if val != "true" && val != "false" {
				return nil, fmt.Errorf("answer to '%s' must be a boolean", q.Name)
			}
		case "text":
			if val == "" {
				return nil, fmt.Errorf("answer to '%s' must not be empty", q.Name)
			}
		}
	}

	for _, q := range questions {
		if q.Required && !seen[q.ID] {
			return nil, fmt.Errorf("answer to '%s' is required", q.Name)
		}
	}

	if scaleCount == 0 {
		return nil, nil
	}
	score := sum / float64(scaleCount)
	return &score, nil
}

// GetTokenFeedbackTargets handles GET /api/token/{token}/feedback/targets
func (api *API) GetTokenFeedbackTargets(w http.ResponseWriter, r *http.Request) {
	info, err := api.resolveTokenHelper(r.PathValue("token"))
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	tdb, err := api.DBMgr.Get(info.Slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	targets, err := tdb.GetFeedbackTargets(info.Type, info.OwnerID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if targets == nil {
		targets = []models.FeedbackTarget{}
	}

	questions, err := tdb.ListFeedbackQuestions(info.Type, "adjudicator")
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if questions == nil {
		questions = []models.FeedbackQuestion{}
	}

	JSONResponse(w, TokenFeedbackTargetsResponse{Questions: questions, Targets: targets}, http.StatusOK)
}

// SubmitTokenFeedback handles POST /api/token/{token}/feedback
func (api *API) SubmitTokenFeedback(w http.ResponseWriter, r *http.Request) {
	info, err := api.resolveTokenHelper(r.PathValue("token"))
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	tdb, err := api.DBMgr.Get(info.Slug)
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	var req SubmitTokenFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	targets, err := tdb.GetFeedbackTargets(info.Type, info.OwnerID)
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	permitted := false
	for _, t := range targets {
		if t.DebateID == req.DebateID && t.AdjudicatorID == req.TargetAdjudicator {
			permitted = true
			break
		}
	}
	if !permitted {
		JSONError(w, "you are not permitted to evaluate this adjudicator in this debate", http.StatusForbidden)
		return
	}

	questions, err := tdb.ListFeedbackQuestions(info.Type, "adjudicator")
	if err != nil {
		JSONError(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	answers := make(map[string]string, len(req.Answers))
	for k, v := range req.Answers {
		answers[k] = v
	}
	score, verr := validateFeedbackAnswers(questions, answers)
	if verr != nil {
		JSONError(w, verr.Error(), http.StatusBadRequest)
		return
	}

	err = tdb.SubmitFeedback(req.DebateID, info.Type, info.OwnerID, req.TargetAdjudicator, score, answers)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "debate not found", http.StatusNotFound)
		} else {
			JSONError(w, "failed to submit feedback: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := tdb.RecalcAdjudicatorRatings(); err != nil {
		JSONError(w, "feedback saved but rating recalc failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSONResponse(w, map[string]string{"status": "submitted"}, http.StatusCreated)
}

// tournamentFromSlug resolves the tournament store for the request slug or writes an error.
func (api *API) tournamentFromSlug(w http.ResponseWriter, r *http.Request) (db.TournamentStore, error) {
	tdb, err := api.DBMgr.Get(r.PathValue("slug"))
	if err != nil {
		JSONError(w, err.Error(), http.StatusNotFound)
		return nil, err
	}
	return tdb, nil
}
