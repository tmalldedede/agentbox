package runtime

import "time"

var builtinRuntimes = []*AgentRuntime{
	{
		ID:          "default",
		Name:        "Default",
		Description: "Standard runtime with 2 CPUs and 4GB memory",
		Image:       "agentbox/agent:v2",
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
		Image:       "agentbox/agent:v2",
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
		Image:       "agentbox/agent:v2",
		CPUs:        4.0,
		MemoryMB:    8192,
		Network:     "bridge",
		Privileged:  true,
		IsBuiltIn:   true,
		CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

// GetBuiltinRuntimes returns all built-in runtimes
func GetBuiltinRuntimes() []*AgentRuntime {
	return builtinRuntimes
}
