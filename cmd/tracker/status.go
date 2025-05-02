// defines the status command
package main

import (
	"fmt"

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
		tr, branchName, err := tracker.Init()
		if err != nil {
			fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
			return
		}

		defer tr.Close()

		status, err := tr.Status()
		if err != nil {
			fmt.Printf("❌ Failed to get status: %v\n", err)
			return
		}

		fmt.Printf("🕒 Total work time: %s on branch '%s'\n", tracker.FormatDuration(status.TotalDuration), branchName)
		if status.IsPaused {
			fmt.Printf("⏸️  Session paused on branch '%s'\n", branchName)
			return
		}

	},
}
