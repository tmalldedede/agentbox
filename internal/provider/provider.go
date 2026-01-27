// Package provider defines API Provider configurations for AgentBox.
// Provider represents a pre-configured API endpoint (official, compatible, or aggregator).
package provider

import (
	"strings"
	"time"
)

// Provider represents an API provider configuration
type Provider struct {
	// Basic info
	ID          string `json:"id"`                      // e.g., "deepseek", "openrouter"
	Name        string `json:"name"`                    // e.g., "DeepSeek"
	Description string `json:"description,omitempty"`   // Provider description
	TemplateID  string `json:"template_id,omitempty"`   // Which template this was created from

	// Agent compatibility (which adapters this provider supports)
	Agents []string `json:"agents"` // subset of: "claude-code", "codex", "opencode"

	// Category
	Category ProviderCategory `json:"category"` // official | cn_official | aggregator | third_party

	// URLs
	WebsiteURL string `json:"website_url,omitempty"` // Provider website
	APIKeyURL  string `json:"api_key_url,omitempty"` // Where to get API key
	DocsURL    string `json:"docs_url,omitempty"`    // Documentation URL

	// API Configuration
	BaseURL string `json:"base_url,omitempty"` // API endpoint

	// Environment variables template
	EnvConfig map[string]string `json:"env_config,omitempty"` // e.g., {"ANTHROPIC_API_KEY": ""}

	// Model defaults
	DefaultModel  string   `json:"default_model,omitempty"`  // Default model name
	DefaultModels []string `json:"default_models,omitempty"` // Available models

	// UI
	Icon      string `json:"icon,omitempty"`       // Icon name or URL
	IconColor string `json:"icon_color,omitempty"` // Icon color (hex)

	// Flags
	IsBuiltIn  bool `json:"is_built_in"`            // Built-in provider
	IsPartner  bool `json:"is_partner,omitempty"`   // Partner provider (featured)
	RequiresAK bool `json:"requires_ak,omitempty"`  // Requires API key
	IsEnabled  bool `json:"is_enabled"`             // Is enabled

	// API Key management (merged from Credential)
	APIKeyMasked    string     `json:"api_key_masked,omitempty"`    // Masked display
	IsConfigured    bool       `json:"is_configured"`               // Whether key is configured
	IsValid         bool       `json:"is_valid"`                    // Whether key is validated
	LastValidatedAt *time.Time `json:"last_validated_at,omitempty"` // Last validation time
}

// ProviderKeyData 存储 Provider 的 API Key 数据（持久化用）
type ProviderKeyData struct {
	ProviderID      string     `json:"provider_id"`
	EncryptedKey    string     `json:"encrypted_key"`
	KeyMasked       string     `json:"key_masked"`
	IsValid         bool       `json:"is_valid"`
	LastValidatedAt *time.Time `json:"last_validated_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ProviderCategory defines the category of a provider
type ProviderCategory string

const (
	// CategoryOfficial - Official providers (Anthropic, OpenAI)
	CategoryOfficial ProviderCategory = "official"

	// CategoryCNOfficial - Chinese official providers (DeepSeek, Zhipu, Qwen, etc.)
	CategoryCNOfficial ProviderCategory = "cn_official"

	// CategoryAggregator - API aggregators (OpenRouter, AiHubMix)
	CategoryAggregator ProviderCategory = "aggregator"

	// CategoryThirdParty - Third-party compatible providers
	CategoryThirdParty ProviderCategory = "third_party"
)

// Agent (adapter) constants
const (
	AgentClaudeCode = "claude-code"
	AgentCodex      = "codex"
	AgentOpenCode   = "opencode"
)

// AllAgents is the list of all supported agents
var AllAgents = []string{AgentClaudeCode, AgentCodex, AgentOpenCode}

// Validate validates the Provider configuration
func (p *Provider) Validate() error {
	if p.ID == "" {
		return ErrProviderIDRequired
	}
	if p.Name == "" {
		return ErrProviderNameRequired
	}
	if len(p.Agents) == 0 {
		return ErrProviderAgentRequired
	}
	return nil
}

// SupportsAgent checks if the provider supports the given agent
func (p *Provider) SupportsAgent(agent string) bool {
	for _, a := range p.Agents {
		if a == agent {
			return true
		}
	}
	return false
}

// GetEnvVars returns environment variables for this provider with the given API key
func (p *Provider) GetEnvVars(apiKey string) map[string]string {
	env := make(map[string]string)

	// Copy template env config
	for k, v := range p.EnvConfig {
		if v == "" && apiKey != "" {
			// Fill in API key placeholder
			env[k] = apiKey
		} else {
			env[k] = v
		}
	}

	// Set base URL if provided
	if p.BaseURL != "" {
		// For Claude Code compatible providers
		if p.SupportsAgent(AgentClaudeCode) {
			env["ANTHROPIC_BASE_URL"] = p.BaseURL
		}
		// For Codex/OpenAI compatible providers
		if p.SupportsAgent(AgentCodex) {
			codexBaseURL := p.BaseURL
			// Codex 使用 OpenAI 兼容 API，需要将 zhipu 的 Anthropic 端点转换为 OpenAI 兼容端点
			// zhipu 的 BaseURL 是 /api/anthropic，但 Codex 需要 OpenAI 兼容端点
			// 优先使用 /api/paas/v4（通用端点），如果需要编码计划权限可使用 /api/coding/paas/v4
			// 检查 ID 或 TemplateID 是否为 zhipu，或者 BaseURL 包含智谱AI的特征
			isZhipu := p.ID == "zhipu" || p.TemplateID == "zhipu" || 
				(strings.Contains(codexBaseURL, "open.bigmodel.cn") && strings.Contains(codexBaseURL, "/api/anthropic"))
			if isZhipu && strings.Contains(codexBaseURL, "/api/anthropic") {
				// 使用通用端点 /api/paas/v4（如果编码计划不可用，可以尝试 /api/coding/paas/v4）
				codexBaseURL = strings.Replace(codexBaseURL, "/api/anthropic", "/api/paas/v4", 1)
			}
			env["OPENAI_BASE_URL"] = codexBaseURL
		}
		// 注意：如果 Provider 同时支持多个 agent，会同时设置多个 BASE_URL
		// 但实际使用的 agent 由 Session 的 adapter 决定，只会创建一个容器
	}

	// Ensure OPENAI_API_KEY is set for Codex-compatible providers
	if apiKey != "" && p.SupportsAgent(AgentCodex) {
		if _, exists := env["OPENAI_API_KEY"]; !exists {
			env["OPENAI_API_KEY"] = apiKey
		}
	}

	return env
}
