package tracker

import (
	"testing"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
)

func TestStart_WhenNoActiveSession_ShouldCreateNewSession(t *testing.T) {
	mock := &mockDB{}

	tracker := NewTracker("lofi-tracker", mock)

	err := tracker.Start("feature/awesome")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !mock.CreateSessionCalled {
		t.Errorf("expected CreateSession to be called, but it was not")
	}

	if mock.ActiveSession == nil || mock.ActiveSession.Branch != "feature/awesome" {
		t.Errorf("expected session branch to be 'feature/awesome', got %+v", mock.ActiveSession)
	}
}

func TestPause_WhenActiveSessionExists_ShouldPauseSession(t *testing.T) {
	mock := &mockDB{
		ActiveSession: &db.Session{
			ID:        1,
			Branch:    "feature/test",
			StartTime: time.Now(),
			IsPaused:  false,
		},
	}

	tracker := NewTracker("lofi-tracker", mock)

	err := tracker.Pause()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !mock.PauseSessionCalled {
		t.Errorf("expected PauseSession to be called, but it was not")
	}

	if !mock.ActiveSession.IsPaused {
		t.Errorf("expected session to be paused, but IsPaused is false")
	}
}

func TestResume_WhenSessionIsPaused_ShouldResumeSession(t *testing.T) {
	mock := &mockDB{
		ActiveSession: &db.Session{
			ID:        1,
			Branch:    "feature/test",
			StartTime: time.Now(),
			IsPaused:  true,
		},
		Paused: true,
	}

	tracker := NewTracker("lofi-tracker", mock)

	err := tracker.Resume()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !mock.ResumeSessionCalled {
		t.Errorf("expected ResumeSession to be called, but it was not")
	}

	if mock.ActiveSession.IsPaused {
		t.Errorf("expected session to be resumed (IsPaused=false), but IsPaused is still true")
	}
}

func TestComplete_WhenActiveSessionExists_ShouldCompleteSession(t *testing.T) {
	mock := &mockDB{
		ActiveSession: &db.Session{
			ID:        1,
			Branch:    "feature/test",
			StartTime: time.Now().Add(-2 * time.Hour), // Assume started 2 hours ago
			IsPaused:  false,
		},
	}

	tracker := NewTracker("lofi-tracker", mock)

	status, err := tracker.Complete()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if status.Branch != "feature/test" {
		t.Errorf("expected branch 'feature/test', got %s", status.Branch)
	}

	if status.TotalDuration < 2*time.Hour {
		t.Errorf("expected at least 2 hours of work, got %v", status.TotalDuration)
	}

	if status.IsPaused {
		t.Errorf("expected session to be not paused after completion")
	}
}
