// This way anytime the db package is imported, the drivers will be registered
// The driver is guaranteed to be registered because it's in a top-level file
package db

import (
	_ "github.com/mattn/go-sqlite3"
)
