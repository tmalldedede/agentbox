package provider

// BuiltinProviders contains all built-in provider configurations
var BuiltinProviders = []*Provider{
	// ==================== Claude Code Providers ====================

	// Official
	{
		ID:          "anthropic",
		Name:        "Anthropic",
		Description: "Anthropic Official API",
		Agent:       AgentClaudeCode,
		Category:    CategoryOfficial,
		WebsiteURL:  "https://www.anthropic.com",
		APIKeyURL:   "https://console.anthropic.com/settings/keys",
		DocsURL:     "https://docs.anthropic.com",
		BaseURL:     "", // Default, no override needed
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
		},
		DefaultModel:  "claude-sonnet-4-20250514",
		DefaultModels: []string{"claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-haiku-3-5-20241022"},
		Icon:          "anthropic",
		IconColor:     "#D97706",
		IsBuiltIn:     true,
		IsPartner:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},

	// Chinese Official Providers
	{
		ID:          "deepseek",
		Name:        "DeepSeek",
		Description: "DeepSeek API (Anthropic Compatible)",
		Agent:       AgentClaudeCode,
		Category:    CategoryCNOfficial,
		WebsiteURL:  "https://www.deepseek.com",
		APIKeyURL:   "https://platform.deepseek.com/api_keys",
		DocsURL:     "https://platform.deepseek.com/api-docs",
		BaseURL:     "https://api.deepseek.com/anthropic",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
		},
		DefaultModel:  "deepseek-chat",
		DefaultModels: []string{"deepseek-chat", "deepseek-coder", "deepseek-reasoner"},
		Icon:          "deepseek",
		IconColor:     "#0066FF",
		IsBuiltIn:     true,
		IsPartner:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
	{
		ID:          "zhipu",
		Name:        "Zhipu GLM",
		Description: "Zhipu AI GLM API (Anthropic Compatible)",
		Agent:       AgentClaudeCode,
		Category:    CategoryCNOfficial,
		WebsiteURL:  "https://www.zhipuai.cn",
		APIKeyURL:   "https://open.bigmodel.cn/usercenter/apikeys",
		DocsURL:     "https://open.bigmodel.cn/dev/howuse/introduction",
		BaseURL:     "https://open.bigmodel.cn/api/anthropic",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
		},
		DefaultModel:  "glm-4-plus",
		DefaultModels: []string{"glm-4-plus", "glm-4", "glm-4-flash"},
		Icon:          "zhipu",
		IconColor:     "#2563EB",
		IsBuiltIn:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
	{
		ID:          "qwen",
		Name:        "Qwen (Tongyi Qianwen)",
		Description: "Alibaba Qwen API (Anthropic Compatible)",
		Agent:       AgentClaudeCode,
		Category:    CategoryCNOfficial,
		WebsiteURL:  "https://tongyi.aliyun.com",
		APIKeyURL:   "https://dashscope.console.aliyun.com/apiKey",
		DocsURL:     "https://help.aliyun.com/zh/dashscope",
		BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
		},
		DefaultModel:  "qwen-max",
		DefaultModels: []string{"qwen-max", "qwen-plus", "qwen-turbo", "qwen-coder-plus"},
		Icon:          "qwen",
		IconColor:     "#6366F1",
		IsBuiltIn:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
	{
		ID:          "kimi",
		Name:        "Kimi (Moonshot)",
		Description: "Moonshot Kimi API (Anthropic Compatible)",
		Agent:       AgentClaudeCode,
		Category:    CategoryCNOfficial,
		WebsiteURL:  "https://www.moonshot.cn",
		APIKeyURL:   "https://platform.moonshot.cn/console/api-keys",
		DocsURL:     "https://platform.moonshot.cn/docs",
		BaseURL:     "https://api.moonshot.cn/anthropic",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
		},
		DefaultModel:  "moonshot-v1-auto",
		DefaultModels: []string{"moonshot-v1-auto", "moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k", "kimi-latest"},
		Icon:          "kimi",
		IconColor:     "#000000",
		IsBuiltIn:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
	{
		ID:          "minimax",
		Name:        "MiniMax",
		Description: "MiniMax API (Anthropic Compatible)",
		Agent:       AgentClaudeCode,
		Category:    CategoryCNOfficial,
		WebsiteURL:  "https://www.minimaxi.com",
		APIKeyURL:   "https://platform.minimaxi.com/user-center/basic-information/interface-key",
		DocsURL:     "https://platform.minimaxi.com/document",
		BaseURL:     "https://api.minimax.chat/v1",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
		},
		DefaultModel:  "MiniMax-Text-01",
		DefaultModels: []string{"MiniMax-Text-01", "abab6.5s-chat", "abab5.5-chat"},
		Icon:          "minimax",
		IconColor:     "#FF6B35",
		IsBuiltIn:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
	{
		ID:          "baichuan",
		Name:        "Baichuan",
		Description: "Baichuan AI API",
		Agent:       AgentClaudeCode,
		Category:    CategoryCNOfficial,
		WebsiteURL:  "https://www.baichuan-ai.com",
		APIKeyURL:   "https://platform.baichuan-ai.com/console/apikey",
		DocsURL:     "https://platform.baichuan-ai.com/docs",
		BaseURL:     "https://api.baichuan-ai.com/v1",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
		},
		DefaultModel:  "Baichuan4",
		DefaultModels: []string{"Baichuan4", "Baichuan3-Turbo", "Baichuan2-Turbo"},
		Icon:          "baichuan",
		IconColor:     "#00D4AA",
		IsBuiltIn:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},

	// Aggregators
	{
		ID:          "openrouter",
		Name:        "OpenRouter",
		Description: "OpenRouter API Aggregator",
		Agent:       AgentAll,
		Category:    CategoryAggregator,
		WebsiteURL:  "https://openrouter.ai",
		APIKeyURL:   "https://openrouter.ai/keys",
		DocsURL:     "https://openrouter.ai/docs",
		BaseURL:     "https://openrouter.ai/api/v1",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
			"OPENAI_API_KEY":    "",
		},
		DefaultModel:  "anthropic/claude-sonnet-4",
		DefaultModels: []string{"anthropic/claude-sonnet-4", "anthropic/claude-opus-4", "openai/gpt-4o", "google/gemini-2.0-flash-exp"},
		Icon:          "openrouter",
		IconColor:     "#6366F1",
		IsBuiltIn:     true,
		IsPartner:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
	{
		ID:          "aihubmix",
		Name:        "AiHubMix",
		Description: "AiHubMix API Aggregator",
		Agent:       AgentAll,
		Category:    CategoryAggregator,
		WebsiteURL:  "https://aihubmix.com",
		APIKeyURL:   "https://aihubmix.com/token",
		DocsURL:     "https://doc.aihubmix.com",
		BaseURL:     "https://aihubmix.com/v1",
		EnvConfig: map[string]string{
			"ANTHROPIC_API_KEY": "",
			"OPENAI_API_KEY":    "",
		},
		DefaultModel:  "claude-sonnet-4-20250514",
		DefaultModels: []string{"claude-sonnet-4-20250514", "gpt-4o", "deepseek-chat"},
		Icon:          "aihubmix",
		IconColor:     "#10B981",
		IsBuiltIn:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},

	// ==================== Codex (OpenAI) Providers ====================

	// Official
	{
		ID:          "openai",
		Name:        "OpenAI",
		Description: "OpenAI Official API",
		Agent:       AgentCodex,
		Category:    CategoryOfficial,
		WebsiteURL:  "https://openai.com",
		APIKeyURL:   "https://platform.openai.com/api-keys",
		DocsURL:     "https://platform.openai.com/docs",
		BaseURL:     "", // Default
		EnvConfig: map[string]string{
			"OPENAI_API_KEY": "",
		},
		DefaultModel:  "o3",
		DefaultModels: []string{"o3", "o4-mini", "gpt-4.1"},
		Icon:          "openai",
		IconColor:     "#10A37F",
		IsBuiltIn:     true,
		IsPartner:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
	{
		ID:          "azure-openai",
		Name:        "Azure OpenAI",
		Description: "Microsoft Azure OpenAI Service",
		Agent:       AgentCodex,
		Category:    CategoryOfficial,
		WebsiteURL:  "https://azure.microsoft.com/products/ai-services/openai-service",
		APIKeyURL:   "https://portal.azure.com",
		DocsURL:     "https://learn.microsoft.com/azure/ai-services/openai",
		BaseURL:     "https://YOUR_RESOURCE.openai.azure.com/openai",
		EnvConfig: map[string]string{
			"OPENAI_API_KEY":   "",
			"AZURE_OPENAI_KEY": "",
		},
		DefaultModel:  "gpt-4o",
		DefaultModels: []string{"gpt-4o", "gpt-4", "gpt-35-turbo"},
		Icon:          "azure",
		IconColor:     "#0078D4",
		IsBuiltIn:     true,
		RequiresAK:    true,
		IsEnabled:     true,
	},
}

// GetBuiltinProviders returns all built-in providers
func GetBuiltinProviders() []*Provider {
	return BuiltinProviders
}

// GetBuiltinProvidersByAgent returns built-in providers for a specific agent
func GetBuiltinProvidersByAgent(agent string) []*Provider {
	var result []*Provider
	for _, p := range BuiltinProviders {
		if p.SupportsAgent(agent) {
			result = append(result, p)
		}
	}
	return result
}

// GetBuiltinProviderByID returns a built-in provider by ID
func GetBuiltinProviderByID(id string) *Provider {
	for _, p := range BuiltinProviders {
		if p.ID == id {
			return p
		}
	}
	return nil
}
