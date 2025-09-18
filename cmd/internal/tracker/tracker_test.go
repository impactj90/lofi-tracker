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

	err := tracker.Pause(false)
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

func TestForceStartCompletesExistingSessionAndStartsNew(t *testing.T) {
	start := time.Now().Add(-90 * time.Minute)
	mock := &mockDB{
		ActiveSession: &db.Session{
			ID:        1,
			Branch:    "feature/old",
			StartTime: start,
			IsPaused:  false,
		},
	}

	tracker := NewTracker("lofi-tracker", mock)

	result, err := tracker.ForceStart("feature/new")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Previous == nil {
		t.Fatal("expected previous session details to be returned")
	}

	if result.Previous.Branch != "feature/old" {
		t.Fatalf("expected previous branch 'feature/old', got %q", result.Previous.Branch)
	}

	if !mock.CompleteSessionCalled {
		t.Errorf("expected CompleteSession to be called")
	}

	if !mock.CreateSessionCalled {
		t.Errorf("expected CreateSession to be called for new session")
	}

	if result.Current.Branch != "feature/new" {
		t.Fatalf("expected new session branch 'feature/new', got %q", result.Current.Branch)
	}
}

func TestForceStartStartsNewSessionWhenNoneActive(t *testing.T) {
	mock := &mockDB{}
	tracker := NewTracker("lofi-tracker", mock)

	result, err := tracker.ForceStart("feature/new")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Previous != nil {
		t.Fatalf("expected no previous session, got %+v", result.Previous)
	}

	if !mock.CreateSessionCalled {
		t.Errorf("expected CreateSession to be called")
	}

	if result.Current.Branch != "feature/new" {
		t.Fatalf("expected new session branch 'feature/new', got %q", result.Current.Branch)
	}
}

func TestGetDateRangeSummaryWithLiveActiveSession(t *testing.T) {
	mock := &mockDB{
		ActiveSession: &db.Session{Branch: "feature/live"},
		Summaries:     []db.SummaryData{{Branch: "feature/live"}},
	}

	tracker := NewTracker("lofi-tracker", mock)

	result, err := tracker.GetDateRangeSummaryWithLive(time.Now(), time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.IncludesActive {
		t.Fatalf("expected IncludesActive to be true")
	}

	if result.ActiveBranch != "feature/live" {
		t.Fatalf("expected ActiveBranch 'feature/live', got %q", result.ActiveBranch)
	}

	if len(result.Summaries) != 1 {
		t.Fatalf("expected summaries to be returned")
	}
}

func TestGetDateRangeSummaryWithLiveNoActiveSession(t *testing.T) {
	mock := &mockDB{
		Summaries: []db.SummaryData{{Branch: "feature/one"}},
	}

	tracker := NewTracker("lofi-tracker", mock)

	result, err := tracker.GetDateRangeSummaryWithLive(time.Now(), time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.IncludesActive {
		t.Fatalf("expected IncludesActive to be false")
	}

	if result.ActiveBranch != "" {
		t.Fatalf("expected empty ActiveBranch, got %q", result.ActiveBranch)
	}

	if len(result.Summaries) != 1 {
		t.Fatalf("expected summaries to be preserved")
	}
}

func TestDailySummariesReturnsPerDayRows(t *testing.T) {
	dayOne := time.Date(2024, 10, 10, 12, 0, 0, 0, time.UTC)
	dayTwo := dayOne.Add(24 * time.Hour)

	mock := &mockDB{
		Daily: map[string][]db.SummaryData{
			"2024-10-10": {
				{Branch: "feature/one", TotalTime: 90 * time.Minute, ActiveTime: time.Hour, PauseTime: 30 * time.Minute},
			},
			"2024-10-11": {
				{Branch: "feature/one", TotalTime: 2 * time.Hour, ActiveTime: 90 * time.Minute, PauseTime: 30 * time.Minute},
				{Branch: "feature/two", TotalTime: 30 * time.Minute, ActiveTime: 25 * time.Minute, PauseTime: 5 * time.Minute},
			},
		},
	}

	tracker := NewTracker("lofi-tracker", mock)

	rows, err := tracker.DailySummaries(dayOne, dayTwo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if !rows[0].Date.Equal(time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)) || rows[0].Branch != "feature/one" {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}

	if rows[2].Branch != "feature/two" {
		t.Fatalf("expected last row for feature/two, got %+v", rows[2])
	}

	if rows[2].Active != 25*time.Minute {
		t.Fatalf("expected active duration 25m, got %v", rows[2].Active)
	}
}
