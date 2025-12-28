/*
Copyright © 2025 Marco Almeida
*/
package cmd

import (
	"fmt"

	"runtime/debug"

	"github.com/spf13/cobra"
)

func vcsVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return version
}

var (
	version = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information and quit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", vcsVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
