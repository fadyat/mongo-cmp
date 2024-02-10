package flags

import (
	"fmt"
	"time"
)

type CompareFlags struct {

	// From is a connection string to the source database
	From string

	// To is a connection string to the destination database
	To string

	// Timeout is the timeout for MongoDB operations (default: 15s)
	Timeout time.Duration
}

func NewCompareFlags() *CompareFlags {
	return &CompareFlags{
		Timeout: 15 * time.Second,
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
