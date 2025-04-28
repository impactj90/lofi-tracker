// defines the pause command
package main

import (
	"fmt"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/git"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pauseCmd)
}

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause tracking",
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

		tracker := tracker.NewTracker(branchName, dbConn)
		tracker.Pause()
		if err != nil {
			fmt.Printf("❌ Failed to pause tracking: %v\n", err)
			return
		}

		fmt.Printf("⏸️  Session paused on branch '%s'\n", branchName)
	},
}
