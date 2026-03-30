package cmd

import (
	"fmt"
	"strings"

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
  agent_infini init --api-key sk-xxx --prefer-language zh_CN`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		preferLang, _ := cmd.Flags().GetString("prefer-language")

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

		values := map[string]string{
			config.KeyServer:         server,
			config.KeyAPIKey:         apiKey,
			config.KeyPreferLanguage: preferLang,
		}

		if err := config.Save(values); err != nil {
			return err
		}

		dir, _ := config.ConfigDir()
		output.PrintSuccess("Configuration saved to %s/config.key", dir)
		fmt.Printf("  server:           %s\n", server)
		fmt.Printf("  api-key:          %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
		fmt.Printf("  prefer-language:  %s\n", preferLang)
		return nil
	},
}

func init() {
	initCmd.Flags().String("server", "https://app.infinisynapse.cn", "Server address")
	initCmd.Flags().String("api-key", "", "API key for authentication (required)")
	initCmd.Flags().String("prefer-language", "zh_CN", fmt.Sprintf("Preferred language (%s)", strings.Join(config.SupportedLanguages, ", ")))
	_ = initCmd.MarkFlagRequired("api-key")

	rootCmd.AddCommand(initCmd)
}
