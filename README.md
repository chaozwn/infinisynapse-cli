# InfiniSynapse CLI (agent_infini)

基于 Go 的命令行工具，通过调用 InfiniSynapse 后端 REST API，在终端中完成 AI 对话、任务管理、数据源管理和系统设置等操作，方便其他应用集成调用。

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
make test      # 运行测试
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

- Go 1.22+
- Make

## 快速开始

```bash
# 1. 初始化配置（服务器与 API key，写入 ~/.agent_infini/config.key）
agent_infini init --api-key sk-xxx

# 2. 发起一次 AI 对话
agent_infini chat "帮我查询本月销售数据"

# 3. 查看任务列表
agent_infini task list
```

## 项目结构

```
infinisynapse-cli/
├── main.go                        # 程序入口
├── go.mod / go.sum                # Go 模块依赖
├── .gitignore
├── cmd/
│   ├── root.go                    # 根命令
│   ├── init.go                    # 初始化本地配置
│   ├── chat.go                    # 聊天与 AI 配置
│   ├── task.go                    # 任务管理
│   ├── database.go                # 数据源管理
│   └── setting.go                 # 系统设置
└── internal/
    ├── client/
    │   ├── client.go              # HTTP 客户端封装
    │   └── sse.go                 # SSE 流式客户端
    ├── config/
    │   └── config.go              # 本地配置管理 (~/.agent_infini/config.key)
    ├── output/
    │   └── output.go              # JSON / Table 输出格式化
    └── types/
        ├── common.go              # 通用类型（API 响应、分页）
        ├── ai.go                  # AI 相关类型
        ├── task.go                # Task 相关类型
        ├── database.go            # Database 相关类型
        └── setting.go             # Setting 相关类型
```

## 命令参考

### 初始化 `agent_infini init`

写入 `~/.agent_infini/config.key`（服务器地址、API key）。默认服务器为 `https://app.infinisynapse.cn`。也可用 `--server`、`--api-key` 非交互。

### 聊天 `agent_infini chat`

```bash
# 发起新对话（支持 SSE 流式输出）
agent_infini chat "帮我分析用户增长趋势"

# 在已有任务中继续对话
agent_infini chat --task-id TASK_ID "请用柱状图展示"

# 查看 AI 当前状态
agent_infini chat state
agent_infini chat state --task-id TASK_ID

# 查看 API 配置
agent_infini chat config get

# 更新 API 配置
agent_infini chat config set --provider openai --model gpt-4 --api-key sk-xxx --base-url https://api.openai.com

# 列出可用模型
agent_infini chat models

# 取消正在执行的任务
agent_infini chat cancel --task-id TASK_ID
```

### 任务管理 `agent_infini task`

```bash
# 任务列表（支持分页和筛选）
agent_infini task list
agent_infini task list --page 1 --size 20 --name "销售"
agent_infini task list --category-id CATEGORY_ID

# 查看任务详情
agent_infini task show TASK_ID

# 查看任务元信息
agent_infini task info TASK_ID

# 删除任务（支持批量）
agent_infini task delete TASK_ID_1 TASK_ID_2

# 取消运行中的任务
agent_infini task cancel --task-id TASK_ID

# --- 分类管理 ---

# 查看所有分类
agent_infini task category list

# 添加分类
agent_infini task category add "月度报表"

# 删除分类
agent_infini task category delete CATEGORY_ID
```

### 数据源管理 `agent_infini db`

```bash
# 数据源列表
agent_infini db list
agent_infini db list --name mysql --type mysql

# 查看数据源详情
agent_infini db get DATABASE_ID

# 添加数据源
agent_infini db add --name "生产库" --type mysql --config '{"host":"localhost","port":3306,"user":"root","password":"xxx","database":"mydb"}'

# 更新数据源
agent_infini db update --id DATABASE_ID --name "新名称" --description "更新描述"

# 删除数据源
agent_infini db delete DATABASE_ID_1 DATABASE_ID_2

# 测试连接
agent_infini db test --type mysql --config '{"host":"localhost","port":3306,"user":"root","password":"xxx"}'

# 启用 / 禁用
agent_infini db enable DATABASE_ID_1 DATABASE_ID_2
agent_infini db disable DATABASE_ID
```

### 系统设置 `agent_infini setting`

```bash
# 获取 / 设置键值配置
agent_infini setting get KEY_NAME
agent_infini setting set KEY_NAME VALUE

# 偏好语言
agent_infini setting language get
agent_infini setting language set zh-CN

# 引擎配置
agent_infini setting engine-config get
agent_infini setting engine-config set CONFIG_KEY CONFIG_VALUE

# 模型信息
agent_infini setting model-info MODEL_ID
```

## 输出格式

默认 JSON；在 `~/.agent_infini/config.key` 中设置 `default-output: table` 则各列表类命令以表格输出。

```bash
agent_infini task list
agent_infini task list | jq '.items[].task_name'
agent_infini chat state > state.json
```

## 配置文件

首次执行 `agent_infini init` 后，配置保存在 `~/.agent_infini/config.key`：

```yaml
global:
  server: https://app.infinisynapse.cn
  api-key: your-bearer-token
  default-output: json
  lang: zh-CN
```

服务器地址与 Token 均来自 `~/.agent_infini/config.key`（通过 `agent_infini init` 或手动编辑）。修改后无需额外命令行参数。

## License

MIT
