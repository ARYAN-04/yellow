package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"yellow/internal/models"

	"github.com/google/uuid"
)

func scanFeedbackQuestions(rows *sql.Rows) ([]models.FeedbackQuestion, error) {
	defer rows.Close()

	var list []models.FeedbackQuestion
	for rows.Next() {
		var q models.FeedbackQuestion
		var options sql.NullString
		if err := rows.Scan(&q.ID, &q.Seq, &q.Type, &q.Name, &options, &q.Required, &q.FromType, &q.ToType); err != nil {
			return nil, err
		}
		q.Options = []string{}
		if options.Valid && options.String != "" {
			if err := json.Unmarshal([]byte(options.String), &q.Options); err != nil {
				return nil, fmt.Errorf("invalid options_json for question %s: %w", q.ID, err)
			}
		}
		list = append(list, q)
	}
	return list, rows.Err()
}

// ListFeedbackQuestions returns feedback questions ordered by seq, optionally
// filtered by from_type/to_type (empty strings mean no filter).
func (s *SQLTournamentStore) ListFeedbackQuestions(fromType, toType string) ([]models.FeedbackQuestion, error) {
	query := "SELECT id, seq, type, name, options_json, COALESCE(required, 0), from_type, to_type FROM feedback_questions"
	var where []string
	var args []interface{}
	if fromType != "" {
		where = append(where, "from_type = ?")
		args = append(args, fromType)
	}
	if toType != "" {
		where = append(where, "to_type = ?")
		args = append(args, toType)
	}
	for i, w := range where {
		if i == 0 {
			query += " WHERE " + w
		} else {
			query += " AND " + w
		}
	}
	query += " ORDER BY seq ASC, rowid ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return scanFeedbackQuestions(rows)
}

func (s *SQLTournamentStore) CreateFeedbackQuestion(q models.FeedbackQuestion) (models.FeedbackQuestion, error) {
	var maxSeq sql.NullInt64
	if err := s.db.QueryRow("SELECT MAX(seq) FROM feedback_questions").Scan(&maxSeq); err != nil {
		return q, err
	}
	q.Seq = int(maxSeq.Int64) + 1
	optionsJSON, err := marshalQuestionOptions(q.Options)
	if err != nil {
		return q, err
	}
	q.ID = uuid.New().String()
	_, err = s.db.Exec(
		"INSERT INTO feedback_questions (id, seq, type, name, options_json, required, from_type, to_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		q.ID, q.Seq, q.Type, q.Name, optionsJSON, q.Required, q.FromType, q.ToType,
	)
	if err != nil {
		return q, err
	}
	return q, nil
}

func marshalQuestionOptions(options []string) (interface{}, error) {
	if len(options) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (s *SQLTournamentStore) UpdateFeedbackQuestion(q models.FeedbackQuestion) error {
	optionsJSON, err := marshalQuestionOptions(q.Options)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(
		"UPDATE feedback_questions SET type = ?, name = ?, options_json = ?, required = ?, from_type = ?, to_type = ? WHERE id = ?",
		q.Type, q.Name, optionsJSON, q.Required, q.FromType, q.ToType, q.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLTournamentStore) DeleteFeedbackQuestion(id string) error {
	res, err := s.db.Exec("DELETE FROM feedback_questions WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MoveFeedbackQuestion swaps a question's position with its neighbor in the seq order.
func (s *SQLTournamentStore) MoveFeedbackQuestion(id, direction string) error {
	questions, err := s.ListFeedbackQuestions("", "")
	if err != nil {
		return err
	}

	idx := -1
	for i, q := range questions {
		if q.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return sql.ErrNoRows
	}

	other := idx - 1
	if direction == "down" {
		other = idx + 1
	}
	if other < 0 || other >= len(questions) {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE feedback_questions SET seq = ? WHERE id = ?", questions[other].Seq, questions[idx].ID); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE feedback_questions SET seq = ? WHERE id = ?", questions[idx].Seq, questions[other].ID); err != nil {
		return err
	}
	return tx.Commit()
}

// ListFeedbackSubmissions returns feedback submissions with resolved source/target
// names and answers, optionally filtered to a round.
func (s *SQLTournamentStore) ListFeedbackSubmissions(roundID string) ([]models.FeedbackSubmission, error) {
	query := `
		SELECT fs.id, fs.round_id, fs.debate_id, fs.source_type, fs.source_id,
		       COALESCE(t.name, a2.name, '(deleted)'), fs.target_adjudicator_id,
		       COALESCE(a.name, '(deleted)'), fs.score, fs.created_at
		FROM feedback_submissions fs
		LEFT JOIN teams t ON fs.source_type = 'team' AND fs.source_id = t.id
		LEFT JOIN adjudicators a2 ON fs.source_type = 'adjudicator' AND fs.source_id = a2.id
		LEFT JOIN adjudicators a ON fs.target_adjudicator_id = a.id
	`
	var args []interface{}
	if roundID != "" {
		query += " WHERE fs.round_id = ?"
		args = append(args, roundID)
	}
	query += " ORDER BY fs.created_at DESC, fs.id ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.FeedbackSubmission
	for rows.Next() {
		var sub models.FeedbackSubmission
		var score sql.NullFloat64
		if err := rows.Scan(&sub.ID, &sub.RoundID, &sub.DebateID, &sub.SourceType, &sub.SourceID,
			&sub.SourceName, &sub.TargetID, &sub.TargetName, &score, &sub.CreatedAt); err != nil {
			return nil, err
		}
		if score.Valid {
			sub.Score = &score.Float64
		}
		sub.Answers = map[string]string{}
		list = append(list, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range list {
		aRows, err := s.db.Query("SELECT question_id, value FROM feedback_answers WHERE submission_id = ?", list[i].ID)
		if err != nil {
			return nil, err
		}
		for aRows.Next() {
			var qid, val string
			if err := aRows.Scan(&qid, &val); err != nil {
				aRows.Close()
				return nil, err
			}
			list[i].Answers[qid] = val
		}
		aRows.Close()
	}

	return list, nil
}

// GetFeedbackTargets returns the adjudicators the given token owner may evaluate:
// team tokens see non-trainee panel members of their debates; adjudicator tokens
// see their chair (as panelist) or panelists (as chair) per debate.
func (s *SQLTournamentStore) GetFeedbackTargets(ownerType, ownerID string) ([]models.FeedbackTarget, error) {
	var query string
	if ownerType == "team" {
		query = `
			SELECT d.id, d.venue, r.name, da.adjudicator_id, a.name, da.role
			FROM debates d
			JOIN rounds r ON d.round_id = r.id
			JOIN debate_teams dt ON dt.debate_id = d.id AND dt.team_id = ?
			JOIN debate_adjudicators da ON da.debate_id = d.id AND da.role != 'trainee'
			JOIN adjudicators a ON a.id = da.adjudicator_id
			ORDER BY r.seq DESC, a.name ASC
		`
	} else if ownerType == "adjudicator" {
		query = `
			SELECT d.id, d.venue, r.name, da.adjudicator_id, a.name, da.role
			FROM debates d
			JOIN rounds r ON d.round_id = r.id
			JOIN debate_adjudicators me ON me.debate_id = d.id AND me.adjudicator_id = ? AND me.role IN ('chair', 'panel')
			JOIN debate_adjudicators da ON da.debate_id = d.id AND da.role IN ('chair', 'panel')
				AND ((me.role = 'chair' AND da.role = 'panel') OR (me.role = 'panel' AND da.role = 'chair'))
			JOIN adjudicators a ON a.id = da.adjudicator_id
			ORDER BY r.seq DESC, a.name ASC
		`
	} else {
		return nil, fmt.Errorf("unsupported feedback source type: %s", ownerType)
	}

	rows, err := s.db.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.FeedbackTarget
	for rows.Next() {
		var tgt models.FeedbackTarget
		if err := rows.Scan(&tgt.DebateID, &tgt.Venue, &tgt.RoundName, &tgt.AdjudicatorID, &tgt.AdjudicatorName, &tgt.Role); err != nil {
			return nil, err
		}
		list = append(list, tgt)
	}
	return list, rows.Err()
}

// SubmitFeedback upserts a feedback submission keyed on
// (debate_id, source_type, source_id, target_adjudicator_id), replacing prior answers.
func (s *SQLTournamentStore) SubmitFeedback(debateID, sourceType, sourceID, targetAdjID string, score *float64, answers map[string]string) error {
	var roundID string
	err := s.db.QueryRow("SELECT round_id FROM debates WHERE id = ?", debateID).Scan(&roundID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	subID := uuid.New().String()
	err = tx.QueryRow(
		"SELECT id FROM feedback_submissions WHERE debate_id = ? AND source_type = ? AND source_id = ? AND target_adjudicator_id = ?",
		debateID, sourceType, sourceID, targetAdjID,
	).Scan(&subID)

	switch {
	case err == sql.ErrNoRows:
		if _, err := tx.Exec(
			"INSERT INTO feedback_submissions (id, round_id, debate_id, source_type, source_id, target_adjudicator_id, score) VALUES (?, ?, ?, ?, ?, ?, ?)",
			subID, roundID, debateID, sourceType, sourceID, targetAdjID, score,
		); err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		if _, err := tx.Exec("DELETE FROM feedback_answers WHERE submission_id = ?", subID); err != nil {
			return err
		}
		if _, err := tx.Exec(
			"UPDATE feedback_submissions SET score = ?, created_at = CURRENT_TIMESTAMP WHERE id = ?",
			score, subID,
		); err != nil {
			return err
		}
	}

	for qid, val := range answers {
		if _, err := tx.Exec(
			"INSERT INTO feedback_answers (submission_id, question_id, value) VALUES (?, ?, ?)",
			subID, qid, val,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RecalcAdjudicatorRatings recomputes each adjudicator's rolling dynamic rating as
// test_score*testWeight + avg(feedback scores)*feedbackWeight. Adjudicators without
// any feedback keep rating = test_score.
func (s *SQLTournamentStore) RecalcAdjudicatorRatings() error {
	testWeight, feedbackWeight := 0.5, 0.5
	if v, err := s.GetConfig("rating_test_weight"); err == nil {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil {
			testWeight = f
		}
	}
	if v, err := s.GetConfig("rating_feedback_weight"); err == nil {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil {
			feedbackWeight = f
		}
	}

	_, err := s.db.Exec(`
		UPDATE adjudicators SET rating =
			test_score * ? + (SELECT AVG(fs.score) FROM feedback_submissions fs WHERE fs.target_adjudicator_id = adjudicators.id) * ?
		WHERE (SELECT COUNT(*) FROM feedback_submissions fs WHERE fs.target_adjudicator_id = adjudicators.id) > 0
	`, testWeight, feedbackWeight)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE adjudicators SET rating = test_score
		WHERE (SELECT COUNT(*) FROM feedback_submissions fs WHERE fs.target_adjudicator_id = adjudicators.id) = 0
	`)
	return err
}
