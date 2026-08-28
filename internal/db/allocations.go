package db

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrAssignmentNotFound = errors.New("assignment not found")
	ErrDebateNotFound     = errors.New("debate not found")
	ErrCrossRoundMove     = errors.New("cannot move an assignment across rounds")
	ErrAlreadyInDebate    = errors.New("entity is already assigned to the target debate")
)

func debateRoundID(tx *sql.Tx, debateID string) (string, error) {
	var roundID string
	err := tx.QueryRow("SELECT round_id FROM debates WHERE id = ?", debateID).Scan(&roundID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrDebateNotFound
	}
	return roundID, err
}

func checkSameRound(tx *sql.Tx, sourceDebateID, targetDebateID string) error {
	srcRound, err := debateRoundID(tx, sourceDebateID)
	if err != nil {
		return err
	}
	tgtRound, err := debateRoundID(tx, targetDebateID)
	if err != nil {
		return err
	}
	if srcRound != tgtRound {
		return ErrCrossRoundMove
	}
	return nil
}

// MoveSwapTeamAssignment moves a team assignment into the target debate by
// swapping with the assignment holding the same side there; if that side is
// free in the target debate it performs a plain move. Returns the round ID.
func (s *SQLTournamentStore) MoveSwapTeamAssignment(assignmentID, targetDebateID string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var srcDebateID, teamID, side string
	err = tx.QueryRow("SELECT debate_id, team_id, side FROM debate_teams WHERE id = ?", assignmentID).
		Scan(&srcDebateID, &teamID, &side)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrAssignmentNotFound
	}
	if err != nil {
		return "", err
	}

	if srcDebateID != targetDebateID {
		if err := checkSameRound(tx, srcDebateID, targetDebateID); err != nil {
			return "", err
		}
	}

	var dupes int
	if err := tx.QueryRow(
		"SELECT COUNT(*) FROM debate_teams WHERE debate_id = ? AND team_id = ? AND id != ?",
		targetDebateID, teamID, assignmentID,
	).Scan(&dupes); err != nil {
		return "", err
	}
	if dupes > 0 {
		return "", ErrAlreadyInDebate
	}

	if srcDebateID == targetDebateID {
		return debateRoundID(tx, targetDebateID)
	}

	var counterpartID string
	var counterpartTeam string
	err = tx.QueryRow(
		"SELECT id, team_id FROM debate_teams WHERE debate_id = ? AND side = ?",
		targetDebateID, side,
	).Scan(&counterpartID, &counterpartTeam)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		if _, err := tx.Exec("UPDATE debate_teams SET debate_id = ? WHERE id = ?", targetDebateID, assignmentID); err != nil {
			return "", err
		}
	} else {
		if _, err := tx.Exec("DELETE FROM debate_teams WHERE id IN (?, ?)", assignmentID, counterpartID); err != nil {
			return "", err
		}
		rows := []struct{ debateID, teamID string }{
			{srcDebateID, counterpartTeam},
			{targetDebateID, teamID},
		}
		for _, rw := range rows {
			if _, err := tx.Exec(
				"INSERT INTO debate_teams (id, debate_id, team_id, side, pull_up) VALUES (?, ?, ?, ?, 0)",
				uuid.New().String(), rw.debateID, rw.teamID, side,
			); err != nil {
				return "", err
			}
		}
	}

	roundID, err := debateRoundID(tx, targetDebateID)
	if err != nil {
		return "", err
	}
	return roundID, tx.Commit()
}

// MoveSwapAdjudicatorAssignment moves an adjudicator assignment into the target
// debate using swap-with-same-role semantics; if the target lacks a slot with
// that role it performs a plain move, applying roleOverride when provided.
func (s *SQLTournamentStore) MoveSwapAdjudicatorAssignment(assignmentID, targetDebateID, roleOverride string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var srcDebateID, adjID, role string
	err = tx.QueryRow("SELECT debate_id, adjudicator_id, role FROM debate_adjudicators WHERE id = ?", assignmentID).
		Scan(&srcDebateID, &adjID, &role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrAssignmentNotFound
	}
	if err != nil {
		return "", err
	}
	if roleOverride != "" {
		role = roleOverride
	}
	if targetDebateID == "" {
		targetDebateID = srcDebateID
	}

	if srcDebateID != targetDebateID {
		if err := checkSameRound(tx, srcDebateID, targetDebateID); err != nil {
			return "", err
		}
	}

	var dupes int
	if err := tx.QueryRow(
		"SELECT COUNT(*) FROM debate_adjudicators WHERE debate_id = ? AND adjudicator_id = ? AND id != ?",
		targetDebateID, adjID, assignmentID,
	).Scan(&dupes); err != nil {
		return "", err
	}
	if dupes > 0 {
		return "", ErrAlreadyInDebate
	}

	if srcDebateID == targetDebateID {
		if _, err := tx.Exec("UPDATE debate_adjudicators SET role = ? WHERE id = ?", role, assignmentID); err != nil {
			return "", err
		}
	} else if err := moveAdjWithinTx(tx, assignmentID, srcDebateID, adjID, targetDebateID, role); err != nil {
		return "", err
	}

	roundID, err := debateRoundID(tx, targetDebateID)
	if err != nil {
		return "", err
	}
	return roundID, tx.Commit()
}

// AddAdjudicatorToDebate adds an unallocated adjudicator into a debate with a specified role ('chair', 'panel', 'trainee').
// Returns the round ID.
func (s *SQLTournamentStore) AddAdjudicatorToDebate(debateID, adjudicatorID, role string) (string, error) {
	if role == "" {
		role = "panel"
	}
	if role != "chair" && role != "panel" && role != "trainee" {
		return "", errors.New("invalid adjudicator role: must be 'chair', 'panel', or 'trainee'")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	roundID, err := debateRoundID(tx, debateID)
	if err != nil {
		return "", err
	}

	// Check if the adjudicator is already assigned to any debate in this round
	var dupes int
	err = tx.QueryRow(`
		SELECT COUNT(*)
		FROM debate_adjudicators da
		JOIN debates d ON da.debate_id = d.id
		WHERE d.round_id = ? AND da.adjudicator_id = ?
	`, roundID, adjudicatorID).Scan(&dupes)
	if err != nil {
		return "", err
	}
	if dupes > 0 {
		return "", ErrAlreadyInDebate
	}

	assignmentID := uuid.New().String()
	_, err = tx.Exec(
		"INSERT INTO debate_adjudicators (id, debate_id, adjudicator_id, role) VALUES (?, ?, ?, ?)",
		assignmentID, debateID, adjudicatorID, role,
	)
	if err != nil {
		return "", err
	}

	return roundID, tx.Commit()
}

// RemoveAdjudicatorFromDebate removes an adjudicator from a debate, returning them to the unallocated scratch pool.
// Returns the round ID.
func (s *SQLTournamentStore) RemoveAdjudicatorFromDebate(assignmentID string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var debateID string
	err = tx.QueryRow("SELECT debate_id FROM debate_adjudicators WHERE id = ?", assignmentID).Scan(&debateID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrAssignmentNotFound
	}
	if err != nil {
		return "", err
	}

	roundID, err := debateRoundID(tx, debateID)
	if err != nil {
		return "", err
	}

	if _, err := tx.Exec("DELETE FROM debate_adjudicators WHERE id = ?", assignmentID); err != nil {
		return "", err
	}

	return roundID, tx.Commit()
}

func moveAdjWithinTx(tx *sql.Tx, assignmentID, srcDebateID, adjID, targetDebateID, role string) error {
	var counterpartID, counterpartAdj string
	err := tx.QueryRow(
		"SELECT id, adjudicator_id FROM debate_adjudicators WHERE debate_id = ? AND role = ?",
		targetDebateID, role,
	).Scan(&counterpartID, &counterpartAdj)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.Exec("UPDATE debate_adjudicators SET debate_id = ? WHERE id = ?", targetDebateID, assignmentID)
		return err
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM debate_adjudicators WHERE id IN (?, ?)", assignmentID, counterpartID); err != nil {
		return err
	}
	rows := []struct{ debateID, adjID string }{
		{srcDebateID, counterpartAdj},
		{targetDebateID, adjID},
	}
	for _, rw := range rows {
		if _, err := tx.Exec(
			"INSERT INTO debate_adjudicators (id, debate_id, adjudicator_id, role) VALUES (?, ?, ?, ?)",
			uuid.New().String(), rw.debateID, rw.adjID, role,
		); err != nil {
			return err
		}
	}
	return nil
}
