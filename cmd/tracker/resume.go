// defines the resume command
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resumeCmd)
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume tracking",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("resume tracking...")
	},
}
