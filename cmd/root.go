package cmd

import (
	"fmt"
	"os"

	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	jsonOutput  bool
	tableOutput bool
	showSkill   bool
)

func getOutputFormat() output.Format {
	if tableOutput {
		return output.FormatTable
	}
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
	Short: "CLI for InfiniSynapse — AI-powered data analysis from the terminal",
	Long: `agent_infini is a CLI tool for interacting with the InfiniSynapse platform,
designed for both human users and AI agent workflows.

Key features:
  - Multi-turn AI task conversations (create, ask, cancel)
  - Database connection management (list, enable, disable)
  - RAG knowledge base management (list, enable, disable)
  - Workspace file operations (list, preview, download)
  - Unified output: JSON (default) or table (--table), ideal for pipelines

Quick start:
  1. Initialize:      agent_infini init --api-key <your-key>
  2. List resources:  agent_infini db ls              (list all databases)
                      agent_infini rag ls             (list all RAG knowledge bases)
  3. Check context:   agent_infini task context       (verify target databases and RAGs are enabled)
     If needed:       agent_infini db enable <id>     (enable a database)
                      agent_infini rag enable <id>    (enable a RAG knowledge base)
  4. Create task:     agent_infini task new "Analyze my data"
  5. Follow up:       agent_infini task ask <taskId> "Show trends"
  6. Browse tasks:    agent_infini task ls

Available command groups:
  init      Configure server address, API key, and preferences
  task      Create and manage multi-turn AI task conversations
  db        List, enable, or disable database connections
  rag       List, enable, or disable RAG knowledge bases
  skill     Show detailed command specification for AI agents
  version   Print version and build information

Use 'agent_infini --skill' or 'agent_infini skill' for the full AI agent specification.
Use 'agent_infini [command] --help' for details on a specific command.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if showSkill || skipConfigCmds[cmd.Name()] {
			return nil
		}
		if err := config.Init(); err != nil {
			return err
		}
		applyFlagOverrides(cmd)
		return nil
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

// injectableKeys lists the config keys that the WinClaw executor may inject as CLI flags.
var injectableKeys = []string{
	config.KeyAPIKey,
	config.KeyServer,
	config.KeyConsole,
	config.KeyPreferLanguage,
	config.KeyOutput,
}

func applyFlagOverrides(cmd *cobra.Command) {
	for _, key := range injectableKeys {
		if cmd.Flags().Changed(key) {
			if val, _ := cmd.Flags().GetString(key); val != "" {
				config.Set(key, val)
			}
		}
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false,
		"Force JSON output: {success, data, error} (default)")
	rootCmd.PersistentFlags().BoolVar(&tableOutput, "table", false,
		"Force table output for list commands")
	rootCmd.Flags().BoolVar(&showSkill, "skill", false,
		"Show the full AI agent command specification")

	pf := rootCmd.PersistentFlags()
	pf.String("api-key", "", "API key (auto-injected by WinClaw executor)")
	pf.String("server", "", "Server address override")
	pf.String("console", "", "Console API base URL override")
	pf.String("prefer-language", "", "Preferred language override")
	pf.String("default-output", "", "Default output format override (json|table)")

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
