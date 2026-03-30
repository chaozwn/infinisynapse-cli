package cmd

import (
	"fmt"

	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize CLI configuration",
	Long: `Initialize the agent_infini CLI by providing the server address and API key.

This writes a config file to ~/.agent_infini/config.key that will be used by all
subsequent commands.

Example:
  agent_infini init --api-key sk-xxx
  agent_infini init --server https://custom-server.example.com --api-key sk-xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")

		values := map[string]string{
			config.KeyServer: server,
			config.KeyAPIKey: apiKey,
		}

		if err := config.Save(values); err != nil {
			return err
		}

		dir, _ := config.ConfigDir()
		output.PrintSuccess("Configuration saved to %s/config.key", dir)
		fmt.Printf("  server:  %s\n", server)
		fmt.Printf("  api-key: %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
		return nil
	},
}

func init() {
	initCmd.Flags().String("server", "https://app.infinisynapse.cn", "Server address")
	initCmd.Flags().String("api-key", "", "API key for authentication (required)")
	_ = initCmd.MarkFlagRequired("api-key")

	rootCmd.AddCommand(initCmd)
}
