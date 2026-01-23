package runtime

import "time"

// AgentRuntime defines a container runtime configuration for agents
type AgentRuntime struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Image       string    `json:"image"`
	CPUs        float64   `json:"cpus"`
	MemoryMB    int       `json:"memory_mb"`
	Network     string    `json:"network"`
	Privileged  bool      `json:"privileged"`
	IsBuiltIn   bool      `json:"is_built_in"`
	IsDefault   bool      `json:"is_default,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate validates the runtime configuration
func (r *AgentRuntime) Validate() error {
	if r.ID == "" {
		return ErrRuntimeIDRequired
	}
	if r.Name == "" {
		return ErrRuntimeNameRequired
	}
	if r.Image == "" {
		return ErrRuntimeImageRequired
	}
	if r.CPUs <= 0 {
		r.CPUs = 2.0
	}
	if r.MemoryMB <= 0 {
		r.MemoryMB = 4096
	}
	if r.Network == "" {
		r.Network = "bridge"
	}
	return nil
}
