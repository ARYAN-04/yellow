package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"yellow/internal/models"

	"github.com/google/uuid"
)

// bracket_position convention: within an elimination round, debates are numbered
// 1..N/2 in seed-pair order (debate 1 = seed 1 vs seed N, i.e. the top half of the
// draw); the last remaining round is the Grand Final. Rounds themselves sequence
// eliminations via rounds.seq.

func (s *SQLTournamentStore) breakPrecedence() []string {
	prec := []string{"points", "speaker_points", "margin"}
	if v, err := s.GetConfig("ranking_precedence"); err == nil {
		var p []string
		if json.Unmarshal([]byte(v), &p) == nil && len(p) > 0 {
			prec = p
		}
	}
	return prec
}

// ComputeBreak resolves the qualifier list for a category: "open" (all teams),
// "novice"/"esl"/"efl" (flag filters), or a break_categories row ID.
func (s *SQLTournamentStore) ComputeBreak(categoryID string) (*models.BreakResult, error) {
	res := &models.BreakResult{CategoryID: categoryID, Qualifiers: []models.BreakTeam{}}

	filter := ""
	var maxPerInst *int
	switch categoryID {
	case "open":
		res.CategoryName = "Open"
	case "novice":
		res.CategoryName = "Novice"
		filter = "novice"
	case "esl":
		res.CategoryName = "ESL"
		filter = "esl"
	case "efl":
		res.CategoryName = "EFL"
		filter = "efl"
	default:
		cats, err := s.ListBreakCategories()
		if err != nil {
			return nil, err
		}
		found := false
		for _, c := range cats {
			if c.ID == categoryID {
				res.CategoryName = c.Name
				res.Size = c.Size
				maxPerInst = c.MaxTeamsPerInstitution
				filter = categoryID
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown break category: %s", categoryID)
		}
	}

	standings, err := s.GetStandingsWithPrecedenceEx(s.breakPrecedence(), filter, true)
	if err != nil {
		return nil, err
	}
	s.applyBreakCutoff(res, standings, maxPerInst)
	return res, nil
}

// applyBreakCutoff caps qualifiers at the category size and flags bubble teams:
// teams tied on points with the last qualifier, both inside and just outside the cutoff.
// If maxPerInst is set, it caps qualifiers per institution, continuing down standings.
func (s *SQLTournamentStore) applyBreakCutoff(res *models.BreakResult, standings []models.Standing, maxPerInst *int) {
	instCounts := make(map[string]int)
	limitPerInst := 0
	if maxPerInst != nil && *maxPerInst > 0 {
		limitPerInst = *maxPerInst
	}

	var eligible []models.Standing
	for _, st := range standings {
		inst := strings.TrimSpace(st.InstitutionCode)
		if limitPerInst > 0 && inst != "" {
			if instCounts[inst] >= limitPerInst {
				continue
			}
			instCounts[inst]++
		}
		eligible = append(eligible, st)
	}

	cutoff := 0
	if len(eligible) > 0 {
		cutoff = eligible[len(eligible)-1].Points
		if res.Size != nil && *res.Size < len(eligible) {
			cutoff = eligible[*res.Size-1].Points
		}
	}

	for i, st := range eligible {
		rank := i + 1
		if res.Size != nil && rank > *res.Size && st.Points != cutoff {
			break
		}
		bt := models.BreakTeam{
			Rank: rank, TeamID: st.TeamID, TeamName: st.TeamName,
			Points: st.Points, SpeakerPoints: st.SpeakerPoints, Margin: st.Margin,
			Bubble: res.Size != nil && st.Points == cutoff,
		}
		if flags, ok := s.teamBreakFlags(st.TeamID); ok {
			bt.IsNovice, bt.IsEsl, bt.IsEfl = flags[0], flags[1], flags[2]
		}
		res.Qualifiers = append(res.Qualifiers, bt)
	}
	res.Cutoff = cutoff
}

func (s *SQLTournamentStore) teamBreakFlags(teamID string) ([3]bool, bool) {
	var f [3]bool
	err := s.db.QueryRow("SELECT COALESCE(is_novice,0), COALESCE(is_esl,0), COALESCE(is_efl,0) FROM teams WHERE id = ?", teamID).Scan(&f[0], &f[1], &f[2])
	return f, err == nil
}

// SaveBreakSnapshot persists the qualifier ranks so generated brackets survive
// later data changes.
func (s *SQLTournamentStore) SaveBreakSnapshot(categoryID string, teams []models.BreakTeam) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM break_qualifiers WHERE category_id = ?", categoryID); err != nil {
		return err
	}
	for _, t := range teams {
		if _, err = tx.Exec("INSERT INTO break_qualifiers (category_id, team_id, rank) VALUES (?, ?, ?)", categoryID, t.TeamID, t.Rank); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetBreakSnapshot returns persisted qualifiers ordered by rank; ok is false when none exist.
func (s *SQLTournamentStore) getBreakSeeds(categoryID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT bq.team_id FROM break_qualifiers bq
		JOIN teams t ON bq.team_id = t.id
		WHERE bq.category_id = ? ORDER BY bq.rank ASC
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seeds []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		seeds = append(seeds, id)
	}
	return seeds, rows.Err()
}

func nextPow2(n int) int {
	p := 1
	for p < n {
		p *= 2
	}
	return p
}

// eliminationName derives the round name from the number of teams it contains.
func eliminationName(teamCount int) string {
	switch teamCount {
	case 2:
		return "Final"
	case 4:
		return "Semifinals"
	case 8:
		return "Quarterfinals"
	case 16:
		return "Octofinals"
	default:
		return fmt.Sprintf("Elimination %d", teamCount)
	}
}

// knockoutSides returns the first two side labels from the tournament sides config.
func (s *SQLTournamentStore) knockoutSides() []string {
	sidesStr, err := s.GetSidesConfig()
	if err != nil || strings.TrimSpace(sidesStr) == "" {
		sidesStr = "OG,OO,CG,CO"
	}
	parts := strings.Split(sidesStr, ",")
	if len(parts) < 2 {
		return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[0])}
	}
	return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
}

// buildKnockoutDebates pairs seeds 1vN, 2v(N-1), ... padding to a power of two;
// missing bottom seeds become single-team bye debates that auto-advance.
func buildKnockoutDebates(seeds []string, sides []string) []models.DebateDrawInput {
	n := nextPow2(len(seeds))
	seedAt := func(pos int) string {
		if pos <= len(seeds) {
			return seeds[pos-1]
		}
		return ""
	}
	debates := make([]models.DebateDrawInput, 0, n/2)
	for pos := 1; pos <= n/2; pos++ {
		a, b := seedAt(pos), seedAt(n+1-pos)
		if a == "" && b == "" {
			continue
		}
		teams := make([]models.TeamAssignment, 0, 2)
		if a != "" {
			teams = append(teams, models.TeamAssignment{TeamID: a, Side: sides[0]})
		}
		if b != "" {
			teams = append(teams, models.TeamAssignment{TeamID: b, Side: sides[1]})
		}
		debates = append(debates, models.DebateDrawInput{
			DebateID:        uuid.New().String(),
			Venue:           fmt.Sprintf("Room %d", pos),
			BracketPosition: pos,
			Teams:           teams,
		})
	}
	return debates
}

// GenerateBracket populates an elimination round from the persisted break snapshot
// for the category, falling back to a live computation when nothing was published.
func (s *SQLTournamentStore) GenerateBracket(roundID, categoryID string) error {
	round, err := s.GetRound(roundID)
	if err != nil {
		return fmt.Errorf("round not found: %w", err)
	}
	if round.Stage != "elimination" {
		return fmt.Errorf("round %q is not an elimination round", round.Name)
	}

	seeds, err := s.getBreakSeeds(categoryID)
	if err != nil {
		return err
	}
	if len(seeds) == 0 {
		res, cerr := s.ComputeBreak(categoryID)
		if cerr != nil {
			return cerr
		}
		for _, q := range res.Qualifiers {
			seeds = append(seeds, q.TeamID)
		}
	}
	if len(seeds) < 2 {
		return fmt.Errorf("not enough qualifiers to generate a bracket")
	}
	return s.SaveDraw(roundID, buildKnockoutDebates(seeds, s.knockoutSides()))
}

type debateScore struct {
	points        float64
	speakerPoints float64
}

// roundDebateScores sums confirmed ballot points per debate and team for one round.
func (s *SQLTournamentStore) roundDebateScores(roundID string) (map[string]map[string]debateScore, error) {
	rows, err := s.db.Query(`
		SELECT d.id, br.team_id, SUM(br.points), SUM(br.speaker_points)
		FROM debates d
		JOIN ballots b ON b.debate_id = d.id AND b.status = 'confirmed'
		JOIN ballot_results br ON br.ballot_id = b.id
		WHERE d.round_id = ?
		GROUP BY d.id, br.team_id
	`, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make(map[string]map[string]debateScore)
	for rows.Next() {
		var debateID, teamID string
		var sc debateScore
		if err := rows.Scan(&debateID, &teamID, &sc.points, &sc.speakerPoints); err != nil {
			return nil, err
		}
		if scores[debateID] == nil {
			scores[debateID] = make(map[string]debateScore)
		}
		scores[debateID][teamID] = sc
	}
	return scores, rows.Err()
}

// pickWinner returns the debate's winning team by confirmed points, breaking ties
// on speaker points; decided is false while results are missing or still tied.
func pickWinner(scores map[string]debateScore, teamIDs []string) (string, bool) {
	if len(teamIDs) == 1 {
		return teamIDs[0], true
	}
	a, okA := scores[teamIDs[0]]
	b, okB := scores[teamIDs[1]]
	if !okA || !okB || a.points == b.points && a.speakerPoints == b.speakerPoints {
		return "", false
	}
	if a.points > b.points || a.points == b.points && a.speakerPoints > b.speakerPoints {
		return teamIDs[0], true
	}
	return teamIDs[1], true
}

// elimDebate is a debate of an elimination round with its teams in seed-slot order
// (first inserted team = higher seed).
type elimDebate struct {
	id      string
	venue   string
	pos     int
	teamIDs []string
}

func (s *SQLTournamentStore) eliminationRoundTeams(roundID string) ([]elimDebate, error) {
	dRows, err := s.db.Query(`
		SELECT id, venue, COALESCE(bracket_position, 0) FROM debates
		WHERE round_id = ?
		ORDER BY bracket_position IS NULL, bracket_position, rowid
	`, roundID)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()

	var debates []elimDebate
	for dRows.Next() {
		var d elimDebate
		if err := dRows.Scan(&d.id, &d.venue, &d.pos); err != nil {
			return nil, err
		}
		debates = append(debates, d)
	}
	if err := dRows.Err(); err != nil {
		return nil, err
	}

	for i := range debates {
		tRows, err := s.db.Query("SELECT team_id FROM debate_teams WHERE debate_id = ? ORDER BY rowid", debates[i].id)
		if err != nil {
			return nil, err
		}
		for tRows.Next() {
			var tid string
			if err := tRows.Scan(&tid); err != nil {
				tRows.Close()
				return nil, err
			}
			debates[i].teamIDs = append(debates[i].teamIDs, tid)
		}
		if err := tRows.Err(); err != nil {
			tRows.Close()
			return nil, err
		}
		tRows.Close()
	}
	return debates, nil
}

// AdvanceEliminationRound verifies every debate has a decided winner, then creates
// the next elimination round seeded by the winners' original break seeds.
func (s *SQLTournamentStore) AdvanceEliminationRound(roundID string) (string, error) {
	round, err := s.GetRound(roundID)
	if err != nil {
		return "", fmt.Errorf("round not found: %w", err)
	}
	if round.Stage != "elimination" {
		return "", fmt.Errorf("round %q is not an elimination round", round.Name)
	}

	debates, err := s.eliminationRoundTeams(roundID)
	if err != nil {
		return "", err
	}
	if len(debates) == 0 {
		return "", fmt.Errorf("round has no debates to advance")
	}
	scores, err := s.roundDebateScores(roundID)
	if err != nil {
		return "", err
	}

	type winner struct {
		teamID string
		seed   int
	}
	winners := make([]winner, 0, len(debates))
	for _, d := range debates {
		w, decided := pickWinner(scores[d.id], d.teamIDs)
		if !decided {
			return "", fmt.Errorf("debate in %q has no confirmed result yet", d.venue)
		}
		var rank sql.NullInt64
		_ = s.db.QueryRow("SELECT rank FROM break_qualifiers WHERE team_id = ? ORDER BY rank ASC LIMIT 1", w).Scan(&rank)
		seed := 9999
		if rank.Valid {
			seed = int(rank.Int64)
		} else {
			seed = d.pos
			if len(d.teamIDs) > 1 && w == d.teamIDs[1] {
				seed = len(debates)*2 + 1 - d.pos
			}
		}
		winners = append(winners, winner{teamID: w, seed: seed})
	}
	sort.Slice(winners, func(i, j int) bool { return winners[i].seed < winners[j].seed })

	nextSeq, err := s.nextRoundSeq()
	if err != nil {
		return "", err
	}
	newID := uuid.New().String()
	if err := s.CreateRound(newID, nextSeq, eliminationName(len(winners)), "elimination"); err != nil {
		return "", err
	}

	ordered := make([]string, len(winners))
	for i, w := range winners {
		ordered[i] = w.teamID
	}
	if err := s.SaveDraw(newID, buildKnockoutDebates(ordered, s.knockoutSides())); err != nil {
		return "", err
	}
	return newID, nil
}

func (s *SQLTournamentStore) nextRoundSeq() (int, error) {
	var maxSeq sql.NullInt64
	if err := s.db.QueryRow("SELECT MAX(seq) FROM rounds").Scan(&maxSeq); err != nil {
		return 0, err
	}
	return int(maxSeq.Int64) + 1, nil
}

// GetBracket returns all elimination rounds with their debates, teams, and
// confirmed winners for the visualizer.
func (s *SQLTournamentStore) GetBracket() ([]models.BracketRound, error) {
	rRows, err := s.db.Query("SELECT id, seq, name FROM rounds WHERE stage = 'elimination' ORDER BY seq ASC")
	if err != nil {
		return nil, err
	}
	defer rRows.Close()

	var rounds []models.BracketRound
	for rRows.Next() {
		var br models.BracketRound
		if err := rRows.Scan(&br.ID, &br.Seq, &br.Name); err != nil {
			return nil, err
		}
		br.Debates = []models.BracketDebate{}
		rounds = append(rounds, br)
	}
	if err := rRows.Err(); err != nil {
		return nil, err
	}

	for i := range rounds {
		if err := s.loadBracketDebates(&rounds[i]); err != nil {
			return nil, err
		}
	}
	return rounds, nil
}

func (s *SQLTournamentStore) loadBracketDebates(br *models.BracketRound) error {
	debates, err := s.eliminationRoundTeams(br.ID)
	if err != nil {
		return err
	}
	scores, err := s.roundDebateScores(br.ID)
	if err != nil {
		return err
	}

	nameRows, err := s.db.Query(`
		SELECT dt.debate_id, dt.team_id, t.name, dt.side
		FROM debate_teams dt
		JOIN teams t ON dt.team_id = t.id
		JOIN debates d ON dt.debate_id = d.id
		WHERE d.round_id = ?
		ORDER BY dt.rowid
	`, br.ID)
	if err != nil {
		return err
	}
	defer nameRows.Close()
	teamsByDebate := make(map[string][]models.TeamAssignment)
	for nameRows.Next() {
		var debateID string
		var ta models.TeamAssignment
		if err := nameRows.Scan(&debateID, &ta.TeamID, &ta.TeamName, &ta.Side); err != nil {
			return err
		}
		teamsByDebate[debateID] = append(teamsByDebate[debateID], ta)
	}
	if err := nameRows.Err(); err != nil {
		return err
	}

	br.Debates = make([]models.BracketDebate, 0, len(debates))
	for _, d := range debates {
		bd := models.BracketDebate{ID: d.id, Venue: d.venue, Bye: len(d.teamIDs) == 1, Teams: teamsByDebate[d.id]}
		if d.pos > 0 {
			pos := d.pos
			bd.BracketPosition = &pos
		}
		if w, decided := pickWinner(scores[d.id], d.teamIDs); decided {
			winner := w
			bd.WinnerTeamID = &winner
		}
		br.Debates = append(br.Debates, bd)
	}
	return nil
}
