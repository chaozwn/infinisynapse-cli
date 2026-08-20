package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const skillSpec = `# agent_infini (InfiniSynapse CLI) Skill Specification

A CLI tool for interacting with the InfiniSynapse platform, designed for AI agent workflows.

================================================================================
AI Agent Recommended Workflow
================================================================================

Step 1 — First-time setup (run once):
  agent_infini init --api-key "your_api_key"

  Or manually create ~/.agent_infini/config.txt:
  mkdir -p ~/.agent_infini && cat > ~/.agent_infini/config.txt << 'EOF'
  global:
    server: "https://app.infinisynapse.cn"
    api-key: "your_api_key"
    console: "https://api.infinisynapse.cn/api"
    prefer-language: "zh_CN"
  EOF

Step 2 — List available resources (use the real name/id, e.g. remote_tmall not t_mall):
  agent_infini db ls                             # List all databases
  agent_infini rag ls                            # List all RAG knowledge bases

Step 3 — Enable the resources this query needs. This is required, not optional:
  agent_infini db enable <id>                    # Enable a database
  agent_infini rag enable <id>                   # Enable a RAG knowledge base
  agent_infini task context                      # Confirm the target appears in the enabled list
  If it is missing, STOP. Do not call task new.

Step 4 — Create the task, then verify it inherited those resources:
  agent_infini task new "Analyze my data"
  The JSON result includes databaseIds and ragIds. task new cannot take --database-id;
  the server snapshots currently enabled resources.
  If databaseIds is [] and the question needs a database, STOP.
  Bind afterwards with:
  agent_infini task resources <taskId> --db <id>
  Then continue:
  agent_infini task ask <taskId> "Show trends"

Step 5 — Manage tasks and workspace files:
  agent_infini task ls                           # List tasks (paginated)
  agent_infini task show <taskId>                # View task details
  agent_infini task cancel <taskId>              # Cancel a running task
  agent_infini task rm <taskId>                  # Delete a task
  agent_infini task file <taskId>                # List workspace files
  agent_infini task preview <taskId> <path>      # Preview file content
  agent_infini task download <taskId> <path>     # Download file to local

================================================================================
Available Commands
================================================================================

  init                                             Configure server, API key, and preferences

  task new <query>                                 Create a new AI task
  task new --query <message>                       (alternative flag form)
  task ask <taskId> <query>                        Continue conversation in a task
  task ask <taskId> --query <message>              (alternative flag form)
  task ls [--page N] [--page-size N] [--search Q]  List tasks with pagination and search
  task show <taskId>                               Show task details
  task rm <taskId> [taskId2...]                    Delete one or more tasks
  task cancel <taskId>                             Cancel a running task
  task context                                     Show enabled databases and RAGs
  task resources <taskId>                          Show databases/RAGs bound to a task
  task resources <taskId> --db <id> [--rag <id>]   Bind databases/RAGs onto an existing task
  task file <taskId>                               List workspace files
  task preview <taskId> <fileName>                 Preview workspace file to stdout
  task download <taskId> <fileName> [-o dir]       Download workspace file to local disk

  db ls [--type T] [--enabled] [--disabled]        List registered database connections
  db enable <id> [id...]                           Enable databases for AI task access
  db disable <id> [id...]                          Disable databases from AI task access

  rag ls [--keyword K] [--enabled] [--disabled]    List registered RAG knowledge bases
  rag enable <id> [id...]                          Enable RAGs for AI task access
  rag disable <id> [id...]                         Disable RAGs from AI task access

  skill                                            Show this specification
  version                                          Print version, commit, build date, Go runtime

================================================================================
Supported Database Types
================================================================================

  mysql, postgres, sqlite, sqlserver, clickhouse, snowflake,
  doris, starrocks, gbase, kingbase, dm, supabase, deltalake, file

================================================================================
Global Flags
================================================================================

  --json               Force JSON output: {"success": true, "data": ...} (default)
  --table              Force table output for list commands
  --skill              Show this detailed specification
  --version, -v        Print version string
  --help, -h           Show help for any command

  Credential overrides (auto-injected by WinClaw executor, also usable manually):
  --api-key <key>      Override API key from config
  --server <url>       Override server address from config
  --console <url>      Override console API base URL from config
  --prefer-language <l> Override preferred language from config
  --default-output <f> Override default output format (json|table)

================================================================================
Common Scenarios
================================================================================

1. Check which databases and RAGs are available:
   agent_infini task context
   agent_infini db ls
   agent_infini rag ls --enabled

2. Enable a database before starting analysis (required for SQL / data questions):
   agent_infini db ls
   agent_infini db enable <id>
   agent_infini task context

3. Start a new analysis task, then confirm databaseIds is not empty:
   agent_infini task new "What tables are in my database?"
   If databaseIds is [], bind explicitly:
   agent_infini task resources <taskId> --db <id>

4. Multi-turn analysis:
   agent_infini task new "Analyze the users table schema"
   agent_infini task ask <taskId> "Now show me the top 10 users by activity"
   agent_infini task ask <taskId> "Generate a summary report"

5. Browse and search tasks:
   agent_infini task ls
   agent_infini task ls --search "analysis" --page 2 --page-size 20
   agent_infini task show <taskId>

6. Cancel a running task:
   agent_infini task cancel <taskId>

7. Work with workspace files:
   agent_infini task file <taskId>
   agent_infini task preview <taskId> analysis.py
   agent_infini task download <taskId> report.csv -o ./results/

================================================================================
Output Format
================================================================================

JSON mode (default, or --json to override config):
  {"success": true, "data": { ... }}
  {"success": false, "error": "error message"}

Table mode (--table, or default-output: "table" in config.txt):
  Formatted table for list commands, JSON for detail commands.

Priority: --table > --json > config default-output > json

================================================================================
Error Handling
================================================================================

  - Token expired:        Update api-key via 'agent_infini init' or edit ~/.agent_infini/config.txt
  - Server unreachable:   Check --server URL and network connectivity
  - Task not found:       Use 'task ls' to find valid task IDs
  - No enabled resources: Use 'task context' to check, then 'db enable' or 'rag enable'
  - Task has empty databaseIds: The query will not see any database. Enable first, or run
    'task resources <taskId> --db <id>'. task new does not accept a database id flag.

================================================================================
Configuration & Credential Chain
================================================================================

Configuration is loaded from the first file found in this order
(per execute_external_tool_resolver.py):

  1. <binary_dir>/agent_infini.key          (tool_basename.key, YAML)
  2. <binary_dir>/<filename>.key            (tool_filename.key, compat)
  3. ~/.agent_infini/config.txt             (YAML, recommended)
  4. ~/.agent_infini/config.json            (JSON)

The most common approach: run 'agent_infini init' or create ~/.agent_infini/config.txt.
config.txt and config.json are alternatives; if config.txt exists, config.json
is not checked.

config.txt format (YAML):
  global:
    server: "https://app.infinisynapse.cn"
    api-key: "your_api_key"
    console: "https://api.infinisynapse.cn/api"
    default-output: "json"
    prefer-language: "zh_CN"

config.json format (JSON):
  {
    "global": {
      "server": "https://app.infinisynapse.cn",
      "api-key": "your_api_key",
      "console": "https://api.infinisynapse.cn/api"
    }
  }

Supported languages: en, zh_CN, ar, ja, ko, ru

WinClaw Marketplace:
  The executor reads config.txt and injects credentials as global CLI flags
  (--api-key, --server, --console, etc.) automatically. These flags are accepted
  by all commands and override the corresponding config file values.
  Help commands (--help, --skill) do not receive injected parameters.
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
