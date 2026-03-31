# InfiniSynapse CLI (agent_infini)

基于 Go 的命令行工具，通过调用 InfiniSynapse 后端 REST API，在终端中完成多轮 AI 任务对话、数据源管理、RAG 知识库管理和工作区文件操作，适用于人工使用和 AI Agent 集成场景。

## 安装

### 从源码编译

```bash
git clone git@github.com:chaozwn/infinisynapse-cli.git
cd infinisynapse-cli
make build
```

构建产物位于 `build/agent_infini`。

### 交叉编译全平台

```bash
make cross
```

在 `build/` 目录下生成：

| 路径 | 平台 |
|---|---|
| `build/linux-amd64/agent_infini` | Linux x86_64 |
| `build/linux-arm64/agent_infini` | Linux ARM64 |
| `build/darwin-amd64/agent_infini` | macOS Intel |
| `build/darwin-arm64/agent_infini` | macOS Apple Silicon |
| `build/windows-amd64/agent_infini.exe` | Windows x86_64 |
| `build/windows-arm64/agent_infini.exe` | Windows ARM64 |

### 其他构建命令

```bash
make install   # 安装到 $GOPATH/bin
make clean     # 清理构建产物
make help      # 查看所有可用命令
```

### 版本号管理

构建时版本号自动从 git tag 获取，无 tag 时默认为 `0.1.0`。发布新版本时打 tag 即可：

```bash
git tag v0.2.0
git push origin v0.2.0
make cross
```

### 环境要求

- Go 1.26+
- Make

## 快速开始

```bash
# 1. 初始化配置（服务器、API key、语言偏好，写入 ~/.agent_infini/config.key）
agent_infini init --api-key sk-xxx

# 2. 查看可用资源
agent_infini db ls            # 列出所有数据库连接
agent_infini rag ls           # 列出所有 RAG 知识库

# 3. 检查当前启用的上下文
agent_infini task context     # 查看已启用的数据库和 RAG

# 4. 按需启用资源
agent_infini db enable <id>   # 启用数据库
agent_infini rag enable <id>  # 启用 RAG 知识库

# 5. 创建 AI 任务并对话
agent_infini task new "帮我分析用户增长趋势"
agent_infini task ask <taskId> "请用柱状图展示"

# 6. 浏览任务列表
agent_infini task ls
```

## 项目结构

```
infinisynapse-cli/
├── main.go                        # 程序入口
├── Makefile                       # 构建脚本
├── go.mod / go.sum                # Go 模块依赖
├── .gitignore
├── cmd/
│   ├── root.go                    # 根命令与全局 flag
│   ├── init.go                    # 初始化本地配置
│   ├── task.go                    # 多轮 AI 任务对话与工作区文件管理
│   ├── database.go                # 数据源管理（列表 / 启用 / 禁用）
│   ├── rag.go                     # RAG 知识库管理（列表 / 启用 / 禁用）
│   ├── skill.go                   # AI Agent 规范说明
│   └── version.go                 # 版本信息
└── internal/
    ├── client/
    │   ├── client.go              # HTTP 客户端封装
    │   ├── sse.go                 # SSE 流式客户端
    │   └── profile.go             # 用户 Profile API
    ├── config/
    │   └── config.go              # 本地配置管理（凭证链加载）
    ├── output/
    │   └── output.go              # JSON / Table 输出格式化
    ├── task/
    │   └── task.go                # SSE 流式任务执行逻辑
    └── types/
        ├── common.go              # 通用类型（API 响应、分页）
        ├── ai.go                  # AI 相关类型
        ├── database.go            # Database 相关类型
        └── rag.go                 # RAG 相关类型
```

## 命令参考

### 初始化 `agent_infini init`

写入 `~/.agent_infini/config.key`（服务器地址、API key、偏好语言、Console URL）。初始化时会自动获取用户 ID。

```bash
agent_infini init --api-key sk-xxx
agent_infini init --server https://custom-server.example.com --api-key sk-xxx
agent_infini init --api-key sk-xxx --prefer-language zh_CN
agent_infini init --api-key sk-xxx --console https://api.infinisynapse.cn/api
```

### 任务管理 `agent_infini task`

```bash
# 创建新任务（SSE 流式输出）
agent_infini task new "帮我分析用户增长趋势"
agent_infini task new --query "检查库存水平"

# 在已有任务中继续对话
agent_infini task ask <taskId> "请用柱状图展示"
agent_infini task ask <taskId> --query "导出报告"

# 任务列表（支持分页和搜索）
agent_infini task ls
agent_infini task ls --page 1 --page-size 20 --search "销售"

# 查看任务详情（包含最后消息和工作区文件）
agent_infini task show <taskId>

# 查看启用的数据库和 RAG 上下文
agent_infini task context
agent_infini task ctx

# 取消运行中的任务
agent_infini task cancel <taskId>

# 删除任务（支持批量，空格或逗号分隔）
agent_infini task rm <taskId>
agent_infini task rm id1 id2 id3
agent_infini task rm id1,id2,id3

# --- 工作区文件 ---

# 列出任务工作区文件
agent_infini task file <taskId>

# 预览文件内容到 stdout
agent_infini task preview <taskId> analysis.py

# 下载文件到本地
agent_infini task download <taskId> report.csv
agent_infini task download <taskId> report.csv -o ./output/
```

### 数据源管理 `agent_infini db`

```bash
# 数据源列表
agent_infini db ls
agent_infini db ls --name mysql --type mysql
agent_infini db ls --enabled
agent_infini db ls --disabled

# 启用 / 禁用（支持批量，空格或逗号分隔）
agent_infini db enable <id>
agent_infini db enable id1 id2 id3
agent_infini db disable <id>
```

支持的数据库类型：`mysql`, `postgres`, `sqlite`, `sqlserver`, `clickhouse`, `snowflake`, `doris`, `starrocks`, `gbase`, `kingbase`, `dm`, `supabase`, `deltalake`, `file`

### RAG 知识库管理 `agent_infini rag`

```bash
# 知识库列表
agent_infini rag ls
agent_infini rag ls --keyword sales
agent_infini rag ls --enabled
agent_infini rag ls --disabled

# 启用 / 禁用（支持批量，空格或逗号分隔）
agent_infini rag enable <id>
agent_infini rag enable uuid-1 uuid-2
agent_infini rag disable <id>
```

### AI Agent 规范 `agent_infini skill`

输出完整的 AI Agent 命令规范说明，供其他 AI Agent 了解和调用本工具。

```bash
agent_infini skill
agent_infini --skill
```

### 版本信息 `agent_infini version`

```bash
agent_infini version
```

输出版本号、Commit、构建时间、Go 版本和操作系统架构。

## 全局 Flag

| Flag | 说明 |
|---|---|
| `--json` | 强制 JSON 输出（默认） |
| `--table` | 强制表格输出 |
| `--skill` | 显示 AI Agent 规范 |
| `--version`, `-v` | 显示版本号 |
| `--help`, `-h` | 显示帮助信息 |
| `--api-key` | 覆盖配置中的 API key |
| `--server` | 覆盖配置中的服务器地址 |
| `--console` | 覆盖配置中的 Console API URL |
| `--prefer-language` | 覆盖配置中的偏好语言 |
| `--default-output` | 覆盖默认输出格式（`json` \| `table`） |

输出格式优先级：`--table` > `--json` > 配置文件 `default-output` > `json`

## 输出格式

默认 JSON 格式。操作类命令（enable / disable / cancel / rm 等）输出统一结构：

```json
{"success": true, "data": { ... }, "message": ""}
{"success": false, "data": null, "message": "error message"}
```

列表类命令直接输出数据结构（不嵌套在 `success` / `data` 中）。表格模式下列表命令以表格输出。

```bash
agent_infini task ls --table
agent_infini task ls | jq '.items[].task_name'
```

## 配置文件

首次执行 `agent_infini init` 后，配置保存在 `~/.agent_infini/config.key`：

```yaml
global:
  server: https://app.infinisynapse.cn
  api-key: your-bearer-token
  console: https://api.infinisynapse.cn/api
  default-output: json
  prefer-language: zh_CN
  user-id: auto-fetched-user-id
```

### 凭证加载链

配置按以下顺序查找，使用第一个找到的文件：

1. `<binary_dir>/agent_infini.key` — 二进制同目录 YAML
2. `<binary_dir>/<filename>.key` — 兼容路径
3. `~/.agent_infini/config.key` — 用户目录 YAML（推荐）
4. `~/.agent_infini/config.json` — 用户目录 JSON

支持语言：`en`, `zh_CN`, `ar`, `ja`, `ko`, `ru`

## License

[MIT](LICENSE)
