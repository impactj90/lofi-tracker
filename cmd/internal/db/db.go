package db

import "time"

type Session struct {
	ID        int64
	Branch    string
	StartTime time.Time
	Endtime   *time.Time
	IsPaused  bool
	IsAfk     bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Pause struct {
	ID         int64
	SessionID  int64
	PauseStart time.Time
	PauseEnd   *time.Time
}

type SummaryData struct {
	Branch       string
	TotalTime    time.Duration
	ActiveTime   time.Duration
	PauseTime    time.Duration
	AfkTime      time.Duration
	SessionCount int
	StartDate    time.Time
	EndDate      time.Time
}

type DB interface {
	CreateSession(branch string, startTime time.Time) (int64, error)
	CompleteSession(sessionID int64, endTime time.Time) error
	GetActiveSession() (*Session, error)
	PauseSession(sessionID int64, pauseStart time.Time, isAfk bool) (int64, error)
	ResumeSession(sessionID int64, pauseEnd time.Time) error
	Close() error

	GetDailySummary(date time.Time) ([]SummaryData, error)
	GetWeeklySummary(startOfWeek time.Time) ([]SummaryData, error)
	GetMonthlySummary(year int, month time.Month) ([]SummaryData, error)
	GetDateRangeSummary(startDate, endDate time.Time) ([]SummaryData, error)
	GetBranchSummary(branch string, days int) (*SummaryData, error)
}
