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
  1. Start a new conversation with a session alias:
     agent_infini chat "Analyze my data" --session main
  2. Continue the conversation (auto-resumes context):
     agent_infini chat "Show me the trends" --session main
  3. Manage sessions:
     agent_infini session ls              # List all sessions
     agent_infini session show main       # View session details
     agent_infini session rm main         # Delete a session

Without --session, each 'agent_infini chat' starts a fresh conversation.

================================================================================
Available Commands
================================================================================

  version                        Print version, commit, build date, Go runtime
  chat <message>                 Chat with AI (streaming), supports --session
  chat state                     Get AI state
  chat config get                Get API configuration
  chat config set                Update API configuration
  chat models                    List available AI models
  chat cancel                    Cancel a running task
  task list                      List tasks with pagination
  task show <id>                 Show task details
  task info <id>                 Get task metadata
  task delete <ids...>           Delete tasks
  task cancel                    Cancel a running task
  task category list             List task categories
  task category add <name>       Add a task category
  task category delete <ids...>  Delete task categories
  setting get <key>              Get a setting value
  setting set <key> <value>      Set a setting value
  setting language get           Get preferred language
  setting language set <lang>    Set preferred language
  setting engine-config get      Get engine configuration
  setting engine-config set      Update engine configuration
  setting model-info <model-id>  Get model information
  session ls                     List all sessions
  session show <name>            Show session details
  session rm <name>              Delete a session

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

1. Quick one-shot question:
   agent_infini chat "What tables are in my database?"

2. Multi-turn analysis with session:
   agent_infini chat "Analyze the users table schema" --session analysis
   agent_infini chat "Now show me the top 10 users by activity" --session analysis
   agent_infini chat "Generate a summary report" --session analysis

3. Task management:
   agent_infini task list --size 20
   agent_infini task show <task-id>
   agent_infini task delete <task-id>

4. Pipeline with JSON output:
   agent_infini task list --json | jq '.data.items[].id'

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
