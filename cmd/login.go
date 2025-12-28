/*
Copyright © 2025 Marco Almeida
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/marco-almeida/trackvault/pkg/core"
)

const providerFlagName = "provider"

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to your music provider account",
	Long:    `Login to your music provider account using your credentials.`,
	Aliases: []string{"l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := cmd.PersistentFlags().GetString(providerFlagName)
		if err != nil {
			return fmt.Errorf("error getting provider flag: %w", err)
		}

		loginArgs := core.LoginArgs{
			Provider: provider,
		}

		_, err = core.Login(cmd.Context(), loginArgs)
		if err != nil {
			return fmt.Errorf("error logging in to provider %s: %w", provider, err)
		}
		return nil
	},
}

func init() {
	loginCmd.PersistentFlags().StringP(providerFlagName, "p", "", "Provider to login to")
	err := loginCmd.MarkPersistentFlagRequired(providerFlagName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error marking flag required:", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(loginCmd)
}
