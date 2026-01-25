package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Progress 进度跟踪器
type Progress struct {
	mu        sync.Mutex
	total     int
	completed int
	failed    int
	current   map[string]string // file -> status
	startTime time.Time
	quiet     bool
}

// NewProgress 创建进度跟踪器
func NewProgress(total int, quiet bool) *Progress {
	return &Progress{
		total:     total,
		current:   make(map[string]string),
		startTime: time.Now(),
		quiet:     quiet,
	}
}

// Start 开始分析文件
func (p *Progress) Start(file string) {
	if p.quiet {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current[file] = "analyzing"
	p.render()
}

// Complete 完成文件分析
func (p *Progress) Complete(file string, result *AnalysisResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.current, file)
	p.completed++

	if p.quiet {
		return
	}

	// 显示结果
	emoji := p.riskEmoji(result.RiskLevel)
	fmt.Printf("\r\033[K%s %s → %s\n", emoji, filepath.Base(file), p.formatRisk(result.RiskLevel))
	p.render()
}

// Fail 分析失败
func (p *Progress) Fail(file string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.current, file)
	p.failed++

	if p.quiet {
		return
	}

	fmt.Printf("\r\033[K❌ %s → 失败: %v\n", filepath.Base(file), err)
	p.render()
}

// Done 完成所有分析
func (p *Progress) Done() {
	if p.quiet {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	elapsed := time.Since(p.startTime).Round(time.Second)
	fmt.Printf("\r\033[K")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Printf("✅ 完成: %d | ❌ 失败: %d | ⏱️  耗时: %s\n", p.completed, p.failed, elapsed)
}

// render 渲染进度条
func (p *Progress) render() {
	done := p.completed + p.failed
	percent := float64(done) / float64(p.total) * 100

	// 进度条
	barWidth := 30
	filled := int(float64(barWidth) * float64(done) / float64(p.total))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// 显示正在分析的文件
	var analyzing []string
	for file := range p.current {
		analyzing = append(analyzing, filepath.Base(file))
	}

	status := ""
	if len(analyzing) > 0 {
		if len(analyzing) > 3 {
			status = fmt.Sprintf("分析中: %s 等 %d 个...", analyzing[0], len(analyzing))
		} else {
			status = fmt.Sprintf("分析中: %s", strings.Join(analyzing, ", "))
		}
	}

	elapsed := time.Since(p.startTime).Round(time.Second)
	fmt.Printf("\r\033[K[%s] %.0f%% (%d/%d) | %s | %s",
		bar, percent, done, p.total, elapsed, status)
}

// riskEmoji 风险等级对应的 emoji
func (p *Progress) riskEmoji(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	case "safe":
		return "✅"
	default:
		return "⚪"
	}
}

// formatRisk 格式化风险等级
func (p *Progress) formatRisk(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return "严重风险"
	case "high":
		return "高风险"
	case "medium":
		return "中风险"
	case "low":
		return "低风险"
	case "safe":
		return "安全"
	default:
		return level
	}
}

// PrintHeader 打印头部信息
func PrintHeader(dir, agentID string, workers, totalFiles int) {
	fmt.Println()
	fmt.Println("📧 批量邮件钓鱼分析")
	fmt.Println(strings.Repeat("━", 50))
	if dir != "" {
		fmt.Printf("📁 目录: %s\n", dir)
	}
	fmt.Printf("🤖 Agent: %s\n", agentID)
	fmt.Printf("⚡ 并行: %d workers\n", workers)
	fmt.Printf("📋 文件: %d 个\n", totalFiles)
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()
}
