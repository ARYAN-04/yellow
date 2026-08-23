package db

import (
	"database/sql"
	"sort"
	"strings"

	"yellow/internal/models"
)

const ballotSummarySelect = `
	SELECT b.id, b.debate_id, d.venue, b.submitter_type,
	       COALESCE(a.name, COALESCE(b.submitter_id, '')), b.status,
	       COALESCE(b.is_split, 0), b.entry_group
	FROM ballots b
	JOIN debates d ON b.debate_id = d.id
	LEFT JOIN adjudicators a ON a.id = b.submitter_id AND b.submitter_type = 'adjudicator'
`

// queryBallotSummaries loads ballot headers plus all their results in two passes.
func (s *SQLTournamentStore) queryBallotSummaries(where string, args ...interface{}) ([]models.BallotSummary, error) {
	rows, err := s.db.Query(ballotSummarySelect+" WHERE "+where+" ORDER BY d.venue ASC, b.id ASC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.BallotSummary
	index := make(map[string]int)
	for rows.Next() {
		var b models.BallotSummary
		if err := rows.Scan(&b.ID, &b.DebateID, &b.DebateVenue, &b.SubmitterType, &b.SubmitterName, &b.Status, &b.IsSplit, &b.EntryGroup); err != nil {
			return nil, err
		}
		b.Results = []models.TeamBallotResult{}
		list = append(list, b)
		index[b.ID] = len(list) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return list, nil
	}

	resRows, err := s.db.Query(`
		SELECT br.ballot_id, br.team_id, br.points, br.speaker_points, br.adjudicator_id
		FROM ballot_results br
		JOIN ballots b ON br.ballot_id = b.id
		WHERE `+where+`
	`, args...)
	if err != nil {
		return nil, err
	}
	defer resRows.Close()

	for resRows.Next() {
		var ballotID string
		var r models.TeamBallotResult
		if err := resRows.Scan(&ballotID, &r.TeamID, &r.Points, &r.SpeakerPoints, &r.AdjudicatorID); err != nil {
			return nil, err
		}
		if i, ok := index[ballotID]; ok {
			list[i].Results = append(list[i].Results, r)
		}
	}
	return list, resRows.Err()
}

// GetBallotsForRound returns every ballot (with results) recorded for the round.
func (s *SQLTournamentStore) GetBallotsForRound(roundID string) ([]models.BallotSummary, error) {
	return s.queryBallotSummaries("b.round_id = ?", roundID)
}

// GetBallotByID returns a single ballot with its results.
func (s *SQLTournamentStore) GetBallotByID(ballotID string) (*models.BallotSummary, error) {
	list, err := s.queryBallotSummaries("b.id = ?", ballotID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, sql.ErrNoRows
	}
	return &list[0], nil
}

// SetBallotStatus transitions a ballot to an arbitrary status ('confirmed', 'discrepancy', ...).
func (s *SQLTournamentStore) SetBallotStatus(ballotID, status string) error {
	res, err := s.db.Exec("UPDATE ballots SET status = ? WHERE id = ?", status, ballotID)
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

// CompareEntryGroup fetches all unconfirmed ballots in a double-entry group and
// reports whether their results match (order-insensitive by team_id + adjudicator_id).
func (s *SQLTournamentStore) CompareEntryGroup(group string) ([]models.BallotSummary, bool, []models.BallotDiff, error) {
	ballots, err := s.queryBallotSummaries("b.entry_group = ? AND b.status != 'confirmed'", group)
	if err != nil {
		return nil, false, nil, err
	}
	if len(ballots) < 2 {
		return ballots, true, nil, nil
	}
	diffs := CompareBallotSummaries(ballots[0], ballots[1])
	return ballots, len(diffs) == 0, diffs, nil
}

// CompareBallotSummaries diffs two ballots' results keyed by team_id + adjudicator_id.
func CompareBallotSummaries(a, b models.BallotSummary) []models.BallotDiff {
	key := func(r models.TeamBallotResult) string {
		adj := ""
		if r.AdjudicatorID != nil {
			adj = *r.AdjudicatorID
		}
		return r.TeamID + "|" + adj
	}

	mapResults := func(rs []models.TeamBallotResult) map[string]models.TeamBallotResult {
		m := make(map[string]models.TeamBallotResult, len(rs))
		for _, r := range rs {
			m[key(r)] = r
		}
		return m
	}

	resultsA, resultsB := mapResults(a.Results), mapResults(b.Results)
	keys := make([]string, 0, len(resultsA)+len(resultsB))
	for k := range resultsA {
		keys = append(keys, k)
	}
	for k := range resultsB {
		if _, ok := resultsA[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	floatPtr := func(v float64) *float64 { return &v }

	var diffs []models.BallotDiff
	for _, k := range keys {
		parts := strings.SplitN(k, "|", 2)
		teamID := parts[0]
		var adjID *string
		if parts[1] != "" {
			v := parts[1]
			adjID = &v
		}
		ra, inA := resultsA[k]
		rb, inB := resultsB[k]
		if ra.Points != rb.Points || !inA || !inB {
			var va, vb *float64
			if inA {
				va = floatPtr(float64(ra.Points))
			}
			if inB {
				vb = floatPtr(float64(rb.Points))
			}
			diffs = append(diffs, models.BallotDiff{TeamID: teamID, AdjudicatorID: adjID, Field: "points", BallotA: va, BallotB: vb})
		}
		if ra.SpeakerPoints != rb.SpeakerPoints || !inA || !inB {
			var va, vb *float64
			if inA {
				va = floatPtr(ra.SpeakerPoints)
			}
			if inB {
				vb = floatPtr(rb.SpeakerPoints)
			}
			diffs = append(diffs, models.BallotDiff{TeamID: teamID, AdjudicatorID: adjID, Field: "speaker_points", BallotA: va, BallotB: vb})
		}
	}
	return diffs
}
