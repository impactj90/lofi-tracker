package db

import "time"

type Session struct {
	ID        int64
	Branch    string
	StartTime time.Time
	Endtime   *time.Time
	IsPaused  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Pause struct {
	ID        int64
	SessionID int64
	PauseStart time.Time
	PauseEnd *time.Time
}

type DB interface {
    CreateSession(branch string, startTime time.Time) (int64, error)
    CompleteSession(sessionID int64, endTime time.Time) error
    GetActiveSession() (*Session, error)
    PauseSession(sessionID int64, pauseStart time.Time) (int64, error)
    ResumeSession(sessionID int64, pauseEnd time.Time) error
	Close() error
}
