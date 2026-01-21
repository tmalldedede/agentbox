package smartagent

import "github.com/tmalldedede/agentbox/internal/apperr"

// Error definitions for smartagent package
var (
	ErrAgentIDRequired      = apperr.BadRequest("agent ID is required")
	ErrAgentNameRequired    = apperr.BadRequest("agent name is required")
	ErrAgentProfileRequired = apperr.BadRequest("agent profile_id is required")
	ErrAgentNotFound        = apperr.NotFound("agent")
	ErrAgentAlreadyExists   = apperr.Conflict("agent already exists")
	ErrAgentInactive        = apperr.BadRequest("agent is not active")
	ErrProfileNotFound      = apperr.BadRequest("referenced profile not found")
)
