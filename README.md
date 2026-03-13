# InfiniSynapse CLI (isc)

基于 Go 的命令行工具，通过调用 InfiniSynapse 后端 REST API，在终端中完成 AI 对话、任务管理、数据源管理和系统设置等操作，方便其他应用集成调用。

## 安装

### 从源码编译

```bash
git clone git@github.com:chaozwn/infinisynapse-cli.git
cd infinisynapse-cli
go build -o isc .
```

Windows 下会生成 `isc.exe`，Linux/macOS 下生成 `isc`。

### 环境要求

- Go 1.22+

## 快速开始

```bash
# 1. 配置服务器地址和认证 Token
isc auth login --server http://localhost:7001 --token YOUR_TOKEN

# 2. 检查连接状态
isc auth status

# 3. 发起一次 AI 对话
isc ai chat "帮我查询本月销售数据"

# 4. 查看任务列表
isc task list --table
```

## 项目结构

```
infinisynapse-cli/
├── main.go                        # 程序入口
├── go.mod / go.sum                # Go 模块依赖
├── .gitignore
├── cmd/
│   ├── root.go                    # 根命令 + 全局 flags
│   ├── auth.go                    # 认证管理
│   ├── ai.go                      # AI 对话与配置
│   ├── task.go                    # 任务管理
│   ├── database.go                # 数据源管理
│   └── setting.go                 # 系统设置
└── internal/
    ├── client/
    │   ├── client.go              # HTTP 客户端封装
    │   └── sse.go                 # SSE 流式客户端
    ├── config/
    │   └── config.go              # 本地配置管理 (~/.isc.yaml)
    ├── output/
    │   └── output.go              # JSON / Table 输出格式化
    └── types/
        ├── common.go              # 通用类型（API 响应、分页）
        ├── ai.go                  # AI 相关类型
        ├── task.go                # Task 相关类型
        ├── database.go            # Database 相关类型
        └── setting.go             # Setting 相关类型
```

## 全局参数

所有命令均支持以下全局参数：

| 参数 | 缩写 | 说明 |
|------|------|------|
| `--server` | `-s` | API 服务器地址（覆盖配置文件） |
| `--token` | `-t` | Bearer Token（覆盖配置文件） |
| `--table` | | 以表格形式输出（默认 JSON） |

## 命令参考

### 认证管理 `isc auth`

```bash
# 配置服务器和 Token（保存到 ~/.isc.yaml）
isc auth login --server http://localhost:7001 --token YOUR_TOKEN

# 仅更新 Token
isc auth login --token NEW_TOKEN

# 检查连接和认证状态
isc auth status

# 清除本地凭据
isc auth logout
```

### AI 对话 `isc ai`

```bash
# 发起新对话（支持 SSE 流式输出）
isc ai chat "帮我分析用户增长趋势"

# 在已有任务中继续对话
isc ai chat --task-id TASK_ID "请用柱状图展示"

# 查看 AI 当前状态
isc ai state
isc ai state --task-id TASK_ID

# 查看 API 配置
isc ai config get

# 更新 API 配置
isc ai config set --provider openai --model gpt-4 --api-key sk-xxx --base-url https://api.openai.com

# 列出可用模型
isc ai models

# 取消正在执行的任务
isc ai cancel --task-id TASK_ID
```

### 任务管理 `isc task`

```bash
# 任务列表（支持分页和筛选）
isc task list
isc task list --page 1 --size 20 --name "销售"
isc task list --category-id CATEGORY_ID --table

# 查看任务详情
isc task show TASK_ID

# 查看任务元信息
isc task info TASK_ID

# 删除任务（支持批量）
isc task delete TASK_ID_1 TASK_ID_2

# 取消运行中的任务
isc task cancel --task-id TASK_ID

# --- 分类管理 ---

# 查看所有分类
isc task category list --table

# 添加分类
isc task category add "月度报表"

# 删除分类
isc task category delete CATEGORY_ID
```

### 数据源管理 `isc db`

```bash
# 数据源列表
isc db list
isc db list --name mysql --type mysql --table

# 查看数据源详情
isc db get DATABASE_ID

# 添加数据源
isc db add --name "生产库" --type mysql --config '{"host":"localhost","port":3306,"user":"root","password":"xxx","database":"mydb"}'

# 更新数据源
isc db update --id DATABASE_ID --name "新名称" --description "更新描述"

# 删除数据源
isc db delete DATABASE_ID_1 DATABASE_ID_2

# 测试连接
isc db test --type mysql --config '{"host":"localhost","port":3306,"user":"root","password":"xxx"}'

# 启用 / 禁用
isc db enable DATABASE_ID_1 DATABASE_ID_2
isc db disable DATABASE_ID
```

### 系统设置 `isc setting`

```bash
# 获取 / 设置键值配置
isc setting get KEY_NAME
isc setting set KEY_NAME VALUE

# 偏好语言
isc setting language get
isc setting language set zh-CN

# 引擎配置
isc setting engine-config get
isc setting engine-config set CONFIG_KEY CONFIG_VALUE

# 模型信息
isc setting model-info MODEL_ID
```

## 输出格式

默认以 JSON 格式输出到 stdout，方便程序化调用和管道处理：

```bash
# JSON 输出（默认）
isc task list

# 表格输出
isc task list --table

# 结合 jq 处理 JSON
isc task list | jq '.items[].task_name'

# 重定向到文件
isc ai state > state.json
```

## 配置文件

首次执行 `isc auth login` 后，配置保存在 `~/.isc.yaml`：

```yaml
server: http://localhost:7001
token: your-bearer-token
default_output: json
lang: zh-CN
```

也可通过环境变量或全局参数 `--server` / `--token` 临时覆盖。

## License

MIT
