package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const skillSpec = `# isc (InfiniSynapse CLI) Skill Specification

A CLI tool for interacting with the InfiniSynapse backend, designed for AI agent workflows.

================================================================================
AI Agent Important Notes
================================================================================

First-time Setup:
  Before using any other commands, you MUST initialize configuration:
  isc init --server "http://app.infinisynapse.cn" --api-key "your_api_key"

Session Workflow (multi-turn chat):
  1. Start a new conversation with a session alias:
     isc chat "Analyze my data" --session main
  2. Continue the conversation (auto-resumes context):
     isc chat "Show me the trends" --session main
  3. Manage sessions:
     isc session ls              # List all sessions
     isc session show main       # View session details
     isc session rm main         # Delete a session

Without --session, each 'isc chat' starts a fresh conversation.

================================================================================
Available Commands
================================================================================

  init                           Interactive setup (server URL + API key)
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
  --help, -h             Show help for any command

================================================================================
Common Scenarios
================================================================================

1. Quick one-shot question:
   isc chat "What tables are in my database?"

2. Multi-turn analysis with session:
   isc chat "Analyze the users table schema" --session analysis
   isc chat "Now show me the top 10 users by activity" --session analysis
   isc chat "Generate a summary report" --session analysis

3. Task management:
   isc task list --size 20
   isc task show <task-id>
   isc task delete <task-id>

4. Pipeline with JSON output:
   isc task list --json | jq '.data.items[].id'

================================================================================
Output Format
================================================================================

JSON mode (--json or default_output=json in ~/.isc.yaml):
  {"success": true, "data": { ... }}
  {"success": false, "error": "error message"}

Table mode (default_output=table in ~/.isc.yaml):
  Formatted table for list commands, JSON for detail commands.

================================================================================
Error Handling
================================================================================

  - Token expired:     Run 'isc init' to reconfigure
  - Server unreachable: Check --server URL and network connectivity
  - Session not found:  A new conversation will be started automatically

================================================================================
Configuration
================================================================================

Config file: ~/.isc.yaml
Session files: ~/.isc/sessions/<name>.json
`

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Show detailed command specification for AI agents",
	Long:  "Display the complete skill specification, designed for AI agents.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(skillSpec)
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
}
