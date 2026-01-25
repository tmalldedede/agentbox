package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	baseURL       string
	apiKey        string
	agentID       string
	outputFmt     string
	outputFile    string
	workers       int
	taskTimeout   time.Duration
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "email-analyzer",
		Short: "批量邮件钓鱼分析工具",
		Long: `基于 AgentBox 的批量邮件钓鱼分析 CLI 工具。

支持功能:
  - 单个/批量邮件钓鱼分析
  - 并行分析提升效率
  - 多种输出格式 (table/json/csv)
  - 自动提取威胁指标 (IOC)

示例:
  # 检查服务状态
  email-analyzer status

  # 登录获取 Token
  email-analyzer login

  # 分析单个邮件
  email-analyzer analyze -F evil.eml -k ab_xxx

  # 批量分析目录
  email-analyzer analyze -d ./emails/ -k ab_xxx -w 10 -o report.json`,
	}

	// analyze 命令
	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "分析邮件文件或目录",
		Long: `分析单个邮件文件或批量分析目录中的所有邮件。

支持的文件格式: .eml, .msg

示例:
  email-analyzer analyze -F sample.eml -k ab_xxx
  email-analyzer analyze -d ./emails/ -k ab_xxx -w 10
  email-analyzer analyze -F sample.eml -k ab_xxx -t 15m  # 设置 15 分钟超时`,
		RunE: runAnalyze,
	}
	analyzeCmd.Flags().StringVarP(&baseURL, "url", "u", "http://localhost:18080", "AgentBox API 地址")
	analyzeCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API Key 或 JWT Token")
	analyzeCmd.Flags().StringVarP(&agentID, "agent", "a", "phishing-analyzer", "Agent ID")
	analyzeCmd.Flags().StringVarP(&outputFmt, "format", "f", "table", "输出格式: table/json/csv")
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径（默认输出到终端）")
	analyzeCmd.Flags().IntVarP(&workers, "workers", "w", 5, "并行任务数")
	analyzeCmd.Flags().DurationVarP(&taskTimeout, "timeout", "t", 10*time.Minute, "单任务超时时间 (如: 5m, 10m, 30m)")

	var dir, file string
	analyzeCmd.Flags().StringVarP(&dir, "dir", "d", "", "邮件目录路径")
	analyzeCmd.Flags().StringVarP(&file, "file", "F", "", "单个邮件文件路径")
	rootCmd.AddCommand(analyzeCmd)

	// login 命令
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "登录获取 API Token",
		RunE:  runLogin,
	}
	loginCmd.Flags().StringVarP(&baseURL, "url", "u", "http://localhost:18080", "AgentBox API 地址")
	rootCmd.AddCommand(loginCmd)

	// status 命令
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "检查 AgentBox 服务状态",
		RunE:  runStatus,
	}
	statusCmd.Flags().StringVarP(&baseURL, "url", "u", "http://localhost:18080", "AgentBox API 地址")
	rootCmd.AddCommand(statusCmd)

	// list 命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出历史分析任务",
		RunE:  runList,
	}
	listCmd.Flags().StringVarP(&baseURL, "url", "u", "http://localhost:18080", "AgentBox API 地址")
	listCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API Key 或 JWT Token")
	var limit int
	listCmd.Flags().IntVarP(&limit, "limit", "n", 20, "显示数量")
	rootCmd.AddCommand(listCmd)

	// version 命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("email-analyzer v0.1.0")
			fmt.Println("基于 AgentBox 的邮件钓鱼分析工具")
		},
	}
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	file, _ := cmd.Flags().GetString("file")

	if dir == "" && file == "" {
		return fmt.Errorf("请指定 --dir 或 --file 参数")
	}

	if apiKey == "" {
		// 尝试从环境变量读取
		apiKey = os.Getenv("AGENTBOX_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("请指定 --api-key 或设置 AGENTBOX_API_KEY 环境变量")
		}
	}

	client := NewClient(baseURL, apiKey)

	// 收集文件列表
	var files []string
	var err error
	if file != "" {
		files = []string{file}
	} else {
		files, err = collectEmailFiles(dir)
		if err != nil {
			return fmt.Errorf("扫描目录失败: %w", err)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("未找到 .eml 文件")
	}

	fmt.Printf("📧 找到 %d 个邮件文件\n", len(files))

	// 执行分析
	analyzer := NewAnalyzer(client, agentID, workers, taskTimeout)
	results, err := analyzer.AnalyzeFiles(files)
	if err != nil {
		return fmt.Errorf("分析失败: %w", err)
	}

	// 输出结果
	return outputResults(results, outputFmt, outputFile)
}

func runLogin(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		// 如果无法隐藏密码（非终端），回退到普通输入
		password, _ := reader.ReadString('\n')
		passwordBytes = []byte(strings.TrimSpace(password))
	}
	fmt.Println() // 换行

	password := string(passwordBytes)

	client := NewClient(baseURL, "")
	token, err := client.Login(username, password)
	if err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	fmt.Println("\n✅ 登录成功！")
	fmt.Printf("Token: %s\n", token)
	fmt.Println("\n使用方式:")
	fmt.Printf("  export AGENTBOX_API_KEY=%s\n", token)
	fmt.Printf("  email-analyzer analyze --dir ./emails\n")

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	client := NewClient(baseURL, "")
	status, err := client.Health()
	if err != nil {
		return fmt.Errorf("❌ 服务不可用: %w", err)
	}

	fmt.Printf("✅ AgentBox 服务正常\n")
	fmt.Printf("   状态: %s\n", status.Status)
	if status.Uptime != "" {
		fmt.Printf("   运行时间: %s\n", status.Uptime)
	}

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	if apiKey == "" {
		apiKey = os.Getenv("AGENTBOX_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("请指定 --api-key 或设置 AGENTBOX_API_KEY 环境变量")
		}
	}

	limit, _ := cmd.Flags().GetInt("limit")
	client := NewClient(baseURL, apiKey)

	tasks, err := client.ListTasks(limit)
	if err != nil {
		return fmt.Errorf("获取任务列表失败: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("暂无分析任务")
		return nil
	}

	fmt.Println("\n📋 历史分析任务")
	fmt.Println(strings.Repeat("━", 80))
	fmt.Printf("%-36s  %-10s  %-15s  %s\n", "任务 ID", "状态", "Agent", "创建时间")
	fmt.Println(strings.Repeat("─", 80))

	for _, t := range tasks {
		status := formatTaskStatus(t.Status)
		created := t.CreatedAt.Format("2006-01-02 15:04")
		fmt.Printf("%-36s  %-10s  %-15s  %s\n", t.ID, status, t.AgentID, created)
	}

	fmt.Println(strings.Repeat("━", 80))
	fmt.Printf("共 %d 个任务\n", len(tasks))

	return nil
}

func formatTaskStatus(status string) string {
	switch status {
	case "completed":
		return "✅ 完成"
	case "running":
		return "⏳ 运行中"
	case "queued":
		return "⌛ 排队中"
	case "failed":
		return "❌ 失败"
	case "cancelled":
		return "⏹️  取消"
	default:
		return status
	}
}
