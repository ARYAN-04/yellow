package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"yellow/internal/models"

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
	UpdateTeam(teamID string, name, code, instID *string, novice, esl, efl, standby *bool) error
	UpsertSpeaker(teamID string, sp models.Speaker) error
	DeleteTeam(id string) error

	// Config & Format Presets
	GetConfig(key string) (string, error)
	SetConfig(key, value string) error
	ApplyFormatPreset(presetName string) error

	// Break Categories
	ListBreakCategories() ([]models.BreakCategory, error)
	CreateBreakCategory(c models.BreakCategory) error
	UpdateBreakCategory(c models.BreakCategory) error
	DeleteBreakCategory(id string) error

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

	// Conflicts
	ListConflicts() ([]models.Conflict, error)
	CreateConflict(subjectType, subjectID, targetType, targetID, weight string) error
	DeleteConflict(id string) error
	GetConflictsForDraw() ([]models.Conflict, error)
	GetDebateConflicts(debateID string) (hard []string, soft []string, err error)

	// Feedback
	ListFeedbackQuestions(fromType, toType string) ([]models.FeedbackQuestion, error)
	CreateFeedbackQuestion(q models.FeedbackQuestion) (models.FeedbackQuestion, error)
	UpdateFeedbackQuestion(q models.FeedbackQuestion) error
	DeleteFeedbackQuestion(id string) error
	MoveFeedbackQuestion(id, direction string) error
	ListFeedbackSubmissions(roundID string) ([]models.FeedbackSubmission, error)
	GetFeedbackTargets(ownerType, ownerID string) ([]models.FeedbackTarget, error)
	SubmitFeedback(debateID, sourceType, sourceID, targetAdjID string, score *float64, answers map[string]string) error
	RecalcAdjudicatorRatings() error

	// Manual allocations
	MoveSwapTeamAssignment(assignmentID, targetDebateID string) (roundID string, err error)
	MoveSwapAdjudicatorAssignment(assignmentID, targetDebateID, role string) (roundID string, err error)
	AddAdjudicatorToDebate(debateID, adjudicatorID, role string) (roundID string, err error)
	RemoveAdjudicatorFromDebate(assignmentID string) (roundID string, err error)

	// Ballots & Standings
	SubmitBallot(debateID, ballotID, submitterType, submitterID, status string, isSplit bool, entryGroup string, results []models.TeamBallotResult) error
	ConfirmBallot(ballotID string) error
	SetBallotStatus(ballotID, status string) error
	GetBallotByID(ballotID string) (*models.BallotSummary, error)
	GetBallotsForRound(roundID string) ([]models.BallotSummary, error)
	CompareEntryGroup(group string) (ballots []models.BallotSummary, match bool, diffs []models.BallotDiff, err error)
	GetStandings() ([]models.Standing, error)
	GetStandingsWithPrecedence(precedence []string, filterCategory string) ([]models.Standing, error)
	GetStandingsWithPrecedenceEx(precedence []string, filterCategory string, includeSilent bool) ([]models.Standing, error)
	GetSpeakerStandings(category string, trimmed bool, includeSilent bool) ([]models.SpeakerStanding, error)
	GetAdjudicatorStandings(includeSilent bool) ([]models.AdjudicatorStanding, error)

	// Breaks & Brackets
	ComputeBreak(categoryID string) (*models.BreakResult, error)
	SaveBreakSnapshot(categoryID string, teams []models.BreakTeam) error
	GenerateBracket(roundID, categoryID string) error
	AdvanceEliminationRound(roundID string) (string, error)
	GetBracket() ([]models.BracketRound, error)

	// Check-ins & Availability
	ListCheckins() ([]models.Checkin, error)
	SetCheckedIn(entityType, entityID string, checkedIn bool) error
	ResolveCheckinToken(token string) (*models.CheckinTokenInfo, error)
	GetRoundAvailability(roundID string) ([]models.AvailabilityOverride, error)
	SetRoundAvailability(roundID, entityType, entityID string, isAvailable bool) error
	SyncAvailabilityFromCheckins(roundID string) error

	// Tokens & Public Portal
	ResolveToken(token string) (*models.TokenInfo, error)
	GetTokenDebates(ownerID, ownerType string) ([]models.DebateInfo, error)
	SubmitTokenBallot(debateID, ballotID, adjID string, results []models.TeamBallotResult) error

	// Middleware & Access checks
	ValidateToken(token string) (*models.TokenOwner, error)
	CanViewRound(roundID string, isAdmin bool) (bool, error)

	// Motions & Vetoes
	ListMotions(roundID string) ([]models.Motion, error)
	CreateMotion(m models.Motion) error
	UpdateMotion(m models.Motion) error
	DeleteMotion(id string) error
	ReleaseMotions(roundID string, release bool) error
	GetDebateVetoes(debateID string) ([]models.MotionVeto, error)
	RecordMotionVeto(debateID, teamID, motionID string, preference int) error
	GetMotionStatistics() ([]models.MotionStatistics, error)

	// Venues
	ListVenues() ([]models.Venue, error)
	CreateVenue(v models.Venue) error
	UpdateVenue(v models.Venue) error
	DeleteVenue(id string) error
	GetAvailableVenuesForRound(roundID string) ([]models.Venue, error)
	ImportVenues(venues []models.Venue) (int, error)

	// CSV Imports
	ImportInstitutions(insts []models.Institution) (int, error)
	ImportTeams(teams []models.TeamImport) (int, error)
	ImportAdjudicators(adjs []models.AdjudicatorImport) (int, error)

	// Trajectories
	GetTeamTrajectory(teamID string, isAdmin bool) (*models.TeamTrajectory, error)
	GetSpeakerTrajectory(speakerID string, isAdmin bool) (*models.SpeakerTrajectory, error)
	GetAdjudicatorTrajectory(adjID string, isAdmin bool) (*models.AdjudicatorTrajectory, error)

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
		SELECT t.id, t.name, t.code, t.institution_id, i.name, i.code, COALESCE(t.is_novice, 0), COALESCE(t.is_esl, 0), COALESCE(t.is_efl, 0), COALESCE(t.is_standby, 0)
		FROM teams t
		LEFT JOIN institutions i ON t.institution_id = i.id
		ORDER BY t.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Team
	teamIndex := make(map[string]int)
	for rows.Next() {
		var t models.Team
		var instID, instName, instCode sql.NullString
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &instID, &instName, &instCode, &t.IsNovice, &t.IsEsl, &t.IsEfl, &t.IsStandby); err != nil {
			return nil, err
		}
		if instID.Valid {
			t.InstitutionID = &instID.String
			t.InstitutionName = &instName.String
			t.InstitutionCode = &instCode.String
		}
		t.Speakers = []models.Speaker{}
		list = append(list, t)
		teamIndex[t.ID] = len(list) - 1
	}

	spRows, err := s.db.Query("SELECT id, name, team_id, is_novice, is_esl, is_efl FROM speakers")
	if err != nil {
		return nil, err
	}
	defer spRows.Close()

	for spRows.Next() {
		var sp models.Speaker
		var teamID string
		if err := spRows.Scan(&sp.ID, &sp.Name, &teamID, &sp.IsNovice, &sp.IsEsl, &sp.IsEfl); err != nil {
			return nil, err
		}
		if idx, ok := teamIndex[teamID]; ok {
			list[idx].Speakers = append(list[idx].Speakers, sp)
		}
	}
	return list, nil
}

func (s *SQLTournamentStore) UpdateTeam(teamID string, name, code, instID *string, novice, esl, efl, standby *bool) error {
	query := "UPDATE teams SET "
	var updates []string
	var params []interface{}

	if name != nil {
		updates = append(updates, "name = ?")
		params = append(params, *name)
	}
	if code != nil {
		updates = append(updates, "code = ?")
		params = append(params, *code)
	}
	if instID != nil {
		updates = append(updates, "institution_id = ?")
		params = append(params, *instID)
	}
	if novice != nil {
		updates = append(updates, "is_novice = ?")
		params = append(params, *novice)
	}
	if esl != nil {
		updates = append(updates, "is_esl = ?")
		params = append(params, *esl)
	}
	if efl != nil {
		updates = append(updates, "is_efl = ?")
		params = append(params, *efl)
	}
	if standby != nil {
		updates = append(updates, "is_standby = ?")
		params = append(params, *standby)
	}

	if len(updates) == 0 {
		return nil
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	params = append(params, teamID)

	res, err := s.db.Exec(query, params...)
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

func (s *SQLTournamentStore) UpsertSpeaker(teamID string, sp models.Speaker) error {
	name := strings.TrimSpace(sp.Name)
	if name == "" && sp.ID != "" {
		var existing string
		err := s.db.QueryRow("SELECT name FROM speakers WHERE id = ?", sp.ID).Scan(&existing)
		if err != nil {
			return err
		}
		name = existing
	}

	var existingTeamID string
	err := s.db.QueryRow("SELECT team_id FROM speakers WHERE id = ?", sp.ID).Scan(&existingTeamID)
	if err == nil {
		if existingTeamID != teamID {
			return fmt.Errorf("speaker %s does not belong to team %s", sp.ID, teamID)
		}
		_, err = s.db.Exec("UPDATE speakers SET name = ?, is_novice = ?, is_esl = ?, is_efl = ? WHERE id = ?", name, sp.IsNovice, sp.IsEsl, sp.IsEfl, sp.ID)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}

	id := sp.ID
	if id == "" {
		id = uuid.New().String()
	}
	_, err = s.db.Exec("INSERT INTO speakers (id, name, team_id, is_novice, is_esl, is_efl) VALUES (?, ?, ?, ?, ?, ?)", id, name, teamID, sp.IsNovice, sp.IsEsl, sp.IsEfl)
	return err
}

func (s *SQLTournamentStore) GetConfig(key string) (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *SQLTournamentStore) SetConfig(key, value string) error {
	_, err := s.db.Exec("INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value)
	return err
}

func (s *SQLTournamentStore) ListBreakCategories() ([]models.BreakCategory, error) {
	rows, err := s.db.Query("SELECT id, name, seq, size, base_points, max_teams_per_institution, rule FROM break_categories ORDER BY seq ASC, name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.BreakCategory
	for rows.Next() {
		var c models.BreakCategory
		var size, basePoints, maxInst sql.NullInt64
		var rule sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Seq, &size, &basePoints, &maxInst, &rule); err != nil {
			return nil, err
		}
		if size.Valid {
			v := int(size.Int64)
			c.Size = &v
		}
		if basePoints.Valid {
			v := int(basePoints.Int64)
			c.BasePoints = &v
		}
		if maxInst.Valid {
			v := int(maxInst.Int64)
			c.MaxTeamsPerInstitution = &v
		}
		if rule.Valid {
			c.Rule = &rule.String
		}
		list = append(list, c)
	}
	return list, nil
}

func (s *SQLTournamentStore) CreateBreakCategory(c models.BreakCategory) error {
	_, err := s.db.Exec(
		"INSERT INTO break_categories (id, name, seq, size, base_points, max_teams_per_institution, rule) VALUES (?, ?, ?, ?, ?, ?, ?)",
		c.ID, c.Name, c.Seq, c.Size, c.BasePoints, c.MaxTeamsPerInstitution, c.Rule,
	)
	return err
}

func (s *SQLTournamentStore) UpdateBreakCategory(c models.BreakCategory) error {
	res, err := s.db.Exec(
		"UPDATE break_categories SET name = ?, seq = ?, size = ?, base_points = ?, max_teams_per_institution = ?, rule = ? WHERE id = ?",
		c.Name, c.Seq, c.Size, c.BasePoints, c.MaxTeamsPerInstitution, c.Rule, c.ID,
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

func (s *SQLTournamentStore) DeleteBreakCategory(id string) error {
	res, err := s.db.Exec("DELETE FROM break_categories WHERE id = ?", id)
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
	rows, err := s.db.Query(`
		SELECT d.id, d.venue, COALESCE(v.is_accessible, 0)
		FROM debates d
		LEFT JOIN venues v ON d.venue = v.name
		WHERE d.round_id = ?
	`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debates []models.DebateDraw

	for rows.Next() {
		var d models.DebateDraw
		if err := rows.Scan(&d.ID, &d.Venue, &d.VenueAccessible); err != nil {
			return nil, err
		}
		d.Teams = []models.TeamAssignment{}
		d.Adjudicators = []models.AdjudicatorAssignment{}
		debates = append(debates, d)
	}

	if len(debates) == 0 {
		return debates, nil
	}

	debateMap := make(map[string]*models.DebateDraw, len(debates))
	for i := range debates {
		debateMap[debates[i].ID] = &debates[i]
	}

	tRows, err := s.db.Query(`
		SELECT dt.debate_id, dt.id, dt.team_id, t.name, dt.side, COALESCE(dt.pull_up, 0)
		FROM debate_teams dt
		JOIN teams t ON dt.team_id = t.id
		JOIN debates d ON dt.debate_id = d.id
		WHERE d.round_id = ?
	`, roundID)
	if err != nil {
		return nil, err
	}
	defer tRows.Close()

	for tRows.Next() {
		var dID, tID, tName, side string
		var assignmentID string
		var pullUp bool
		if err := tRows.Scan(&dID, &assignmentID, &tID, &tName, &side, &pullUp); err != nil {
			return nil, err
		}
		if d, ok := debateMap[dID]; ok {
			d.Teams = append(d.Teams, models.TeamAssignment{ID: assignmentID, TeamID: tID, TeamName: tName, Side: side, PullUp: pullUp})
		}
	}

	aRows, err := s.db.Query(`
		SELECT da.debate_id, da.id, da.adjudicator_id, a.name, da.role
		FROM debate_adjudicators da
		JOIN adjudicators a ON da.adjudicator_id = a.id
		JOIN debates d ON da.debate_id = d.id
		WHERE d.round_id = ?
	`, roundID)
	if err != nil {
		return nil, err
	}
	defer aRows.Close()

	for aRows.Next() {
		var dID, aID, aName, role string
		var assignmentID string
		if err := aRows.Scan(&dID, &assignmentID, &aID, &aName, &role); err != nil {
			return nil, err
		}
		if d, ok := debateMap[dID]; ok {
			d.Adjudicators = append(d.Adjudicators, models.AdjudicatorAssignment{ID: assignmentID, AdjudicatorID: aID, AdjudicatorName: aName, Role: role})
		}
	}

	return debates, nil
}

func (s *SQLTournamentStore) GetSidesConfig() (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = 'sides'").Scan(&val)
	return val, err
}

func (s *SQLTournamentStore) GetTeamsForDraw() ([]models.TeamDrawInfo, error) {
	rows, err := s.db.Query("SELECT id, institution_id, COALESCE(is_standby, 0) FROM teams")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.TeamDrawInfo
	for rows.Next() {
		var t models.TeamDrawInfo
		var instID sql.NullString
		if err := rows.Scan(&t.ID, &instID, &t.Standby); err != nil {
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
	rows, err := s.db.Query("SELECT id, institution_id, COALESCE(rating, test_score) FROM adjudicators")
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
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		existingDebateIDs = append(existingDebateIDs, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, dID := range existingDebateIDs {
		if _, err := tx.Exec("DELETE FROM debate_teams WHERE debate_id = ?", dID); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM debate_adjudicators WHERE debate_id = ?", dID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec("DELETE FROM debates WHERE round_id = ?", roundID); err != nil {
		return err
	}

	for _, d := range debates {
		var posVal interface{}
		if d.BracketPosition > 0 {
			posVal = d.BracketPosition
		}
		_, err = tx.Exec("INSERT INTO debates (id, round_id, venue, bracket_position) VALUES (?, ?, ?, ?)", d.DebateID, roundID, d.Venue, posVal)
		if err != nil {
			return err
		}

		for _, t := range d.Teams {
			dtID := uuid.New().String()
			_, err = tx.Exec(
				"INSERT INTO debate_teams (id, debate_id, team_id, side, pull_up) VALUES (?, ?, ?, ?, ?)",
				dtID, d.DebateID, t.TeamID, t.Side, t.PullUp,
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

func (s *SQLTournamentStore) SubmitBallot(debateID, ballotID, submitterType, submitterID, status string, isSplit bool, entryGroup string, results []models.TeamBallotResult) error {
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

	if entryGroup != "" {
		_, err = tx.Exec(
			"DELETE FROM ballots WHERE entry_group = ? AND submitter_type = ? AND COALESCE(submitter_id, '') = ? AND status != 'confirmed'",
			entryGroup, submitterType, submitterID,
		)
		if err != nil {
			return err
		}
	}

	var subIDVal interface{}
	if submitterID != "" {
		subIDVal = submitterID
	} else {
		subIDVal = nil
	}

	var groupVal interface{}
	if entryGroup != "" {
		groupVal = entryGroup
	} else {
		groupVal = nil
	}

	_, err = tx.Exec(
		"INSERT INTO ballots (id, debate_id, round_id, submitter_type, submitter_id, status, is_split, entry_group) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		ballotID, debateID, roundID, submitterType, subIDVal, status, isSplit, groupVal,
	)
	if err != nil {
		return err
	}

	for _, res := range results {
		brID := uuid.New().String()
		var adjVal interface{}
		if res.AdjudicatorID != nil && *res.AdjudicatorID != "" {
			adjVal = *res.AdjudicatorID
		} else {
			adjVal = nil
		}
		_, err = tx.Exec(
			"INSERT INTO ballot_results (id, ballot_id, team_id, points, speaker_points, adjudicator_id) VALUES (?, ?, ?, ?, ?, ?)",
			brID, ballotID, res.TeamID, res.Points, res.SpeakerPoints, adjVal,
		)
		if err != nil {
			return err
		}

		for idx, sp := range res.SpeakerScores {
			if sp.SpeakerID == "" {
				continue
			}
			spID := uuid.New().String()
			order := sp.SpeechOrder
			if order <= 0 {
				order = idx + 1
			}
			var roleVal interface{}
			if sp.Role != "" {
				roleVal = sp.Role
			} else {
				roleVal = nil
			}
			_, err = tx.Exec(
				"INSERT INTO speaker_scores (id, ballot_id, speaker_id, team_id, score, is_reply, speech_order, role) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				spID, ballotID, sp.SpeakerID, res.TeamID, sp.Score, sp.IsReply, order, roleVal,
			)
			if err != nil {
				return err
			}
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

func (s *SQLTournamentStore) computeStandings(includeSilent bool) ([]models.Standing, map[string][3]bool, error) {
	tRows, err := s.db.Query(`
		SELECT t.id, t.name, t.code, i.code, t.is_novice, t.is_esl, t.is_efl
		FROM teams t
		LEFT JOIN institutions i ON t.institution_id = i.id
	`)
	if err != nil {
		return nil, nil, err
	}
	defer tRows.Close()

	var list []models.Standing
	flags := make(map[string][3]bool)
	for tRows.Next() {
		var ts models.Standing
		var instCode sql.NullString
		var f [3]bool
		if err := tRows.Scan(&ts.TeamID, &ts.TeamName, &ts.TeamCode, &instCode, &f[0], &f[1], &f[2]); err != nil {
			return nil, nil, err
		}
		if instCode.Valid {
			ts.InstitutionCode = instCode.String
		}
		list = append(list, ts)
		flags[ts.TeamID] = f
	}

	type resultRow struct {
		ballotID string
		teamID   string
		points   int
		spts     float64
	}
	var rows []resultRow
	pRows, err := s.db.Query(`
		SELECT b.id, br.team_id, br.points, br.speaker_points
		FROM ballot_results br
		JOIN ballots b ON br.ballot_id = b.id
		JOIN rounds r ON b.round_id = r.id
		WHERE b.status = 'confirmed' AND (? OR r.silent = 0 OR r.results_released = 1)
		ORDER BY b.id ASC
	`, includeSilent)
	if err != nil {
		return nil, nil, err
	}
	defer pRows.Close()
	for pRows.Next() {
		var rr resultRow
		if err := pRows.Scan(&rr.ballotID, &rr.teamID, &rr.points, &rr.spts); err != nil {
			return nil, nil, err
		}
		rows = append(rows, rr)
	}

	standByTeam := make(map[string]*models.Standing)
	for i := range list {
		standByTeam[list[i].TeamID] = &list[i]
	}

	i := 0
	for i < len(rows) {
		j := i
		for j < len(rows) && rows[j].ballotID == rows[i].ballotID {
			j++
		}
		total := 0
		for _, rr := range rows[i:j] {
			total += rr.points
		}
		n := j - i
		for _, rr := range rows[i:j] {
			ts, ok := standByTeam[rr.teamID]
			if !ok {
				continue
			}
			ts.Points += rr.points
			ts.SpeakerPoints += rr.spts
			if n > 1 {
				ts.Margin += float64(rr.points) - float64(total-rr.points)/float64(n-1)
			}
		}
		i = j
	}

	return list, flags, nil
}

var validPrecedenceKeys = map[string]bool{
	"points":         true,
	"speaker_points": true,
	"margin":         true,
}

func (s *SQLTournamentStore) GetStandings() ([]models.Standing, error) {
	list, _, err := s.computeStandings(true)
	return list, err
}

// GetStandingsWithPrecedence computes standings with full visibility (admin semantics).
func (s *SQLTournamentStore) GetStandingsWithPrecedence(precedence []string, filterCategory string) ([]models.Standing, error) {
	return s.GetStandingsWithPrecedenceEx(precedence, filterCategory, true)
}

// GetStandingsWithPrecedenceEx optionally excludes confirmed ballots from silent,
// unreleased rounds so non-admin viewers cannot infer hidden results.
func (s *SQLTournamentStore) GetStandingsWithPrecedenceEx(precedence []string, filterCategory string, includeSilent bool) ([]models.Standing, error) {
	list, flags, err := s.computeStandings(includeSilent)
	if err != nil {
		return nil, err
	}

	switch filterCategory {
	case "":
	case "novice", "esl", "efl":
		idx := map[string]int{"novice": 0, "esl": 1, "efl": 2}[filterCategory]
		filtered := make([]models.Standing, 0)
		for _, ts := range list {
			if flags[ts.TeamID][idx] {
				filtered = append(filtered, ts)
			}
		}
		list = filtered
	default:
		cats, err := s.ListBreakCategories()
		if err != nil {
			return nil, err
		}
		var basePoints *int
		for _, c := range cats {
			if c.ID == filterCategory {
				basePoints = c.BasePoints
				break
			}
		}
		if basePoints != nil {
			filtered := make([]models.Standing, 0)
			for _, ts := range list {
				if ts.Points >= *basePoints {
					filtered = append(filtered, ts)
				}
			}
			list = filtered
		}
	}

	keys := make([]string, 0, len(precedence))
	for _, k := range precedence {
		if validPrecedenceKeys[k] {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		keys = []string{"points", "speaker_points", "margin"}
	}

	sort.SliceStable(list, func(a, b int) bool {
		for _, k := range keys {
			var av, bv float64
			switch k {
			case "points":
				av, bv = float64(list[a].Points), float64(list[b].Points)
			case "speaker_points":
				av, bv = list[a].SpeakerPoints, list[b].SpeakerPoints
			case "margin":
				av, bv = list[a].Margin, list[b].Margin
			}
			if av != bv {
				return av > bv
			}
		}
		return false
	})

	return list, nil
}

// GetSpeakerStandings computes individual speaker rankings based on confirmed substantive ballot scores.
func (s *SQLTournamentStore) GetSpeakerStandings(category string, trimmed bool, includeSilent bool) ([]models.SpeakerStanding, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.name, s.team_id, t.name, COALESCE(i.code, ''),
		       COALESCE(s.is_novice, 0), COALESCE(s.is_esl, 0), COALESCE(s.is_efl, 0),
		       COALESCE(t.is_novice, 0), COALESCE(t.is_esl, 0), COALESCE(t.is_efl, 0)
		FROM speakers s
		JOIN teams t ON s.team_id = t.id
		LEFT JOIN institutions i ON t.institution_id = i.id
		ORDER BY s.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type spMeta struct {
		id, name, teamID, teamName, instCode string
		isNovice, isEsl, isEfl               bool
		tNovice, tEsl, tEfl                  bool
	}
	var speakers []spMeta
	for rows.Next() {
		var sm spMeta
		if err := rows.Scan(&sm.id, &sm.name, &sm.teamID, &sm.teamName, &sm.instCode,
			&sm.isNovice, &sm.isEsl, &sm.isEfl,
			&sm.tNovice, &sm.tEsl, &sm.tEfl); err != nil {
			return nil, err
		}
		speakers = append(speakers, sm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	cat := strings.ToLower(strings.TrimSpace(category))
	var filtered []spMeta
	for _, sm := range speakers {
		match := false
		switch cat {
		case "", "all", "open":
			match = true
		case "novice":
			match = sm.isNovice || sm.tNovice
		case "esl":
			match = sm.isEsl || sm.tEsl
		case "efl":
			match = sm.isEfl || sm.tEfl
		default:
			match = true
		}
		if match {
			filtered = append(filtered, sm)
		}
	}

	scoreRows, err := s.db.Query(`
		SELECT ss.speaker_id, ss.score
		FROM speaker_scores ss
		JOIN ballots b ON ss.ballot_id = b.id
		JOIN rounds r ON b.round_id = r.id
		WHERE b.status = 'confirmed' AND (ss.is_reply = 0 OR ss.is_reply IS NULL) AND (? OR r.silent = 0 OR r.results_released = 1)
		ORDER BY ss.speaker_id, ss.score ASC
	`, includeSilent)
	if err != nil {
		return nil, err
	}
	defer scoreRows.Close()

	scoresBySpeaker := make(map[string][]float64)
	for scoreRows.Next() {
		var spID string
		var sc float64
		if err := scoreRows.Scan(&spID, &sc); err != nil {
			return nil, err
		}
		scoresBySpeaker[spID] = append(scoresBySpeaker[spID], sc)
	}
	if err := scoreRows.Err(); err != nil {
		return nil, err
	}

	standings := make([]models.SpeakerStanding, 0, len(filtered))
	for _, sm := range filtered {
		scores := scoresBySpeaker[sm.id]
		sort.Float64s(scores)
		count := len(scores)
		total := 0.0
		for _, sc := range scores {
			total += sc
		}
		avg := 0.0
		if count > 0 {
			avg = total / float64(count)
		}
		trimmedScore := avg
		if count >= 3 {
			trimmedSum := 0.0
			for k := 1; k < count-1; k++ {
				trimmedSum += scores[k]
			}
			trimmedScore = trimmedSum / float64(count-2)
		}

		standings = append(standings, models.SpeakerStanding{
			SpeakerID:       sm.id,
			SpeakerName:     sm.name,
			TeamID:          sm.teamID,
			TeamName:        sm.teamName,
			InstitutionCode: sm.instCode,
			TotalScore:      total,
			AverageScore:    avg,
			TrimmedScore:    trimmedScore,
			SpeechCount:     count,
			IsNovice:        sm.isNovice || sm.tNovice,
			IsEsl:           sm.isEsl || sm.tEsl,
			IsEfl:           sm.isEfl || sm.tEfl,
		})
	}

	if trimmed {
		sort.Slice(standings, func(i, j int) bool {
			if standings[i].TrimmedScore != standings[j].TrimmedScore {
				return standings[i].TrimmedScore > standings[j].TrimmedScore
			}
			if standings[i].TotalScore != standings[j].TotalScore {
				return standings[i].TotalScore > standings[j].TotalScore
			}
			if standings[i].AverageScore != standings[j].AverageScore {
				return standings[i].AverageScore > standings[j].AverageScore
			}
			return standings[i].SpeakerName < standings[j].SpeakerName
		})
	} else {
		sort.Slice(standings, func(i, j int) bool {
			if standings[i].TotalScore != standings[j].TotalScore {
				return standings[i].TotalScore > standings[j].TotalScore
			}
			if standings[i].AverageScore != standings[j].AverageScore {
				return standings[i].AverageScore > standings[j].AverageScore
			}
			if standings[i].TrimmedScore != standings[j].TrimmedScore {
				return standings[i].TrimmedScore > standings[j].TrimmedScore
			}
			return standings[i].SpeakerName < standings[j].SpeakerName
		})
	}

	for i := range standings {
		standings[i].Rank = i + 1
	}

	return standings, nil
}

type FormatPresetConfig struct {
	Sides             string
	ScoreMin          float64
	ScoreMax          float64
	SpeakersPerTeam   int
	HasReplySpeeches  bool
	ReplyScoreMin     float64
	ReplyScoreMax     float64
	RankingPrecedence []string
}

var FormatPresets = map[string]FormatPresetConfig{
	"bp": {
		Sides:             "OG,OO,CG,CO",
		ScoreMin:          50,
		ScoreMax:          100,
		SpeakersPerTeam:   2,
		HasReplySpeeches:  false,
		ReplyScoreMin:     0,
		ReplyScoreMax:     0,
		RankingPrecedence: []string{"points", "speaker_points", "margin"},
	},
	"australs": {
		Sides:             "Aff,Neg",
		ScoreMin:          65,
		ScoreMax:          85,
		SpeakersPerTeam:   3,
		HasReplySpeeches:  true,
		ReplyScoreMin:     30,
		ReplyScoreMax:     45,
		RankingPrecedence: []string{"points", "speaker_points", "margin"},
	},
	"asians": {
		Sides:             "Gov,Opp",
		ScoreMin:          65,
		ScoreMax:          85,
		SpeakersPerTeam:   3,
		HasReplySpeeches:  true,
		ReplyScoreMin:     30,
		ReplyScoreMax:     45,
		RankingPrecedence: []string{"points", "margin", "speaker_points"},
	},
	"wsdc": {
		Sides:             "Prop,Opp",
		ScoreMin:          60,
		ScoreMax:          80,
		SpeakersPerTeam:   3,
		HasReplySpeeches:  true,
		ReplyScoreMin:     30,
		ReplyScoreMax:     40,
		RankingPrecedence: []string{"points", "speaker_points", "margin"},
	},
	"apda": {
		Sides:             "Gov,Opp",
		ScoreMin:          20,
		ScoreMax:          30,
		SpeakersPerTeam:   2,
		HasReplySpeeches:  false,
		ReplyScoreMin:     0,
		ReplyScoreMax:     0,
		RankingPrecedence: []string{"points", "speaker_points", "margin"},
	},
}

func (s *SQLTournamentStore) ApplyFormatPreset(presetName string) error {
	p := strings.ToLower(strings.TrimSpace(presetName))
	if p == "uadc" {
		p = "asians"
	}
	preset, ok := FormatPresets[p]
	if !ok {
		return fmt.Errorf("unknown format preset: %s (supported: bp, australs, asians, wsdc, apda)", presetName)
	}

	precBytes, err := json.Marshal(preset.RankingPrecedence)
	if err != nil {
		return err
	}

	configs := map[string]string{
		"preset":             p,
		"sides":              preset.Sides,
		"score_min":          fmt.Sprintf("%.1f", preset.ScoreMin),
		"score_max":          fmt.Sprintf("%.1f", preset.ScoreMax),
		"speakers_per_team":  fmt.Sprintf("%d", preset.SpeakersPerTeam),
		"has_reply_speeches": fmt.Sprintf("%t", preset.HasReplySpeeches),
		"reply_score_min":    fmt.Sprintf("%.1f", preset.ReplyScoreMin),
		"reply_score_max":    fmt.Sprintf("%.1f", preset.ReplyScoreMax),
		"ranking_precedence": string(precBytes),
	}

	for k, v := range configs {
		if err := s.SetConfig(k, v); err != nil {
			return fmt.Errorf("failed to set config %s: %w", k, err)
		}
	}
	return nil
}

func (s *SQLTournamentStore) ResolveToken(token string) (*models.TokenInfo, error) {
	var tType, ownerID string
	err := s.db.QueryRow("SELECT type, owner_id FROM access_tokens WHERE token = ?", token).Scan(&tType, &ownerID)
	if err != nil {
		return nil, err
	}

	var name string
	var speakers []models.Speaker
	if tType == "team" {
		_ = s.db.QueryRow("SELECT name FROM teams WHERE id = ?", ownerID).Scan(&name)
		spRows, err := s.db.Query("SELECT id, name, is_novice, is_esl, is_efl FROM speakers WHERE team_id = ? ORDER BY name ASC", ownerID)
		if err == nil {
			for spRows.Next() {
				var sp models.Speaker
				if err := spRows.Scan(&sp.ID, &sp.Name, &sp.IsNovice, &sp.IsEsl, &sp.IsEfl); err == nil {
					speakers = append(speakers, sp)
				}
			}
			spRows.Close()
		}
	} else if tType == "adjudicator" {
		_ = s.db.QueryRow("SELECT name FROM adjudicators WHERE id = ?", ownerID).Scan(&name)
	}

	return &models.TokenInfo{
		Type:      tType,
		OwnerID:   ownerID,
		OwnerName: name,
		Speakers:  speakers,
	}, nil
}

func (s *SQLTournamentStore) GetTokenDebates(ownerID, ownerType string) ([]models.DebateInfo, error) {
	var debates []models.DebateInfo

	if ownerType == "adjudicator" {
		rows, err := s.db.Query(`
			SELECT d.id, d.venue, r.name, r.seq, da.role, COALESCE(b.status, 'unsubmitted'),
			       COALESCE((SELECT text FROM motions WHERE round_id = r.id ORDER BY seq ASC LIMIT 1), '')
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
			var motionText string
			if err := rows.Scan(&d.ID, &d.Venue, &d.RoundName, &d.RoundSeq, &d.Role, &d.BallotStatus, &motionText); err != nil {
				return nil, err
			}
			d.Motion = motionText
			d.Teams = []models.TeamAssignment{}
			d.Adjudicators = []string{}
			d.Panellists = []string{}
			debates = append(debates, d)
		}
	} else if ownerType == "team" {
		rows, err := s.db.Query(`
			SELECT d.id, d.venue, r.name, r.seq, dt.side, r.results_released,
			       COALESCE(br.points, -1), COALESCE(br.speaker_points, -1.0),
			       COALESCE(b.status, ''),
			       COALESCE((SELECT text FROM motions WHERE round_id = r.id AND (released_at IS NOT NULL OR r.draw_released = 1) ORDER BY seq ASC LIMIT 1), '')
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
			var bStatus string
			var motionText string
			if err := rows.Scan(&d.ID, &d.Venue, &d.RoundName, &d.RoundSeq, &d.Side, &resultsReleased, &pts, &spPts, &bStatus, &motionText); err != nil {
				return nil, err
			}
			if resultsReleased && pts != -1 {
				d.Points = &pts
				d.SpeakerPoints = &spPts
			}
			d.BallotStatus = bStatus
			d.Motion = motionText
			d.Teams = []models.TeamAssignment{}
			d.Adjudicators = []string{}
			d.Panellists = []string{}
			debates = append(debates, d)
		}
	}

	for idx := range debates {
		var teamList []models.TeamAssignment
		tRows, err := s.db.Query(`
			SELECT dt.team_id, t.name, dt.side
			FROM debate_teams dt
			JOIN teams t ON dt.team_id = t.id
			WHERE dt.debate_id = ?
			ORDER BY dt.side ASC
		`, debates[idx].ID)
		if err == nil {
			for tRows.Next() {
				var ta models.TeamAssignment
				if err := tRows.Scan(&ta.TeamID, &ta.TeamName, &ta.Side); err == nil {
					teamList = append(teamList, ta)
				}
			}
			tRows.Close()

			for tIdx := range teamList {
				spRows, spErr := s.db.Query(`
					SELECT id, name, is_novice, is_esl, is_efl
					FROM speakers
					WHERE team_id = ?
					ORDER BY name ASC
				`, teamList[tIdx].TeamID)
				if spErr == nil {
					for spRows.Next() {
						var sp models.Speaker
						if err := spRows.Scan(&sp.ID, &sp.Name, &sp.IsNovice, &sp.IsEsl, &sp.IsEfl); err == nil {
							teamList[tIdx].Speakers = append(teamList[tIdx].Speakers, sp)
						}
					}
					spRows.Close()
				}
			}
			debates[idx].Teams = teamList
		}

		aRows, aErr := s.db.Query(`
			SELECT a.name, da.role
			FROM debate_adjudicators da
			JOIN adjudicators a ON da.adjudicator_id = a.id
			WHERE da.debate_id = ?
			ORDER BY CASE da.role WHEN 'chair' THEN 1 WHEN 'panel' THEN 2 ELSE 3 END
		`, debates[idx].ID)
		if aErr == nil {
			for aRows.Next() {
				var aName, aRole string
				if err := aRows.Scan(&aName, &aRole); err == nil {
					if aRole == "chair" && debates[idx].Chair == "" {
						debates[idx].Chair = aName
					} else {
						debates[idx].Panellists = append(debates[idx].Panellists, aName)
					}
					debates[idx].Adjudicators = append(debates[idx].Adjudicators, aName)
				}
			}
			aRows.Close()
		}

		if ownerType == "team" && debates[idx].Points != nil {
			ssRows, ssErr := s.db.Query(`
				SELECT ss.id, ss.ballot_id, ss.speaker_id, s.name, ss.team_id, ss.score, ss.is_reply, ss.speech_order, COALESCE(ss.role, '')
				FROM speaker_scores ss
				JOIN ballots b ON ss.ballot_id = b.id
				JOIN speakers s ON ss.speaker_id = s.id
				WHERE b.debate_id = ? AND ss.team_id = ? AND b.status = 'confirmed'
				ORDER BY ss.speech_order ASC, ss.is_reply ASC
			`, debates[idx].ID, ownerID)
			if ssErr == nil {
				for ssRows.Next() {
					var ss models.SpeakerScore
					if err := ssRows.Scan(&ss.ID, &ss.BallotID, &ss.SpeakerID, &ss.SpeakerName, &ss.TeamID, &ss.Score, &ss.IsReply, &ss.SpeechOrder, &ss.Role); err == nil {
						debates[idx].SpeakerScores = append(debates[idx].SpeakerScores, ss)
					}
				}
				ssRows.Close()
			}
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

	return s.SubmitBallot(debateID, ballotID, "adjudicator", adjID, "submitted", false, "", results)
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

// SQLiteStore is the concrete SQLite implementation of TournamentStore.
type SQLiteStore struct {
	*SQLTournamentStore
}

// NewSQLiteStore opens a connection to a local SQLite database and returns a SQLiteStore.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := InitTournamentDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{
		SQLTournamentStore: NewSQLTournamentStore(db),
	}, nil
}
