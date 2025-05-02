package tracker

import (
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
)

type mockDB struct {
	ActiveSession *db.Session
	Paused        bool
	IsAfk         bool

	CreateSessionCalled bool
	PauseSessionCalled  bool
	ResumeSessionCalled bool
}

func (m *mockDB) CreateSession(branch string, startTime time.Time) (int64, error) {
	m.CreateSessionCalled = true
	m.ActiveSession = &db.Session{
		ID:        1,
		Branch:    branch,
		StartTime: startTime,
		IsPaused:  false,
	}
	return 1, nil
}

func (m *mockDB) CompleteSession(sessionID int64, endTime time.Time) error {
	return nil
}

func (m *mockDB) GetActiveSession() (*db.Session, error) {
	if m.ActiveSession == nil {
		return nil, db.ErrNoActiveSession
	}
	return m.ActiveSession, nil
}

func (m *mockDB) PauseSession(sessionID int64, pauseStart time.Time, isAfk bool) (int64, error) {
	m.Paused = true
	m.PauseSessionCalled = true
	m.ActiveSession.IsPaused = true
	return 1, nil
}

func (m *mockDB) ResumeSession(sessionID int64, pauseEnd time.Time) error {
	m.Paused = false
	m.ResumeSessionCalled = true
	m.ActiveSession.IsPaused = false
	return nil
}

func (m *mockDB) Close() error {
	return nil
}
