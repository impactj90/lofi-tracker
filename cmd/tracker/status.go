// defines the status command
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("status")
	},
}
