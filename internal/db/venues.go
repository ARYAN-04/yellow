package db

import (
	"database/sql"

	"yellow/internal/models"

	"github.com/google/uuid"
)

// ListVenues returns all venues ordered by priority descending, then name ascending.
func (s *SQLTournamentStore) ListVenues() ([]models.Venue, error) {
	rows, err := s.db.Query(`
		SELECT id, name, priority, category_id, is_accessible
		FROM venues
		ORDER BY priority DESC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Venue
	for rows.Next() {
		var v models.Venue
		var catID sql.NullString
		if err := rows.Scan(&v.ID, &v.Name, &v.Priority, &catID, &v.IsAccessible); err != nil {
			return nil, err
		}
		if catID.Valid {
			v.CategoryID = &catID.String
		}
		list = append(list, v)
	}
	if list == nil {
		list = []models.Venue{}
	}
	return list, nil
}

// CreateVenue inserts a new venue.
func (s *SQLTournamentStore) CreateVenue(v models.Venue) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	var catVal interface{}
	if v.CategoryID != nil && *v.CategoryID != "" {
		catVal = *v.CategoryID
	}

	_, err := s.db.Exec(`
		INSERT INTO venues (id, name, priority, category_id, is_accessible)
		VALUES (?, ?, ?, ?, ?)
	`, v.ID, v.Name, v.Priority, catVal, v.IsAccessible)
	return err
}

// UpdateVenue updates an existing venue's properties.
func (s *SQLTournamentStore) UpdateVenue(v models.Venue) error {
	var catVal interface{}
	if v.CategoryID != nil && *v.CategoryID != "" {
		catVal = *v.CategoryID
	}

	res, err := s.db.Exec(`
		UPDATE venues
		SET name = ?, priority = ?, category_id = ?, is_accessible = ?
		WHERE id = ?
	`, v.Name, v.Priority, catVal, v.IsAccessible, v.ID)
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

// DeleteVenue removes a venue by ID.
func (s *SQLTournamentStore) DeleteVenue(id string) error {
	res, err := s.db.Exec("DELETE FROM venues WHERE id = ?", id)
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

// GetAvailableVenuesForRound returns venues available for the specified round (excluding any marked unavailable in round_availability).
func (s *SQLTournamentStore) GetAvailableVenuesForRound(roundID string) ([]models.Venue, error) {
	rows, err := s.db.Query(`
		SELECT v.id, v.name, v.priority, v.category_id, v.is_accessible
		FROM venues v
		WHERE NOT EXISTS (
			SELECT 1 FROM round_availability ra
			WHERE ra.round_id = ? AND ra.entity_type = 'venue' AND ra.entity_id = v.id AND ra.is_available = 0
		)
		ORDER BY v.priority DESC, v.name ASC
	`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Venue
	for rows.Next() {
		var v models.Venue
		var catID sql.NullString
		if err := rows.Scan(&v.ID, &v.Name, &v.Priority, &catID, &v.IsAccessible); err != nil {
			return nil, err
		}
		if catID.Valid {
			v.CategoryID = &catID.String
		}
		list = append(list, v)
	}
	if list == nil {
		list = []models.Venue{}
	}
	return list, nil
}

// ImportVenues inserts or updates venues from imported records.
func (s *SQLTournamentStore) ImportVenues(venues []models.Venue) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inserted := 0
	for _, v := range venues {
		id := v.ID
		if id == "" {
			id = uuid.New().String()
		}
		var catVal interface{}
		if v.CategoryID != nil && *v.CategoryID != "" {
			catVal = *v.CategoryID
		}
		_, err = tx.Exec(`
			INSERT INTO venues (id, name, priority, category_id, is_accessible)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				priority = excluded.priority,
				category_id = excluded.category_id,
				is_accessible = excluded.is_accessible
		`, id, v.Name, v.Priority, catVal, v.IsAccessible)
		if err != nil {
			return 0, err
		}
		inserted++
	}
	return inserted, tx.Commit()
}
