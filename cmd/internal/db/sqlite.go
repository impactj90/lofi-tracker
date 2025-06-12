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
	_, err := s.db.Exec(`UPDATE sessions SET end_time = ?, is_paused = 0, is_afk = 0 WHERE id = ?`, endTime, sessionID)
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
	var isAfk bool

	err := s.db.QueryRow(`
		SELECT id, branch, start_time, end_time, is_paused, is_afk
		FROM sessions
		WHERE end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
		`).Scan(&sessionID, &branch, &startTime, &endTime, &isPaused, &isAfk)
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
			IsAfk:     isAfk,
		}, nil
	}

	return &Session{
		ID:        sessionID,
		Branch:    branch,
		StartTime: startTime,
		IsPaused:  isPaused,
		IsAfk:     isAfk,
	}, nil
}

// PauseSession implements DB.
func (s *sqliteDB) PauseSession(sessionID int64, pauseStart time.Time, isAfk bool) (int64, error) {
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

	_, err = s.db.Exec(`UPDATE sessions SET is_paused = 1, is_afk = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, sessionID)
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

	_, err = s.db.Exec(`UPDATE sessions SET is_paused = 0, is_afk = 0 WHERE id = ?`, sessionID)
	if err != nil {
		return err
	}

	return nil
}

// GetDailySummary returns work summary for a specific date
func (s *sqliteDB) GetDailySummary(date time.Time) ([]SummaryData, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	return s.GetDateRangeSummary(startOfDay, endOfDay)
}

// GetWeeklySummary returns work summary for a week starting from startOfWeek
func (s *sqliteDB) GetWeeklySummary(startOfWeek time.Time) ([]SummaryData, error) {
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)
	return s.GetDateRangeSummary(startOfWeek, endOfWeek)
}

// GetMonthlySummary returns work summary for a specific month
func (s *sqliteDB) GetMonthlySummary(year int, month time.Month) ([]SummaryData, error) {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	
	return s.GetDateRangeSummary(startOfMonth, endOfMonth)
}

// GetDateRangeSummary returns work summary for a date range, grouped by branch
func (s *sqliteDB) GetDateRangeSummary(startDate, endDate time.Time) ([]SummaryData, error) {
	query := `
	SELECT 
		s.branch,
		COUNT(s.id) as session_count,
		MIN(s.start_time) as earliest_start,
		MAX(COALESCE(s.end_time, datetime('now'))) as latest_end,
		-- Total session time (from start to end, including pauses)
		SUM(
			CASE 
				WHEN s.end_time IS NOT NULL 
				THEN (julianday(s.end_time) - julianday(s.start_time)) * 86400
				ELSE (julianday('now') - julianday(s.start_time)) * 86400
			END
		) as total_seconds,
		-- Total pause time
		COALESCE(SUM(
			CASE 
				WHEN p.pause_end IS NOT NULL 
				THEN (julianday(p.pause_end) - julianday(p.pause_start)) * 86400
				ELSE 0
			END
		), 0) as pause_seconds
	FROM sessions s
	LEFT JOIN pauses p ON s.id = p.session_id
	WHERE s.start_time >= ? AND s.start_time < ?
	GROUP BY s.branch
	ORDER BY s.branch`
	
	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var summaries []SummaryData
	for rows.Next() {
		var summary SummaryData
		var totalSeconds, pauseSeconds float64
		var earliestStart, latestEnd string
		
		err := rows.Scan(
			&summary.Branch,
			&summary.SessionCount,
			&earliestStart,
			&latestEnd,
			&totalSeconds,
			&pauseSeconds,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse times
		summary.StartDate, _ = time.Parse("2006-01-02 15:04:05", earliestStart)
		summary.EndDate, _ = time.Parse("2006-01-02 15:04:05", latestEnd)
		
		// Convert seconds to durations
		summary.TotalTime = time.Duration(totalSeconds) * time.Second
		summary.PauseTime = time.Duration(pauseSeconds) * time.Second
		summary.ActiveTime = summary.TotalTime - summary.PauseTime
		
		// For now, we'll calculate AFK time separately if needed
		// This would require more complex queries to distinguish AFK vs manual pauses
		summary.AfkTime = 0
		
		summaries = append(summaries, summary)
	}
	
	return summaries, nil
}

// GetBranchSummary returns summary for a specific branch over the last N days
func (s *sqliteDB) GetBranchSummary(branch string, days int) (*SummaryData, error) {
	startDate := time.Now().UTC().AddDate(0, 0, -days)
	endDate := time.Now().UTC()
	
	query := `
	SELECT 
		s.branch,
		COUNT(s.id) as session_count,
		MIN(s.start_time) as earliest_start,
		MAX(COALESCE(s.end_time, datetime('now'))) as latest_end,
		SUM(
			CASE 
				WHEN s.end_time IS NOT NULL 
				THEN (julianday(s.end_time) - julianday(s.start_time)) * 86400
				ELSE (julianday('now') - julianday(s.start_time)) * 86400
			END
		) as total_seconds,
		COALESCE(SUM(
			CASE 
				WHEN p.pause_end IS NOT NULL 
				THEN (julianday(p.pause_end) - julianday(p.pause_start)) * 86400
				ELSE 0
			END
		), 0) as pause_seconds
	FROM sessions s
	LEFT JOIN pauses p ON s.id = p.session_id
	WHERE s.branch = ? AND s.start_time >= ? AND s.start_time < ?
	GROUP BY s.branch`
	
	var summary SummaryData
	var totalSeconds, pauseSeconds float64
	var earliestStart, latestEnd string
	
	err := s.db.QueryRow(query, branch, startDate, endDate).Scan(
		&summary.Branch,
		&summary.SessionCount,
		&earliestStart,
		&latestEnd,
		&totalSeconds,
		&pauseSeconds,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return &SummaryData{Branch: branch}, nil
		}
		return nil, err
	}
	
	// Parse times
	summary.StartDate, _ = time.Parse("2006-01-02 15:04:05", earliestStart)
	summary.EndDate, _ = time.Parse("2006-01-02 15:04:05", latestEnd)
	
	// Convert to durations
	summary.TotalTime = time.Duration(totalSeconds) * time.Second
	summary.PauseTime = time.Duration(pauseSeconds) * time.Second
	summary.ActiveTime = summary.TotalTime - summary.PauseTime
	summary.AfkTime = 0 // Would need additional logic to track AFK specifically
	
	return &summary, nil
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
		is_afk BOOLEAN default 0, 
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
