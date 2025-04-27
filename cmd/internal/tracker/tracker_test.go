package tracker

import (
    "testing"
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
