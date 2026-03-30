package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultInitServer = "http://app.infinisynapse.cn"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup: save server URL and API key to ~/.isc.yaml",
	Long: `Prompts for API server address and API key, then writes ~/.isc.yaml.

Default server when you press Enter: ` + defaultInitServer + `

You can also pass --server and/or --api-key to skip prompts.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	serverFlag, _ := cmd.Flags().GetString("server")
	apiKeyFlag, _ := cmd.Flags().GetString("api-key")

	server := strings.TrimSpace(serverFlag)
	token := strings.TrimSpace(apiKeyFlag)

	if server == "" {
		var err error
		server, err = promptServer()
		if err != nil {
			return err
		}
	}
	if server == "" {
		server = defaultInitServer
	}
	server = strings.TrimRight(server, "/")

	if token == "" {
		var err error
		token, err = promptAPIKey()
		if err != nil {
			return err
		}
	}
	token = strings.TrimPrefix(strings.TrimSpace(token), "Bearer ")
	if token == "" {
		return fmt.Errorf("API key is required")
	}

	values := map[string]string{
		config.KeyServer: server,
		config.KeyToken:  token,
	}
	if err := config.SetMultiple(values); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	output.PrintSuccess("Configuration saved to %s", config.ConfigFilePath())
	output.PrintSuccess("  Server: %s", server)
	output.PrintSuccess("  API key: %s...%s", token[:min(4, len(token))], token[max(0, len(token)-4):])
	return nil
}

func promptServer() (string, error) {
	fmt.Fprintf(os.Stderr, "Server address [%s]: ", defaultInitServer)
	line, err := readStdinLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptAPIKey() (string, error) {
	fmt.Fprint(os.Stderr, "API key: ")
	line, err := readStdinLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readStdinLine() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("no input")
	}
	return scanner.Text(), nil
}

func init() {
	initCmd.Flags().String("server", "", fmt.Sprintf("API server URL (default when omitted in prompt: %s)", defaultInitServer))
	initCmd.Flags().String("api-key", "", "API key / bearer token (skips prompt when set)")

	rootCmd.AddCommand(initCmd)
}
