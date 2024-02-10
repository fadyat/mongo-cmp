package flags

import (
	"fmt"
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

	return nil
}
