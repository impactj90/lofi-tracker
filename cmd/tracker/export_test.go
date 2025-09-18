package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

type exportTrackerStub struct {
	days             []tracker.DaySummary
	forceResult      tracker.ForceStartResult
	forceError       error
	lastForcedBranch string
}

func (e *exportTrackerStub) Start(string) error { return nil }
func (e *exportTrackerStub) ForceStart(branch string) (tracker.ForceStartResult, error) {
	e.lastForcedBranch = branch
	return e.forceResult, e.forceError
}
func (e *exportTrackerStub) Pause(bool) error { return nil }
func (e *exportTrackerStub) Resume() error    { return nil }
func (e *exportTrackerStub) Status() (tracker.SessionStatus, error) {
	return tracker.SessionStatus{}, nil
}
func (e *exportTrackerStub) Complete() (tracker.SessionStatus, error) {
	return tracker.SessionStatus{}, nil
}
func (e *exportTrackerStub) Close() error                                         { return nil }
func (e *exportTrackerStub) GetDailySummary(time.Time) ([]db.SummaryData, error)  { return nil, nil }
func (e *exportTrackerStub) GetWeeklySummary(time.Time) ([]db.SummaryData, error) { return nil, nil }
func (e *exportTrackerStub) GetMonthlySummary(int, time.Month) ([]db.SummaryData, error) {
	return nil, nil
}
func (e *exportTrackerStub) GetBranchSummary(string, int) (*db.SummaryData, error) { return nil, nil }
func (e *exportTrackerStub) GetDateRangeSummaryWithLive(time.Time, time.Time) (tracker.LiveSummary, error) {
	return tracker.LiveSummary{}, nil
}
func (e *exportTrackerStub) DailySummaries(time.Time, time.Time) ([]tracker.DaySummary, error) {
	return e.days, nil
}

func TestWriteCSVIncludesHeaderAndRows(t *testing.T) {
	rows := []tracker.DaySummary{
		{Date: time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC), Branch: "feature/a", Active: 90 * time.Minute},
		{Date: time.Date(2024, 10, 11, 0, 0, 0, 0, time.UTC), Branch: "feature/b", Active: 45 * time.Minute},
	}

	var buf bytes.Buffer
	if err := writeCSV(rows, &buf); err != nil {
		t.Fatalf("writeCSV returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "date,ticket,minutes") {
		t.Fatalf("expected header, got %q", output)
	}

	if !strings.Contains(output, "2024-10-10,feature/a,90") {
		t.Fatalf("expected first row, got %q", output)
	}

	if !strings.Contains(output, "2024-10-11,feature/b,45") {
		t.Fatalf("expected second row, got %q", output)
	}
}

func TestExportCommandWritesFile(t *testing.T) {
	originalInit := initTracker
	defer func() { initTracker = originalInit }()

	stub := &exportTrackerStub{
		days: []tracker.DaySummary{
			{Date: time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC), Branch: "feature/x", Active: 30 * time.Minute},
		},
	}

	initTracker = func() (tracker.Tracker, string, error) {
		return stub, "", nil
	}

	defer func() {
		exportFrom = ""
		exportTo = ""
		exportOutput = ""
	}()

	exportFrom = "2024-10-12"
	exportTo = "2024-10-12"
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "export.csv")
	exportOutput = path

	if err := exportCmd.RunE(exportCmd, nil); err != nil {
		t.Fatalf("export command failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "2024-10-12,feature/x,30") {
		t.Fatalf("unexpected file content: %q", content)
	}
}

func TestExportCommandWritesToStdout(t *testing.T) {
	originalInit := initTracker
	defer func() { initTracker = originalInit }()

	stub := &exportTrackerStub{
		days: []tracker.DaySummary{{Date: time.Date(2024, 10, 13, 0, 0, 0, 0, time.UTC), Branch: "feature/y", Active: time.Minute}},
	}

	initTracker = func() (tracker.Tracker, string, error) {
		return stub, "", nil
	}

	exportFrom = "2024-10-13"
	exportTo = "2024-10-13"
	exportOutput = ""

	var buf bytes.Buffer
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer

	err = exportCmd.RunE(exportCmd, nil)
	writer.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatalf("export command failed: %v", err)
	}

	_, _ = io.Copy(&buf, reader)
	output := buf.String()
	if !strings.Contains(output, "2024-10-13,feature/y,1") {
		t.Fatalf("expected stdout export, got %q", output)
	}

	exportFrom = ""
	exportTo = ""
}
