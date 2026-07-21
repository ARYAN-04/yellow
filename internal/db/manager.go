package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConnectionManager manages pool connections for active tournament SQLite databases.
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]TournamentStore
	baseDir     string
	globalDB    *sql.DB
}

// NewConnectionManager initializes and returns a ConnectionManager instance.
func NewConnectionManager(globalDB *sql.DB, baseDir string) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]TournamentStore),
		baseDir:     baseDir,
		globalDB:    globalDB,
	}
}

// Get retrieves or opens the SQLite database connection for the requested tournament slug.
func (m *ConnectionManager) Get(slug string) (TournamentStore, error) {
	m.mu.RLock()
	dbConn, ok := m.connections[slug]
	m.mu.RUnlock()
	if ok {
		return dbConn, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check inside lock
	if dbConn, ok = m.connections[slug]; ok {
		return dbConn, nil
	}

	// Verify the tournament exists in global DB and fetch archive state
	var dbPath string
	var isArchived bool
	err := m.globalDB.QueryRow("SELECT db_path, is_archived FROM tournaments WHERE slug = ?", slug).Scan(&dbPath, &isArchived)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tournament '%s' not found", slug)
		}
		return nil, fmt.Errorf("failed to query tournament info: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tournament db directory: %w", err)
	}

	var newDB *sql.DB
	if isArchived {
		// Open in read-only mode using SQLite URI parameters
		uri := fmt.Sprintf("file:%s?mode=ro", dbPath)
		newDB, err = sql.Open("sqlite", uri)
		if err != nil {
			return nil, fmt.Errorf("failed to open read-only database: %w", err)
		}
		newDB.SetMaxOpenConns(1)
	} else {
		newDB, err = InitTournamentDB(dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tournament database: %w", err)
		}
	}

	store := &SQLiteStore{SQLTournamentStore: NewSQLTournamentStore(newDB)}
	m.connections[slug] = store
	return store, nil
}

// CloseAll closes all open SQLite database connections in the manager.
func (m *ConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, dbConn := range m.connections {
		_ = dbConn.Close()
	}
	m.connections = make(map[string]TournamentStore)
}
