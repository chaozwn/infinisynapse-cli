package cmd

import (
	"fmt"
	"strings"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
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
  agent_infini init --server https://custom-server.example.com --api-key sk-xxx
  agent_infini init --api-key sk-xxx --prefer-language zh_CN
  agent_infini init --api-key sk-xxx --console https://api.infinisynapse.cn/api/user/profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		preferLang, _ := cmd.Flags().GetString("prefer-language")
		consoleURL, _ := cmd.Flags().GetString("console")

		valid := false
		for _, l := range config.SupportedLanguages {
			if l == preferLang {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported language %q, supported: %s", preferLang, strings.Join(config.SupportedLanguages, ", "))
		}

		userID, err := client.FetchUserID(consoleURL, apiKey)
		if err != nil {
			return fmt.Errorf("failed to fetch user profile: %w", err)
		}

		values := map[string]string{
			config.KeyServer:         server,
			config.KeyAPIKey:         apiKey,
			config.KeyPreferLanguage: preferLang,
			config.KeyConsole:        consoleURL,
			config.KeyUserID:         userID,
		}

		if err := config.Save(values); err != nil {
			return err
		}

		dir, _ := config.ConfigDir()
		output.PrintSuccess("Configuration saved to %s/config.key", dir)
		fmt.Printf("  server:           %s\n", server)
		fmt.Printf("  api-key:          %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
		fmt.Printf("  prefer-language:  %s\n", preferLang)
		fmt.Printf("  console:          %s\n", consoleURL)
		fmt.Printf("  user-id:          %s\n", userID)
		return nil
	},
}

func init() {
	initCmd.Flags().String("server", "https://app.infinisynapse.cn", "Server address")
	initCmd.Flags().String("api-key", "", "API key for authentication (required)")
	initCmd.Flags().String("prefer-language", "zh_CN", fmt.Sprintf("Preferred language (%s)", strings.Join(config.SupportedLanguages, ", ")))
	initCmd.Flags().String("console", config.DefaultConsoleURL, "Console API base URL")
	_ = initCmd.MarkFlagRequired("api-key")

	rootCmd.AddCommand(initCmd)
}
