package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/google/uuid"
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
	code TEXT,
	is_novice BOOLEAN DEFAULT 0,
	is_esl BOOLEAN DEFAULT 0,
	is_efl BOOLEAN DEFAULT 0,
	is_standby BOOLEAN DEFAULT 0
);

CREATE TABLE IF NOT EXISTS speakers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,
	is_novice BOOLEAN DEFAULT 0,
	is_esl BOOLEAN DEFAULT 0,
	is_efl BOOLEAN DEFAULT 0
);

CREATE TABLE IF NOT EXISTS adjudicators (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	institution_id TEXT REFERENCES institutions(id) ON DELETE SET NULL,
	test_score REAL DEFAULT 0.0,
	rating REAL
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
	venue TEXT NOT NULL,
	bracket_position INTEGER
);

CREATE TABLE IF NOT EXISTS debate_teams (
	id TEXT PRIMARY KEY,
	debate_id TEXT REFERENCES debates(id) ON DELETE CASCADE,
	team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,
	side TEXT NOT NULL, -- e.g., 'OG', 'OO', 'CG', 'CO' or 'Gov', 'Opp'
	pull_up BOOLEAN DEFAULT 0,
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
	status TEXT NOT NULL, -- 'draft', 'submitted', 'confirmed', 'discrepancy'
	is_split BOOLEAN DEFAULT 0,
	entry_group TEXT -- links paired drafts in double-entry mode
);

CREATE TABLE IF NOT EXISTS ballot_results (
	id TEXT PRIMARY KEY,
	ballot_id TEXT REFERENCES ballots(id) ON DELETE CASCADE,
	team_id TEXT REFERENCES teams(id) ON DELETE CASCADE,
	points INTEGER NOT NULL,
	speaker_points REAL NOT NULL,
	adjudicator_id TEXT REFERENCES adjudicators(id) ON DELETE CASCADE, -- NULL = consensus result
	UNIQUE(ballot_id, team_id, adjudicator_id)
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

CREATE TABLE IF NOT EXISTS break_categories (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	seq INTEGER NOT NULL DEFAULT 0,
	size INTEGER,
	base_points INTEGER
);

CREATE TABLE IF NOT EXISTS feedback_questions (
	id TEXT PRIMARY KEY,
	seq INTEGER NOT NULL DEFAULT 0,
	type TEXT NOT NULL CHECK (type IN ('scale','text','checkbox','select')),
	name TEXT NOT NULL,
	options_json TEXT,
	required BOOLEAN DEFAULT 0,
	from_type TEXT NOT NULL,
	to_type TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS feedback_submissions (
	id TEXT PRIMARY KEY,
	round_id TEXT REFERENCES rounds(id) ON DELETE CASCADE,
	debate_id TEXT REFERENCES debates(id) ON DELETE CASCADE,
	source_type TEXT NOT NULL,
	source_id TEXT NOT NULL,
	target_adjudicator_id TEXT REFERENCES adjudicators(id) ON DELETE CASCADE,
	score REAL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(debate_id, source_type, source_id, target_adjudicator_id)
);

CREATE TABLE IF NOT EXISTS feedback_answers (
	submission_id TEXT REFERENCES feedback_submissions(id) ON DELETE CASCADE,
	question_id TEXT REFERENCES feedback_questions(id) ON DELETE CASCADE,
	value TEXT NOT NULL,
	PRIMARY KEY (submission_id, question_id)
);

CREATE TABLE IF NOT EXISTS checkins (
	id TEXT PRIMARY KEY,
	entity_type TEXT NOT NULL CHECK (entity_type IN ('team','adjudicator','venue')),
	entity_id TEXT NOT NULL,
	checked_in_at TIMESTAMP,
	checkin_token TEXT UNIQUE NOT NULL,
	UNIQUE(entity_type, entity_id)
);

CREATE TABLE IF NOT EXISTS round_availability (
	round_id TEXT REFERENCES rounds(id) ON DELETE CASCADE,
	entity_type TEXT NOT NULL CHECK (entity_type IN ('team','adjudicator','venue')),
	entity_id TEXT NOT NULL,
	is_available BOOLEAN NOT NULL DEFAULT 1,
	PRIMARY KEY (round_id, entity_type, entity_id)
);

CREATE TABLE IF NOT EXISTS break_qualifiers (
	category_id TEXT NOT NULL,
	team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
	rank INTEGER NOT NULL,
	PRIMARY KEY (category_id, team_id)
);

CREATE TABLE IF NOT EXISTS conflicts (
	id TEXT PRIMARY KEY,
	subject_type TEXT NOT NULL CHECK (subject_type IN ('adjudicator','team')),
	subject_id TEXT NOT NULL,
	target_type TEXT NOT NULL CHECK (target_type IN ('team','speaker','adjudicator','institution')),
	target_id TEXT NOT NULL,
	weight TEXT NOT NULL DEFAULT 'soft' CHECK (weight IN ('hard','soft')),
	UNIQUE(subject_type, subject_id, target_type, target_id)
);
`

// columnMigrations contains idempotent ALTER TABLE statements for pre-existing
// tournament databases, where CREATE TABLE IF NOT EXISTS cannot add columns.
var columnMigrations = []string{
	"ALTER TABLE teams ADD COLUMN is_novice BOOLEAN DEFAULT 0",
	"ALTER TABLE teams ADD COLUMN is_esl BOOLEAN DEFAULT 0",
	"ALTER TABLE teams ADD COLUMN is_efl BOOLEAN DEFAULT 0",
	"ALTER TABLE teams ADD COLUMN is_standby BOOLEAN DEFAULT 0",
	"ALTER TABLE speakers ADD COLUMN is_novice BOOLEAN DEFAULT 0",
	"ALTER TABLE speakers ADD COLUMN is_esl BOOLEAN DEFAULT 0",
	"ALTER TABLE speakers ADD COLUMN is_efl BOOLEAN DEFAULT 0",
	"ALTER TABLE debate_teams ADD COLUMN pull_up BOOLEAN DEFAULT 0",
	"ALTER TABLE ballots ADD COLUMN is_split BOOLEAN DEFAULT 0",
	"ALTER TABLE ballots ADD COLUMN entry_group TEXT",
	"ALTER TABLE ballot_results ADD COLUMN adjudicator_id TEXT REFERENCES adjudicators(id) ON DELETE CASCADE",
	"ALTER TABLE adjudicators ADD COLUMN rating REAL",
	"ALTER TABLE debates ADD COLUMN bracket_position INTEGER",
}

// applyColumnMigrations runs column migrations, tolerating already-applied ones.
func applyColumnMigrations(db *sql.DB) error {
	for _, stmt := range columnMigrations {
		if _, err := db.Exec(stmt); err != nil {
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return fmt.Errorf("column migration failed: %w", err)
		}
	}
	return nil
}

// InitTournamentDB opens a tournament database, configures the single-writer connection, and initializes its tables.
func InitTournamentDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open tournament db: %w", err)
	}

	// Single-writer discipline to prevent SQLITE_BUSY errors
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure tournament db pragmas: %w", err)
	}

	if _, err := db.Exec(TournamentSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply tournament db schema: %w", err)
	}

	if err := applyColumnMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply column migrations: %w", err)
	}

	if err := seedDefaultFeedbackQuestions(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to seed default feedback questions: %w", err)
	}

	return db, nil
}

// seedDefaultFeedbackQuestions inserts the two default team→adjudicator
// questions when no feedback questions exist yet.
func seedDefaultFeedbackQuestions(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM feedback_questions").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	defaults := []struct {
		seq                   int
		qType, name, fromType string
		required              bool
	}{
		{1, "scale", "Adjudicator rating", "team", true},
		{2, "text", "Comments", "team", false},
	}
	for _, d := range defaults {
		if _, err := tx.Exec(
			"INSERT INTO feedback_questions (id, seq, type, name, required, from_type, to_type) VALUES (?, ?, ?, ?, ?, ?, 'adjudicator')",
			uuid.New().String(), d.seq, d.qType, d.name, d.required, d.fromType,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}
