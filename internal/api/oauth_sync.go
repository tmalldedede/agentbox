package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/oauth"
	"github.com/tmalldedede/agentbox/internal/provider"
)

// OAuthSyncAPI handles OAuth token synchronization with external CLIs
type OAuthSyncAPI struct {
	syncMgr     *oauth.SyncManager
	providerMgr *provider.Manager
}

// NewOAuthSyncAPI creates a new OAuth sync API handler
func NewOAuthSyncAPI(syncMgr *oauth.SyncManager, providerMgr *provider.Manager) *OAuthSyncAPI {
	return &OAuthSyncAPI{
		syncMgr:     syncMgr,
		providerMgr: providerMgr,
	}
}

// RegisterRoutes registers OAuth sync routes
func (api *OAuthSyncAPI) RegisterRoutes(r *gin.RouterGroup) {
	oauth := r.Group("/oauth")
	{
		oauth.GET("/sync-status", api.GetSyncStatus)
		oauth.POST("/sync-from-claude-cli", api.SyncFromClaudeCli)
		oauth.POST("/sync-from-codex-cli", api.SyncFromCodexCli)
		oauth.POST("/sync-to-claude-cli/:provider_id", api.SyncToClaudeCli)
	}
}

// SyncStatusResponse represents OAuth sync status
type SyncStatusResponse struct {
	ClaudeCliAvailable bool       `json:"claude_cli_available"`
	CodexCliAvailable  bool       `json:"codex_cli_available"`
	LastSyncAt         *time.Time `json:"last_sync_at,omitempty"`
	Platform           string     `json:"platform"`
}

// GetSyncStatus checks availability of external CLI credentials
func (api *OAuthSyncAPI) GetSyncStatus(c *gin.Context) {
	status := &SyncStatusResponse{
		ClaudeCliAvailable: api.syncMgr.CheckClaudeCliAvailable(),
		CodexCliAvailable:  api.syncMgr.CheckCodexCliAvailable(),
		Platform:           "darwin", // Will be dynamic in production
	}

	c.JSON(http.StatusOK, status)
}

// SyncFromClaudeCliRequest represents sync request from Claude CLI
type SyncFromClaudeCliRequest struct {
	ProviderID string `json:"provider_id" binding:"required"`
}

// SyncFromClaudeCliResponse represents sync result
type SyncFromClaudeCliResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// SyncFromClaudeCli syncs OAuth credentials from Claude Code CLI to a provider
func (api *OAuthSyncAPI) SyncFromClaudeCli(c *gin.Context) {
	var req SyncFromClaudeCliRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation("invalid request: "+err.Error()))
		return
	}

	// Read credentials from Claude CLI
	cred, err := api.syncMgr.ReadClaudeCliCredentials()
	if err != nil {
		HandleError(c, apperr.NotFound("Claude Code CLI credentials not found: "+err.Error()))
		return
	}

	// Get provider
	prov, err := api.providerMgr.Get(req.ProviderID)
	if err != nil {
		HandleError(c, apperr.NotFound("provider not found: "+err.Error()))
		return
	}

	// Verify provider is Anthropic
	if prov.BaseURL != "https://api.anthropic.com" {
		HandleError(c, apperr.BadRequest("provider is not Anthropic"))
		return
	}

	// Use OAuth access token as the "API key" for now
	// Priority 0 = highest priority
	_, err = api.providerMgr.AddAuthProfile(req.ProviderID, cred.Access, 0)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to add auth profile"))
		return
	}

	c.JSON(http.StatusOK, &SyncFromClaudeCliResponse{
		Success:   true,
		Message:   "Synced OAuth credentials from Claude Code CLI",
		ExpiresAt: time.UnixMilli(cred.Expires),
	})
}

// SyncFromCodexCli syncs OAuth credentials from Codex CLI
func (api *OAuthSyncAPI) SyncFromCodexCli(c *gin.Context) {
	var req SyncFromClaudeCliRequest // Reuse same struct
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, apperr.Validation("invalid request: "+err.Error()))
		return
	}

	// Read credentials from Codex CLI
	cred, err := api.syncMgr.ReadCodexCliCredentials()
	if err != nil {
		HandleError(c, apperr.NotFound("Codex CLI credentials not found: "+err.Error()))
		return
	}

	// Get provider
	prov, err := api.providerMgr.Get(req.ProviderID)
	if err != nil {
		HandleError(c, apperr.NotFound("provider not found: "+err.Error()))
		return
	}

	// Verify provider is OpenAI
	if prov.Name != "OpenAI" && prov.BaseURL != "https://api.openai.com" {
		HandleError(c, apperr.BadRequest("provider is not OpenAI"))
		return
	}

	// Use OAuth access token as the "API key" for now
	_, err = api.providerMgr.AddAuthProfile(req.ProviderID, cred.Access, 0)
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to add auth profile"))
		return
	}

	c.JSON(http.StatusOK, &SyncFromClaudeCliResponse{
		Success:   true,
		Message:   "Synced OAuth credentials from Codex CLI",
		ExpiresAt: time.UnixMilli(cred.Expires),
	})
}

// SyncToClaudeCli writes refreshed tokens back to Claude Code CLI storage
func (api *OAuthSyncAPI) SyncToClaudeCli(c *gin.Context) {
	providerID := c.Param("provider_id")

	// Get provider
	prov, err := api.providerMgr.Get(providerID)
	if err != nil {
		HandleError(c, apperr.NotFound("provider not found: "+err.Error()))
		return
	}

	// Verify provider is Anthropic
	if prov.BaseURL != "https://api.anthropic.com" {
		HandleError(c, apperr.BadRequest("provider is not Anthropic"))
		return
	}

	// Read current Claude CLI credentials
	cred, err := api.syncMgr.ReadClaudeCliCredentials()
	if err != nil {
		HandleError(c, apperr.NotFound("Claude Code CLI credentials not found: "+err.Error()))
		return
	}

	// For now, we just verify the credentials exist
	// In a full implementation, we would:
	// 1. Get the OAuth profile from the provider
	// 2. Check if the token has been refreshed
	// 3. Write the new token back to CLI if it's different

	// Write back to Claude CLI (using existing credentials for now)
	if err := api.syncMgr.WriteClaudeCliCredentials(
		cred.Access,
		cred.Refresh,
		cred.Expires,
	); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to write credentials to Claude CLI"))
		return
	}

	c.JSON(http.StatusOK, &SyncFromClaudeCliResponse{
		Success:   true,
		Message:   "Synced OAuth credentials to Claude Code CLI",
		ExpiresAt: time.UnixMilli(cred.Expires),
	})
}
