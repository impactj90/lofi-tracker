// defines the start command
package main

import (
	"fmt"
	"time"

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
		tr, branchName, err := tracker.Init()
		if err != nil {
			fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
			return
		}

		defer tr.Close()

		tr.Start(branchName)
		if err != nil {
			fmt.Printf("❌ Failed to start tracking: %v\n", err)
			return
		}

		fmt.Printf("✅ Started tracking on branch '%s' at %s\n", branchName, time.Now().UTC().Format(time.RFC3339))
	},
}

