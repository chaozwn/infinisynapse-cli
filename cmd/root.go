package cmd

import (
	"fmt"
	"os"

	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	showSkill  bool
)

func getOutputFormat() output.Format {
	if jsonOutput {
		return output.FormatJSON
	}
	if config.GetDefaultOutput() == "table" {
		return output.FormatTable
	}
	return output.FormatJSON
}

var skipConfigCmds = map[string]bool{
	"init":    true,
	"skill":   true,
	"version": true,
	"help":    true,
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

func printInitHint() {
	fmt.Fprintf(os.Stderr, "\n── Checking Environment ──\n")

	config.Init()

	if !config.IsInitialized() {
		fmt.Fprintf(os.Stderr, "  Status   : NOT CONFIGURED\n")
		fmt.Fprintf(os.Stderr, "\n  Run 'agent_infini init --help' to get started.\n\n")
		return
	}

	server := config.GetServer()
	if server == "" {
		server = "(default)"
	}
	token := config.GetToken()
	userID := config.GetUserID()
	if userID == "" {
		userID = "(not set)"
	}

	fmt.Fprintf(os.Stderr, "  Status   : Ready\n")
	fmt.Fprintf(os.Stderr, "  Server   : %s\n", server)
	fmt.Fprintf(os.Stderr, "  API Key  : %s\n", maskToken(token))
	fmt.Fprintf(os.Stderr, "  Console  : %s\n", config.GetConsole())
	fmt.Fprintf(os.Stderr, "  Language : %s\n", config.GetPreferLanguage())
	fmt.Fprintf(os.Stderr, "  User ID  : %s\n", userID)
	fmt.Fprintf(os.Stderr, "\n  All set — you can start working now.\n\n")
}

var rootCmd = &cobra.Command{
	Use:   "agent_infini",
	Short: "InfiniSynapse CLI - command line tool for InfiniSynapse",
	Long: `agent_infini is a CLI tool that allows you to interact with InfiniSynapse backend APIs
from the terminal, designed for both human users and AI agent workflows.

Key Features:
  - Chat with AI with task-based multi-turn support
  - Unified JSON output for pipeline composability

Quick Start:
  agent_infini task new "Analyze my data"
  agent_infini task ask <taskId> "Show trends"
  agent_infini task ls

Use 'agent_infini --skill' or 'agent_infini skill' for detailed command specifications.

For more information about a specific command, use:
  agent_infini [command] --help`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if showSkill || skipConfigCmds[cmd.Name()] {
			return nil
		}
		return config.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if showSkill {
			fmt.Print(skillSpec)
			printInitHint()
			return
		}
		cmd.Help()
	},
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false,
		"Output in JSON format: {success, data, error}")
	rootCmd.Flags().BoolVar(&showSkill, "skill", false,
		"Show the detailed command specification")

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
		printInitHint()
	})
}

func Execute() error {
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--skill" && i == 1 {
			fmt.Print(skillSpec)
			printInitHint()
			return nil
		}
	}
	return rootCmd.Execute()
}
