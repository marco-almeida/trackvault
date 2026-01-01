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

const sourcePathFlagName = "source"

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore music from a backup",
	Long:  `Restore music from a backup. This command will create all the playlists in the logged in account and save/like all the tracks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := cmd.PersistentFlags().GetString(providerFlagName)
		if err != nil {
			return fmt.Errorf("error getting provider flag: %w", err)
		}

		outputPath, err := cmd.PersistentFlags().GetString(sourcePathFlagName)
		if err != nil {
			var perr *pflag.NotExistError
			if !errors.As(err, &perr) {
				return fmt.Errorf("error getting destination path flag: %w", err)
			}
		}

		restoreArgs := core.RestoreArgs{
			Provider:   provider,
			BackupPath: outputPath,
		}

		err = core.Restore(cmd.Context(), restoreArgs)
		if err != nil {
			return fmt.Errorf("error restoring %s data: %w", provider, err)
		}
		return nil
	},
}

func init() {
	restoreCmd.PersistentFlags().StringP(providerFlagName, "p", "", "Selected music provider")
	restoreCmd.PersistentFlags().StringP(sourcePathFlagName, "s", "", "Source path for the backup folder")

	err := restoreCmd.MarkPersistentFlagRequired(providerFlagName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error marking flag required:", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(restoreCmd)
}
