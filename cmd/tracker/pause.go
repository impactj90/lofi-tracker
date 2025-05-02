// defines the pause command
package main

import (
	"fmt"
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
		tr, branchName, err := tracker.Init()
		if err != nil {
			fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
			return
		}

		defer tr.Close()

		err = tr.Pause()
		if err != nil {
			fmt.Printf("❌ Failed to pause tracking: %v\n", err)
			return
		}

		fmt.Printf("⏸️  Session paused on branch '%s'\n", branchName)
	},
}
