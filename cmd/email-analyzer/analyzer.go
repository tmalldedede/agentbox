package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Analyzer 邮件分析器
type Analyzer struct {
	client   *Client
	agentID  string
	workers  int
	timeout  time.Duration
	progress *Progress
}

// NewAnalyzer 创建分析器
func NewAnalyzer(client *Client, agentID string, workers int, timeout time.Duration) *Analyzer {
	if timeout <= 0 {
		timeout = 10 * time.Minute // 默认 10 分钟
	}
	return &Analyzer{
		client:  client,
		agentID: agentID,
		workers: workers,
		timeout: timeout,
	}
}

// AnalyzeFiles 分析多个文件
func (a *Analyzer) AnalyzeFiles(files []string) ([]AnalysisResult, error) {
	a.progress = NewProgress(len(files), false)

	results := make([]AnalysisResult, len(files))
	resultsMu := sync.Mutex{}

	// 创建任务通道
	jobs := make(chan int, len(files))
	for i := range files {
		jobs <- i
	}
	close(jobs)

	// 启动 worker
	var wg sync.WaitGroup
	for w := 0; w < a.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				result := a.analyzeFile(files[i])
				resultsMu.Lock()
				results[i] = result
				resultsMu.Unlock()
			}
		}()
	}

	wg.Wait()
	a.progress.Done()

	return results, nil
}

// analyzeFile 分析单个文件
func (a *Analyzer) analyzeFile(filePath string) AnalysisResult {
	start := time.Now()
	result := AnalysisResult{
		File:   filepath.Base(filePath),
		Status: "failed",
	}

	a.progress.Start(filePath)

	// 1. 上传文件
	uploaded, err := a.client.UploadFile(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("上传失败: %v", err)
		a.progress.Fail(filePath, err)
		return result
	}

	// 2. 创建分析任务
	prompt := fmt.Sprintf(`请分析这封邮件文件 %s，判断是否为钓鱼邮件。

请按以下格式返回分析结果：

## 风险评估
- 风险等级: [critical/high/medium/low/safe]
- 风险评分: [0-100]

## 威胁类型
- [列出发现的威胁类型，如: 钓鱼链接、伪造发件人、恶意附件等]

## IOC (威胁指标)
- [列出发现的可疑 URL、域名、IP、哈希等]

## 分析摘要
[简要说明分析结论]`, uploaded.Name)

	task, err := a.client.CreateTask(&CreateTaskRequest{
		AgentID:     a.agentID,
		Prompt:      prompt,
		Attachments: []string{uploaded.ID},
		Timeout:     int(a.timeout.Seconds()),
	})
	if err != nil {
		result.Error = fmt.Sprintf("创建任务失败: %v", err)
		a.progress.Fail(filePath, err)
		return result
	}
	result.TaskID = task.ID

	// 3. 等待任务完成
	task, err = a.client.WaitTask(task.ID, a.timeout)
	if err != nil {
		result.Error = fmt.Sprintf("等待任务失败: %v", err)
		a.progress.Fail(filePath, err)
		return result
	}

	result.Duration = time.Since(start)

	// 4. 解析结果
	if task.Status == "completed" {
		result.Status = "completed"
		a.parseTaskResult(task, &result)
		a.progress.Complete(filePath, &result)
	} else {
		result.Status = task.Status
		result.Error = task.Error
		a.progress.Fail(filePath, fmt.Errorf(task.Error))
	}

	return result
}

// parseTaskResult 解析任务结果
func (a *Analyzer) parseTaskResult(task *Task, result *AnalysisResult) {
	if task.Result == nil {
		result.RiskLevel = "unknown"
		return
	}

	// 从 result 中提取文本内容
	var content string
	if output, ok := task.Result["output"].(string); ok {
		content = output
	} else if text, ok := task.Result["text"].(string); ok {
		content = text
	} else if message, ok := task.Result["message"].(string); ok {
		content = message
	}

	result.Summary = content

	// 解析风险等级
	result.RiskLevel = extractRiskLevel(content)
	result.RiskScore = extractRiskScore(content)
	result.Threats = extractThreats(content)
	result.IOCs = extractIOCs(content)
}

// extractRiskLevel 提取风险等级
func extractRiskLevel(content string) string {
	content = strings.ToLower(content)

	patterns := []struct {
		level   string
		matches []string
	}{
		{"critical", []string{"critical", "严重", "紧急", "极高"}},
		{"high", []string{"high", "高风险", "高危"}},
		{"medium", []string{"medium", "中风险", "中等"}},
		{"low", []string{"low", "低风险", "低危"}},
		{"safe", []string{"safe", "安全", "正常", "无风险"}},
	}

	for _, p := range patterns {
		for _, m := range p.matches {
			if strings.Contains(content, m) {
				return p.level
			}
		}
	}

	return "unknown"
}

// extractRiskScore 提取风险评分
func extractRiskScore(content string) int {
	// 匹配评分模式：风险评分: 85, score: 85, 85/100, 85分
	re := regexp.MustCompile(`(?:风险评分|score|评分)[:\s]*(\d{1,3})`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		var score int
		fmt.Sscanf(matches[1], "%d", &score)
		if score >= 0 && score <= 100 {
			return score
		}
	}

	// 根据风险等级估算评分
	level := extractRiskLevel(content)
	switch level {
	case "critical":
		return 95
	case "high":
		return 75
	case "medium":
		return 50
	case "low":
		return 25
	case "safe":
		return 5
	}
	return 0
}

// extractThreats 提取威胁类型
func extractThreats(content string) []string {
	var threats []string
	seen := make(map[string]bool)

	threatPatterns := []struct {
		pattern string
		threat  string
	}{
		{`钓鱼|phishing`, "钓鱼链接"},
		{`伪造|spoof|冒充`, "伪造发件人"},
		{`恶意附件|malicious attachment`, "恶意附件"},
		{`BEC|商业邮件|CEO欺诈`, "BEC攻击"},
		{`紧急|urgent|立即`, "紧急诱导"},
		{`密码|credential|登录`, "凭证窃取"},
		{`恶意链接|malicious link`, "恶意链接"},
		{`社会工程|social engineering`, "社会工程"},
	}

	contentLower := strings.ToLower(content)
	for _, p := range threatPatterns {
		re := regexp.MustCompile(`(?i)` + p.pattern)
		if re.MatchString(contentLower) && !seen[p.threat] {
			threats = append(threats, p.threat)
			seen[p.threat] = true
		}
	}

	return threats
}

// extractIOCs 提取 IOC
func extractIOCs(content string) []IOC {
	var iocs []IOC
	seen := make(map[string]bool)

	// URL 提取
	urlRe := regexp.MustCompile(`https?://[^\s\])"'<>]+`)
	urls := urlRe.FindAllString(content, -1)
	for _, u := range urls {
		if !seen[u] {
			iocs = append(iocs, IOC{Type: "url", Value: u, Risk: "suspicious"})
			seen[u] = true
		}
	}

	// 域名提取
	domainRe := regexp.MustCompile(`(?:域名|domain)[:\s]*([a-zA-Z0-9][-a-zA-Z0-9]*\.[a-zA-Z]{2,})`)
	domains := domainRe.FindAllStringSubmatch(content, -1)
	for _, m := range domains {
		if len(m) > 1 && !seen[m[1]] {
			iocs = append(iocs, IOC{Type: "domain", Value: m[1], Risk: "suspicious"})
			seen[m[1]] = true
		}
	}

	// IP 提取
	ipRe := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	ips := ipRe.FindAllStringSubmatch(content, -1)
	for _, m := range ips {
		if len(m) > 1 && !seen[m[1]] && !isPrivateIP(m[1]) {
			iocs = append(iocs, IOC{Type: "ip", Value: m[1], Risk: "suspicious"})
			seen[m[1]] = true
		}
	}

	return iocs
}

// isPrivateIP 检查是否为私有 IP
func isPrivateIP(ip string) bool {
	return strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "127.")
}

// collectEmailFiles 收集邮件文件
func collectEmailFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".eml" || ext == ".msg" {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
