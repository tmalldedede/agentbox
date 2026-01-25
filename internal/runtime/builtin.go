package runtime

import (
	"time"

	"github.com/tmalldedede/agentbox/internal/config"
)

// GetBuiltinRuntimes returns all built-in runtimes with images from config
func GetBuiltinRuntimes(cfg *config.Config) []*AgentRuntime {
	// 如果没有传入配置，使用默认配置
	if cfg == nil {
		cfg = config.Default()
	}

	return []*AgentRuntime{
		{
			ID:          "default",
			Name:        "Default",
			Description: "Standard runtime with 2 CPUs and 4GB memory",
			Image:       cfg.Runtime.DefaultImage,
			CPUs:        2.0,
			MemoryMB:    4096,
			Network:     "bridge",
			Privileged:  false,
			IsBuiltIn:   true,
			IsDefault:   true,
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:          "light",
			Name:        "Light",
			Description: "Lightweight runtime with 1 CPU and 2GB memory",
			Image:       cfg.Runtime.LightImage,
			CPUs:        1.0,
			MemoryMB:    2048,
			Network:     "bridge",
			Privileged:  false,
			IsBuiltIn:   true,
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:          "heavy",
			Name:        "Heavy",
			Description: "High-performance runtime with 4 CPUs and 8GB memory (privileged)",
			Image:       cfg.Runtime.HeavyImage,
			CPUs:        4.0,
			MemoryMB:    8192,
			Network:     "bridge",
			Privileged:  true,
			IsBuiltIn:   true,
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:          "binary-re",
			Name:        "Binary Reverse Engineering",
			Description: "Specialized runtime for binary analysis with radare2, ghidra, yara, pwntools (4 CPUs, 8GB RAM, privileged)",
			Image:       cfg.Runtime.BinaryREImage,
			CPUs:        4.0,
			MemoryMB:    8192,
			Network:     "bridge",
			Privileged:  true,
			IsBuiltIn:   true,
			CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
}
