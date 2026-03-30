package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const skillSpec = `# agent_infini (InfiniSynapse CLI) Skill Specification

A CLI tool for interacting with the InfiniSynapse backend, designed for AI agent workflows.

================================================================================
AI Agent Important Notes
================================================================================

First-time Setup:
  Before using any other commands, run init to configure:
  agent_infini init --api-key "your_api_key"

  Or manually create ~/.agent_infini/config.key:
  mkdir -p ~/.agent_infini && cat > ~/.agent_infini/config.key << 'EOF'
  global:
    server: "https://app.infinisynapse.cn"
    api-key: "your_api_key"
  EOF

Session Workflow (multi-turn chat):
  1. Create a session and start a new task:
     agent_infini session new --name "analysis" --query "Analyze my data"
  2. Continue the conversation (respond to server ask):
     agent_infini session -s "analysis" --query "Show me the trends"
  3. Manage sessions:
     agent_infini session ls              # List all sessions
     agent_infini session show analysis   # View session details
     agent_infini session rm analysis     # Delete a session

================================================================================
Available Commands
================================================================================

  version                                          Print version, commit, build date, Go runtime
  session new --name <name> --query <message>      Create session and start a new task (newTask)
  session -s <name> --query <message>              Continue conversation in session (askResponse)
  session ls                                       List all sessions with status
  session show <name>                              Show session details
  session rm <name>                                Delete a session
  session state -s <name>                          Get AI state for a session
  session cancel -s <name>                         Cancel a running task in a session

================================================================================
Global Flags
================================================================================

  --session, -s <name>   Session alias for automatic resume across chat calls
  --json                 Force JSON output: {"success": true, "data": ...}
  --skill                Show this detailed specification
  --version, -v          Print version string
  --help, -h             Show help for any command

================================================================================
Common Scenarios
================================================================================

1. Start a new analysis session:
   agent_infini session new --name "analysis" --query "What tables are in my database?"

2. Multi-turn analysis:
   agent_infini session new --name "analysis" --query "Analyze the users table schema"
   agent_infini session -s "analysis" -q "Now show me the top 10 users by activity"
   agent_infini session -s "analysis" -q "Generate a summary report"

3. Check session state:
   agent_infini session ls
   agent_infini session show "analysis"

4. Cancel a running task:
   agent_infini session cancel -s "analysis"

================================================================================
Output Format
================================================================================

JSON mode (--json or default-output: "json" in config.key):
  {"success": true, "data": { ... }}
  {"success": false, "error": "error message"}

Table mode (default-output: "table" in config.key):
  Formatted table for list commands, JSON for detail commands.

================================================================================
Error Handling
================================================================================

  - Token expired:     Update api-key in ~/.agent_infini/config.key
  - Server unreachable: Check --server URL and network connectivity
  - Session not found:  A new conversation will be started automatically

================================================================================
Configuration & Credential Chain
================================================================================

Session files: ~/.agent_infini/sessions/<name>.json

Configuration is loaded from the first file found in this order
(per execute_external_tool_resolver.py):

  1. <binary_dir>/agent_infini.key          (tool_basename.key, YAML)
  2. <binary_dir>/<filename>.key            (tool_filename.key, compat)
  3. ~/.agent_infini/config.key             (YAML, recommended)
  4. ~/.agent_infini/config.json            (JSON)

The most common approach: run 'agent_infini init' or create ~/.agent_infini/ and place config.key or config.json.
config.key and config.json are alternatives; if config.key exists, config.json
is not checked.

config.key format (YAML):
  global:
    server: "https://app.infinisynapse.cn"
    api-key: "your_api_key"
    default-output: "json"
    lang: "zh-CN"

config.json format (JSON):
  {
    "global": {
      "server": "https://app.infinisynapse.cn",
      "api-key": "your_api_key"
    }
  }

WinClaw Marketplace:
  The executor (execute_external_tool_resolver.py) reads config.key and injects
  credentials as CLI flags automatically. Help commands (--help, --skill) do not
  receive injected parameters.
`

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Show detailed command specification for AI agents",
	Long:  "Display the complete skill specification, designed for AI agents.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(skillSpec)
		printInitHint()
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
}
