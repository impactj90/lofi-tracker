package tracker

import (
	"errors"
	"fmt"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
)

type Tracker interface {
	Start(branch string) error
	Pause(isAfk bool) error
	Resume() error
	Status() (SessionStatus, error)
	Complete() (SessionStatus, error)
	Close() error

	GetDailySummary(date time.Time) ([]db.SummaryData, error)
	GetWeeklySummary(startOfWeek time.Time) ([]db.SummaryData, error)
	GetMonthlySummary(year int, month time.Month) ([]db.SummaryData, error)
	GetBranchSummary(branch string, days int) (*db.SummaryData, error)
}

type SessionStatus struct {
	Branch        string
	StartedAt     time.Time
	TotalDuration time.Duration
	IsPaused      bool
	IsAfk         bool
}

type tracker struct {
	repoName string
	db       db.DB
}

func NewTracker(repoName string, db db.DB) Tracker {
	return &tracker{
		repoName: repoName,
		db:       db,
	}
}

// Complete implements Tracker.
func (t *tracker) Complete() (SessionStatus, error) {
	activeSession, err := t.db.GetActiveSession()
	if err != nil && !errors.Is(err, db.ErrNoActiveSession) {
		return SessionStatus{}, err
	}

	if activeSession == nil {
		return SessionStatus{}, db.ErrNoActiveSession
	}

	endTime := time.Now().UTC()
	t.db.CompleteSession(activeSession.ID, endTime)
	if err != nil {
		return SessionStatus{}, err
	}

	return SessionStatus{
		Branch:        activeSession.Branch,
		StartedAt:     activeSession.StartTime,
		TotalDuration: endTime.Sub(activeSession.StartTime),
		IsPaused:      false,
		IsAfk:         false,
	}, nil
}

// Pause implements Tracker.
func (t *tracker) Pause(isAfk bool) error {
	activeSession, err := t.db.GetActiveSession()
	if err != nil && !errors.Is(err, db.ErrNoActiveSession) {
		return err
	}

	if activeSession == nil {
		return db.ErrNoActiveSession
	}

	_, err = t.db.PauseSession(activeSession.ID, time.Now().UTC(), isAfk)
	if err != nil {
		return err
	}

	return nil
}

// Resume implements Tracker.
func (t *tracker) Resume() error {
	activeSession, err := t.db.GetActiveSession()
	if err != nil && !errors.Is(err, db.ErrNoActiveSession) {
		return err
	}

	if activeSession == nil {
		return db.ErrNoActiveSession
	}

	err = t.db.ResumeSession(activeSession.ID, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

// Start implements Tracker.
func (t *tracker) Start(branch string) error {
	activeSession, err := t.db.GetActiveSession()
	if err != nil && !errors.Is(err, db.ErrNoActiveSession) {
		return err
	}

	if activeSession != nil {
		return db.ErrActiveSessionAlreadyActive
	}

	_, err = t.db.CreateSession(branch, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

// Status implements Tracker.
func (t *tracker) Status() (SessionStatus, error) {
	activeSession, err := t.db.GetActiveSession()
	if err != nil && !errors.Is(err, db.ErrNoActiveSession) {
		return SessionStatus{}, err
	}

	if activeSession == nil {
		return SessionStatus{}, db.ErrNoActiveSession
	}

	return SessionStatus{
		Branch:        activeSession.Branch,
		StartedAt:     activeSession.StartTime,
		TotalDuration: time.Now().UTC().Sub(activeSession.StartTime),
		IsPaused:      activeSession.IsPaused,
		IsAfk:         activeSession.IsAfk,
	}, nil
}

func (t *tracker) Close() error {
	return t.db.Close()
}

func (t *tracker) GetDailySummary(date time.Time) ([]db.SummaryData, error) {
	return t.db.GetDailySummary(date)
}

func (t *tracker) GetWeeklySummary(startOfWeek time.Time) ([]db.SummaryData, error) {
	return t.db.GetWeeklySummary(startOfWeek)
}

func (t *tracker) GetMonthlySummary(year int, month time.Month) ([]db.SummaryData, error) {
	return t.db.GetMonthlySummary(year, month)
}

func (t *tracker) GetBranchSummary(branch string, days int) (*db.SummaryData, error) {
	return t.db.GetBranchSummary(branch, days)
}

// Helper functions for summary formatting
func FormatSummaryDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func FormatEfficiency(activeTime, totalTime time.Duration) string {
	if totalTime == 0 {
		return "0%"
	}

	efficiency := float64(activeTime) / float64(totalTime) * 100
	return fmt.Sprintf("%.1f%%", efficiency)
}
