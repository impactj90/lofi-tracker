// defines the complete command
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completeCmd)
}

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Complete tracking",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("complete")
	},
}
