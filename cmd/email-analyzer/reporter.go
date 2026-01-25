package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// outputResults 输出分析结果
func outputResults(results []AnalysisResult, format, outputFile string) error {
	var output string
	var err error

	switch strings.ToLower(format) {
	case "json":
		output, err = formatJSON(results)
	case "csv":
		output, err = formatCSV(results)
	default:
		output = formatTable(results)
	}

	if err != nil {
		return err
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(output), 0644)
	}

	fmt.Println(output)
	return nil
}

// formatTable 表格格式输出
func formatTable(results []AnalysisResult) string {
	var sb strings.Builder

	// 统计
	stats := calculateStats(results)

	// 表头
	sb.WriteString("\n")
	sb.WriteString("┌─────────────────────────────────────┬──────────┬───────┬─────────────────────────────────┐\n")
	sb.WriteString("│ 文件                                │ 风险等级 │ 评分  │ 威胁类型                        │\n")
	sb.WriteString("├─────────────────────────────────────┼──────────┼───────┼─────────────────────────────────┤\n")

	// 表体
	for _, r := range results {
		file := truncate(r.File, 35)
		level := formatRiskLevel(r.RiskLevel)
		score := fmt.Sprintf("%d", r.RiskScore)
		threats := truncate(strings.Join(r.Threats, ", "), 31)
		if len(r.Threats) == 0 {
			threats = "-"
		}

		sb.WriteString(fmt.Sprintf("│ %-35s │ %-8s │ %-5s │ %-31s │\n",
			file, level, score, threats))
	}

	sb.WriteString("└─────────────────────────────────────┴──────────┴───────┴─────────────────────────────────┘\n")

	// 统计摘要
	sb.WriteString("\n📊 统计: ")
	summaryParts := []string{}
	if stats["critical"] > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("🔴严重 %d", stats["critical"]))
	}
	if stats["high"] > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("🟠高 %d", stats["high"]))
	}
	if stats["medium"] > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("🟡中 %d", stats["medium"]))
	}
	if stats["low"] > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("🟢低 %d", stats["low"]))
	}
	if stats["safe"] > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("✅安全 %d", stats["safe"]))
	}
	if stats["unknown"] > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("⚪未知 %d", stats["unknown"]))
	}
	sb.WriteString(strings.Join(summaryParts, " | "))
	sb.WriteString("\n")

	return sb.String()
}

// formatJSON JSON 格式输出
func formatJSON(results []AnalysisResult) (string, error) {
	stats := calculateStats(results)

	report := BatchReport{
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		TotalFiles:  len(results),
		Completed:   countByStatus(results, "completed"),
		Failed:      countByStatus(results, "failed"),
		RiskSummary: stats,
		Results:     results,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// formatCSV CSV 格式输出
func formatCSV(results []AnalysisResult) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// 写入表头
	header := []string{"文件", "状态", "风险等级", "评分", "威胁类型", "IOC数量", "摘要", "错误"}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	// 写入数据
	for _, r := range results {
		threats := strings.Join(r.Threats, ";")
		summary := truncate(r.Summary, 100)

		row := []string{
			r.File,
			r.Status,
			r.RiskLevel,
			fmt.Sprintf("%d", r.RiskScore),
			threats,
			fmt.Sprintf("%d", len(r.IOCs)),
			summary,
			r.Error,
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	return sb.String(), writer.Error()
}

// calculateStats 计算统计信息
func calculateStats(results []AnalysisResult) map[string]int {
	stats := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"safe":     0,
		"unknown":  0,
	}

	for _, r := range results {
		level := strings.ToLower(r.RiskLevel)
		if _, ok := stats[level]; ok {
			stats[level]++
		} else {
			stats["unknown"]++
		}
	}

	return stats
}

// countByStatus 按状态统计数量
func countByStatus(results []AnalysisResult, status string) int {
	count := 0
	for _, r := range results {
		if r.Status == status {
			count++
		}
	}
	return count
}

// formatRiskLevel 格式化风险等级
func formatRiskLevel(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return "🔴 严重"
	case "high":
		return "🟠 高"
	case "medium":
		return "🟡 中"
	case "low":
		return "🟢 低"
	case "safe":
		return "✅ 安全"
	default:
		return "⚪ 未知"
	}
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	// 移除换行符
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")

	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
