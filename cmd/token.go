package cmd

import (
	"errors"
	"fmt"

	"switchtube-downloader/internal/token"

	"github.com/spf13/cobra"
)

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
			fmt.Printf("Error displaying help: %v\n", err)

			return
		}
	},
}

var tokenGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current access token",
	Long:  "Checks if an access token is currently stored in the system keyring and returns it if there is one",
	Run: func(_ *cobra.Command, _ []string) {
		tokenMgr := token.NewTokenManager()

		token, err := tokenMgr.Get()
		if err != nil {
			fmt.Printf("Error getting token: %v\n", err)

			return
		}

		fmt.Printf("Token: %s\n", token)
	},
}

var tokenSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a new access token",
	Long:  "Create and store a new SwitchTube access token in the system keyring",
	Run: func(_ *cobra.Command, _ []string) {
		tokenMgr := token.NewTokenManager()

		if err := tokenMgr.Set(); errors.Is(err, token.ErrTokenAlreadyExists) {
			return
		} else if err != nil {
			fmt.Printf("Error setting token: %v\n", err)

			return
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
			fmt.Printf("Error deleting token: %v\n", err)

			return
		}
	},
}

var tokenValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the current access token",
	Long:  "Checks if an access token is currently stored in the system keyring and validates it if there is one",
	Run: func(_ *cobra.Command, _ []string) {
		tokenMgr := token.NewTokenManager()

		if err := tokenMgr.Validate(); err != nil {
			fmt.Printf("Error validating token: %v\n", err)

			return
		}
	},
}
