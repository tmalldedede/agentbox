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
		c.JSON(http.StatusBadRequest, apperr.NewValidationError("invalid request", err))
		return
	}

	// Read credentials from Claude CLI
	cred, err := api.syncMgr.ReadClaudeCliCredentials()
	if err != nil {
		c.JSON(http.StatusNotFound, apperr.NewError(
			apperr.ErrCodeNotFound,
			"Claude Code CLI credentials not found",
			err,
		))
		return
	}

	// Get provider
	prov, err := api.providerMgr.Get(req.ProviderID)
	if err != nil {
		c.JSON(http.StatusNotFound, apperr.NewError(
			apperr.ErrCodeNotFound,
			"provider not found",
			err,
		))
		return
	}

	// Verify provider is Anthropic
	if prov.BaseURL != "https://api.anthropic.com" {
		c.JSON(http.StatusBadRequest, apperr.NewError(
			apperr.ErrCodeBadRequest,
			"provider is not Anthropic",
			nil,
		))
		return
	}

	// Create or update auth profile
	profileID := "anthropic:claude-cli"
	profileReq := &provider.CreateAuthProfileRequest{
		Priority:     0,
		Mode:         "oauth",
		IsEnabled:    true,
		OAuthAccess:  cred.Access,
		OAuthRefresh: cred.Refresh,
		OAuthExpires: time.UnixMilli(cred.Expires),
	}

	// Try to find existing profile
	profiles, _ := api.providerMgr.ListAuthProfiles(req.ProviderID)
	var existingProfile *provider.AuthProfile
	for _, p := range profiles {
		if p.KeyMasked == profileID || p.Mode == "oauth" {
			existingProfile = p
			break
		}
	}

	if existingProfile != nil {
		// Update existing
		updateReq := &provider.UpdateAuthProfileRequest{
			IsEnabled:    &profileReq.IsEnabled,
			OAuthAccess:  &profileReq.OAuthAccess,
			OAuthRefresh: &profileReq.OAuthRefresh,
			OAuthExpires: &profileReq.OAuthExpires,
		}
		if err := api.providerMgr.UpdateAuthProfile(req.ProviderID, existingProfile.ID, updateReq); err != nil {
			c.JSON(http.StatusInternalServerError, apperr.NewError(
				apperr.ErrCodeInternal,
				"failed to update auth profile",
				err,
			))
			return
		}
	} else {
		// Create new
		if _, err := api.providerMgr.AddAuthProfile(req.ProviderID, profileReq); err != nil {
			c.JSON(http.StatusInternalServerError, apperr.NewError(
				apperr.ErrCodeInternal,
				"failed to create auth profile",
				err,
			))
			return
		}
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
		c.JSON(http.StatusBadRequest, apperr.NewValidationError("invalid request", err))
		return
	}

	// Read credentials from Codex CLI
	cred, err := api.syncMgr.ReadCodexCliCredentials()
	if err != nil {
		c.JSON(http.StatusNotFound, apperr.NewError(
			apperr.ErrCodeNotFound,
			"Codex CLI credentials not found",
			err,
		))
		return
	}

	// Get provider
	prov, err := api.providerMgr.Get(req.ProviderID)
	if err != nil {
		c.JSON(http.StatusNotFound, apperr.NewError(
			apperr.ErrCodeNotFound,
			"provider not found",
			err,
		))
		return
	}

	// Verify provider is OpenAI
	if prov.Name != "OpenAI" && prov.BaseURL != "https://api.openai.com" {
		c.JSON(http.StatusBadRequest, apperr.NewError(
			apperr.ErrCodeBadRequest,
			"provider is not OpenAI",
			nil,
		))
		return
	}

	// Create or update auth profile
	profileReq := &provider.CreateAuthProfileRequest{
		Priority:     0,
		Mode:         "oauth",
		IsEnabled:    true,
		OAuthAccess:  cred.Access,
		OAuthRefresh: cred.Refresh,
		OAuthExpires: time.UnixMilli(cred.Expires),
	}

	// Try to find existing OAuth profile
	profiles, _ := api.providerMgr.ListAuthProfiles(req.ProviderID)
	var existingProfile *provider.AuthProfile
	for _, p := range profiles {
		if p.Mode == "oauth" {
			existingProfile = p
			break
		}
	}

	if existingProfile != nil {
		// Update existing
		updateReq := &provider.UpdateAuthProfileRequest{
			IsEnabled:    &profileReq.IsEnabled,
			OAuthAccess:  &profileReq.OAuthAccess,
			OAuthRefresh: &profileReq.OAuthRefresh,
			OAuthExpires: &profileReq.OAuthExpires,
		}
		if err := api.providerMgr.UpdateAuthProfile(req.ProviderID, existingProfile.ID, updateReq); err != nil {
			c.JSON(http.StatusInternalServerError, apperr.NewError(
				apperr.ErrCodeInternal,
				"failed to update auth profile",
				err,
			))
			return
		}
	} else {
		// Create new
		if _, err := api.providerMgr.AddAuthProfile(req.ProviderID, profileReq); err != nil {
			c.JSON(http.StatusInternalServerError, apperr.NewError(
				apperr.ErrCodeInternal,
				"failed to create auth profile",
				err,
			))
			return
		}
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
		c.JSON(http.StatusNotFound, apperr.NewError(
			apperr.ErrCodeNotFound,
			"provider not found",
			err,
		))
		return
	}

	// Verify provider is Anthropic
	if prov.BaseURL != "https://api.anthropic.com" {
		c.JSON(http.StatusBadRequest, apperr.NewError(
			apperr.ErrCodeBadRequest,
			"provider is not Anthropic",
			nil,
		))
		return
	}

	// Find OAuth profile
	profiles, err := api.providerMgr.ListAuthProfiles(providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apperr.NewError(
			apperr.ErrCodeInternal,
			"failed to list auth profiles",
			err,
		))
		return
	}

	var oauthProfile *provider.AuthProfile
	for _, p := range profiles {
		if p.Mode == "oauth" && p.OAuthAccess != "" && p.OAuthRefresh != "" {
			oauthProfile = p
			break
		}
	}

	if oauthProfile == nil {
		c.JSON(http.StatusNotFound, apperr.NewError(
			apperr.ErrCodeNotFound,
			"no OAuth profile found for this provider",
			nil,
		))
		return
	}

	// Write to Claude CLI
	expiresMs := oauthProfile.OAuthExpires.UnixMilli()
	if err := api.syncMgr.WriteClaudeCliCredentials(
		oauthProfile.OAuthAccess,
		oauthProfile.OAuthRefresh,
		expiresMs,
	); err != nil {
		c.JSON(http.StatusInternalServerError, apperr.NewError(
			apperr.ErrCodeInternal,
			"failed to write credentials to Claude CLI",
			err,
		))
		return
	}

	c.JSON(http.StatusOK, &SyncFromClaudeCliResponse{
		Success:   true,
		Message:   "Synced OAuth credentials to Claude Code CLI",
		ExpiresAt: oauthProfile.OAuthExpires,
	})
}
