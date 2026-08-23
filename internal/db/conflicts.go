package db

import (
	"database/sql"
	"fmt"

	"yellow/internal/models"

	"github.com/google/uuid"
)

const deletedEntityName = "(deleted)"

// entityIndex holds resolved entity names plus speaker-to-team and
// team-to-institution mappings used to resolve conflicts.
type entityIndex struct {
	names     map[string]string
	speakerTo map[string]string
	instOf    map[string]string
}

func (s *SQLTournamentStore) loadEntityIndex() (*entityIndex, error) {
	idx := &entityIndex{
		names:     make(map[string]string),
		speakerTo: make(map[string]string),
		instOf:    make(map[string]string),
	}

	rows, err := s.db.Query("SELECT id, name FROM teams")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return nil, err
		}
		idx.names[id] = name
	}
	rows.Close()

	rows, err = s.db.Query("SELECT id, name FROM adjudicators")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return nil, err
		}
		idx.names[id] = name
	}
	rows.Close()

	rows, err = s.db.Query("SELECT id, name, team_id FROM speakers")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, name string
		var teamID sql.NullString
		if err := rows.Scan(&id, &name, &teamID); err != nil {
			rows.Close()
			return nil, err
		}
		idx.names[id] = name
		if teamID.Valid {
			idx.speakerTo[id] = teamID.String
		}
	}
	rows.Close()

	rows, err = s.db.Query("SELECT id, name FROM institutions")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return nil, err
		}
		idx.names[id] = name
	}
	rows.Close()

	rows, err = s.db.Query("SELECT id, institution_id FROM teams WHERE institution_id IS NOT NULL")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var teamID, instID string
		if err := rows.Scan(&teamID, &instID); err != nil {
			rows.Close()
			return nil, err
		}
		idx.instOf[teamID] = instID
	}
	rows.Close()

	return idx, nil
}

func (idx *entityIndex) resolveName(entityType, id string) string {
	if name, ok := idx.names[id]; ok && name != "" {
		return name
	}
	return deletedEntityName
}

func scanConflicts(rows *sql.Rows) ([]models.Conflict, error) {
	defer rows.Close()

	var list []models.Conflict
	for rows.Next() {
		var c models.Conflict
		if err := rows.Scan(&c.ID, &c.SubjectType, &c.SubjectID, &c.TargetType, &c.TargetID, &c.Weight); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *SQLTournamentStore) ListConflicts() ([]models.Conflict, error) {
	rows, err := s.db.Query("SELECT id, subject_type, subject_id, target_type, target_id, weight FROM conflicts ORDER BY rowid ASC")
	if err != nil {
		return nil, err
	}

	list, err := scanConflicts(rows)
	if err != nil {
		return nil, err
	}

	idx, err := s.loadEntityIndex()
	if err != nil {
		return nil, err
	}

	for i := range list {
		list[i].SubjectName = idx.resolveName(list[i].SubjectType, list[i].SubjectID)
		list[i].TargetName = idx.resolveName(list[i].TargetType, list[i].TargetID)
	}
	return list, nil
}

func (s *SQLTournamentStore) CreateConflict(subjectType, subjectID, targetType, targetID, weight string) error {
	id := uuid.New().String()
	_, err := s.db.Exec(
		"INSERT INTO conflicts (id, subject_type, subject_id, target_type, target_id, weight) VALUES (?, ?, ?, ?, ?, ?)",
		id, subjectType, subjectID, targetType, targetID, weight,
	)
	return err
}

func (s *SQLTournamentStore) DeleteConflict(id string) error {
	res, err := s.db.Exec("DELETE FROM conflicts WHERE id = ?", id)
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

// GetConflictsForDraw returns conflicts normalized for draw consumption:
// speaker targets are expanded to their owning team so the draw package
// only needs team-level checks.
func (s *SQLTournamentStore) GetConflictsForDraw() ([]models.Conflict, error) {
	rows, err := s.db.Query("SELECT id, subject_type, subject_id, target_type, target_id, weight FROM conflicts")
	if err != nil {
		return nil, err
	}

	list, err := scanConflicts(rows)
	if err != nil {
		return nil, err
	}

	idx, err := s.loadEntityIndex()
	if err != nil {
		return nil, err
	}

	var normalized []models.Conflict
	for _, c := range list {
		if c.TargetType == "speaker" {
			teamID, ok := idx.speakerTo[c.TargetID]
			if !ok {
				continue
			}
			c.TargetType = "team"
			c.TargetID = teamID
		}
		normalized = append(normalized, c)
	}
	return normalized, nil
}

type debateEntities struct {
	teams     []models.TeamAssignment
	teamInst  map[string]string
	adjs      []models.AdjudicatorAssignment
	speakerOf map[string]string
}

func (s *SQLTournamentStore) loadDebateEntities(debateID string) (*debateEntities, bool, error) {
	var exists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM debates WHERE id = ?", debateID).Scan(&exists)
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}

	ent := &debateEntities{
		teamInst:  make(map[string]string),
		speakerOf: make(map[string]string),
	}

	tRows, err := s.db.Query(`
		SELECT dt.team_id, t.name, COALESCE(t.institution_id, '')
		FROM debate_teams dt
		JOIN teams t ON dt.team_id = t.id
		WHERE dt.debate_id = ?
	`, debateID)
	if err != nil {
		return nil, true, err
	}
	for tRows.Next() {
		var ta models.TeamAssignment
		var instID string
		if err := tRows.Scan(&ta.TeamID, &ta.TeamName, &instID); err != nil {
			tRows.Close()
			return nil, true, err
		}
		ent.teams = append(ent.teams, ta)
		ent.teamInst[ta.TeamID] = instID
	}
	tRows.Close()

	aRows, err := s.db.Query(`
		SELECT da.adjudicator_id, a.name, da.role
		FROM debate_adjudicators da
		JOIN adjudicators a ON da.adjudicator_id = a.id
		WHERE da.debate_id = ?
	`, debateID)
	if err != nil {
		return nil, true, err
	}
	for aRows.Next() {
		var aa models.AdjudicatorAssignment
		if err := aRows.Scan(&aa.AdjudicatorID, &aa.AdjudicatorName, &aa.Role); err != nil {
			aRows.Close()
			return nil, true, err
		}
		ent.adjs = append(ent.adjs, aa)
	}
	aRows.Close()

	spRows, err := s.db.Query("SELECT id, team_id FROM speakers")
	if err != nil {
		return nil, true, err
	}
	for spRows.Next() {
		var spID, teamID string
		if err := spRows.Scan(&spID, &teamID); err != nil {
			spRows.Close()
			return nil, true, err
		}
		ent.speakerOf[spID] = teamID
	}
	spRows.Close()

	return ent, true, nil
}

func describeConflict(c models.Conflict, ent *debateEntities, idx *entityIndex, adjIdx map[string]int) (string, bool) {
	teamNameByID := make(map[string]string)
	for _, t := range ent.teams {
		teamNameByID[t.TeamID] = t.TeamName
	}

	targetTeamName := func() (string, bool) {
		switch c.TargetType {
		case "team":
			name, ok := teamNameByID[c.TargetID]
			return name, ok
		case "speaker":
			tid, ok := ent.speakerOf[c.TargetID]
			if !ok {
				return "", false
			}
			name, ok := teamNameByID[tid]
			return name, ok
		case "institution":
			for _, t := range ent.teams {
				if ent.teamInst[t.TeamID] == c.TargetID {
					return t.TeamName, true
				}
			}
			return "", false
		}
		return "", false
	}

	switch {
	case c.SubjectType == "adjudicator" && c.TargetType == "adjudicator":
		if _, ok := adjIdx[c.SubjectID]; !ok {
			return "", false
		}
		if _, ok := adjIdx[c.TargetID]; !ok {
			return "", false
		}
		return fmt.Sprintf("Adjudicator conflict: %s ↔ %s",
			idx.resolveName("adjudicator", c.SubjectID), idx.resolveName("adjudicator", c.TargetID)), true

	case c.SubjectType == "adjudicator":
		i, ok := adjIdx[c.SubjectID]
		if !ok {
			return "", false
		}
		tName, ok := targetTeamName()
		if !ok {
			return "", false
		}
		if c.TargetType == "institution" {
			return fmt.Sprintf("%s — institution clash with Team %s", ent.adjs[i].AdjudicatorName, tName), true
		}
		label := ""
		if c.TargetType == "speaker" {
			label = "Speaker "
		}
		return fmt.Sprintf("Personal conflict: %s ↔ %s%s", ent.adjs[i].AdjudicatorName, label,
			idx.resolveName(c.TargetType, c.TargetID)), true

	case c.SubjectType == "team" && c.TargetType == "team":
		sName, ok := teamNameByID[c.SubjectID]
		if !ok {
			return "", false
		}
		tName, ok := teamNameByID[c.TargetID]
		if !ok || c.TargetID == c.SubjectID {
			return "", false
		}
		return fmt.Sprintf("Team conflict: %s ↔ %s", sName, tName), true
	}
	return "", false
}

func findTeam(ent *debateEntities, teamID string) int {
	for i, t := range ent.teams {
		if t.TeamID == teamID {
			return i
		}
	}
	return 0
}

// GetDebateConflicts returns human-readable hard and soft conflict descriptions
// for the teams and adjudicators assigned to the given debate.
func (s *SQLTournamentStore) GetDebateConflicts(debateID string) ([]string, []string, error) {
	ent, ok, err := s.loadDebateEntities(debateID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, sql.ErrNoRows
	}

	rows, err := s.db.Query("SELECT id, subject_type, subject_id, target_type, target_id, weight FROM conflicts")
	if err != nil {
		return nil, nil, err
	}

	list, err := scanConflicts(rows)
	if err != nil {
		return nil, nil, err
	}

	idx, err := s.loadEntityIndex()
	if err != nil {
		return nil, nil, err
	}

	adjIdx := make(map[string]int)
	for i, a := range ent.adjs {
		adjIdx[a.AdjudicatorID] = i
	}

	var hard, soft []string
	for _, c := range list {
		desc, applies := describeConflict(c, ent, idx, adjIdx)
		if !applies {
			continue
		}
		if c.Weight == "hard" {
			hard = append(hard, desc)
		} else {
			soft = append(soft, desc)
		}
	}
	return hard, soft, nil
}
