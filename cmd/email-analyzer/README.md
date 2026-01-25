# Email Analyzer CLI

基于 AgentBox 的批量邮件钓鱼分析命令行工具。

## 功能特性

- 单个/批量邮件钓鱼分析
- 并行分析提升效率
- 多种输出格式 (table/json/csv)
- 自动提取威胁指标 (IOC)
- 实时进度显示

## 安装

```bash
# 从源码编译
cd /path/to/agentbox
go build -o bin/email-analyzer ./cmd/email-analyzer

# 或使用 make
make build-email-analyzer
```

## 快速开始

### 1. 检查服务状态

```bash
email-analyzer status
# 或指定服务地址
email-analyzer status -u http://localhost:18080
```

### 2. 获取 API Token

```bash
# 交互式登录
email-analyzer login

# 登录成功后会显示 Token，设置环境变量
export AGENTBOX_API_KEY=<your-token>
```

### 3. 分析邮件

```bash
# 分析单个邮件
email-analyzer analyze -F sample.eml

# 批量分析目录
email-analyzer analyze -d ./emails/

# 指定并行数和超时
email-analyzer analyze -d ./emails/ -w 10 -t 15m

# 输出到文件
email-analyzer analyze -d ./emails/ -o report.json -f json
```

## 命令参考

### analyze - 分析邮件

```bash
email-analyzer analyze [flags]

Flags:
  -F, --file string      单个邮件文件路径
  -d, --dir string       邮件目录路径
  -a, --agent string     Agent ID (default "phishing-analyzer")
  -k, --api-key string   API Key 或 JWT Token
  -u, --url string       AgentBox API 地址 (default "http://localhost:18080")
  -w, --workers int      并行任务数 (default 5)
  -t, --timeout duration 单任务超时时间 (default 10m)
  -f, --format string    输出格式: table/json/csv (default "table")
  -o, --output string    输出文件路径（默认输出到终端）
```

### login - 登录获取 Token

```bash
email-analyzer login [flags]

Flags:
  -u, --url string   AgentBox API 地址 (default "http://localhost:18080")
```

### status - 检查服务状态

```bash
email-analyzer status [flags]

Flags:
  -u, --url string   AgentBox API 地址 (default "http://localhost:18080")
```

### list - 列出历史任务

```bash
email-analyzer list [flags]

Flags:
  -k, --api-key string   API Key 或 JWT Token
  -u, --url string       AgentBox API 地址 (default "http://localhost:18080")
  -n, --limit int        显示数量 (default 20)
```

## 输出格式

### Table (默认)

```
┌─────────────────────────┬──────────┬───────┬─────────────────────────┐
│ 文件                    │ 风险等级 │ 评分  │ 威胁类型                │
├─────────────────────────┼──────────┼───────┼─────────────────────────┤
│ sample_phishing.eml     │ 🔴 高    │ 85    │ 钓鱼链接, 伪造发件人    │
│ ceo_fraud.eml           │ 🔴 严重  │ 95    │ BEC 攻击, 紧急诱导      │
│ newsletter.eml          │ 🟢 安全  │ 10    │ -                       │
└─────────────────────────┴──────────┴───────┴─────────────────────────┘

📊 统计: 严重 2 | 高 5 | 中 3 | 低 8 | 安全 7
```

### JSON

```json
{
  "summary": {
    "total": 25,
    "completed": 24,
    "failed": 1,
    "risk_distribution": {
      "critical": 2,
      "high": 5,
      "medium": 3,
      "low": 8,
      "safe": 6
    }
  },
  "results": [
    {
      "file": "sample_phishing.eml",
      "risk_level": "high",
      "risk_score": 85,
      "threats": ["钓鱼链接", "伪造发件人"],
      "iocs": [
        {"type": "url", "value": "https://evil.com/login", "risk": "malicious"}
      ],
      "summary": "该邮件包含伪造的发件人地址和钓鱼链接..."
    }
  ]
}
```

### CSV

```csv
文件,风险等级,评分,威胁类型,IOC数量,摘要
sample_phishing.eml,high,85,"钓鱼链接,伪造发件人",3,"该邮件包含..."
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `AGENTBOX_API_KEY` | API Key 或 JWT Token | - |

## 前置条件

1. AgentBox 服务运行中
2. 配置了 `phishing-analyzer` Agent（或使用 `-a` 指定其他 Agent）
3. Agent 使用的 Provider 已配置 API Key

## 支持的文件格式

- `.eml` - 标准邮件格式
- `.msg` - Outlook 邮件格式

## 示例

```bash
# 完整示例：批量分析并生成报告
export AGENTBOX_API_KEY=$(email-analyzer login 2>/dev/null | grep Token | awk '{print $2}')

email-analyzer analyze \
  -d /path/to/emails/ \
  -w 10 \
  -t 15m \
  -f json \
  -o analysis_report.json

echo "分析完成，报告保存到 analysis_report.json"
```
