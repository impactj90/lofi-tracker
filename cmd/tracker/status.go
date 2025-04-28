// defines the status command
package main

import (
	"fmt"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/git"
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
		dbPath := getDBPath()
		dbConn, err := db.NewSQLiteDB(dbPath)
		if err != nil {
			fmt.Printf("Failed to open database: %v\n", err)
			return
		}

		branchName, err := git.GetCurrentBranchName()
		if err != nil {
			fmt.Printf("Failed to get current branch name: %v\n", err)
			return
		}

		t := tracker.NewTracker(branchName, dbConn)
		status, err := t.Status()
		if err != nil {
			fmt.Printf("❌ Failed to get status: %v\n", err)
			return
		}

		fmt.Printf("🕒 Total work time: %s\n", tracker.FormatDuration(status.TotalDuration))
		if status.IsPaused {
			fmt.Printf("⏸️  Session paused on branch '%s'\n", branchName)
			return
		}

	},
}
