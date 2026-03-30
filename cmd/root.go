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

var rootCmd = &cobra.Command{
	Use:   "isc",
	Short: "InfiniSynapse CLI - command line tool for InfiniSynapse",
	Long: `isc is a CLI tool that allows you to interact with InfiniSynapse backend APIs
from the terminal, designed for both human users and AI agent workflows.

Key Features:
  - Chat with AI (streaming) with session-based multi-turn support
  - Task, database, and settings management
  - Unified JSON output for pipeline composability

Quick Start:
  isc init
  isc chat "Hello, analyze my data" --session main
  isc chat "Show me the trends" --session main

Use 'isc --skill' or 'isc skill' for detailed command specifications.

For more information about a specific command, use:
  isc [command] --help`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "skill" || showSkill {
			return nil
		}
		return config.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if showSkill {
			fmt.Print(skillSpec)
			return
		}
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&sessionName, "session", "s", "",
		"Session alias name for automatic resume")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false,
		"Output in JSON format: {success, data, error}")
	rootCmd.Flags().BoolVar(&showSkill, "skill", false,
		"Show the detailed command specification")
}

func Execute() error {
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--skill" && i == 1 {
			fmt.Print(skillSpec)
			return nil
		}
	}
	return rootCmd.Execute()
}
