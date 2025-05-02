//go:build linux

package afk

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type linuxIdle struct{}

func (l *linuxIdle) GetIdleTime() (time.Duration, error) {
	// Optional: Check for unsupported Wayland
	if os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		return 0, errors.New(`❌ Wayland is currently unsupported for AFK detection.
Please switch to an X11 session or disable AFK tracking.`)
	}

	// Check if xprintidle is available
	if _, err := exec.LookPath("xprintidle"); err != nil {
		return 0, fmt.Errorf(`❌ "xprintidle" not found.

To enable AFK detection on Linux (X11), install it with:
  sudo apt install xprintidle`)
	}

	// Run the command
	cmd := exec.Command("xprintidle")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to run xprintidle: %w", err)
	}

	// Convert output (ms as string) to int
	outStr := strings.TrimSpace(out.String())
	idleMs, err := strconv.Atoi(outStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse xprintidle output: %w", err)
	}

	return time.Duration(idleMs) * time.Millisecond, nil
}

func init() {
	idleProvider = &linuxIdle{}
}
