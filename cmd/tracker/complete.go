// defines the complete command
package main

import (
	"fmt"

	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completeCmd)
}

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Complete tracking",
	Run: func(cmd *cobra.Command, args []string) {
		tr, _, err := tracker.Init()
		if err != nil {
			fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
			return
		}

		defer tr.Close()

		status, err := tr.Complete()
		if err != nil {
			fmt.Printf("❌ Failed to complete tracking: %v\n", err)
			return
		}

		fmt.Printf("✅ Completed session on branch '%s'\n", status.Branch)
		fmt.Printf("🕒 Total work time: %s\n", tracker.FormatDuration(status.TotalDuration))
	},
}
