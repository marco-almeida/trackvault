/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trackvault",
	Short: "Trackvault is a CLI tool to manage your music tracks.",
	Long:  `Trackvault is a CLI tool to manage your music tracks. You can use it to login to different music providers, backup your liked tracks and playlists and restore them into a new account.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {}
