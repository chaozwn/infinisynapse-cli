#!/usr/bin/env bash
#
# Install the InfiniSynapse CLI (agent_infini) on macOS / Linux and distribute
# its Skill file to any detected AI tool (.cursor / .codex / .gemini / .claude).
#
#   1. Detect OS/arch and download the matching agent_infini binary from OSS.
#   2. Install to ~/.infini/bin, chmod +x, and add the dir to PATH (rc files).
#   3. Write a SKILL.md into each detected AI tool's skills directory.
#
# Usage:
#   bash install.sh
#   bash install.sh --version 0.9.0
#   curl -fsSL <url>/install.sh | bash

set -euo pipefail

DEFAULT_VERSION="0.9.0"
VERSION=""
VERSION_EXPLICIT=0
BASE_URL="https://infinisynapse.oss-cn-shanghai.aliyuncs.com/plugins/infini_cli"
MANIFEST_URL="$BASE_URL/manifest.json"
SKILL_URL="https://infinisynapse.oss-cn-shanghai.aliyuncs.com/cli-install/SKILL.md"
INSTALL_DIR="$HOME/.infini/bin"
SKILL_NAME="agent_infini"

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; VERSION_EXPLICIT=1; shift 2 ;;
    --version=*) VERSION="${1#*=}"; VERSION_EXPLICIT=1; shift ;;
    -h|--help) grep '^#' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# Resolve the directory of this script (for local SKILL.md), tolerate piping.
SCRIPT_DIR=""
if [ -n "${BASH_SOURCE:-}" ] && [ -f "${BASH_SOURCE[0]}" ]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

step() { printf '\033[36m==> %s\033[0m\n' "$1"; }
ok()   { printf '\033[32m    %s\033[0m\n' "$1"; }
warn() { printf '\033[33m    %s\033[0m\n' "$1"; }
die()  { printf '\033[31mError: %s\033[0m\n' "$1" >&2; exit 1; }

# ---------------------------------------------------------------------------
# 0. Resolve version: explicit --version wins; otherwise read the published
#    manifest and fall back to the bundled default if it cannot be read.
# ---------------------------------------------------------------------------
resolve_version() {
  if [ "$VERSION_EXPLICIT" = "1" ] && [ -n "$VERSION" ]; then
    return
  fi

  step "Resolving latest version"
  local manifest=""
  if command -v curl >/dev/null 2>&1; then
    manifest="$(curl -fsSL --connect-timeout 15 "$MANIFEST_URL" 2>/dev/null || true)"
  elif command -v wget >/dev/null 2>&1; then
    manifest="$(wget -q -O - "$MANIFEST_URL" 2>/dev/null || true)"
  fi

  # Extract the first top-level "version": "X.Y.Z" field without jq.
  local latest=""
  latest="$(printf '%s' "$manifest" \
    | sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n 1)"

  if [ -n "$latest" ]; then
    VERSION="$latest"
    ok "Latest version: $VERSION"
  else
    VERSION="$DEFAULT_VERSION"
    warn "Could not read manifest ($MANIFEST_URL); using default version $VERSION"
  fi
}

resolve_version

# ---------------------------------------------------------------------------
# 1. Detect platform -> OSS directory
# ---------------------------------------------------------------------------
detect_platform_dir() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"
  case "$os" in
    Darwin)
      case "$arch" in
        arm64|aarch64) echo "darwin-arm64" ;;
        x86_64|amd64)  echo "darwin-x64" ;;
        *) die "Unsupported macOS arch: $arch" ;;
      esac ;;
    Linux)
      case "$arch" in
        x86_64|amd64)  echo "linux-x64" ;;
        arm64|aarch64) echo "linux-arm64" ;;
        *) die "Unsupported Linux arch: $arch" ;;
      esac ;;
    *) die "Unsupported OS: $os (use install.ps1 on Windows)" ;;
  esac
}

PLATFORM_DIR="$(detect_platform_dir)"
FILE_NAME="agent_infini"
DOWNLOAD_URL="$BASE_URL/$PLATFORM_DIR/$VERSION/$FILE_NAME"
TARGET_PATH="$INSTALL_DIR/$FILE_NAME"

# ---------------------------------------------------------------------------
# 2. Download binary
# ---------------------------------------------------------------------------
step "Downloading agent_infini $VERSION ($PLATFORM_DIR)"
echo "    $DOWNLOAD_URL"
mkdir -p "$INSTALL_DIR"

if command -v curl >/dev/null 2>&1; then
  curl -fSL --connect-timeout 30 -o "$TARGET_PATH" "$DOWNLOAD_URL" || die "Download failed (curl)."
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$TARGET_PATH" "$DOWNLOAD_URL" || die "Download failed (wget)."
else
  die "Neither curl nor wget is available."
fi

[ -s "$TARGET_PATH" ] || die "Downloaded file is missing or empty: $TARGET_PATH"
chmod +x "$TARGET_PATH"
# Clear macOS quarantine attribute so the binary can run without Gatekeeper prompt.
if [ "$(uname -s)" = "Darwin" ] && command -v xattr >/dev/null 2>&1; then
  xattr -d com.apple.quarantine "$TARGET_PATH" 2>/dev/null || true
fi
ok "Installed to $TARGET_PATH"

# ---------------------------------------------------------------------------
# 3. Add install dir to PATH via shell rc files
# ---------------------------------------------------------------------------
step "Updating PATH"
PATH_LINE="export PATH=\"\$HOME/.infini/bin:\$PATH\""
added=0
for rc in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
  [ -f "$rc" ] || continue
  if ! grep -qF '.infini/bin' "$rc" 2>/dev/null; then
    printf '\n# InfiniSynapse CLI\n%s\n' "$PATH_LINE" >> "$rc"
    ok "Added PATH entry to $rc"
    added=1
  fi
done
if [ "$added" -eq 0 ]; then
  ok "PATH entry already present (or no rc file found)."
fi
export PATH="$INSTALL_DIR:$PATH"

# ---------------------------------------------------------------------------
# 4. Distribute SKILL.md to detected AI tools
# ---------------------------------------------------------------------------
step "Distributing Skill file to detected AI tools"

# Resolve a single SKILL.md source: local file > download from server > embedded.
SKILL_SRC="$(mktemp 2>/dev/null || echo "${TMPDIR:-/tmp}/agent_infini_SKILL.md")"
if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/SKILL.md" ]; then
  cp "$SCRIPT_DIR/SKILL.md" "$SKILL_SRC"
  ok "Using local SKILL.md"
elif command -v curl >/dev/null 2>&1 && curl -fsSL --connect-timeout 30 -o "$SKILL_SRC" "$SKILL_URL" && [ -s "$SKILL_SRC" ]; then
  ok "Downloaded SKILL.md from server"
elif command -v wget >/dev/null 2>&1 && wget -q -O "$SKILL_SRC" "$SKILL_URL" && [ -s "$SKILL_SRC" ]; then
  ok "Downloaded SKILL.md from server"
else
  warn "Falling back to embedded SKILL content."
  cat > "$SKILL_SRC" <<'SKILL_EOF'
---
name: agent_infini
description: Use the InfiniSynapse CLI (agent_infini) to run multi-turn AI data-analysis tasks, manage database/RAG context, and work with task workspace files from the terminal. Use when the user mentions InfiniSynapse, agent_infini, or wants AI-driven database / RAG analysis from the command line.
---

# agent_infini (InfiniSynapse CLI)

`agent_infini` is a CLI that talks to the InfiniSynapse backend REST API to run multi-turn AI tasks, manage data sources / RAG knowledge bases, and handle task workspace files.

Binary default location: `~/.infini/bin/agent_infini`. If `agent_infini` is not on PATH, call it by full path.

## Setup (run once)
    agent_infini init --api-key "your_api_key"
Writes ~/.agent_infini/config.txt (server, api-key, console, prefer-language).

## Recommended workflow
1. agent_infini init --api-key "your_api_key"
2. agent_infini db ls / agent_infini rag ls
3. agent_infini task context  (enable with db enable / rag enable if needed)
4. agent_infini task new "..."  then  agent_infini task ask <taskId> "..."
5. agent_infini task ls / show / file / download

## Commands
    init --api-key sk-xxx [--server URL] [--prefer-language zh_CN]
    task new <query> | task ask <taskId> <query>
    task ls [--page N] [--page-size N] [--search Q]
    task show <taskId> | task context | task cancel <taskId> | task rm <id...>
    task file <taskId> | task preview <taskId> <file> | task download <taskId> <file> [-o dir]
    db ls [--type T] [--enabled] [--disabled] | db enable <id...> | db disable <id...>
    rag ls [--keyword K] [--enabled] [--disabled] | rag enable <id...> | rag disable <id...>
    skill | version

DB types: mysql, postgres, sqlite, sqlserver, clickhouse, snowflake, doris,
starrocks, gbase, kingbase, dm, supabase, deltalake, file

## Global flags
--json (default) | --table | --skill | --version,-v | --help,-h
--api-key | --server | --console | --prefer-language | --default-output
Priority: --table > --json > config default-output > json

## Output (JSON default)
    {"success": true, "data": { ... }}
    {"success": false, "error": "message"}
Pipe to jq, e.g.: agent_infini task ls | jq '.items[].task_name'

## Errors
- Token expired: re-run agent_infini init or edit ~/.agent_infini/config.txt
- Server unreachable: check --server URL and network
- Task not found: use task ls
- No enabled resources: task context, then db enable / rag enable

Run `agent_infini skill` for the full specification.
SKILL_EOF
fi

distributed=0
for tool in .cursor .codex .gemini .claude; do
  tool_dir="$HOME/$tool"
  if [ -d "$tool_dir" ]; then
    skill_dir="$tool_dir/skills/$SKILL_NAME"
    mkdir -p "$skill_dir"
    cp "$SKILL_SRC" "$skill_dir/SKILL.md"
    ok "$tool  -> $skill_dir/SKILL.md"
    distributed=$((distributed + 1))
  fi
done
rm -f "$SKILL_SRC" 2>/dev/null || true
if [ "$distributed" -eq 0 ]; then
  warn "No AI tool folders (.cursor/.codex/.gemini/.claude) found under $HOME; skipped Skill distribution."
fi

echo ""
step "Done. Next steps:"
echo "    1) Reload your shell:  source ~/.bashrc   (or ~/.zshrc)"
echo "    2) Run: agent_infini init --api-key \"your_api_key\""
echo "    3) Run: agent_infini version"
