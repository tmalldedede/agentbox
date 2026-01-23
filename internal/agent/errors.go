package agent

import "github.com/tmalldedede/agentbox/internal/apperr"

// Error definitions for agent package
var (
	ErrAgentIDRequired       = apperr.BadRequest("agent ID is required")
	ErrAgentNameRequired     = apperr.BadRequest("agent name is required")
	ErrAgentAdapterRequired  = apperr.BadRequest("agent adapter is required")
	ErrAgentInvalidAdapter   = apperr.BadRequest("agent adapter must be one of: claude-code, codex, opencode")
	ErrAgentProviderRequired = apperr.BadRequest("agent provider_id is required")
	ErrAgentNotFound         = apperr.NotFound("agent")
	ErrAgentAlreadyExists    = apperr.Conflict("agent already exists")
	ErrAgentInactive         = apperr.BadRequest("agent is not active")
	ErrProviderNotFound      = apperr.BadRequest("referenced provider not found")
	ErrRuntimeNotFound       = apperr.BadRequest("referenced runtime not found")
)
