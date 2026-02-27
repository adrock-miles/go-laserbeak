package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Laserbeak",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Laserbeak %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
