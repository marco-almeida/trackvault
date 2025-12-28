/*
Copyright © 2025 Marco Almeida
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information and quit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
