package cmd

import (
	"fmt"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure server and authentication token",
	Long:  `Save server URL and bearer token to local config file for subsequent requests.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server, _ := cmd.Flags().GetString("server")
		token, _ := cmd.Flags().GetString("token")

		if server == "" && token == "" {
			return fmt.Errorf("at least one of --server or --token is required")
		}

		values := make(map[string]string)
		if server != "" {
			values[config.KeyServer] = server
		}
		if token != "" {
			values[config.KeyToken] = token
		}

		if err := config.SetMultiple(values); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		output.PrintSuccess("Login configuration saved to %s", config.ConfigFilePath())
		if server != "" {
			output.PrintSuccess("  Server: %s", server)
		}
		if token != "" {
			output.PrintSuccess("  Token:  %s...%s", token[:4], token[len(token)-4:])
		}
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check connection and authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		_, err = c.Get("/api/ai/ping", nil)
		if err != nil {
			output.PrintError("Connection failed: %v", err)
			return err
		}

		output.PrintSuccess("Connected to %s", c.BaseURL())
		output.PrintSuccess("Authentication: OK")
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear local authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return fmt.Errorf("failed to clear config: %w", err)
		}
		output.PrintSuccess("Credentials cleared from %s", config.ConfigFilePath())
		return nil
	},
}

func init() {
	authLoginCmd.Flags().StringP("server", "s", "", "API server URL (e.g. http://localhost:7001)")
	authLoginCmd.Flags().StringP("token", "t", "", "Bearer token for authentication")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)

	rootCmd.AddCommand(authCmd)
}
