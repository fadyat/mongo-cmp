package cmd

import (
	"github.com/fadyat/mongo-cmp/cmd/flags"
	"github.com/fadyat/mongo-cmp/cmd/log"
	"github.com/fadyat/mongo-cmp/internal"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var rootFlags = flags.NewCompareFlags()

var rootCmd = &cobra.Command{
	Use:          "mongo-cmp",
	SilenceUsage: true,
	Short:        "Run some compare operations on MongoDB",
	Long: `This tool is used to compare two MongoDB databases.
It can be used to compare the data, indexes, and collections between two databases.

The tool requires two connection strings to the source and destination databases.
The connection strings should be in default MongoDB URI format.

For example:
    mongodb://username:password@localhost:27017/mydb

Examples:
    mongo-cmp --from="mongodb://localhost:27017/mydb" --to="mongodb://localhost:27017/mydb2"
	mongo-cmp -f="mongodb://localhost:27017/mydb" -t="mongodb://localhost:27017/mydb2" --timeout=30s
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		slog.SetDefault(log.JsonLogger(rootFlags.LogLevel))
		return rootFlags.Validate()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.Compare(rootFlags)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().
		StringVarP(&rootFlags.From, "from", "f", "", "Connection string to the source database")
	rootCmd.Flags().
		StringVarP(&rootFlags.To, "to", "t", "", "Connection string to the destination database")
	rootCmd.Flags().
		DurationVarP(&rootFlags.Timeout, "timeout", "", rootFlags.Timeout, "Timeout for MongoDB operations")
	rootCmd.Flags().
		StringVarP(&rootFlags.Database, "database", "d", rootFlags.Database, "Name of the database to compare")
	rootCmd.Flags().
		StringVarP(&rootFlags.LogLevel, "log-level", "l", rootFlags.LogLevel, "Log level (debug, info, warn, error)")
	rootCmd.Flags().
		BoolVarP(&rootFlags.ShowDetails, "show-details", "s", rootFlags.ShowDetails, "Show detailed information about the differences")

	_ = rootCmd.MarkFlagRequired("from")
	_ = rootCmd.MarkFlagRequired("to")
}
