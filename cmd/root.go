package cmd

import (
	"fmt"
	"os"

	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	sessionName string
	jsonOutput  bool
	showSkill   bool
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

func printInitHint() {
	if config.IsInitialized() {
		fmt.Fprintf(os.Stderr, "\n[OK] Configuration initialized.\n")
	} else {
		fmt.Fprintf(os.Stderr, "\n[WARNING] Not initialized. Run 'agent_infini init --help' to get started.\n")
	}
}

var rootCmd = &cobra.Command{
	Use:   "agent_infini",
	Short: "InfiniSynapse CLI - command line tool for InfiniSynapse",
	Long: `agent_infini is a CLI tool that allows you to interact with InfiniSynapse backend APIs
from the terminal, designed for both human users and AI agent workflows.

Key Features:
  - Chat with AI with session-based multi-turn support
  - Unified JSON output for pipeline composability

Quick Start:
  agent_infini session current                  # View the current active session
  agent_infini session use "data-analysis"      # Create & activate a session named "data-analysis"
  agent_infini chat "Hello, analyze my data"    # Chat using the current session

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
	rootCmd.PersistentFlags().StringVarP(&sessionName, "session", "s", "",
		"Session alias name for automatic resume")
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
