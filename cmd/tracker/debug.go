package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Show debug information about sessions",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database directly for debug access
		dbPath, err := getDBPath()
		if err != nil {
			fmt.Printf("❌ Failed to get database path: %v\n", err)
			return
		}

		database, err := db.NewSQLiteDB(dbPath)
		if err != nil {
			fmt.Printf("❌ Failed to open database: %v\n", err)
			return
		}
		defer database.Close()

		fmt.Printf("🔍 Debug Information\n")
		fmt.Printf("Database path: %s\n\n", dbPath)

		// Get session count
		count, err := database.GetSessionCount()
		if err != nil {
			fmt.Printf("❌ Failed to get session count: %v\n", err)
			return
		}
		fmt.Printf("Total sessions in database: %d\n\n", count)

		// Get recent sessions
		sessions, err := database.GetAllSessions(10)
		if err != nil {
			fmt.Printf("❌ Failed to get sessions: %v\n", err)
			return
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found in database.")
			return
		}

		fmt.Println("📋 Recent Sessions:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tBranch\tStart Time\tEnd Time\tPaused\tAFK")
		fmt.Fprintln(w, "--\t------\t----------\t--------\t------\t---")

		for _, session := range sessions {
			endTime := "Active"
			if session.EndTime != nil {
				endTime = *session.EndTime
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%v\t%v\n",
				session.ID,
				session.Branch,
				session.StartTime,
				endTime,
				session.IsPaused,
				session.IsAfk,
			)
		}
		w.Flush()
	},
}

func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".lofi-tracker", "lofi-tracker.db"), nil
}
