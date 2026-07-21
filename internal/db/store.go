package db

import (
	"database/sql"
	"strings"

	"GoTabs/internal/models"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// TournamentStore defines the unified storage interface for a single tournament.
type TournamentStore interface {
	// Institutions
	ListInstitutions() ([]models.Institution, error)
	CreateInstitution(id, name, code string) error
	UpdateInstitution(id, name, code string) error
	DeleteInstitution(id string) error

	// Teams & Speakers
	ListTeams() ([]models.Team, error)
	CreateTeam(teamID, name, code, instID string, speakers []models.SpeakerRequest, token string) error
	DeleteTeam(id string) error

	// Adjudicators
	ListAdjudicators() ([]models.Adjudicator, error)
	CreateAdjudicator(id, name, instID string, score float64, token string) error
	DeleteAdjudicator(id string) error

	// Rounds
	ListRounds() ([]models.Round, error)
	GetRound(roundID string) (*models.Round, error)
	CreateRound(id string, seq int, name, stage string) error
	UpdateRound(roundID string, silent, drawReleased, resultsReleased *bool) error

	// Draw
	GetRoundDraw(roundID string) ([]models.DebateDraw, error)
	GetSidesConfig() (string, error)
	GetTeamsForDraw() ([]models.TeamDrawInfo, error)
	GetSideHistory(seq int) (map[models.SideHistKey]int, error)
	GetConfirmedPoints() (map[string]int, error)
	GetAdjudicatorsForDraw() ([]models.AdjDrawInfo, error)
	SaveDraw(roundID string, debates []models.DebateDrawInput) error

	// Ballots & Standings
	SubmitBallot(debateID, ballotID, submitterType, submitterID, status string, results []models.TeamBallotResult) error
	ConfirmBallot(ballotID string) error
	GetStandings() ([]models.Standing, error)

	// Tokens & Public Portal
	ResolveToken(token string) (*models.TokenInfo, error)
	GetTokenDebates(ownerID, ownerType string) ([]models.DebateInfo, error)
	SubmitTokenBallot(debateID, ballotID, adjID string, results []models.TeamBallotResult) error

	// Middleware & Access checks
	ValidateToken(token string) (*models.TokenOwner, error)
	CanViewRound(roundID string, isAdmin bool) (bool, error)

	// Stats
	GetStats() (int, int, error) // teamCount, roundCount

	// CSV Imports
	ImportInstitutions(insts []models.Institution) (int, error)
	ImportTeams(teams []models.TeamImport) (int, error)
	ImportAdjudicators(adjs []models.AdjudicatorImport) (int, error)

	// Close database connection
	Close() error
}

// SQLTournamentStore implements TournamentStore using standard library database/sql queries.
type SQLTournamentStore struct {
	db *sql.DB
}

// NewSQLTournamentStore returns an initialized SQLTournamentStore wrapper.
func NewSQLTournamentStore(db *sql.DB) *SQLTournamentStore {
	return &SQLTournamentStore{db: db}
}

func (s *SQLTournamentStore) Close() error {
	return s.db.Close()
}

func (s *SQLTournamentStore) ListInstitutions() ([]models.Institution, error) {
	rows, err := s.db.Query("SELECT id, name, code FROM institutions ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Institution
	for rows.Next() {
		var inst models.Institution
		if err := rows.Scan(&inst.ID, &inst.Name, &inst.Code); err != nil {
			return nil, err
		}
		list = append(list, inst)
	}
	return list, nil
}

func (s *SQLTournamentStore) CreateInstitution(id, name, code string) error {
	_, err := s.db.Exec("INSERT INTO institutions (id, name, code) VALUES (?, ?, ?)", id, name, code)
	return err
}

func (s *SQLTournamentStore) UpdateInstitution(id, name, code string) error {
	res, err := s.db.Exec("UPDATE institutions SET name = ?, code = ? WHERE id = ?", name, code, id)
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

func (s *SQLTournamentStore) DeleteInstitution(id string) error {
	res, err := s.db.Exec("DELETE FROM institutions WHERE id = ?", id)
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

func (s *SQLTournamentStore) ListTeams() ([]models.Team, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name, t.code, t.institution_id, i.name, i.code 
		FROM teams t
		LEFT JOIN institutions i ON t.institution_id = i.id
		ORDER BY t.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Team
	teamMap := make(map[string]*models.Team)
	for rows.Next() {
		var t models.Team
		var instID, instName, instCode sql.NullString
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &instID, &instName, &instCode); err != nil {
			return nil, err
		}
		if instID.Valid {
			t.InstitutionID = &instID.String
			t.InstitutionName = &instName.String
			t.InstitutionCode = &instCode.String
		}
		t.Speakers = []models.Speaker{}
		list = append(list, t)
		teamMap[t.ID] = &list[len(list)-1]
	}

	spRows, err := s.db.Query("SELECT id, name, team_id FROM speakers")
	if err != nil {
		return nil, err
	}
	defer spRows.Close()

	for spRows.Next() {
		var spID, name, teamID string
		if err := spRows.Scan(&spID, &name, &teamID); err != nil {
			return nil, err
		}
		if t, ok := teamMap[teamID]; ok {
			t.Speakers = append(t.Speakers, models.Speaker{ID: spID, Name: name})
		}
	}
	return list, nil
}

func (s *SQLTournamentStore) CreateTeam(teamID, name, code, instID string, speakers []models.SpeakerRequest, token string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var instVal interface{}
	if instID != "" {
		instVal = instID
	} else {
		instVal = nil
	}

	_, err = tx.Exec("INSERT INTO teams (id, name, institution_id, code) VALUES (?, ?, ?, ?)", teamID, name, instVal, code)
	if err != nil {
		return err
	}

	for _, sp := range speakers {
		sName := strings.TrimSpace(sp.Name)
		if sName == "" {
			continue
		}
		spID := uuid.New().String()
		_, err = tx.Exec("INSERT INTO speakers (id, name, team_id) VALUES (?, ?, ?)", spID, sName, teamID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("INSERT INTO access_tokens (token, type, owner_id) VALUES (?, 'team', ?)", token, teamID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLTournamentStore) DeleteTeam(id string) error {
	res, err := s.db.Exec("DELETE FROM teams WHERE id = ?", id)
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

func (s *SQLTournamentStore) ListAdjudicators() ([]models.Adjudicator, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.name, a.institution_id, i.name, i.code, a.test_score 
		FROM adjudicators a
		LEFT JOIN institutions i ON a.institution_id = i.id
		ORDER BY a.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Adjudicator
	for rows.Next() {
		var a models.Adjudicator
		var instID, instName, instCode sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &instID, &instName, &instCode, &a.TestScore); err != nil {
			return nil, err
		}
		if instID.Valid {
			a.InstitutionID = &instID.String
			a.InstitutionName = &instName.String
			a.InstitutionCode = &instCode.String
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *SQLTournamentStore) CreateAdjudicator(id, name, instID string, score float64, token string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var instVal interface{}
	if instID != "" {
		instVal = instID
	} else {
		instVal = nil
	}

	_, err = tx.Exec("INSERT INTO adjudicators (id, name, institution_id, test_score) VALUES (?, ?, ?, ?)", id, name, instVal, score)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO access_tokens (token, type, owner_id) VALUES (?, 'adjudicator', ?)", token, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLTournamentStore) DeleteAdjudicator(id string) error {
	res, err := s.db.Exec("DELETE FROM adjudicators WHERE id = ?", id)
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

func (s *SQLTournamentStore) ListRounds() ([]models.Round, error) {
	rows, err := s.db.Query("SELECT id, seq, name, stage, silent, draw_released, results_released FROM rounds ORDER BY seq ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Round
	for rows.Next() {
		var r models.Round
		if err := rows.Scan(&r.ID, &r.Seq, &r.Name, &r.Stage, &r.Silent, &r.DrawReleased, &r.ResultsReleased); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *SQLTournamentStore) CreateRound(id string, seq int, name, stage string) error {
	_, err := s.db.Exec("INSERT INTO rounds (id, seq, name, stage, silent, draw_released, results_released) VALUES (?, ?, ?, ?, 0, 0, 0)", id, seq, name, stage)
	return err
}

func (s *SQLTournamentStore) UpdateRound(roundID string, silent, drawReleased, resultsReleased *bool) error {
	query := "UPDATE rounds SET "
	var params []interface{}
	var updates []string

	if silent != nil {
		updates = append(updates, "silent = ?")
		params = append(params, *silent)
	}
	if drawReleased != nil {
		updates = append(updates, "draw_released = ?")
		params = append(params, *drawReleased)
	}
	if resultsReleased != nil {
		updates = append(updates, "results_released = ?")
		params = append(params, *resultsReleased)
	}

	if len(updates) == 0 {
		return nil
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	params = append(params, roundID)

	_, err := s.db.Exec(query, params...)
	return err
}

func (s *SQLTournamentStore) GetRoundDraw(roundID string) ([]models.DebateDraw, error) {
	rows, err := s.db.Query("SELECT id, venue FROM debates WHERE round_id = ?", roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debates []models.DebateDraw
	debateMap := make(map[string]*models.DebateDraw)

	for rows.Next() {
		var d models.DebateDraw
		if err := rows.Scan(&d.ID, &d.Venue); err != nil {
			return nil, err
		}
		d.Teams = []models.TeamAssignment{}
		d.Adjudicators = []models.AdjudicatorAssignment{}
		debates = append(debates, d)
		debateMap[d.ID] = &debates[len(debates)-1]
	}

	if len(debates) == 0 {
		return debates, nil
	}

	tRows, err := s.db.Query(`
		SELECT dt.debate_id, dt.team_id, t.name, dt.side 
		FROM debate_teams dt
		JOIN teams t ON dt.team_id = t.id
	`)
	if err == nil {
		for tRows.Next() {
			var dID, tID, tName, side string
			if err := tRows.Scan(&dID, &tID, &tName, &side); err == nil {
				if d, ok := debateMap[dID]; ok {
					d.Teams = append(d.Teams, models.TeamAssignment{TeamID: tID, TeamName: tName, Side: side})
				}
			}
		}
		tRows.Close()
	}

	aRows, err := s.db.Query(`
		SELECT da.debate_id, da.adjudicator_id, a.name, da.role 
		FROM debate_adjudicators da
		JOIN adjudicators a ON da.adjudicator_id = a.id
	`)
	if err == nil {
		for aRows.Next() {
			var dID, aID, aName, role string
			if err := aRows.Scan(&dID, &aID, &aName, &role); err == nil {
				if d, ok := debateMap[dID]; ok {
					d.Adjudicators = append(d.Adjudicators, models.AdjudicatorAssignment{AdjudicatorID: aID, AdjudicatorName: aName, Role: role})
				}
			}
		}
		aRows.Close()
	}

	return debates, nil
}

func (s *SQLTournamentStore) GetSidesConfig() (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = 'sides'").Scan(&val)
	return val, err
}

func (s *SQLTournamentStore) GetTeamsForDraw() ([]models.TeamDrawInfo, error) {
	rows, err := s.db.Query("SELECT id, institution_id FROM teams")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.TeamDrawInfo
	for rows.Next() {
		var t models.TeamDrawInfo
		var instID sql.NullString
		if err := rows.Scan(&t.ID, &instID); err != nil {
			return nil, err
		}
		if instID.Valid {
			t.InstitutionID = instID.String
		}
		list = append(list, t)
	}
	return list, nil
}

func (s *SQLTournamentStore) GetSideHistory(seq int) (map[models.SideHistKey]int, error) {
	rows, err := s.db.Query(`
		SELECT dt.team_id, dt.side, COUNT(*)
		FROM debate_teams dt
		JOIN debates d ON dt.debate_id = d.id
		JOIN rounds r ON d.round_id = r.id
		WHERE r.seq < ?
		GROUP BY dt.team_id, dt.side
	`, seq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hist := make(map[models.SideHistKey]int)
	for rows.Next() {
		var tid, side string
		var count int
		if err := rows.Scan(&tid, &side, &count); err != nil {
			return nil, err
		}
		hist[models.SideHistKey{TeamID: tid, Side: side}] = count
	}
	return hist, nil
}

func (s *SQLTournamentStore) GetConfirmedPoints() (map[string]int, error) {
	rows, err := s.db.Query(`
		SELECT br.team_id, SUM(br.points)
		FROM ballot_results br
		JOIN ballots b ON br.ballot_id = b.id
		WHERE b.status = 'confirmed'
		GROUP BY br.team_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pts := make(map[string]int)
	for rows.Next() {
		var tid string
		var count int
		if err := rows.Scan(&tid, &count); err != nil {
			return nil, err
		}
		pts[tid] = count
	}
	return pts, nil
}

func (s *SQLTournamentStore) GetAdjudicatorsForDraw() ([]models.AdjDrawInfo, error) {
	rows, err := s.db.Query("SELECT id, institution_id, test_score FROM adjudicators")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.AdjDrawInfo
	for rows.Next() {
		var a models.AdjDrawInfo
		var instID sql.NullString
		if err := rows.Scan(&a.ID, &instID, &a.Score); err != nil {
			return nil, err
		}
		if instID.Valid {
			a.InstitutionID = instID.String
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *SQLTournamentStore) SaveDraw(roundID string, debates []models.DebateDrawInput) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existingDebateIDs []string
	rows, err := tx.Query("SELECT id FROM debates WHERE round_id = ?", roundID)
	if err == nil {
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil {
				existingDebateIDs = append(existingDebateIDs, id)
			}
		}
		rows.Close()
	}

	for _, dID := range existingDebateIDs {
		_, _ = tx.Exec("DELETE FROM debate_teams WHERE debate_id = ?", dID)
		_, _ = tx.Exec("DELETE FROM debate_adjudicators WHERE debate_id = ?", dID)
	}
	_, _ = tx.Exec("DELETE FROM debates WHERE round_id = ?", roundID)

	for _, d := range debates {
		_, err = tx.Exec("INSERT INTO debates (id, round_id, venue) VALUES (?, ?, ?)", d.DebateID, roundID, d.Venue)
		if err != nil {
			return err
		}

		for _, t := range d.Teams {
			dtID := uuid.New().String()
			_, err = tx.Exec(
				"INSERT INTO debate_teams (id, debate_id, team_id, side) VALUES (?, ?, ?, ?)",
				dtID, d.DebateID, t.TeamID, t.Side,
			)
			if err != nil {
				return err
			}
		}

		for _, a := range d.Adjudicators {
			daID := uuid.New().String()
			_, err = tx.Exec(
				"INSERT INTO debate_adjudicators (id, debate_id, adjudicator_id, role) VALUES (?, ?, ?, ?)",
				daID, d.DebateID, a.AdjudicatorID, a.Role,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *SQLTournamentStore) SubmitBallot(debateID, ballotID, submitterType, submitterID, status string, results []models.TeamBallotResult) error {
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

	var subIDVal interface{}
	if submitterID != "" {
		subIDVal = submitterID
	} else {
		subIDVal = nil
	}

	_, err = tx.Exec(
		"INSERT INTO ballots (id, debate_id, round_id, submitter_type, submitter_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		ballotID, debateID, roundID, submitterType, subIDVal, status,
	)
	if err != nil {
		return err
	}

	for _, res := range results {
		brID := uuid.New().String()
		_, err = tx.Exec(
			"INSERT INTO ballot_results (id, ballot_id, team_id, points, speaker_points) VALUES (?, ?, ?, ?, ?)",
			brID, ballotID, res.TeamID, res.Points, res.SpeakerPoints,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLTournamentStore) ConfirmBallot(ballotID string) error {
	res, err := s.db.Exec("UPDATE ballots SET status = 'confirmed' WHERE id = ?", ballotID)
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

func (s *SQLTournamentStore) GetStandings() ([]models.Standing, error) {
	tRows, err := s.db.Query(`
		SELECT t.id, t.name, t.code, i.code 
		FROM teams t
		LEFT JOIN institutions i ON t.institution_id = i.id
	`)
	if err != nil {
		return nil, err
	}
	defer tRows.Close()

	var list []models.Standing
	teamMap := make(map[string]*models.Standing)

	for tRows.Next() {
		var ts models.Standing
		var instCode sql.NullString
		if err := tRows.Scan(&ts.TeamID, &ts.TeamName, &ts.TeamCode, &instCode); err != nil {
			return nil, err
		}
		if instCode.Valid {
			ts.InstitutionCode = instCode.String
		}
		list = append(list, ts)
		teamMap[ts.TeamID] = &list[len(list)-1]
	}

	pRows, err := s.db.Query(`
		SELECT br.team_id, br.points, br.speaker_points
		FROM ballot_results br
		JOIN ballots b ON br.ballot_id = b.id
		WHERE b.status = 'confirmed'
	`)
	if err == nil {
		for pRows.Next() {
			var tid string
			var pts int
			var spts float64
			if err := pRows.Scan(&tid, &pts, &spts); err == nil {
				if ts, ok := teamMap[tid]; ok {
					ts.Points += pts
					ts.SpeakerPoints += spts
				}
			}
		}
		pRows.Close()
	}

	return list, nil
}

func (s *SQLTournamentStore) ResolveToken(token string) (*models.TokenInfo, error) {
	var tType, ownerID string
	err := s.db.QueryRow("SELECT type, owner_id FROM access_tokens WHERE token = ?", token).Scan(&tType, &ownerID)
	if err != nil {
		return nil, err
	}

	var name string
	if tType == "team" {
		_ = s.db.QueryRow("SELECT name FROM teams WHERE id = ?", ownerID).Scan(&name)
	} else if tType == "adjudicator" {
		_ = s.db.QueryRow("SELECT name FROM adjudicators WHERE id = ?", ownerID).Scan(&name)
	}

	return &models.TokenInfo{
		Type:      tType,
		OwnerID:   ownerID,
		OwnerName: name,
	}, nil
}

func (s *SQLTournamentStore) GetTokenDebates(ownerID, ownerType string) ([]models.DebateInfo, error) {
	var debates []models.DebateInfo

	if ownerType == "adjudicator" {
		rows, err := s.db.Query(`
			SELECT d.id, d.venue, r.name, r.seq, da.role, COALESCE(b.status, 'unsubmitted')
			FROM debates d
			JOIN debate_adjudicators da ON d.id = da.debate_id
			JOIN rounds r ON d.round_id = r.id
			LEFT JOIN ballots b ON d.id = b.debate_id
			WHERE da.adjudicator_id = ?
			ORDER BY r.seq DESC
		`, ownerID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var d models.DebateInfo
			if err := rows.Scan(&d.ID, &d.Venue, &d.RoundName, &d.RoundSeq, &d.Role, &d.BallotStatus); err != nil {
				return nil, err
			}
			d.Teams = []models.TeamAssignment{}
			debates = append(debates, d)
		}
	} else if ownerType == "team" {
		rows, err := s.db.Query(`
			SELECT d.id, d.venue, r.name, r.seq, dt.side, r.results_released, COALESCE(br.points, -1), COALESCE(br.speaker_points, -1.0)
			FROM debates d
			JOIN debate_teams dt ON d.id = dt.debate_id
			JOIN rounds r ON d.round_id = r.id
			LEFT JOIN ballots b ON d.id = b.debate_id AND b.status = 'confirmed'
			LEFT JOIN ballot_results br ON b.id = br.ballot_id AND br.team_id = dt.team_id
			WHERE dt.team_id = ? AND r.draw_released = 1 AND (r.silent = 0 OR r.results_released = 1)
			ORDER BY r.seq DESC
		`, ownerID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var d models.DebateInfo
			var resultsReleased bool
			var pts int
			var spPts float64
			if err := rows.Scan(&d.ID, &d.Venue, &d.RoundName, &d.RoundSeq, &d.Side, &resultsReleased, &pts, &spPts); err != nil {
				return nil, err
			}
			if resultsReleased && pts != -1 {
				d.Points = &pts
				d.SpeakerPoints = &spPts
			}
			d.Teams = []models.TeamAssignment{}
			debates = append(debates, d)
		}
	}

	for idx := range debates {
		tRows, err := s.db.Query(`
			SELECT dt.team_id, t.name, dt.side
			FROM debate_teams dt
			JOIN teams t ON dt.team_id = t.id
			WHERE dt.debate_id = ?
		`, debates[idx].ID)
		if err == nil {
			for tRows.Next() {
				var ta models.TeamAssignment
				if err := tRows.Scan(&ta.TeamID, &ta.TeamName, &ta.Side); err == nil {
					debates[idx].Teams = append(debates[idx].Teams, ta)
				}
			}
			tRows.Close()
		}
	}

	return debates, nil
}

func (s *SQLTournamentStore) SubmitTokenBallot(debateID, ballotID, adjID string, results []models.TeamBallotResult) error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM debate_adjudicators WHERE debate_id = ? AND adjudicator_id = ?", debateID, adjID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return models.ErrNotAssigned
	}

	return s.SubmitBallot(debateID, ballotID, "adjudicator", adjID, "submitted", results)
}

func (s *SQLTournamentStore) ValidateToken(token string) (*models.TokenOwner, error) {
	var owner models.TokenOwner
	err := s.db.QueryRow("SELECT type, owner_id FROM access_tokens WHERE token = ?", token).Scan(&owner.Type, &owner.OwnerID)
	if err != nil {
		return nil, err
	}
	return &owner, nil
}

func (s *SQLTournamentStore) CanViewRound(roundID string, isAdmin bool) (bool, error) {
	if isAdmin {
		return true, nil
	}

	var silent, drawReleased, resultsReleased bool
	err := s.db.QueryRow("SELECT silent, draw_released, results_released FROM rounds WHERE id = ?", roundID).Scan(&silent, &drawReleased, &resultsReleased)
	if err != nil {
		return false, err
	}

	if silent && !resultsReleased {
		return false, nil
	}

	return drawReleased, nil
}

func (s *SQLTournamentStore) GetStats() (int, int, error) {
	var teamCount, roundCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM teams").Scan(&teamCount)
	if err != nil {
		return 0, 0, err
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM rounds").Scan(&roundCount)
	if err != nil {
		return 0, 0, err
	}
	return teamCount, roundCount, nil
}

func (s *SQLTournamentStore) ImportInstitutions(insts []models.Institution) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inserted := 0
	for _, inst := range insts {
		id := uuid.New().String()
		_, err = tx.Exec("INSERT INTO institutions (id, name, code) VALUES (?, ?, ?) ON CONFLICT(code) DO NOTHING", id, inst.Name, inst.Code)
		if err != nil {
			return 0, err
		}
		inserted++
	}
	return inserted, tx.Commit()
}

func (s *SQLTournamentStore) ImportTeams(teams []models.TeamImport) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inserted := 0
	for _, team := range teams {
		var instID interface{}
		if team.InstitutionCode != "" {
			code := strings.ToUpper(strings.TrimSpace(team.InstitutionCode))
			var id string
			err := tx.QueryRow("SELECT id FROM institutions WHERE code = ?", code).Scan(&id)
			if err == nil {
				instID = id
			} else if err == sql.ErrNoRows {
				id = uuid.New().String()
				name := code + " Institution"
				_, err = tx.Exec("INSERT INTO institutions (id, name, code) VALUES (?, ?, ?)", id, name, code)
				if err != nil {
					return 0, err
				}
				instID = id
			} else {
				return 0, err
			}
		} else {
			instID = nil
		}

		teamID := uuid.New().String()
		_, err = tx.Exec("INSERT INTO teams (id, name, institution_id, code) VALUES (?, ?, ?, ?) ON CONFLICT(name) DO NOTHING", teamID, team.Name, instID, team.Code)
		if err != nil {
			return 0, err
		}

		for _, spName := range team.Speakers {
			spID := uuid.New().String()
			_, err = tx.Exec("INSERT INTO speakers (id, name, team_id) VALUES (?, ?, ?)", spID, spName, teamID)
			if err != nil {
				return 0, err
			}
		}

		token := uuid.New().String()
		_, err = tx.Exec("INSERT INTO access_tokens (token, type, owner_id) VALUES (?, 'team', ?)", token, teamID)
		if err != nil {
			return 0, err
		}

		inserted++
	}
	return inserted, tx.Commit()
}

func (s *SQLTournamentStore) ImportAdjudicators(adjs []models.AdjudicatorImport) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inserted := 0
	for _, adj := range adjs {
		var instID interface{}
		if adj.InstitutionCode != "" {
			code := strings.ToUpper(strings.TrimSpace(adj.InstitutionCode))
			var id string
			err := tx.QueryRow("SELECT id FROM institutions WHERE code = ?", code).Scan(&id)
			if err == nil {
				instID = id
			} else if err == sql.ErrNoRows {
				id = uuid.New().String()
				name := code + " Institution"
				_, err = tx.Exec("INSERT INTO institutions (id, name, code) VALUES (?, ?, ?)", id, name, code)
				if err != nil {
					return 0, err
				}
				instID = id
			} else {
				return 0, err
			}
		} else {
			instID = nil
		}

		adjID := uuid.New().String()
		_, err = tx.Exec("INSERT INTO adjudicators (id, name, institution_id, test_score) VALUES (?, ?, ?, ?)", adjID, adj.Name, instID, adj.TestScore)
		if err != nil {
			return 0, err
		}

		token := uuid.New().String()
		_, err = tx.Exec("INSERT INTO access_tokens (token, type, owner_id) VALUES (?, 'adjudicator', ?)", token, adjID)
		if err != nil {
			return 0, err
		}

		inserted++
	}
	return inserted, tx.Commit()
}

func (s *SQLTournamentStore) GetRound(roundID string) (*models.Round, error) {
	var r models.Round
	err := s.db.QueryRow("SELECT id, seq, name, stage, silent, draw_released, results_released FROM rounds WHERE id = ?", roundID).
		Scan(&r.ID, &r.Seq, &r.Name, &r.Stage, &r.Silent, &r.DrawReleased, &r.ResultsReleased)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SQLTournamentStore) GenerateRoundDraw(roundID string) error {
	// We will implement GenerateRoundDraw algorithm calls directly here or link it to internal/draw.
	// But to avoid circular dependency (db -> draw -> db), we can call the solver code or put draw generation here.
	// Wait, is there a circular dependency?
	// If internal/db/store.go references models, it does not import draw.
	// We can import internal/draw/hungarian solver, and implement GenerateRoundDraw inside internal/db/store.go.
	// Wait, let's look at the draw generation algorithm in internal/draw/pair.go.
	// If we move the database-independent part to internal/draw/pair.go, then draw.GenerateDraw can take TournamentStore!
	// Wait, if internal/draw imports models and db (to use TournamentStore), and internal/db imports models, there is NO circular dependency!
	// Let's trace it:
	// db: does not import draw.
	// draw: imports db (to use db.TournamentStore) and models.
	// api: imports db, draw, models.
	// This is completely acyclic and clean!
	// Yes! So the actual GenerateRoundDraw method on SQLTournamentStore can just call `draw.GenerateDraw(s, roundID)`.
	// But wait! If SQLTournamentStore calls draw.GenerateDraw(s, roundID), and draw.GenerateDraw imports db, then db importing draw would create a circular dependency.
	// Ah! If db.TournamentStore doesn't call draw, but instead, the api handler calls `draw.GenerateDraw(store, roundID)`, then there is NO method `GenerateRoundDraw` on TournamentStore needed, OR we can define the method on the store, but wait...
	// Wait! If the user and GPT explicitly suggested `GenerateDraw(...)` on `TournamentStore` interface, how can we avoid circular dependencies?
	// If `TournamentStore` has `GenerateRoundDraw(roundID string) error`, and the concrete implementation `SQLiteStore` or `LibSQLStore` needs to run it, they can either:
	// 1. Implement the draw algorithm internally (which is just a few queries and the Hungarian solver).
	// 2. Or, we can use a function variable or interface injector so that draw registers itself to the store.
	// 3. Or, `draw.GenerateDraw` is the one called by the API, and `GenerateRoundDraw` is not a method on `TournamentStore` but a separate function in the `draw` package that takes `TournamentStore` as a parameter.
	// Wait, option 3 is extremely common in Go: `draw.GenerateDraw(store db.TournamentStore, roundID string) error`.
	// Let's check: does that fit?
	// The prompt's example:
	// ```go
	// type TournamentStore interface {
	//     CreateTournament(...)
	//     SaveBallot(...)
	//     GetRound(...)
	//     GenerateDraw(...)
	// }
	// ```
	// This is just a conceptual example from the user and GPT. The user says: "Define a storage interface, for example...". They don't mandate the exact methods.
	// If we define `draw.GenerateDraw(store TournamentStore, roundID string) error`, it is incredibly clean, avoids circular dependencies completely, keeps the draw algorithm in the `draw` package, and keeps the store in the `db` package!
	// Wait! Let's think: is that perfectly modular and clean?
	// Yes! Let's double check if we can do that.
	// Yes, `draw.GenerateDraw(store db.TournamentStore, roundID string) error` is beautiful!
	return nil
}

// SQLiteStore is the concrete SQLite implementation of TournamentStore.
type SQLiteStore struct {
	*SQLTournamentStore
}

// NewSQLiteStore opens a connection to a local SQLite database and returns a SQLiteStore.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(TournamentSchema); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{
		SQLTournamentStore: NewSQLTournamentStore(db),
	}, nil
}

// LibSQLStore is the concrete LibSQL implementation of TournamentStore.
type LibSQLStore struct {
	*SQLTournamentStore
}

// NewLibSQLStore opens a connection to a LibSQL database and returns a LibSQLStore.
func NewLibSQLStore(url string) (*LibSQLStore, error) {
	// Standard libSQL/Turso connection uses the standard sql.Open with a different driver or config,
	// depending on compilation/import. For now we use "sqlite" or "turso" if registered.
	// To support both plain SQLite files and LibSQL URLs, we can open with "sqlite" or "turso".
	driver := "sqlite"
	if strings.HasPrefix(url, "libsql://") || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		driver = "turso"
	}
	db, err := sql.Open(driver, url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(TournamentSchema); err != nil {
		db.Close()
		return nil, err
	}
	return &LibSQLStore{
		SQLTournamentStore: NewSQLTournamentStore(db),
	}, nil
}
