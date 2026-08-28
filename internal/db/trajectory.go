package db

import (
	"database/sql"
	"errors"

	"yellow/internal/models"
)

// GetTeamTrajectory compiles a round-by-round debate timeline and performance breakdown for a team.
func (s *SQLTournamentStore) GetTeamTrajectory(teamID string, isAdmin bool) (*models.TeamTrajectory, error) {
	// 1. Fetch team metadata
	var team models.Team
	var instID, instName, instCode *string
	err := s.db.QueryRow(`
		SELECT t.id, t.name, COALESCE(t.code, ''), t.institution_id, i.name, i.code,
		       t.is_novice, t.is_esl, t.is_efl, t.is_standby
		FROM teams t
		LEFT JOIN institutions i ON t.institution_id = i.id
		WHERE t.id = ?
	`, teamID).Scan(
		&team.ID, &team.Name, &team.Code, &instID, &instName, &instCode,
		&team.IsNovice, &team.IsEsl, &team.IsEfl, &team.IsStandby,
	)
	if err != nil {
		return nil, err
	}
	team.InstitutionID = instID
	team.InstitutionName = instName
	team.InstitutionCode = instCode

	// Load team speakers
	spRows, err := s.db.Query(`
		SELECT id, name, is_novice, is_esl, is_efl
		FROM speakers
		WHERE team_id = ?
		ORDER BY name
	`, teamID)
	if err == nil {
		for spRows.Next() {
			var sp models.Speaker
			if err := spRows.Scan(&sp.ID, &sp.Name, &sp.IsNovice, &sp.IsEsl, &sp.IsEfl); err == nil {
				team.Speakers = append(team.Speakers, sp)
			}
		}
		spRows.Close()
	}

	// 2. Query round debates for this team - scan into slice first to release cursor
	debateRows, err := s.db.Query(`
		SELECT d.id, d.venue, dt.side, dt.pull_up,
		       r.id, r.seq, r.name, r.stage, r.silent, r.draw_released, r.results_released
		FROM debate_teams dt
		JOIN debates d ON dt.debate_id = d.id
		JOIN rounds r ON d.round_id = r.id
		WHERE dt.team_id = ?
		ORDER BY r.seq ASC
	`, teamID)
	if err != nil {
		return nil, err
	}

	type rawDebate struct {
		debateID, venue, side                 string
		pullUp                                bool
		roundID, roundName, roundStage        string
		roundSeq                              int
		silent, drawReleased, resultsReleased bool
	}
	var rawList []rawDebate
	for debateRows.Next() {
		var r rawDebate
		if err := debateRows.Scan(
			&r.debateID, &r.venue, &r.side, &r.pullUp,
			&r.roundID, &r.roundSeq, &r.roundName, &r.roundStage,
			&r.silent, &r.drawReleased, &r.resultsReleased,
		); err == nil {
			rawList = append(rawList, r)
		}
	}
	debateRows.Close()

	debates := make([]models.TeamTrajectoryDebate, 0, len(rawList))
	for _, raw := range rawList {
		// If draw is not released and requester is not admin, hide this debate
		if !isAdmin && !raw.drawReleased {
			continue
		}

		entry := models.TeamTrajectoryDebate{
			RoundID:         raw.roundID,
			RoundSeq:        raw.roundSeq,
			RoundName:       raw.roundName,
			RoundStage:      raw.roundStage,
			Silent:          raw.silent,
			DrawReleased:    raw.drawReleased,
			ResultsReleased: raw.resultsReleased,
			DebateID:        raw.debateID,
			Venue:           raw.venue,
			Side:            raw.side,
			PullUp:          raw.pullUp,
			Opponents:       make([]models.TeamTrajectoryOpponent, 0),
			Adjudicators:    make([]models.TeamTrajectoryAdjudicator, 0),
			SpeakerScores:   make([]models.SpeakerScoreEntry, 0),
		}

		// Fetch opponents in this debate
		oppRows, err := s.db.Query(`
			SELECT dt.team_id, t.name, dt.side
			FROM debate_teams dt
			JOIN teams t ON dt.team_id = t.id
			WHERE dt.debate_id = ? AND dt.team_id != ?
			ORDER BY dt.side
		`, entry.DebateID, teamID)
		if err == nil {
			for oppRows.Next() {
				var opp models.TeamTrajectoryOpponent
				if err := oppRows.Scan(&opp.TeamID, &opp.TeamName, &opp.Side); err == nil {
					entry.Opponents = append(entry.Opponents, opp)
				}
			}
			oppRows.Close()
		}

		// Fetch adjudicators for this debate
		adjRows, err := s.db.Query(`
			SELECT da.adjudicator_id, a.name, da.role
			FROM debate_adjudicators da
			JOIN adjudicators a ON da.adjudicator_id = a.id
			WHERE da.debate_id = ?
			ORDER BY CASE da.role WHEN 'chair' THEN 1 WHEN 'panel' THEN 2 ELSE 3 END, a.name
		`, entry.DebateID)
		if err == nil {
			for adjRows.Next() {
				var adj models.TeamTrajectoryAdjudicator
				if err := adjRows.Scan(&adj.ID, &adj.Name, &adj.Role); err == nil {
					entry.Adjudicators = append(entry.Adjudicators, adj)
				}
			}
			adjRows.Close()
		}

		// Results visibility check
		resultsVisible := isAdmin || (entry.ResultsReleased && !entry.Silent)
		if resultsVisible {
			var ballotID, status string
			var pts int
			var spPts float64
			err := s.db.QueryRow(`
				SELECT b.id, b.status, br.points, br.speaker_points
				FROM ballots b
				JOIN ballot_results br ON b.id = br.ballot_id
				WHERE b.debate_id = ? AND br.team_id = ?
				  AND (b.status = 'confirmed' OR (? AND b.status IN ('submitted', 'draft')))
				ORDER BY CASE b.status WHEN 'confirmed' THEN 1 WHEN 'submitted' THEN 2 ELSE 3 END
				LIMIT 1
			`, entry.DebateID, teamID, isAdmin).Scan(&ballotID, &status, &pts, &spPts)

			if err == nil {
				entry.BallotStatus = &status
				entry.Points = &pts
				entry.SpeakerPoints = &spPts

				// Fetch individual speaker scores if recorded
				scRows, err := s.db.Query(`
					SELECT ss.speaker_id, s.name, ss.score, ss.is_reply, ss.speech_order, COALESCE(ss.role, '')
					FROM speaker_scores ss
					JOIN speakers s ON ss.speaker_id = s.id
					WHERE ss.ballot_id = ? AND ss.team_id = ?
					ORDER BY ss.speech_order, ss.is_reply
				`, ballotID, teamID)
				if err == nil {
					for scRows.Next() {
						var sc models.SpeakerScoreEntry
						if err := scRows.Scan(&sc.SpeakerID, &sc.SpeakerName, &sc.Score, &sc.IsReply, &sc.SpeechOrder, &sc.Role); err == nil {
							entry.SpeakerScores = append(entry.SpeakerScores, sc)
						}
					}
					scRows.Close()
				}
			}
		}

		debates = append(debates, entry)
	}

	return &models.TeamTrajectory{
		Team:    team,
		Debates: debates,
	}, nil
}

// GetSpeakerTrajectory compiles a round-by-round speech history and score progression for an individual speaker.
func (s *SQLTournamentStore) GetSpeakerTrajectory(speakerID string, isAdmin bool) (*models.SpeakerTrajectory, error) {
	// 1. Fetch speaker metadata and team info
	var sp models.Speaker
	var teamID, teamName string
	err := s.db.QueryRow(`
		SELECT s.id, s.name, s.is_novice, s.is_esl, s.is_efl, t.id, t.name
		FROM speakers s
		JOIN teams t ON s.team_id = t.id
		WHERE s.id = ?
	`, speakerID).Scan(&sp.ID, &sp.Name, &sp.IsNovice, &sp.IsEsl, &sp.IsEfl, &teamID, &teamName)
	if err != nil {
		return nil, err
	}

	// 2. Query team's debates - scan into slice first to release cursor
	debateRows, err := s.db.Query(`
		SELECT d.id, d.venue, dt.side,
		       r.id, r.seq, r.name, r.stage, r.silent, r.draw_released, r.results_released
		FROM debate_teams dt
		JOIN debates d ON dt.debate_id = d.id
		JOIN rounds r ON d.round_id = r.id
		WHERE dt.team_id = ?
		ORDER BY r.seq ASC
	`, teamID)
	if err != nil {
		return nil, err
	}

	type rawDebate struct {
		debateID, venue, side                 string
		roundID, roundName, roundStage        string
		roundSeq                              int
		silent, drawReleased, resultsReleased bool
	}
	var rawList []rawDebate
	for debateRows.Next() {
		var r rawDebate
		if err := debateRows.Scan(
			&r.debateID, &r.venue, &r.side,
			&r.roundID, &r.roundSeq, &r.roundName, &r.roundStage,
			&r.silent, &r.drawReleased, &r.resultsReleased,
		); err == nil {
			rawList = append(rawList, r)
		}
	}
	debateRows.Close()

	speeches := make([]models.SpeakerTrajectorySpeech, 0, len(rawList))
	for _, raw := range rawList {
		if !isAdmin && !raw.drawReleased {
			continue
		}

		resultsVisible := isAdmin || (raw.resultsReleased && !raw.silent)
		var ballotID, status string
		var teamPts int
		hasBallot := false

		if resultsVisible {
			err := s.db.QueryRow(`
				SELECT b.id, b.status, br.points
				FROM ballots b
				JOIN ballot_results br ON b.id = br.ballot_id
				WHERE b.debate_id = ? AND br.team_id = ?
				  AND (b.status = 'confirmed' OR (? AND b.status IN ('submitted', 'draft')))
				ORDER BY CASE b.status WHEN 'confirmed' THEN 1 WHEN 'submitted' THEN 2 ELSE 3 END
				LIMIT 1
			`, raw.debateID, teamID, isAdmin).Scan(&ballotID, &status, &teamPts)
			if err == nil {
				hasBallot = true
			}
		}

		if hasBallot {
			// Query speaker scores for this speaker in this ballot
			scRows, err := s.db.Query(`
				SELECT score, is_reply, speech_order, COALESCE(role, '')
				FROM speaker_scores
				WHERE ballot_id = ? AND speaker_id = ?
				ORDER BY speech_order, is_reply
			`, ballotID, speakerID)

			speechCount := 0
			if err == nil {
				for scRows.Next() {
					var score float64
					var isReply bool
					var speechOrder int
					var role string
					if err := scRows.Scan(&score, &isReply, &speechOrder, &role); err == nil {
						speechCount++
						st := status
						tp := teamPts
						scVal := score
						speeches = append(speeches, models.SpeakerTrajectorySpeech{
							RoundID:         raw.roundID,
							RoundSeq:        raw.roundSeq,
							RoundName:       raw.roundName,
							RoundStage:      raw.roundStage,
							Silent:          raw.silent,
							DrawReleased:    raw.drawReleased,
							ResultsReleased: raw.resultsReleased,
							DebateID:        raw.debateID,
							Venue:           raw.venue,
							Side:            raw.side,
							SpeechOrder:     speechOrder,
							IsReply:         isReply,
							Role:            role,
							Score:           &scVal,
							TeamPoints:      &tp,
							BallotStatus:    &st,
						})
					}
				}
				scRows.Close()
			}

			// If no individual speaker scores were found, record a general debate participation entry
			if speechCount == 0 {
				st := status
				tp := teamPts
				speeches = append(speeches, models.SpeakerTrajectorySpeech{
					RoundID:         raw.roundID,
					RoundSeq:        raw.roundSeq,
					RoundName:       raw.roundName,
					RoundStage:      raw.roundStage,
					Silent:          raw.silent,
					DrawReleased:    raw.drawReleased,
					ResultsReleased: raw.resultsReleased,
					DebateID:        raw.debateID,
					Venue:           raw.venue,
					Side:            raw.side,
					SpeechOrder:     1,
					IsReply:         false,
					Score:           nil,
					TeamPoints:      &tp,
					BallotStatus:    &st,
				})
			}
		} else {
			// Round draw released, no ballot/results yet
			speeches = append(speeches, models.SpeakerTrajectorySpeech{
				RoundID:         raw.roundID,
				RoundSeq:        raw.roundSeq,
				RoundName:       raw.roundName,
				RoundStage:      raw.roundStage,
				Silent:          raw.silent,
				DrawReleased:    raw.drawReleased,
				ResultsReleased: raw.resultsReleased,
				DebateID:        raw.debateID,
				Venue:           raw.venue,
				Side:            raw.side,
				SpeechOrder:     1,
				IsReply:         false,
				Score:           nil,
				TeamPoints:      nil,
				BallotStatus:    nil,
			})
		}
	}

	return &models.SpeakerTrajectory{
		Speaker:  sp,
		TeamID:   teamID,
		TeamName: teamName,
		Speeches: speeches,
	}, nil
}

// GetAdjudicatorTrajectory compiles a round-by-round judging history for an individual adjudicator.
func (s *SQLTournamentStore) GetAdjudicatorTrajectory(adjID string, isAdmin bool) (*models.AdjudicatorTrajectory, error) {
	var adj models.Adjudicator
	var instID, instName, instCode sql.NullString
	var rating sql.NullFloat64
	err := s.db.QueryRow(`
		SELECT a.id, a.name, a.institution_id, i.name, i.code, a.test_score, a.rating
		FROM adjudicators a
		LEFT JOIN institutions i ON a.institution_id = i.id
		WHERE a.id = ?
	`, adjID).Scan(&adj.ID, &adj.Name, &instID, &instName, &instCode, &adj.TestScore, &rating)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("adjudicator not found")
	}
	if err != nil {
		return nil, err
	}
	if instID.Valid && instID.String != "" {
		adj.InstitutionID = &instID.String
	}
	if instName.Valid && instName.String != "" {
		adj.InstitutionName = &instName.String
	}
	if instCode.Valid && instCode.String != "" {
		adj.InstitutionCode = &instCode.String
	}

	rows, err := s.db.Query(`
		SELECT r.id, r.seq, r.name, r.stage, r.silent, r.draw_released, r.results_released,
		       d.id, d.venue, da.role, b.status
		FROM debate_adjudicators da
		JOIN debates d ON da.debate_id = d.id
		JOIN rounds r ON d.round_id = r.id
		LEFT JOIN ballots b ON b.debate_id = d.id AND b.status = 'confirmed'
		WHERE da.adjudicator_id = ?
		ORDER BY r.seq ASC, d.id ASC
	`, adjID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rawDebate struct {
		roundID, roundName, roundStage        string
		roundSeq                              int
		silent, drawReleased, resultsReleased bool
		debateID, venue, role                 string
		ballotStatus                          sql.NullString
	}
	var rawList []rawDebate
	for rows.Next() {
		var rd rawDebate
		if err := rows.Scan(&rd.roundID, &rd.roundSeq, &rd.roundName, &rd.roundStage,
			&rd.silent, &rd.drawReleased, &rd.resultsReleased,
			&rd.debateID, &rd.venue, &rd.role, &rd.ballotStatus); err != nil {
			return nil, err
		}
		rawList = append(rawList, rd)
	}

	debates := make([]models.AdjudicatorTrajectoryDebate, 0, len(rawList))
	for _, raw := range rawList {
		if !isAdmin && (raw.silent || !raw.drawReleased) {
			continue
		}

		entry := models.AdjudicatorTrajectoryDebate{
			RoundID:         raw.roundID,
			RoundSeq:        raw.roundSeq,
			RoundName:       raw.roundName,
			RoundStage:      raw.roundStage,
			Silent:          raw.silent,
			DrawReleased:    raw.drawReleased,
			ResultsReleased: raw.resultsReleased,
			DebateID:        raw.debateID,
			Venue:           raw.venue,
			Role:            raw.role,
			Teams:           make([]models.TeamTrajectoryOpponent, 0),
			CoAdjudicators:  make([]models.TeamTrajectoryAdjudicator, 0),
		}
		if raw.ballotStatus.Valid && (isAdmin || raw.resultsReleased || !raw.silent) {
			entry.BallotStatus = &raw.ballotStatus.String
		}

		// Teams
		tRows, err := s.db.Query(`
			SELECT t.id, t.name, dt.side
			FROM debate_teams dt
			JOIN teams t ON dt.team_id = t.id
			WHERE dt.debate_id = ?
			ORDER BY dt.side ASC
		`, raw.debateID)
		if err == nil {
			for tRows.Next() {
				var tOpp models.TeamTrajectoryOpponent
				if err := tRows.Scan(&tOpp.TeamID, &tOpp.TeamName, &tOpp.Side); err == nil {
					entry.Teams = append(entry.Teams, tOpp)
				}
			}
			tRows.Close()
		}

		// Co-adjudicators
		aRows, err := s.db.Query(`
			SELECT a.id, a.name, da.role
			FROM debate_adjudicators da
			JOIN adjudicators a ON da.adjudicator_id = a.id
			WHERE da.debate_id = ? AND da.adjudicator_id != ?
			ORDER BY da.role ASC, a.name ASC
		`, raw.debateID, adjID)
		if err == nil {
			for aRows.Next() {
				var aAdj models.TeamTrajectoryAdjudicator
				if err := aRows.Scan(&aAdj.ID, &aAdj.Name, &aAdj.Role); err == nil {
					entry.CoAdjudicators = append(entry.CoAdjudicators, aAdj)
				}
			}
			aRows.Close()
		}

		debates = append(debates, entry)
	}

	return &models.AdjudicatorTrajectory{
		Adjudicator: adj,
		Debates:     debates,
	}, nil
}
