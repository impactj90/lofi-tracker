// defines the start command
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/git"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tracking",
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
		tracker.Start(branchName)
		if err != nil {
			fmt.Printf("❌ Failed to start tracking: %v\n", err)
			return
		}

		fmt.Printf("✅ Started tracking on branch '%s' at %s\n", branchName, time.Now().UTC().Format(time.RFC3339))
	},
}

func getDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Could not get home directory")
	}

	return filepath.Join(homeDir, ".lofi-tracker", "db.sqlite")
}
