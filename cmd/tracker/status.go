// defines the status command
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status",
	Run: func(cmd *cobra.Command, args []string) {
		tr, _, err := tracker.Init()
		if err != nil {
			fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
			return
		}

		defer tr.Close()

		status, err := tr.Status()
		if err != nil {
			if errors.Is(err, db.ErrNoActiveSession) {
				fmt.Println("ℹ️  No active session.")
				return
			}
			fmt.Printf("❌ Failed to get status: %v\n", err)
			return
		}

		elapsed := tracker.FormatSummaryDuration(status.TotalDuration)
		started := status.StartedAt.Local().Format(time.RFC3339)

		if status.IsPaused {
			fmt.Printf("⏸️  Session paused on '%s' (elapsed %s)\n", status.Branch, elapsed)
		} else {
			fmt.Printf("🕒 Tracking '%s' (elapsed %s)\n", status.Branch, elapsed)
		}
		fmt.Printf("    Started at: %s\n", started)

		today := time.Now().UTC()
		startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := startOfDay.Add(24 * time.Hour)

		liveSummary, err := tr.GetDateRangeSummaryWithLive(startOfDay, endOfDay)
		if err != nil {
			fmt.Printf("❌ Failed to load daily totals: %v\n", err)
			return
		}

		var branchSummary *db.SummaryData
		for _, summary := range liveSummary.Summaries {
			if summary.Branch == status.Branch {
				copy := summary
				branchSummary = &copy
				break
			}
		}

		if branchSummary == nil {
			fmt.Println("📊 Today: no tracked time yet.")
			return
		}

		marker := ""
		if liveSummary.IncludesActive && liveSummary.ActiveBranch == status.Branch {
			marker = " (live)"
		}

		total := tracker.FormatSummaryDuration(branchSummary.TotalTime)
		active := tracker.FormatSummaryDuration(branchSummary.ActiveTime)
		paused := tracker.FormatSummaryDuration(branchSummary.PauseTime)
		fmt.Printf("📊 Today%s: %s total / %s active / %s paused\n", marker, total, active, paused)
	},
}
