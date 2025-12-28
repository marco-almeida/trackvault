/*
Copyright © 2025 Marco Almeida
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/marco-almeida/trackvault/pkg/core"
	"github.com/marco-almeida/trackvault/pkg/music"

	"github.com/spf13/cobra"
)

const providerFlagName = "provider"

var musicProviderClient music.Provider

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to your music provider account",
	Long:    `Login to your music provider account using your credentials.`,
	Aliases: []string{"l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := cmd.PersistentFlags().GetString(providerFlagName)
		if err != nil {
			return fmt.Errorf("error getting provider flag: %v", err)
		}

		loginArgs := core.LoginArgs{
			Provider: provider,
		}

		musicProviderClient, err = core.Login(cmd.Context(), loginArgs)
		if err != nil {
			return fmt.Errorf("error logging in to provider %s: %v", provider, err)
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
