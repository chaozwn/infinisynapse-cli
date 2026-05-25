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

This writes a config file to ~/.agent_infini/config.txt that will be used by all
subsequent commands.

agent_infini uses two base URLs:
  --server   Business API base URL used by task, db, and rag commands.
  --console  Console API base URL used by init to call /user/profile.

For local development, pass both --server and --console.

Examples:
  agent_infini init --api-key sk-xxx
  agent_infini init --server https://custom-server.example.com --api-key sk-xxx
  agent_infini init --api-key sk-xxx --server http://localhost:8088 --console http://localhost:3000/api
  agent_infini init --api-key sk-xxx --prefer-language zh_CN
  agent_infini init --api-key sk-xxx --server http://app.infinisynapse.cn --console https://api.infinisynapse.cn/api`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = false

		server, _ := cmd.Flags().GetString("server")
		apiKey, _ := cmd.Flags().GetString("api-key")
		preferLang, _ := cmd.Flags().GetString("prefer-language")
		consoleURL, _ := cmd.Flags().GetString("console")

		if server == "" {
			server = "https://app.infinisynapse.cn"
		}
		if apiKey == "" {
			return fmt.Errorf("required flag \"api-key\" not set")
		}
		if preferLang == "" {
			preferLang = "zh_CN"
		}
		if consoleURL == "" {
			consoleURL = config.DefaultConsoleURL
		}

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

		cmd.SilenceUsage = true

		userID, err := client.FetchUserID(consoleURL, apiKey)
		if err != nil {
			return fmt.Errorf(`failed to fetch user profile with console URL %s: %w

Hint:
  --server is used for task, database, and RAG APIs.
  --console is used by init to call /user/profile.
  For local development, pass both URLs, for example:
    agent_infini init --api-key sk-xxx --server http://localhost:8088 --console http://localhost:3000/api`, consoleURL, err)
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
		output.PrintSuccess("Configuration saved to %s/config.txt", dir)
		fmt.Printf("  server:           %s\n", server)
		fmt.Printf("  api-key:          %s\n", maskToken(apiKey))
		fmt.Printf("  prefer-language:  %s\n", preferLang)
		fmt.Printf("  console:          %s\n", consoleURL)
		fmt.Printf("  user-id:          %s\n", userID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
