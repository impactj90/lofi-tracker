package tracker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/git"
)

func Init() (Tracker, string, error) {
	dbPath, err := getDBPath()
	if err != nil {
		fmt.Printf("Failed to get database path: %v\n", err)
		return nil, "", err
	}
	dbConn, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return nil, "", err
	}

	branchName, err := git.GetCurrentBranchName()
	if err != nil {
		fmt.Printf("Failed to get current branch name: %v\n", err)
		return nil, "", err
	}

	return NewTracker(branchName, dbConn), branchName, nil
}

func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".lofi-tracker", "lofi-tracker.db"), nil
}
