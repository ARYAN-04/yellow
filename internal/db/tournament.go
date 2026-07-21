package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// TournamentSchema contains the DDL to bootstrap a per-tournament SQLite database.
const TournamentSchema = `
CREATE TABLE IF NOT EXISTS institutions (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	code TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS teams (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	institution_id TEXT REFERENCES institutions(id) ON DELETE SET NULL,
	code TEXT
);

CREATE TABLE IF NOT EXISTS speakers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	team_id TEXT REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS adjudicators (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	institution_id TEXT REFERENCES institutions(id) ON DELETE SET NULL,
	test_score REAL DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS rounds (
	id TEXT PRIMARY KEY,
	seq INTEGER UNIQUE NOT NULL,
	name TEXT NOT NULL,
	stage TEXT NOT NULL, -- 'preliminary', 'elimination'
	silent BOOLEAN DEFAULT 0,
	draw_released BOOLEAN DEFAULT 0,
	results_released BOOLEAN DEFAULT 0
);

CREATE TABLE IF NOT EXISTS motions (
	id TEXT PRIMARY KEY,
	text TEXT NOT NULL,
	round_id TEXT REFERENCES rounds(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS debates (
	id TEXT PRIMARY KEY,
	round_id TEXT REFERENCES rounds(id) ON DELETE CASCADE,
	venue TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS debate_teams (
	id TEXT PRIMARY KEY,
	debate_id TEXT REFERENCES debates(id) ON DELETE CASCADE,
	team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,
	side TEXT NOT NULL, -- e.g., 'OG', 'OO', 'CG', 'CO' or 'Gov', 'Opp'
	UNIQUE(debate_id, team_id),
	UNIQUE(debate_id, side)
);

CREATE TABLE IF NOT EXISTS debate_adjudicators (
	id TEXT PRIMARY KEY,
	debate_id TEXT REFERENCES debates(id) ON DELETE CASCADE,
	adjudicator_id TEXT REFERENCES adjudicators(id) ON DELETE CASCADE,
	role TEXT NOT NULL, -- 'chair', 'panel', 'trainee'
	UNIQUE(debate_id, adjudicator_id)
);

CREATE TABLE IF NOT EXISTS ballots (
	id TEXT PRIMARY KEY,
	debate_id TEXT REFERENCES debates(id) ON DELETE CASCADE,
	round_id TEXT REFERENCES rounds(id) ON DELETE CASCADE,
	submitter_type TEXT NOT NULL, -- 'adjudicator', 'organizer'
	submitter_id TEXT, -- ID of adjudicator or organizer
	status TEXT NOT NULL -- 'draft', 'submitted', 'confirmed'
);

CREATE TABLE IF NOT EXISTS ballot_results (
	id TEXT PRIMARY KEY,
	ballot_id TEXT REFERENCES ballots(id) ON DELETE CASCADE,
	team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,
	points INTEGER NOT NULL,
	speaker_points REAL NOT NULL,
	UNIQUE(ballot_id, team_id)
);

CREATE TABLE IF NOT EXISTS feedback (
	id TEXT PRIMARY KEY,
	round_id TEXT REFERENCES rounds(id) ON DELETE CASCADE,
	source_adjudicator_id TEXT REFERENCES adjudicators(id) ON DELETE CASCADE,
	target_adjudicator_id TEXT REFERENCES adjudicators(id) ON DELETE CASCADE,
	score REAL NOT NULL,
	comments TEXT
);

CREATE TABLE IF NOT EXISTS access_tokens (
	token TEXT PRIMARY KEY,
	type TEXT NOT NULL, -- 'speaker', 'adjudicator', 'team'
	owner_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
`

// InitTournamentDB opens a tournament database, configures the single-writer connection, and initializes its tables.
func InitTournamentDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open tournament db: %w", err)
	}

	// Single-writer discipline to prevent SQLITE_BUSY errors
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(TournamentSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply tournament db schema: %w", err)
	}

	return db, nil
}
