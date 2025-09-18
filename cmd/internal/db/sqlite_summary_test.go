package db

import (
	"path/filepath"
	"testing"
	"time"
)

func setupSummaryTestDB(t *testing.T) (*sqliteDB, time.Time) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "summary.db")
	database, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}

	sqlite := database.(*sqliteDB)
	now := time.Date(2024, 10, 10, 12, 0, 0, 0, time.UTC)
	sqlite.clock = func() time.Time {
		return now
	}

	t.Cleanup(func() {
		_ = sqlite.Close()
	})

	return sqlite, now
}

func TestGetDateRangeSummaryIncludesActiveSessions(t *testing.T) {
	sqlite, now := setupSummaryTestDB(t)

	completedStart := now.Add(-2 * time.Hour)
	completedEnd := completedStart.Add(75 * time.Minute)

	_, err := sqlite.db.Exec(`
		INSERT INTO sessions (branch, start_time, end_time, is_paused, is_afk, created_at, updated_at)
		VALUES (?, ?, ?, 0, 0, ?, ?)
	`, "feature/complete", completedStart, completedEnd, completedStart, completedEnd)
	if err != nil {
		t.Fatalf("failed to insert completed session: %v", err)
	}

	activeStart := now.Add(-30 * time.Minute)
	res, err := sqlite.db.Exec(`
		INSERT INTO sessions (branch, start_time, end_time, is_paused, is_afk, created_at, updated_at)
		VALUES (?, ?, NULL, 1, 1, ?, ?)
	`, "feature/active", activeStart, activeStart, activeStart, activeStart)
	if err != nil {
		t.Fatalf("failed to insert active session: %v", err)
	}

	activeID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to grab active session id: %v", err)
	}

	pauseStart := now.Add(-10 * time.Minute)
	_, err = sqlite.db.Exec(`
		INSERT INTO pauses (session_id, pause_start, pause_end)
		VALUES (?, ?, NULL)
	`, activeID, pauseStart)
	if err != nil {
		t.Fatalf("failed to insert pause: %v", err)
	}

	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	summaries, err := sqlite.GetDateRangeSummary(startOfDay, endOfDay)
	if err != nil {
		t.Fatalf("GetDateRangeSummary failed: %v", err)
	}

	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	find := func(branch string) SummaryData {
		for _, summary := range summaries {
			if summary.Branch == branch {
				return summary
			}
		}
		t.Fatalf("missing summary for branch %s", branch)
		return SummaryData{}
	}

	complete := find("feature/complete")
	if complete.SessionCount != 1 {
		t.Fatalf("expected session count 1 for complete branch, got %d", complete.SessionCount)
	}

	expectedComplete := completedEnd.Sub(completedStart)
	if !durationRoughlyEqual(complete.TotalTime, expectedComplete, time.Second) {
		t.Fatalf("expected total %v, got %v", expectedComplete, complete.TotalTime)
	}

	active := find("feature/active")
	if active.SessionCount != 1 {
		t.Fatalf("expected session count 1 for active branch, got %d", active.SessionCount)
	}

	expectedActiveTotal := now.Sub(activeStart)
	if !durationRoughlyEqual(active.TotalTime, expectedActiveTotal, 2*time.Second) {
		t.Fatalf("expected active total %v, got %v", expectedActiveTotal, active.TotalTime)
	}

	expectedPause := now.Sub(pauseStart)
	if !durationRoughlyEqual(active.PauseTime, expectedPause, 2*time.Second) {
		t.Fatalf("expected pause %v, got %v", expectedPause, active.PauseTime)
	}

	expectedActiveTime := expectedActiveTotal - expectedPause
	if !durationRoughlyEqual(active.ActiveTime, expectedActiveTime, 2*time.Second) {
		t.Fatalf("expected active time %v, got %v", expectedActiveTime, active.ActiveTime)
	}
}

func TestGetBranchSummaryIncludesActiveSession(t *testing.T) {
	sqlite, now := setupSummaryTestDB(t)

	activeStart := now.Add(-45 * time.Minute)
	_, err := sqlite.db.Exec(`
		INSERT INTO sessions (branch, start_time, end_time, is_paused, is_afk, created_at, updated_at)
		VALUES (?, ?, NULL, 0, 0, ?, ?)
	`, "feature/live", activeStart, activeStart, activeStart, activeStart)
	if err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}

	summary, err := sqlite.GetBranchSummary("feature/live", 1)
	if err != nil {
		t.Fatalf("GetBranchSummary failed: %v", err)
	}

	if summary.SessionCount != 1 {
		t.Fatalf("expected 1 session, got %d", summary.SessionCount)
	}

	expected := now.Sub(activeStart)
	if !durationRoughlyEqual(summary.TotalTime, expected, 2*time.Second) {
		t.Fatalf("expected total %v, got %v", expected, summary.TotalTime)
	}
}

func durationRoughlyEqual(actual, expected, tolerance time.Duration) bool {
	delta := actual - expected
	if delta < 0 {
		delta = -delta
	}
	return delta <= tolerance
}
