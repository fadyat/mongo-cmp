package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Version is the version of the application
	Version = "dev"
)

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("mongo-cmp version: %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
