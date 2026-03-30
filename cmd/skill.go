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

Task Workflow (multi-turn chat):
  1. Create a task:
     agent_infini task new "Analyze my data"
  2. Continue the conversation (respond to server ask):
     agent_infini task ask <taskId> "Show me the trends"
  3. Manage tasks:
     agent_infini task ls                        # List tasks (paginated)
     agent_infini task ls --search "analysis"    # Search by name
     agent_infini task show <taskId>             # View task details
     agent_infini task rm <taskId>               # Delete a task
  4. Workspace files:
     agent_infini task file <taskId>             # List workspace files
     agent_infini task preview <taskId> <path>   # Preview file content
     agent_infini task download <taskId> <path>  # Download file to local

================================================================================
Available Commands
================================================================================

  version                                          Print version, commit, build date, Go runtime
  task new <query>                                 Create a new task (newTask)
  task new --query <message>                       (alternative flag form)
  task ask <taskId> <query>                        Continue conversation in task (askResponse)
  task ask <taskId> --query <message>              (alternative flag form)
  task ls [--page N] [--page-size N] [--search Q]  List tasks with pagination and search
  task show <taskId>                               Show task details
  task rm <taskId> [taskId2...]                    Delete one or more tasks
  task cancel <taskId>                             Cancel a running task
  task file <taskId>                               List workspace files
  task preview <taskId> <fileName>                  Preview workspace file to stdout
  task download <taskId> <fileName> [-o dir]       Download workspace file to local disk

================================================================================
Global Flags
================================================================================

  --json               Force JSON output: {"success": true, "data": ...}
  --skill              Show this detailed specification
  --version, -v        Print version string
  --help, -h           Show help for any command

================================================================================
Common Scenarios
================================================================================

1. Start a new analysis task:
   agent_infini task new "What tables are in my database?"

2. Multi-turn analysis:
   agent_infini task new "Analyze the users table schema"
   agent_infini task ask <taskId> "Now show me the top 10 users by activity"
   agent_infini task ask <taskId> "Generate a summary report"

3. Browse tasks:
   agent_infini task ls
   agent_infini task ls --search "analysis" --page 2 --page-size 20
   agent_infini task show <taskId>

4. Cancel a running task:
   agent_infini task cancel <taskId>

5. Work with workspace files:
   agent_infini task file <taskId>
   agent_infini task preview <taskId> analysis.py
   agent_infini task download <taskId> report.csv -o ./results/

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

  - Token expired:      Update api-key in ~/.agent_infini/config.key
  - Server unreachable: Check --server URL and network connectivity
  - Task not found:     Use 'task ls' to find valid task IDs

================================================================================
Configuration & Credential Chain
================================================================================

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
