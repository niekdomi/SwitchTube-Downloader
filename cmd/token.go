package cmd

import (
	"errors"
	"fmt"
	"os"

	"switchtube-downloader/internal/token"

	charm "github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var log = charm.NewWithOptions(os.Stderr, charm.Options{
	ReportTimestamp: false,
	ReportCaller:    false,
})

// init initializes the token command and its subcommands, adding them to the root command.
func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(tokenGetCmd)
	tokenCmd.AddCommand(tokenSetCmd)
	tokenCmd.AddCommand(tokenDeleteCmd)
	tokenCmd.AddCommand(tokenValidateCmd)
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage the SwitchTube access token",
	Long:  "Manage the SwitchTube access token stored in the system keyring",
	Run: func(cmd *cobra.Command, _ []string) {
		if err := cmd.Help(); err != nil {
			log.Error("Error displaying help", "err", err)
		}
	},
}

var tokenGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current access token",
	Long:  "Reads and prints the raw token stored in the system keyring",
	Run: func(_ *cobra.Command, _ []string) {
		tokenMgr := token.NewTokenManager()

		t, err := tokenMgr.GetRaw()
		if err != nil {
			log.Error("Error getting token", "err", err)

			return
		}

		fmt.Println(t)
	},
}

var tokenSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a new access token",
	Long:  "Create and store a new SwitchTube access token in the system keyring",
	Run: func(_ *cobra.Command, _ []string) {
		tokenMgr := token.NewTokenManager()

		if err := tokenMgr.Set(); err != nil && !errors.Is(err, token.ErrTokenAlreadyExists) {
			log.Error("Error setting token", "err", err)
		}
	},
}

var tokenDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete access token from the keyring",
	Long:  "Delete the SwitchTube access token stored the system keyring",
	Run: func(_ *cobra.Command, _ []string) {
		tokenMgr := token.NewTokenManager()

		if err := tokenMgr.Delete(); err != nil {
			log.Error("Error deleting token", "err", err)
		}
	},
}

var tokenValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the current access token",
	Long:  "Checks if an access token is currently stored in the system keyring and validates it",
	Run: func(_ *cobra.Command, _ []string) {
		tokenMgr := token.NewTokenManager()

		if err := tokenMgr.Validate(); err != nil {
			log.Error("Error validating token", "err", err)
		}
	},
}
