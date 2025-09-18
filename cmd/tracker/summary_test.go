package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

type fakeTracker struct {
	live tracker.LiveSummary
}

func (f *fakeTracker) Start(string) error { return nil }
func (f *fakeTracker) ForceStart(string) (tracker.ForceStartResult, error) {
	return tracker.ForceStartResult{}, nil
}
func (f *fakeTracker) Pause(bool) error                                     { return nil }
func (f *fakeTracker) Resume() error                                        { return nil }
func (f *fakeTracker) Status() (tracker.SessionStatus, error)               { return tracker.SessionStatus{}, nil }
func (f *fakeTracker) Complete() (tracker.SessionStatus, error)             { return tracker.SessionStatus{}, nil }
func (f *fakeTracker) Close() error                                         { return nil }
func (f *fakeTracker) GetDailySummary(time.Time) ([]db.SummaryData, error)  { return nil, nil }
func (f *fakeTracker) GetWeeklySummary(time.Time) ([]db.SummaryData, error) { return nil, nil }
func (f *fakeTracker) GetMonthlySummary(int, time.Month) ([]db.SummaryData, error) {
	return nil, nil
}
func (f *fakeTracker) GetBranchSummary(string, int) (*db.SummaryData, error) { return nil, nil }
func (f *fakeTracker) GetDateRangeSummaryWithLive(time.Time, time.Time) (tracker.LiveSummary, error) {
	return f.live, nil
}
func (f *fakeTracker) DailySummaries(time.Time, time.Time) ([]tracker.DaySummary, error) {
	return nil, nil
}

func TestShowPeriodSummaryMarksLiveBranch(t *testing.T) {
	fake := &fakeTracker{
		live: tracker.LiveSummary{
			Summaries: []db.SummaryData{
				{Branch: "feature/live", TotalTime: time.Hour, ActiveTime: 50 * time.Minute, PauseTime: 10 * time.Minute, SessionCount: 1},
			},
			ActiveBranch:   "feature/live",
			IncludesActive: true,
		},
	}

	output := captureOutput(func() {
		showPeriodSummary(fake, "today", "feature/live")
	})

	if !strings.Contains(output, "feature/live * +") {
		t.Fatalf("expected live branch marker in output, got:\n%s", output)
	}

	if !strings.Contains(output, "+ = includes active session time") {
		t.Fatalf("expected live legend in output, got:\n%s", output)
	}
}

func captureOutput(fn func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	old := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = old
	}()

	fn()
	writer.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	return buf.String()
}
