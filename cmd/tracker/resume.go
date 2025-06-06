// defines the resume command
package main

import (
	"fmt"

	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resumeCmd)
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume tracking",
	Run: func(cmd *cobra.Command, args []string) {
		tr, branchName, err := tracker.Init()
		if err != nil {
			fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
			return
		}

		defer tr.Close()

		err = tr.Resume()
		if err != nil {
			fmt.Printf("❌ Failed to resume tracking: %v\n", err)
			return
		}

		fmt.Printf("▶️  Session resumed on branch '%s'\n", branchName)
	},
}
