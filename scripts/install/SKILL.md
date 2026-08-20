---
name: agent_infini
description: Use the InfiniSynapse CLI (agent_infini) to run multi-turn AI data-analysis tasks, manage database/RAG context, and work with task workspace files from the terminal. Use when the user mentions InfiniSynapse, agent_infini, or wants AI-driven database / RAG analysis from the command line.
---

# agent_infini (InfiniSynapse CLI)

`agent_infini` is a command-line tool that talks to the InfiniSynapse backend REST API to run multi-turn AI tasks, manage data sources, manage RAG knowledge bases, and handle task workspace files. This Skill explains how an AI agent should call it.

Default binary location: `~/.infini/bin/agent_infini` (Windows: `%USERPROFILE%\.infini\bin\agent_infini.exe`). If `agent_infini` is not on PATH, call it by its full path.

## Prerequisites

Initialize the config once before first use:

```bash
agent_infini init --api-key "your_api_key"
```

This writes `~/.agent_infini/config.txt`:

```yaml
global:
  server: "https://app.infinisynapse.cn"
  api-key: "your_api_key"
  console: "https://api.infinisynapse.cn/api"
  prefer-language: "zh_CN"
  default-output: "json"
```

Credentials are loaded from the first file found in this order:

1. `<binary_dir>/agent_infini.key` (YAML, next to the binary)
2. `<binary_dir>/<filename>.key` (compat path)
3. `~/.agent_infini/config.txt` (recommended)
4. `~/.agent_infini/config.json`

## Recommended workflow

These steps are required for any question that needs a database or RAG. Skipping enable is why SQL tasks fail.

1. Initialize: `agent_infini init --api-key "your_api_key"`
2. List resources: `agent_infini db ls` / `agent_infini rag ls` (use the real name, e.g. `remote_tmall`)
3. Enable + verify: `agent_infini db enable <id>` then `agent_infini task context`
4. Create a task: `agent_infini task new "..."` — this snapshots currently enabled resources. It does **not** accept `--database-id`.
5. Inspect the JSON `databaseIds` / `ragIds`. If they are empty and the question needs data, STOP and run `agent_infini task resources <taskId> --db <id>`.
6. Continue: `agent_infini task ask <taskId> "..."`

## Command reference

### Init

```bash
agent_infini init --api-key sk-xxx
agent_infini init --server https://custom.example.com --api-key sk-xxx --prefer-language zh_CN
```

### Tasks: task

```bash
agent_infini task new "Analyze user growth trend"     # create a task (SSE streaming)
agent_infini task ask <taskId> "Show it as a bar chart"  # continue the conversation
agent_infini task ls [--page N] [--page-size N] [--search Q]
agent_infini task show <taskId>                       # task details
agent_infini task context                             # show enabled DBs and RAGs (alias: ctx)
agent_infini task resources <taskId>                  # show databases/RAGs bound to a task
agent_infini task resources <taskId> --db <id>        # bind databases onto an existing task
agent_infini task cancel <taskId>                     # cancel a running task
agent_infini task rm <id1> [id2 ...]                  # delete tasks (batch: space/comma separated)
agent_infini task file <taskId>                       # list workspace files
agent_infini task preview <taskId> <fileName>         # preview file content to stdout
agent_infini task download <taskId> <fileName> [-o dir]  # download file to local disk
```

### Databases: db

```bash
agent_infini db ls [--name N] [--type T] [--enabled] [--disabled]
agent_infini db enable <id> [id...]
agent_infini db disable <id> [id...]
```

Supported database types: `mysql, postgres, sqlite, sqlserver, clickhouse, snowflake, doris, starrocks, gbase, kingbase, dm, supabase, deltalake, file`

### RAG knowledge bases: rag

```bash
agent_infini rag ls [--keyword K] [--enabled] [--disabled]
agent_infini rag enable <id> [id...]
agent_infini rag disable <id> [id...]
```

### Other

```bash
agent_infini skill        # print the full AI-agent specification
agent_infini version      # version info
```

## Global flags

| Flag | Description |
|---|---|
| `--json` | Force JSON output (default) |
| `--table` | Force table output (list commands) |
| `--skill` | Show the AI-agent specification |
| `--version`, `-v` | Print version |
| `--help`, `-h` | Show help |
| `--api-key <key>` | Override API key from config |
| `--server <url>` | Override server address |
| `--console <url>` | Override Console API URL |
| `--prefer-language <l>` | Override preferred language |
| `--default-output <f>` | Override default output format (json\|table) |

Output priority: `--table` > `--json` > config `default-output` > `json`.

## Output format

JSON mode (default):

```json
{"success": true, "data": { ... }}
{"success": false, "error": "error message"}
```

List commands print the data structure directly and can be piped to `jq`:

```bash
agent_infini task ls | jq '.items[].task_name'
```

## Error handling

- Token expired: re-run `agent_infini init` or edit `~/.agent_infini/config.txt`
- Server unreachable: check the `--server` URL and network
- Task not found: use `task ls` to find a valid taskId
- No enabled resources: use `task context`, then `db enable` / `rag enable`
- Task `databaseIds` is empty: the task cannot query any database. Enable first, or `task resources <taskId> --db <id>`. `task new` does not accept a database id flag.

## Common scenarios

```bash
# Enable a database, then start analysis
agent_infini db ls
agent_infini db enable <id>
agent_infini task context
agent_infini task new "What tables are in my database?"
# Confirm the result includes databaseIds. If empty:
agent_infini task resources <taskId> --db <id>

# Multi-turn analysis
agent_infini task new "Analyze the users table schema"
agent_infini task ask <taskId> "Now show me the top 10 users by activity"
agent_infini task ask <taskId> "Generate a summary report"

# Work with workspace files
agent_infini task file <taskId>
agent_infini task preview <taskId> analysis.py
agent_infini task download <taskId> report.csv -o ./results/
```

Supported languages: `en, zh_CN, ar, ja, ko, ru`
