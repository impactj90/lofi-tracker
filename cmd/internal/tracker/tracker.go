package tracker

import (
	"errors"
	"fmt"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
)

type Tracker interface {
	Start(branch string) error
	Pause() error
	Resume() error
	Status() (SessionStatus, error)
	Complete() (SessionStatus, error)
}

type SessionStatus struct {
	Branch        string
	StartedAt     time.Time
	TotalDuration time.Duration
	IsPaused      bool
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
	err = t.db.CompleteSession(activeSession.ID, endTime)
	if err != nil {
		return SessionStatus{}, err
	}

	updatedSession, err := t.db.GetActiveSession()
	if err != nil {
		return SessionStatus{}, err
	}

	fmt.Printf("✅ Completed session on branch '%s'\n", updatedSession.Branch)
	fmt.Printf("🕒 Total work time: %s\n", FormatDuration(endTime.Sub(updatedSession.StartTime)))

	return SessionStatus{
		Branch:        updatedSession.Branch,
		StartedAt:     updatedSession.StartTime,
		TotalDuration: endTime.Sub(updatedSession.StartTime),
		IsPaused:      updatedSession.IsPaused,
	}, nil
}

// Pause implements Tracker.
func (t *tracker) Pause() error {
	activeSession, err := t.db.GetActiveSession()
	if err != nil && !errors.Is(err, db.ErrNoActiveSession) {
		return err
	}

	if activeSession == nil {
		return db.ErrNoActiveSession
	}

	_, err = t.db.PauseSession(activeSession.ID, time.Now().UTC())
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
	panic("unimplemented")
}
