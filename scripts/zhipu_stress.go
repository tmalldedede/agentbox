// Zhipu API 压力测试工具
//
// 用法:
//   go run scripts/zhipu_stress.go                          # 默认测试
//   go run scripts/zhipu_stress.go -ladder                  # 阶梯压力测试
//   go run scripts/zhipu_stress.go -c 10 -n 50 -endpoint codex
//
// 参数:
//   -key       智谱 API Key (可选，自动从 provider 数据读取)
//   -c         并发数 (默认: 5)
//   -n         总请求数 (默认: 20)
//   -endpoint  测试端点: codex, claude, both (默认: both)
//   -timeout   单次请求超时秒数 (默认: 60)
//   -ladder    启用阶梯压力测试模式
//   -report    输出 Markdown 报告文件路径

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tmalldedede/agentbox/internal/provider"
)

// 端点配置
type EndpointConfig struct {
	Name    string
	BaseURL string
	Path    string
	Model   string
}

var endpoints = map[string]EndpointConfig{
	"codex": {
		Name:    "Codex (Coding)",
		BaseURL: "https://open.bigmodel.cn/api/coding/paas/v4",
		Path:    "/chat/completions",
		Model:   "glm-4.7",
	},
	"claude": {
		Name:    "Claude Code (Anthropic)",
		BaseURL: "https://open.bigmodel.cn/api/anthropic",
		Path:    "/v1/messages",
		Model:   "claude-sonnet-4-20250514",
	},
}

// 请求结果
type Result struct {
	Endpoint   string
	StatusCode int
	Latency    time.Duration
	Error      string
	Success    bool
}

// 统计数据
type Stats struct {
	Endpoint     string
	Concurrency  int
	Total        int
	Success      int
	Failed       int
	SuccessRate  float64
	MinLatency   time.Duration
	MaxLatency   time.Duration
	AvgLatency   time.Duration
	P50Latency   time.Duration
	P95Latency   time.Duration
	P99Latency   time.Duration
	TotalTime    time.Duration
	RPS          float64
	ErrorCounts  map[string]int
	StatusCounts map[int]int
}

// 阶梯测试结果
type LadderResult struct {
	Endpoint string
	Steps    []*Stats
}

// 测试配置
type TestConfig struct {
	APIKey        string
	Timeout       time.Duration
	Prompt        string
	MaxTokens     int
	TestTime      time.Time
	GoVersion     string
	Platform      string
}

var testConfig TestConfig

func main() {
	// 解析命令行参数
	apiKey := flag.String("key", "", "Zhipu API Key")
	concurrency := flag.Int("c", 5, "Concurrency level")
	totalRequests := flag.Int("n", 20, "Total number of requests")
	endpoint := flag.String("endpoint", "both", "Endpoint to test: codex, claude, both")
	timeout := flag.Int("timeout", 60, "Request timeout in seconds")
	ladder := flag.Bool("ladder", false, "Enable ladder stress test mode")
	reportPath := flag.String("report", "", "Output Markdown report file path")
	flag.Parse()

	// 获取 API Key
	if *apiKey == "" {
		providerDataDir := "/tmp/agentbox/workspaces/providers"
		encKey := "agentbox-default-encryption-key-32b"
		provMgr := provider.NewManager(providerDataDir, encKey)
		if key, err := provMgr.GetDecryptedKey("zhipu"); err == nil && key != "" {
			*apiKey = key
			fmt.Printf("从 provider 数据读取到 API Key (masked: %s...%s)\n", key[:4], key[len(key)-4:])
		}
	}

	if *apiKey == "" {
		*apiKey = os.Getenv("ZHIPU_API_KEY")
		if *apiKey != "" {
			fmt.Println("从 ZHIPU_API_KEY 环境变量读取 API Key")
		}
	}

	if *apiKey == "" {
		fmt.Println("Error: API key is required.")
		os.Exit(1)
	}

	// 初始化测试配置
	testConfig = TestConfig{
		APIKey:    *apiKey,
		Timeout:   time.Duration(*timeout) * time.Second,
		Prompt:    "What is 2+2? Reply with just the number.",
		MaxTokens: 10,
		TestTime:  time.Now(),
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// 确定要测试的端点
	var testEndpoints []string
	if *endpoint == "both" {
		testEndpoints = []string{"codex", "claude"}
	} else {
		if _, ok := endpoints[*endpoint]; !ok {
			fmt.Printf("Error: Unknown endpoint '%s'\n", *endpoint)
			os.Exit(1)
		}
		testEndpoints = []string{*endpoint}
	}

	var report strings.Builder

	if *ladder {
		// 阶梯压力测试
		results := runLadderTest(testEndpoints, *apiKey, *timeout)
		generateLadderReport(&report, results)
	} else {
		// 普通测试
		allStats := make(map[string]*Stats)
		for _, ep := range testEndpoints {
			stats := runTest(ep, *apiKey, *concurrency, *totalRequests, time.Duration(*timeout)*time.Second)
			allStats[ep] = stats
		}
		generateSimpleReport(&report, allStats, *concurrency, *totalRequests)
	}

	// 输出报告
	fmt.Println(report.String())

	// 保存报告文件
	if *reportPath != "" {
		if err := os.WriteFile(*reportPath, []byte(report.String()), 0644); err != nil {
			fmt.Printf("Error writing report: %v\n", err)
		} else {
			fmt.Printf("\n报告已保存到: %s\n", *reportPath)
		}
	}
}

func runLadderTest(testEndpoints []string, apiKey string, timeout int) map[string]*LadderResult {
	// 阶梯配置：并发数 -> 每个并发的请求数
	ladderSteps := []struct {
		Concurrency int
		Requests    int
	}{
		{1, 10},   // 预热
		{2, 20},
		{5, 30},
		{10, 50},
		{20, 100},
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║         智谱 API 阶梯压力测试 / Ladder Stress Test             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n阶梯配置:\n")
	for i, step := range ladderSteps {
		fmt.Printf("  Step %d: 并发=%d, 请求数=%d\n", i+1, step.Concurrency, step.Requests)
	}
	fmt.Println()

	results := make(map[string]*LadderResult)

	for _, epName := range testEndpoints {
		ep := endpoints[epName]
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("测试端点: %s\n", ep.Name)
		fmt.Printf("URL: %s%s\n", ep.BaseURL, ep.Path)
		fmt.Printf("Model: %s\n", ep.Model)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		lr := &LadderResult{Endpoint: epName}

		for i, step := range ladderSteps {
			fmt.Printf("\n[Step %d/%d] 并发: %d, 请求数: %d\n", i+1, len(ladderSteps), step.Concurrency, step.Requests)

			stats := runTest(epName, apiKey, step.Concurrency, step.Requests, time.Duration(timeout)*time.Second)
			stats.Concurrency = step.Concurrency
			lr.Steps = append(lr.Steps, stats)

			// 阶梯间休息 2 秒
			if i < len(ladderSteps)-1 {
				fmt.Println("休息 2 秒...")
				time.Sleep(2 * time.Second)
			}
		}

		results[epName] = lr
	}

	return results
}

func runTest(endpointName, apiKey string, concurrency, totalRequests int, timeout time.Duration) *Stats {
	ep := endpoints[endpointName]

	results := make([]Result, 0, totalRequests)
	var mu sync.Mutex
	var completed int64

	jobs := make(chan int, totalRequests)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: timeout}

			for range jobs {
				result := makeRequest(client, ep, apiKey)
				result.Endpoint = endpointName

				mu.Lock()
				results = append(results, result)
				mu.Unlock()

				done := atomic.AddInt64(&completed, 1)
				printProgress(int(done), totalRequests, result)
			}
		}()
	}

	startTime := time.Now()
	for i := 0; i < totalRequests; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	totalTime := time.Since(startTime)

	fmt.Printf("\n总耗时: %v\n", totalTime)

	stats := calculateStats(endpointName, results)
	stats.TotalTime = totalTime
	stats.RPS = float64(stats.Success) / totalTime.Seconds()

	return stats
}

func makeRequest(client *http.Client, ep EndpointConfig, apiKey string) Result {
	start := time.Now()

	var bodyBytes []byte
	isAnthropic := ep.BaseURL == "https://open.bigmodel.cn/api/anthropic"

	reqBody := map[string]interface{}{
		"model": ep.Model,
		"messages": []map[string]string{
			{"role": "user", "content": testConfig.Prompt},
		},
		"max_tokens": testConfig.MaxTokens,
	}
	bodyBytes, _ = json.Marshal(reqBody)

	req, err := http.NewRequest("POST", ep.BaseURL+ep.Path, bytes.NewReader(bodyBytes))
	if err != nil {
		return Result{
			Latency: time.Since(start),
			Error:   fmt.Sprintf("create request: %v", err),
			Success: false,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	if isAnthropic {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Latency: latency,
			Error:   fmt.Sprintf("request: %v", err),
			Success: false,
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := Result{
		StatusCode: resp.StatusCode,
		Latency:    latency,
		Success:    resp.StatusCode == 200,
	}

	if !result.Success {
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if msg, ok := errResp["error"].(map[string]interface{}); ok {
				if m, ok := msg["message"].(string); ok {
					result.Error = m
				}
			} else if msg, ok := errResp["error"].(string); ok {
				result.Error = msg
			} else if msg, ok := errResp["message"].(string); ok {
				result.Error = msg
			}
		}
		if result.Error == "" {
			result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 100))
		}
	}

	return result
}

func printProgress(done, total int, result Result) {
	status := "✓"
	extra := ""
	if !result.Success {
		status = "✗"
		extra = fmt.Sprintf(" [%d: %s]", result.StatusCode, truncate(result.Error, 40))
	}
	fmt.Printf("\r[%d/%d] %s %v%s", done, total, status, result.Latency.Round(time.Millisecond), extra)
	if done == total || !result.Success {
		fmt.Println()
	}
}

func calculateStats(endpoint string, results []Result) *Stats {
	stats := &Stats{
		Endpoint:     endpoint,
		Total:        len(results),
		ErrorCounts:  make(map[string]int),
		StatusCounts: make(map[int]int),
	}

	if len(results) == 0 {
		return stats
	}

	var latencies []time.Duration
	var totalLatency time.Duration

	for _, r := range results {
		if r.Success {
			stats.Success++
			latencies = append(latencies, r.Latency)
			totalLatency += r.Latency
		} else {
			stats.Failed++
			if r.Error != "" {
				stats.ErrorCounts[truncate(r.Error, 60)]++
			}
		}
		if r.StatusCode > 0 {
			stats.StatusCounts[r.StatusCode]++
		}
	}

	stats.SuccessRate = float64(stats.Success) / float64(stats.Total) * 100

	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

		stats.MinLatency = latencies[0]
		stats.MaxLatency = latencies[len(latencies)-1]
		stats.AvgLatency = totalLatency / time.Duration(len(latencies))

		p50Idx := len(latencies) * 50 / 100
		p95Idx := len(latencies) * 95 / 100
		p99Idx := len(latencies) * 99 / 100
		if p50Idx >= len(latencies) { p50Idx = len(latencies) - 1 }
		if p95Idx >= len(latencies) { p95Idx = len(latencies) - 1 }
		if p99Idx >= len(latencies) { p99Idx = len(latencies) - 1 }

		stats.P50Latency = latencies[p50Idx]
		stats.P95Latency = latencies[p95Idx]
		stats.P99Latency = latencies[p99Idx]
	}

	return stats
}

func generateLadderReport(w *strings.Builder, results map[string]*LadderResult) {
	w.WriteString("\n# 智谱 API 阶梯压力测试报告\n\n")
	w.WriteString(fmt.Sprintf("**测试时间**: %s\n\n", testConfig.TestTime.Format("2006-01-02 15:04:05")))

	// 测试环境
	w.WriteString("## 1. 测试环境\n\n")
	w.WriteString("| 项目 | 值 |\n")
	w.WriteString("|------|----|\n")
	w.WriteString(fmt.Sprintf("| 平台 | %s |\n", testConfig.Platform))
	w.WriteString(fmt.Sprintf("| Go 版本 | %s |\n", testConfig.GoVersion))
	w.WriteString(fmt.Sprintf("| 请求超时 | %v |\n", testConfig.Timeout))
	w.WriteString("\n")

	// 测试条件
	w.WriteString("## 2. 测试条件\n\n")
	w.WriteString("### 2.1 请求参数\n\n")
	w.WriteString("| 参数 | 值 |\n")
	w.WriteString("|------|----|\n")
	w.WriteString(fmt.Sprintf("| Prompt | `%s` |\n", testConfig.Prompt))
	w.WriteString(fmt.Sprintf("| max_tokens | %d |\n", testConfig.MaxTokens))
	w.WriteString("\n")

	w.WriteString("### 2.2 端点配置\n\n")
	w.WriteString("| 端点 | URL | 模型 |\n")
	w.WriteString("|------|-----|------|\n")
	for name, ep := range endpoints {
		w.WriteString(fmt.Sprintf("| %s | `%s%s` | %s |\n", ep.Name, ep.BaseURL, ep.Path, ep.Model))
		_ = name
	}
	w.WriteString("\n")

	w.WriteString("### 2.3 阶梯配置\n\n")
	w.WriteString("| 阶段 | 并发数 | 请求数 |\n")
	w.WriteString("|------|--------|--------|\n")
	steps := []struct{ c, n int }{{1, 10}, {2, 20}, {5, 30}, {10, 50}, {20, 100}}
	for i, s := range steps {
		w.WriteString(fmt.Sprintf("| Step %d | %d | %d |\n", i+1, s.c, s.n))
	}
	w.WriteString("\n")

	// 测试结果
	w.WriteString("## 3. 测试结果\n\n")

	for epName, lr := range results {
		ep := endpoints[epName]
		w.WriteString(fmt.Sprintf("### 3.%d %s\n\n", 1, ep.Name))

		// 汇总表格
		w.WriteString("#### 性能指标汇总\n\n")
		w.WriteString("| 并发 | 请求数 | 成功 | 失败 | 成功率 | 平均延迟 | P50 | P95 | P99 | RPS |\n")
		w.WriteString("|------|--------|------|------|--------|----------|-----|-----|-----|-----|\n")

		for _, stats := range lr.Steps {
			w.WriteString(fmt.Sprintf("| %d | %d | %d | %d | %.1f%% | %v | %v | %v | %v | %.2f |\n",
				stats.Concurrency,
				stats.Total,
				stats.Success,
				stats.Failed,
				stats.SuccessRate,
				stats.AvgLatency.Round(time.Millisecond),
				stats.P50Latency.Round(time.Millisecond),
				stats.P95Latency.Round(time.Millisecond),
				stats.P99Latency.Round(time.Millisecond),
				stats.RPS,
			))
		}
		w.WriteString("\n")

		// 错误统计
		hasErrors := false
		for _, stats := range lr.Steps {
			if len(stats.ErrorCounts) > 0 {
				hasErrors = true
				break
			}
		}

		if hasErrors {
			w.WriteString("#### 错误统计\n\n")
			w.WriteString("| 并发 | 错误类型 | 次数 |\n")
			w.WriteString("|------|----------|------|\n")
			for _, stats := range lr.Steps {
				for err, count := range stats.ErrorCounts {
					w.WriteString(fmt.Sprintf("| %d | %s | %d |\n", stats.Concurrency, err, count))
				}
			}
			w.WriteString("\n")
		}
	}

	// 健康评估
	w.WriteString("## 4. 健康状态评估\n\n")
	w.WriteString("| 端点 | 状态 | 最高稳定并发 | 备注 |\n")
	w.WriteString("|------|------|--------------|------|\n")

	for epName, lr := range results {
		ep := endpoints[epName]
		maxStableConcurrency := 0
		status := "🟢 健康"
		note := ""

		for _, stats := range lr.Steps {
			if stats.SuccessRate >= 95 {
				maxStableConcurrency = stats.Concurrency
			} else if stats.SuccessRate >= 80 {
				if note == "" {
					note = fmt.Sprintf("并发 %d 时出现少量错误 (%.1f%%)", stats.Concurrency, 100-stats.SuccessRate)
				}
			} else {
				if status == "🟢 健康" {
					status = "🟡 需关注"
				}
				if stats.SuccessRate < 50 {
					status = "🔴 不稳定"
				}
			}
		}

		if note == "" {
			note = "所有阶段正常"
		}

		w.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n", ep.Name, status, maxStableConcurrency, note))
	}
	w.WriteString("\n")

	// 结论
	w.WriteString("## 5. 结论\n\n")
	for epName, lr := range results {
		ep := endpoints[epName]
		lastStats := lr.Steps[len(lr.Steps)-1]

		w.WriteString(fmt.Sprintf("### %s\n\n", ep.Name))

		if lastStats.SuccessRate >= 95 {
			w.WriteString(fmt.Sprintf("- **状态**: 稳定\n"))
			w.WriteString(fmt.Sprintf("- **最大测试并发**: %d (成功率 %.1f%%)\n", lastStats.Concurrency, lastStats.SuccessRate))
		} else {
			w.WriteString(fmt.Sprintf("- **状态**: 存在限流\n"))
			// 找到最后一个稳定的并发数
			stableConcurrency := 1
			for _, s := range lr.Steps {
				if s.SuccessRate >= 95 {
					stableConcurrency = s.Concurrency
				}
			}
			w.WriteString(fmt.Sprintf("- **推荐并发**: %d\n", stableConcurrency))
		}
		w.WriteString(fmt.Sprintf("- **平均延迟**: %v (并发=%d)\n", lastStats.AvgLatency.Round(time.Millisecond), lastStats.Concurrency))
		w.WriteString("\n")
	}
}

func generateSimpleReport(w *strings.Builder, allStats map[string]*Stats, concurrency, totalRequests int) {
	w.WriteString("\n# 智谱 API 压力测试报告\n\n")
	w.WriteString(fmt.Sprintf("**测试时间**: %s\n\n", testConfig.TestTime.Format("2006-01-02 15:04:05")))

	w.WriteString("## 测试条件\n\n")
	w.WriteString("| 参数 | 值 |\n")
	w.WriteString("|------|----|\n")
	w.WriteString(fmt.Sprintf("| 并发数 | %d |\n", concurrency))
	w.WriteString(fmt.Sprintf("| 总请求数 | %d |\n", totalRequests))
	w.WriteString(fmt.Sprintf("| 超时 | %v |\n", testConfig.Timeout))
	w.WriteString(fmt.Sprintf("| Prompt | `%s` |\n", testConfig.Prompt))
	w.WriteString(fmt.Sprintf("| max_tokens | %d |\n", testConfig.MaxTokens))
	w.WriteString("\n")

	w.WriteString("## 测试结果\n\n")
	w.WriteString("| 端点 | 成功率 | 平均延迟 | P50 | P95 | P99 | RPS |\n")
	w.WriteString("|------|--------|----------|-----|-----|-----|-----|\n")

	for name, stats := range allStats {
		ep := endpoints[name]
		w.WriteString(fmt.Sprintf("| %s | %.1f%% | %v | %v | %v | %v | %.2f |\n",
			ep.Name,
			stats.SuccessRate,
			stats.AvgLatency.Round(time.Millisecond),
			stats.P50Latency.Round(time.Millisecond),
			stats.P95Latency.Round(time.Millisecond),
			stats.P99Latency.Round(time.Millisecond),
			stats.RPS,
		))
	}
	w.WriteString("\n")

	w.WriteString("## 健康状态\n\n")
	for name, stats := range allStats {
		ep := endpoints[name]
		status := "🟢 健康"
		if stats.SuccessRate < 50 {
			status = "🔴 严重问题"
		} else if stats.SuccessRate < 90 {
			status = "🟡 需要关注"
		}
		w.WriteString(fmt.Sprintf("- **%s**: %s (成功率: %.1f%%)\n", ep.Name, status, stats.SuccessRate))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
