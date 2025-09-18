package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	meetingBranch string
)

func init() {
	mmCmd.Flags().StringVarP(&meetingBranch, "branch", "b", "meeting", "Branch name to use for meeting mode tracking")
	rootCmd.AddCommand(mmCmd)
}

var mmCmd = &cobra.Command{
	Use:   "mm",
	Short: "Start meeting mode tracking on a dedicated branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := strings.TrimSpace(meetingBranch)
		if branch == "" {
			branch = "meeting"
		}

		tr, _, err := initTracker()
		if err != nil {
			return fmt.Errorf("failed to initialize tracker: %w", err)
		}
		defer tr.Close()

		result, err := tr.ForceStart(branch)
		if err != nil {
			return fmt.Errorf("failed to start meeting mode: %w", err)
		}

		if result.Previous != nil {
			duration := tracker.FormatSummaryDuration(result.Previous.TotalDuration)
			fmt.Printf("✅ Completed session on '%s' after %s\n", result.Previous.Branch, duration)
		}

		fmt.Printf("🤝 Meeting mode active on '%s'\n", branch)
		fmt.Printf("    Started at: %s\n", result.Current.StartedAt.Local().Format(time.RFC3339))

		return nil
	},
}
