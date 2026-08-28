package db

import (
	"database/sql"
	"math"
	"sort"
	"time"

	"yellow/internal/models"

	"github.com/google/uuid"
)

// ListMotions returns all motions configured for a given round, ordered by sequence.
func (s *SQLTournamentStore) ListMotions(roundID string) ([]models.Motion, error) {
	rows, err := s.db.Query(`
		SELECT id, round_id, seq, reference, text, info_slide, released_at
		FROM motions
		WHERE round_id = ?
		ORDER BY seq ASC, id ASC
	`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Motion
	for rows.Next() {
		var m models.Motion
		var ref, info, rel sql.NullString
		if err := rows.Scan(&m.ID, &m.RoundID, &m.Seq, &ref, &m.Text, &info, &rel); err != nil {
			return nil, err
		}
		if ref.Valid {
			m.Reference = ref.String
		}
		if info.Valid {
			m.InfoSlide = info.String
		}
		if rel.Valid {
			m.ReleasedAt = &rel.String
		}
		list = append(list, m)
	}
	if list == nil {
		list = []models.Motion{}
	}
	return list, nil
}

// CreateMotion creates a new motion for a round.
func (s *SQLTournamentStore) CreateMotion(m models.Motion) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.Seq <= 0 {
		var maxSeq sql.NullInt64
		_ = s.db.QueryRow("SELECT MAX(seq) FROM motions WHERE round_id = ?", m.RoundID).Scan(&maxSeq)
		if maxSeq.Valid {
			m.Seq = int(maxSeq.Int64) + 1
		} else {
			m.Seq = 1
		}
	}

	var refVal, infoVal, relVal interface{}
	if m.Reference != "" {
		refVal = m.Reference
	}
	if m.InfoSlide != "" {
		infoVal = m.InfoSlide
	}
	if m.ReleasedAt != nil && *m.ReleasedAt != "" {
		relVal = *m.ReleasedAt
	}

	_, err := s.db.Exec(`
		INSERT INTO motions (id, round_id, seq, reference, text, info_slide, released_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, m.ID, m.RoundID, m.Seq, refVal, m.Text, infoVal, relVal)
	return err
}

// UpdateMotion updates an existing motion's fields.
func (s *SQLTournamentStore) UpdateMotion(m models.Motion) error {
	var refVal, infoVal, relVal interface{}
	if m.Reference != "" {
		refVal = m.Reference
	}
	if m.InfoSlide != "" {
		infoVal = m.InfoSlide
	}
	if m.ReleasedAt != nil && *m.ReleasedAt != "" {
		relVal = *m.ReleasedAt
	}

	res, err := s.db.Exec(`
		UPDATE motions
		SET seq = ?, reference = ?, text = ?, info_slide = ?, released_at = ?
		WHERE id = ?
	`, m.Seq, refVal, m.Text, infoVal, relVal, m.ID)
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

// DeleteMotion removes a motion by ID.
func (s *SQLTournamentStore) DeleteMotion(id string) error {
	res, err := s.db.Exec("DELETE FROM motions WHERE id = ?", id)
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

// ReleaseMotions sets or clears the released_at timestamp on all motions in a round.
func (s *SQLTournamentStore) ReleaseMotions(roundID string, release bool) error {
	var err error
	if release {
		now := time.Now().UTC().Format("2006-01-02 15:04:05")
		_, err = s.db.Exec("UPDATE motions SET released_at = ? WHERE round_id = ?", now, roundID)
	} else {
		_, err = s.db.Exec("UPDATE motions SET released_at = NULL WHERE round_id = ?", roundID)
	}
	return err
}

// GetDebateVetoes returns all motion veto preferences recorded for a debate.
func (s *SQLTournamentStore) GetDebateVetoes(debateID string) ([]models.MotionVeto, error) {
	rows, err := s.db.Query(`
		SELECT mv.id, mv.debate_id, mv.team_id, t.name, mv.motion_id, COALESCE(m.reference, ''), m.text, mv.preference
		FROM motion_vetoes mv
		JOIN teams t ON mv.team_id = t.id
		JOIN motions m ON mv.motion_id = m.id
		WHERE mv.debate_id = ?
		ORDER BY t.name ASC, mv.preference ASC
	`, debateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.MotionVeto
	for rows.Next() {
		var v models.MotionVeto
		if err := rows.Scan(&v.ID, &v.DebateID, &v.TeamID, &v.TeamName, &v.MotionID, &v.MotionReference, &v.MotionText, &v.Preference); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	if list == nil {
		list = []models.MotionVeto{}
	}
	return list, nil
}

// RecordMotionVeto inserts or updates a team's preference ranking for a motion in a debate.
func (s *SQLTournamentStore) RecordMotionVeto(debateID, teamID, motionID string, preference int) error {
	id := uuid.New().String()
	_, err := s.db.Exec(`
		INSERT INTO motion_vetoes (id, debate_id, team_id, motion_id, preference)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(debate_id, team_id, motion_id) DO UPDATE SET preference = excluded.preference
	`, id, debateID, teamID, motionID, preference)
	return err
}

// GetMotionStatistics computes side win rates and positional distributions across confirmed ballots for all motions.
func (s *SQLTournamentStore) GetMotionStatistics() ([]models.MotionStatistics, error) {
	// 1. Fetch all motions
	mRows, err := s.db.Query(`
		SELECT m.id, m.round_id, m.seq, COALESCE(m.reference, ''), m.text, r.name
		FROM motions m
		JOIN rounds r ON m.round_id = r.id
		ORDER BY r.seq ASC, m.seq ASC
	`)
	if err != nil {
		return nil, err
	}
	defer mRows.Close()

	type motionMeta struct {
		id, roundID string
		seq         int
		reference   string
		text        string
		roundName   string
	}
	var motions []motionMeta
	roundMotionMap := make(map[string][]string) // roundID -> []motionID

	for mRows.Next() {
		var mm motionMeta
		if err := mRows.Scan(&mm.id, &mm.roundID, &mm.seq, &mm.reference, &mm.text, &mm.roundName); err != nil {
			return nil, err
		}
		motions = append(motions, mm)
		roundMotionMap[mm.roundID] = append(roundMotionMap[mm.roundID], mm.id)
	}

	// 2. Prepare statistics lookup per motion ID
	statMap := make(map[string]*models.MotionStatistics, len(motions))
	for _, mm := range motions {
		statMap[mm.id] = &models.MotionStatistics{
			MotionID:         mm.id,
			Reference:        mm.reference,
			Text:             mm.text,
			RoundName:        mm.roundName,
			TotalDebates:     0,
			SideWins:         make(map[string]int),
			SidePercentages:  make(map[string]float64),
			PositionalCounts: make(map[string]map[int]int),
		}
	}

	// 3. Find confirmed debates across the tournament
	dRows, err := s.db.Query(`
		SELECT DISTINCT d.id, d.round_id
		FROM debates d
		JOIN ballots b ON b.debate_id = d.id
		WHERE b.status = 'confirmed'
	`)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()

	type debateInfo struct {
		debateID string
		roundID  string
	}
	var confirmedDebates []debateInfo
	for dRows.Next() {
		var di debateInfo
		if err := dRows.Scan(&di.debateID, &di.roundID); err != nil {
			return nil, err
		}
		confirmedDebates = append(confirmedDebates, di)
	}

	// 4. For each confirmed debate, determine the motion and tally positions
	for _, di := range confirmedDebates {
		roundMotions := roundMotionMap[di.roundID]
		if len(roundMotions) == 0 {
			continue
		}

		chosenMotionID := roundMotions[0]
		if len(roundMotions) > 1 {
			// Check veto preferences for this debate
			vRows, vErr := s.db.Query(`
				SELECT motion_id, SUM(preference) as pref_sum
				FROM motion_vetoes
				WHERE debate_id = ?
				GROUP BY motion_id
				ORDER BY pref_sum ASC
			`, di.debateID)
			if vErr == nil {
				if vRows.Next() {
					var mid string
					var psum int
					if err := vRows.Scan(&mid, &psum); err == nil && mid != "" {
						chosenMotionID = mid
					}
				}
				vRows.Close()
			}
		}

		stat, ok := statMap[chosenMotionID]
		if !ok {
			continue
		}

		// Get team points and sides for this confirmed debate
		// Use the confirmed ballot results
		resRows, err := s.db.Query(`
			SELECT dt.team_id, dt.side, MAX(br.points)
			FROM debate_teams dt
			JOIN ballots b ON b.debate_id = dt.debate_id AND b.status = 'confirmed'
			JOIN ballot_results br ON br.ballot_id = b.id AND br.team_id = dt.team_id
			WHERE dt.debate_id = ?
			GROUP BY dt.team_id, dt.side
		`, di.debateID)
		if err != nil {
			continue
		}

		type teamRes struct {
			teamID string
			side   string
			points int
		}
		var teamResults []teamRes
		for resRows.Next() {
			var tr teamRes
			if err := resRows.Scan(&tr.teamID, &tr.side, &tr.points); err != nil {
				continue
			}
			teamResults = append(teamResults, tr)
		}
		resRows.Close()

		if len(teamResults) == 0 {
			continue
		}

		// Sort teams by points descending
		sort.SliceStable(teamResults, func(i, j int) bool {
			return teamResults[i].points > teamResults[j].points
		})

		stat.TotalDebates++

		for idx, tr := range teamResults {
			rank := idx + 1
			if stat.PositionalCounts[tr.side] == nil {
				stat.PositionalCounts[tr.side] = make(map[int]int)
			}
			stat.PositionalCounts[tr.side][rank]++
			if rank == 1 {
				stat.SideWins[tr.side]++
			}
		}
	}

	// 5. Calculate percentages and construct output
	results := make([]models.MotionStatistics, 0, len(motions))
	for _, mm := range motions {
		stat := statMap[mm.id]
		if stat.TotalDebates > 0 {
			for side, wins := range stat.SideWins {
				stat.SidePercentages[side] = math.Round((float64(wins)/float64(stat.TotalDebates))*1000.0) / 10.0
			}
		}
		results = append(results, *stat)
	}

	return results, nil
}
