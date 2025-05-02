package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
		fmt.Printf("❌ Failed to open database: %v\n", err)
		return nil, ErrFailedToOpenDatabase
	}

	sdb := &sqliteDB{db: db}
	if err := sdb.migrate(); err != nil {
		fmt.Printf("❌ Failed to migrate database: %v\n", err)
		return nil, ErrFailedToMigrateDatabase
	}

	return sdb, nil
}

// CompleteSession implements DB.
func (s *sqliteDB) CompleteSession(sessionID int64, endTime time.Time) error {
	fmt.Printf("CompleteSession(%d, %v)\n", sessionID, endTime)
	_, err := s.db.Exec(`UPDATE sessions SET end_time = ?, is_paused = 0 WHERE id = ?`, endTime, sessionID)
	if err != nil {
		return err
	}

	return nil
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
	var sessionID int64
	var branch string
	var startTime time.Time
	var endTime *time.Time
	var isPaused bool

	err := s.db.QueryRow(`
		SELECT id, branch, start_time, end_time, is_paused
		FROM sessions
		WHERE end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
		`).Scan(&sessionID, &branch, &startTime, &endTime, &isPaused)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActiveSession
		}
		return nil, err
	}

	if endTime != nil {
		return &Session{
			ID:        sessionID,
			Branch:    branch,
			StartTime: startTime,
			Endtime:   endTime,
			IsPaused:  isPaused,
		}, nil
	}

	return &Session{
		ID:        sessionID,
		Branch:    branch,
		StartTime: startTime,
		IsPaused:  isPaused,
	}, nil
}

// PauseSession implements DB.
func (s *sqliteDB) PauseSession(sessionID int64, pauseStart time.Time) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO pauses (pause_start, pause_end, session_id)
		VALUES (?, NULL, ?)
		`, pauseStart, sessionID)
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

	_, err = s.db.Exec(`UPDATE sessions SET is_paused = 0 WHERE id = ?`, sessionID)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqliteDB) Close() error {
	return s.db.Close()
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
