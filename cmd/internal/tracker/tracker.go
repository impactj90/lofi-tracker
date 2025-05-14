package tracker

import (
	"errors"
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
	ResumeOrCreateSession(branchName string) (int64, error)
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

func (t *tracker) ResumeOrCreateSession(branchName string) (int64, error) {
	activeSession, err := t.db.GetActiveSession()
	if err != nil && !errors.Is(err, db.ErrNoActiveSession) {
		return 0, err
	}

	if activeSession == nil {
		return 0, db.ErrNoActiveSession
	}

	sessionId, err := t.db.ResumeOrCreateSession(branchName, time.Now())
	if err != nil {
		return 0, err
	}

	return sessionId, nil
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
