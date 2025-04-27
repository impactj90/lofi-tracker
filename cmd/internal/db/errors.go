package db

import "errors"

var (
	ErrNoActiveSession = errors.New("no active session found")
	ErrNoPausedSession = errors.New("no paused session found")
	ErrFailedToCreateDirectoryForDatabase = errors.New("failed to create directory for database")
	ErrFailedToOpenDatabase = errors.New("failed to open database")
	ErrFailedToMigrateDatabase = errors.New("failed to migrate database")
	ErrActiveSessionAlreadyActive = errors.New("⚠️active session is already active")
)
