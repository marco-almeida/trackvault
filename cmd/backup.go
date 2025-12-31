/*
Copyright © 2025 Marco Almeida
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/marco-almeida/trackvault/pkg/core"
)

const destinationPathFlagName = "dest"

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup all your playlists and liked tracks",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := cmd.PersistentFlags().GetString(providerFlagName)
		if err != nil {
			return fmt.Errorf("error getting provider flag: %w", err)
		}

		destinationPath, err := cmd.PersistentFlags().GetString(destinationPathFlagName)
		if err != nil {
			var perr *pflag.NotExistError
			if !errors.As(err, &perr) {
				return fmt.Errorf("error getting destination path flag: %w", err)
			}
		}

		backupArgs := core.BackupArgs{
			Provider:        provider,
			DestinationPath: destinationPath,
		}

		playlists, err := core.ListPlaylistsAndLikes(cmd.Context(), backupArgs)
		if err != nil {
			return fmt.Errorf("error backing up %s data: %w", provider, err)
		}
		fmt.Println("Playlists ", playlists)
		return nil
	},
}

func init() {
	backupCmd.PersistentFlags().StringP(providerFlagName, "p", "", "Provider to login to")
	backupCmd.PersistentFlags().StringP(destinationPathFlagName, "d", "", "Destination path for the backup")

	err := backupCmd.MarkPersistentFlagRequired(providerFlagName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error marking flag required:", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(backupCmd)
}
