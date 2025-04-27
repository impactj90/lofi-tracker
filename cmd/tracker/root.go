package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "lofi-tracker",
    Short: "Track your work time per Git branch",
    Long:  `Lofi Tracker is a CLI tool to help you track working time on Git branches.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
