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
	db    *sql.DB
	clock func() time.Time
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

	sdb := &sqliteDB{
		db: db,
		clock: func() time.Time {
			return time.Now().UTC()
		},
	}
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

func (s *sqliteDB) now() time.Time {
	if s.clock != nil {
		return s.clock()
	}
	return time.Now().UTC()
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
	now := s.now()

	query := `
WITH session_totals AS (
    SELECT 
        s.id,
        s.branch,
        s.start_time,
        COALESCE(s.end_time, ?) AS computed_end,
        (julianday(COALESCE(s.end_time, ?)) - julianday(s.start_time)) * 86400 AS total_seconds
    FROM sessions s
    WHERE s.start_time >= ? AND s.start_time < ?
),
pause_totals AS (
    SELECT 
        s.id,
        SUM(
            CASE
                WHEN p.id IS NULL THEN 0
                WHEN p.pause_end IS NOT NULL THEN (julianday(p.pause_end) - julianday(p.pause_start)) * 86400
                ELSE (julianday(?) - julianday(p.pause_start)) * 86400
            END
        ) AS pause_seconds
    FROM sessions s
    LEFT JOIN pauses p ON s.id = p.session_id
    WHERE s.start_time >= ? AND s.start_time < ?
    GROUP BY s.id
)
SELECT
    st.branch,
    COUNT(st.id) AS session_count,
    MIN(st.start_time) AS earliest_start,
    MAX(st.computed_end) AS latest_end,
    SUM(st.total_seconds) AS total_seconds,
    SUM(COALESCE(pt.pause_seconds, 0)) AS pause_seconds
FROM session_totals st
LEFT JOIN pause_totals pt ON st.id = pt.id
GROUP BY st.branch
ORDER BY st.branch`

	rows, err := s.db.Query(query, now, now, startDate, endDate, now, startDate, endDate)
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

		summary.StartDate, err = parseFlexibleTime(earliestStart)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date '%s': %w", earliestStart, err)
		}

		summary.EndDate, err = parseFlexibleTime(latestEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end date '%s': %w", latestEnd, err)
		}

		// Convert seconds to durations
		summary.TotalTime = time.Duration(totalSeconds) * time.Second
		summary.PauseTime = time.Duration(pauseSeconds) * time.Second
		summary.ActiveTime = summary.TotalTime - summary.PauseTime
		summary.AfkTime = 0

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Add this helper function to handle different time formats
func parseFlexibleTime(timeStr string) (time.Time, error) {
	// Try different time formats that SQLite might return
	formats := []string{
		"2006-01-02 15:04:05.999999999-07:00", // With timezone and nanoseconds
		"2006-01-02 15:04:05.999999999+00:00", // UTC with nanoseconds
		"2006-01-02 15:04:05-07:00",           // With timezone
		"2006-01-02 15:04:05+00:00",           // UTC
		"2006-01-02 15:04:05.999999999",       // No timezone, with nanoseconds
		"2006-01-02 15:04:05",                 // Basic format
		time.RFC3339Nano,                      // Standard RFC3339 with nanoseconds
		time.RFC3339,                          // Standard RFC3339
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}

func (s *sqliteDB) GetBranchSummary(branch string, days int) (*SummaryData, error) {
	endDate := s.now()
	startDate := endDate.AddDate(0, 0, -days)

	summaries, err := s.GetDateRangeSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}

	for _, summary := range summaries {
		if summary.Branch == branch {
			result := summary
			return &result, nil
		}
	}

	return &SummaryData{Branch: branch}, nil
}

func (s *sqliteDB) GetAllSessions(limit int) ([]DebugSession, error) {
	query := `
	SELECT id, branch, start_time, end_time, is_paused, is_afk, created_at, updated_at
	FROM sessions 
	ORDER BY created_at DESC
	LIMIT ?`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []DebugSession
	for rows.Next() {
		var session DebugSession
		var endTime sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.Branch,
			&session.StartTime,
			&endTime,
			&session.IsPaused,
			&session.IsAfk,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.String
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *sqliteDB) GetSessionCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	return count, err
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
