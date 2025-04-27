// defines the pause command
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pauseCmd)
}

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause tracking",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pause tracking...")
	},
}
