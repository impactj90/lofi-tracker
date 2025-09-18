package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

func TestMeetingModeUsesForceStart(t *testing.T) {
	originalInit := initTracker
	defer func() { initTracker = originalInit }()

	stub := &exportTrackerStub{
		forceResult: tracker.ForceStartResult{
			Previous: &tracker.SessionStatus{Branch: "feature/main", TotalDuration: 90 * time.Minute},
			Current:  tracker.SessionStatus{Branch: "meeting", StartedAt: time.Date(2024, 10, 30, 15, 4, 5, 0, time.UTC)},
		},
	}

	initTracker = func() (tracker.Tracker, string, error) {
		return stub, "", nil
	}

	meetingBranch = "P2PMER-4"

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer

	err = mmCmd.RunE(mmCmd, nil)
	writer.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatalf("meeting mode command failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	output := buf.String()

	if stub.lastForcedBranch != "P2PMER-4" {
		t.Fatalf("expected ForceStart on 'P2PMER-4', got %q", stub.lastForcedBranch)
	}

	if !strings.Contains(output, "Meeting mode active on 'P2PMER-4'") {
		t.Fatalf("expected meeting mode message in output, got %q", output)
	}

	// reset default for other tests
	meetingBranch = "meeting"
}
