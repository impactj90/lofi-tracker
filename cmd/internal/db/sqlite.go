package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type sqliteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, ErrFailedToCreateDirectoryForDatabase
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, ErrFailedToOpenDatabase
	}

	sdb := &sqliteDB{db: db}
	if err := sdb.migrate(); err != nil {
		return nil, ErrFailedToMigrateDatabase
	}

	return sdb, nil
}

// CompleteSession implements DB.
func (s *sqliteDB) CompleteSession(sessionID int64, endTime time.Time) error {
	panic("unimplemented")
}

// CreateSession implements DB.
func (s *sqliteDB) CreateSession(branch string, startTime time.Time) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO sessions (branch, start_time, is_paused, created_at, updated_at)
		VALUES (?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, branch, startTime)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetActiveSession implements DB.
func (s *sqliteDB) GetActiveSession() (*Session, error) {
	panic("unimplemented")
}

// PauseSession implements DB.
func (s *sqliteDB) PauseSession(sessionID int64, pauseStart time.Time) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO pauses (session_id, pause_start, pause_end, created_at, updated_at)
		VALUES (?, ?, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, sessionID, pauseStart)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = s.db.Exec(`UPDATE sessions SET is_paused = 1 WHERE id = ?`, sessionID)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// ResumeSession implements DB.
func (s *sqliteDB) ResumeSession(sessionID int64, pauseEnd time.Time) error {
	var pauseID int64

	err := s.db.QueryRow(`
		SELECT id FROM pauses 
		WHERE session_id = ? AND pause_end IS NULL
		ORDER BY pause_start DESC
		LIMIT 1
		`, sessionID).Scan(&pauseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoPausedSession
		}
		return err
	}

	_, err = s.db.Exec(`UPDATE pauses SET pause_end = ? WHERE id = ?`, pauseEnd, pauseID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE sessions SET is_paused = 0 WHERE id = ?`, pauseID)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqliteDB) migrate() error {
	schema := `
    CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        branch TEXT NOT NULL,
        start_time TIMESTAMP NOT NULL,
        end_time TIMESTAMP,
        is_paused BOOLEAN DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS pauses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id INTEGER NOT NULL,
        pause_start TIMESTAMP NOT NULL,
        pause_end TIMESTAMP,
        FOREIGN KEY(session_id) REFERENCES sessions(id)
    );
    `
	_, err := s.db.Exec(schema)
	return err
}
