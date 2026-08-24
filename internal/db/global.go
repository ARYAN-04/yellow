package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// GlobalSchema contains the DDL to bootstrap the global server database.
const GlobalSchema = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	username TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	email TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tournaments (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	slug TEXT UNIQUE NOT NULL,
	db_path TEXT NOT NULL,
	is_archived BOOLEAN DEFAULT 0,
	team_count INTEGER DEFAULT 0,
	round_count INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS participation (
	id TEXT PRIMARY KEY,
	user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
	tournament_id TEXT REFERENCES tournaments(id) ON DELETE CASCADE,
	role TEXT NOT NULL,
	UNIQUE(user_id, tournament_id)
);
`

// InitGlobalDB opens the global database and initializes its tables.
func InitGlobalDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open global db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure global db pragmas: %w", err)
	}

	if _, err := db.Exec(GlobalSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply global db schema: %w", err)
	}

	return db, nil
}
