package tracker

import (
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
)

type mockDB struct {
	ActiveSession *db.Session
	Paused        bool
	IsAfk         bool
	Summaries     []db.SummaryData
	Daily         map[string][]db.SummaryData

	CreateSessionCalled   bool
	CompleteSessionCalled bool
	PauseSessionCalled    bool
	ResumeSessionCalled   bool
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
	m.CompleteSessionCalled = true
	if m.ActiveSession != nil && m.ActiveSession.ID == sessionID {
		m.ActiveSession.Endtime = &endTime
		m.ActiveSession = nil
	}
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

func (m *mockDB) GetDailySummary(date time.Time) ([]db.SummaryData, error) {
	if m.Daily == nil {
		return nil, nil
	}

	key := date.Format("2006-01-02")
	if summaries, ok := m.Daily[key]; ok {
		return summaries, nil
	}

	return nil, nil
}

func (m *mockDB) GetWeeklySummary(time.Time) ([]db.SummaryData, error) {
	return nil, nil
}

func (m *mockDB) GetMonthlySummary(int, time.Month) ([]db.SummaryData, error) {
	return nil, nil
}

func (m *mockDB) GetDateRangeSummary(time.Time, time.Time) ([]db.SummaryData, error) {
	return m.Summaries, nil
}

func (m *mockDB) GetBranchSummary(string, int) (*db.SummaryData, error) {
	return nil, nil
}

func (m *mockDB) GetAllSessions(int) ([]db.DebugSession, error) {
	return nil, nil
}

func (m *mockDB) GetSessionCount() (int, error) {
	return 0, nil
}
