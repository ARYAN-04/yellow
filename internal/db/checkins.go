package db

import (
	"database/sql"
	"fmt"
	"time"

	"yellow/internal/models"

	"github.com/google/uuid"
)

// ensureCheckinRows lazily creates a checkins row with a random QR token for
// every team and adjudicator that does not have one yet.
func (s *SQLTournamentStore) ensureCheckinRows() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, entityType := range []string{"team", "adjudicator"} {
		table := entityType + "s"
		rows, err := tx.Query(fmt.Sprintf(
			"SELECT id FROM %s WHERE id NOT IN (SELECT entity_id FROM checkins WHERE entity_type = ?)", table,
		), entityType)
		if err != nil {
			return err
		}
		var missing []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			missing = append(missing, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		for _, id := range missing {
			if _, err := tx.Exec(
				"INSERT OR IGNORE INTO checkins (id, entity_type, entity_id, checkin_token) VALUES (?, ?, ?, ?)",
				uuid.New().String(), entityType, id, uuid.New().String(),
			); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *SQLTournamentStore) ListCheckins() ([]models.Checkin, error) {
	if err := s.ensureCheckinRows(); err != nil {
		return nil, fmt.Errorf("failed to create checkin rows: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT c.entity_type, c.entity_id, COALESCE(t.name, a.name, '(deleted)'), COALESCE(c.checked_in_at IS NOT NULL, 0), c.checkin_token
		FROM checkins c
		LEFT JOIN teams t ON c.entity_type = 'team' AND t.id = c.entity_id
		LEFT JOIN adjudicators a ON c.entity_type = 'adjudicator' AND a.id = c.entity_id
		ORDER BY c.entity_type ASC, 3 ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Checkin
	for rows.Next() {
		var c models.Checkin
		if err := rows.Scan(&c.EntityType, &c.EntityID, &c.EntityName, &c.CheckedIn, &c.CheckinToken); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *SQLTournamentStore) SetCheckedIn(entityType, entityID string, checkedIn bool) error {
	var checkedAt interface{}
	if checkedIn {
		checkedAt = time.Now()
	}
	res, err := s.db.Exec(
		"UPDATE checkins SET checked_in_at = ? WHERE entity_type = ? AND entity_id = ?",
		checkedAt, entityType, entityID,
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

func (s *SQLTournamentStore) ResolveCheckinToken(token string) (*models.CheckinTokenInfo, error) {
	var info models.CheckinTokenInfo
	var checkedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT c.entity_type, c.entity_id, c.checked_in_at
		FROM checkins c WHERE c.checkin_token = ?
	`, token).Scan(&info.EntityType, &info.EntityID, &checkedAt)
	if err != nil {
		return nil, err
	}
	info.CheckedIn = checkedAt.Valid

	nameQuery := "SELECT name FROM teams WHERE id = ?"
	if info.EntityType == "adjudicator" {
		nameQuery = "SELECT name FROM adjudicators WHERE id = ?"
	}
	if err := s.db.QueryRow(nameQuery, info.EntityID).Scan(&info.EntityName); err != nil {
		info.EntityName = "(deleted)"
	}
	return &info, nil
}

func (s *SQLTournamentStore) GetRoundAvailability(roundID string) ([]models.AvailabilityOverride, error) {
	rows, err := s.db.Query(
		"SELECT entity_type, entity_id, is_available FROM round_availability WHERE round_id = ?", roundID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.AvailabilityOverride
	for rows.Next() {
		var o models.AvailabilityOverride
		if err := rows.Scan(&o.EntityType, &o.EntityID, &o.IsAvailable); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (s *SQLTournamentStore) SetRoundAvailability(roundID, entityType, entityID string, isAvailable bool) error {
	_, err := s.db.Exec(`
		INSERT INTO round_availability (round_id, entity_type, entity_id, is_available)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(round_id, entity_type, entity_id) DO UPDATE SET is_available = excluded.is_available
	`, roundID, entityType, entityID, isAvailable)
	return err
}

// SyncAvailabilityFromCheckins replaces the round's availability overrides with
// the current check-in state of every team and adjudicator.
func (s *SQLTournamentStore) SyncAvailabilityFromCheckins(roundID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM round_availability WHERE round_id = ?", roundID); err != nil {
		return err
	}
	for _, entityType := range []string{"team", "adjudicator"} {
		if _, err := tx.Exec(`
			INSERT INTO round_availability (round_id, entity_type, entity_id, is_available)
			SELECT ?, ?, entity_id, COALESCE(checked_in_at IS NOT NULL, 0)
			FROM checkins WHERE entity_type = ?
		`, roundID, entityType, entityType); err != nil {
			return err
		}
	}
	return tx.Commit()
}
