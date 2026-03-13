package cmd

import (
	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagServer string
	flagToken  string
	flagTable  bool
)

func getOutputFormat() output.Format {
	if flagTable {
		return output.FormatTable
	}
	if config.GetDefaultOutput() == "table" {
		return output.FormatTable
	}
	return output.FormatJSON
}

var rootCmd = &cobra.Command{
	Use:   "isc",
	Short: "InfiniSynapse CLI - command line tool for InfiniSynapse",
	Long:  `isc is a CLI tool that allows you to interact with InfiniSynapse backend APIs from the terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.Init()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagServer, "server", "s", "", "API server URL (overrides config)")
	rootCmd.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "Bearer token (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&flagTable, "table", false, "Output in table format (default: JSON)")
}

func Execute() error {
	return rootCmd.Execute()
}
