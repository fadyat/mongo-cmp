package flags

import (
	"fmt"
	"log/slog"
	"slices"
	"time"
)

const (
	DefaultTimeout  = 15 * time.Second
	DefaultDatabase = "__all__"
)

type CompareFlags struct {

	// From is a connection string to the source database
	From string

	// To is a connection string to the destination database
	To string

	// Timeout is the timeout for MongoDB operations (default: 15s)
	Timeout time.Duration

	// Database is the name of the database to compare (default: all databases)
	Database string

	// LogLevel is the log level
	LogLevel string

	// ShowDetails is a flag to show detailed information about the differences
	// Like the number of documents in each collection and size of indexes.
	ShowDetails bool
}

func NewCompareFlags() *CompareFlags {
	return &CompareFlags{
		Timeout:  DefaultTimeout,
		Database: DefaultDatabase,
		LogLevel: "info",
	}
}

func (f *CompareFlags) Validate() error {
	if f.From == "" {
		return fmt.Errorf("from is required")
	}

	if f.To == "" {
		return fmt.Errorf("to is required")
	}

	if f.Timeout <= 0 {
		return fmt.Errorf("timeout should be a positive duration")
	}

	if f.Database == "" {
		return fmt.Errorf("database is required")
	}

	if !slices.Contains([]string{"debug", "info", "warn", "error"}, f.LogLevel) {
		slog.Warn("invalid log level, using info", "level", f.LogLevel)
	}

	return nil
}
