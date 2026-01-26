package oauth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("oauth")
}

// ClaudeCliCredential represents Claude Code CLI OAuth credentials
type ClaudeCliCredential struct {
	Type     string `json:"type"`     // "oauth" or "token"
	Provider string `json:"provider"` // "anthropic"
	Access   string `json:"access"`
	Refresh  string `json:"refresh,omitempty"` // OAuth only
	Token    string `json:"token,omitempty"`   // Token only
	Expires  int64  `json:"expires"`           // Unix timestamp (ms)
}

// CodexCliCredential represents Codex CLI OAuth credentials
type CodexCliCredential struct {
	Type      string `json:"type"`      // "oauth"
	Provider  string `json:"provider"`  // "openai-codex"
	Access    string `json:"access"`
	Refresh   string `json:"refresh"`
	Expires   int64  `json:"expires"`   // Unix timestamp (ms)
	AccountID string `json:"accountId"` // Optional
}

// ClaudeCliFileFormat represents ~/.claude/.credentials.json structure
type ClaudeCliFileFormat struct {
	ClaudeAiOauth struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken,omitempty"`
		ExpiresAt    int64  `json:"expiresAt"`
	} `json:"claudeAiOauth"`
}

// CodexCliFileFormat represents ~/.codex/auth.json structure
type CodexCliFileFormat struct {
	Tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		AccountID    string `json:"account_id,omitempty"`
	} `json:"tokens"`
	LastRefresh string `json:"last_refresh,omitempty"`
}

// SyncManager handles OAuth token sync with external CLIs
type SyncManager struct {
	platform string
}

// NewSyncManager creates a new OAuth sync manager
func NewSyncManager() *SyncManager {
	return &SyncManager{
		platform: runtime.GOOS,
	}
}

// ReadClaudeCliCredentials reads credentials from Claude Code CLI storage
func (m *SyncManager) ReadClaudeCliCredentials() (*ClaudeCliCredential, error) {
	// Try Keychain first on macOS
	if m.platform == "darwin" {
		cred, err := m.readClaudeCliKeychain()
		if err == nil && cred != nil {
			log.Info("read anthropic credentials from claude cli keychain", "type", cred.Type)
			return cred, nil
		}
		log.Debug("failed to read from keychain, falling back to file", "error", err)
	}

	// Fallback to file
	return m.readClaudeCliFile()
}

// readClaudeCliKeychain reads credentials from macOS Keychain
func (m *SyncManager) readClaudeCliKeychain() (*ClaudeCliCredential, error) {
	if m.platform != "darwin" {
		return nil, fmt.Errorf("keychain only available on macOS")
	}

	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read keychain: %w", err)
	}

	var data ClaudeCliFileFormat
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse keychain data: %w", err)
	}

	return m.parseClaudeCredentials(&data)
}

// readClaudeCliFile reads credentials from ~/.claude/.credentials.json
func (m *SyncManager) readClaudeCliFile() (*ClaudeCliCredential, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	credPath := filepath.Join(homeDir, ".claude", ".credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var fileData ClaudeCliFileFormat
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	log.Info("read anthropic credentials from claude cli file")
	return m.parseClaudeCredentials(&fileData)
}

// parseClaudeCredentials parses Claude CLI credential format
func (m *SyncManager) parseClaudeCredentials(data *ClaudeCliFileFormat) (*ClaudeCliCredential, error) {
	oauth := data.ClaudeAiOauth
	if oauth.AccessToken == "" || oauth.ExpiresAt == 0 {
		return nil, fmt.Errorf("invalid credentials: missing access token or expiry")
	}

	cred := &ClaudeCliCredential{
		Provider: "anthropic",
		Access:   oauth.AccessToken,
		Expires:  oauth.ExpiresAt,
	}

	if oauth.RefreshToken != "" {
		cred.Type = "oauth"
		cred.Refresh = oauth.RefreshToken
	} else {
		cred.Type = "token"
		cred.Token = oauth.AccessToken
	}

	return cred, nil
}

// WriteClaudeCliCredentials writes credentials back to Claude Code CLI storage
func (m *SyncManager) WriteClaudeCliCredentials(access, refresh string, expiresAt int64) error {
	// Try Keychain first on macOS
	if m.platform == "darwin" {
		err := m.writeClaudeCliKeychain(access, refresh, expiresAt)
		if err == nil {
			log.Info("wrote refreshed credentials to claude cli keychain", "expires", time.UnixMilli(expiresAt))
			return nil
		}
		log.Debug("failed to write to keychain, falling back to file", "error", err)
	}

	// Fallback to file
	return m.writeClaudeCliFile(access, refresh, expiresAt)
}

// writeClaudeCliKeychain writes credentials to macOS Keychain
func (m *SyncManager) writeClaudeCliKeychain(access, refresh string, expiresAt int64) error {
	if m.platform != "darwin" {
		return fmt.Errorf("keychain only available on macOS")
	}

	// Read existing keychain entry
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read existing keychain entry: %w", err)
	}

	var existingData ClaudeCliFileFormat
	if err := json.Unmarshal(output, &existingData); err != nil {
		return fmt.Errorf("failed to parse existing data: %w", err)
	}

	// Update credentials
	existingData.ClaudeAiOauth.AccessToken = access
	existingData.ClaudeAiOauth.RefreshToken = refresh
	existingData.ClaudeAiOauth.ExpiresAt = expiresAt

	newValue, err := json.Marshal(existingData)
	if err != nil {
		return fmt.Errorf("failed to marshal new data: %w", err)
	}

	// Escape single quotes for shell
	escapedValue := strings.ReplaceAll(string(newValue), "'", "'\"'\"'")

	// Update keychain
	cmd = exec.Command("security", "add-generic-password", "-U",
		"-s", "Claude Code-credentials",
		"-a", "Claude Code",
		"-w", escapedValue)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update keychain: %w", err)
	}

	return nil
}

// writeClaudeCliFile writes credentials to ~/.claude/.credentials.json
func (m *SyncManager) writeClaudeCliFile(access, refresh string, expiresAt int64) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	credPath := filepath.Join(homeDir, ".claude", ".credentials.json")

	// Read existing file
	data, err := os.ReadFile(credPath)
	if err != nil {
		return fmt.Errorf("credentials file does not exist: %w", err)
	}

	var fileData ClaudeCliFileFormat
	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to parse existing credentials: %w", err)
	}

	// Update credentials
	fileData.ClaudeAiOauth.AccessToken = access
	fileData.ClaudeAiOauth.RefreshToken = refresh
	fileData.ClaudeAiOauth.ExpiresAt = expiresAt

	// Write back
	newData, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(credPath, newData, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	log.Info("wrote refreshed credentials to claude cli file", "expires", time.UnixMilli(expiresAt))
	return nil
}

// ReadCodexCliCredentials reads credentials from Codex CLI storage
func (m *SyncManager) ReadCodexCliCredentials() (*CodexCliCredential, error) {
	// Try Keychain first on macOS
	if m.platform == "darwin" {
		cred, err := m.readCodexCliKeychain()
		if err == nil && cred != nil {
			log.Info("read openai-codex credentials from codex cli keychain")
			return cred, nil
		}
	}

	// Fallback to file
	return m.readCodexCliFile()
}

// readCodexCliKeychain reads Codex credentials from macOS Keychain
func (m *SyncManager) readCodexCliKeychain() (*CodexCliCredential, error) {
	if m.platform != "darwin" {
		return nil, fmt.Errorf("keychain only available on macOS")
	}

	// Compute Codex home path
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		homeDir, _ := os.UserHomeDir()
		codexHome = filepath.Join(homeDir, ".codex")
	}

	// Compute account hash (first 16 chars of sha256)
	// Simplified: just use fixed account for now
	account := "cli|default"

	cmd := exec.Command("security", "find-generic-password",
		"-s", "Codex Auth",
		"-a", account,
		"-w")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read keychain: %w", err)
	}

	var data struct {
		Tokens struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			AccountID    string `json:"account_id,omitempty"`
		} `json:"tokens"`
		LastRefresh interface{} `json:"last_refresh,omitempty"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse keychain data: %w", err)
	}

	if data.Tokens.AccessToken == "" || data.Tokens.RefreshToken == "" {
		return nil, fmt.Errorf("invalid credentials: missing tokens")
	}

	// Estimate expiry (1 hour from last refresh or now)
	expires := time.Now().Add(time.Hour).UnixMilli()

	return &CodexCliCredential{
		Type:      "oauth",
		Provider:  "openai-codex",
		Access:    data.Tokens.AccessToken,
		Refresh:   data.Tokens.RefreshToken,
		Expires:   expires,
		AccountID: data.Tokens.AccountID,
	}, nil
}

// readCodexCliFile reads credentials from ~/.codex/auth.json
func (m *SyncManager) readCodexCliFile() (*CodexCliCredential, error) {
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		homeDir, _ := os.UserHomeDir()
		codexHome = filepath.Join(homeDir, ".codex")
	}

	authPath := filepath.Join(codexHome, "auth.json")
	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	var fileData CodexCliFileFormat
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("failed to parse auth file: %w", err)
	}

	if fileData.Tokens.AccessToken == "" || fileData.Tokens.RefreshToken == "" {
		return nil, fmt.Errorf("invalid credentials: missing tokens")
	}

	// Use file mtime as expiry heuristic
	fileInfo, _ := os.Stat(authPath)
	expires := fileInfo.ModTime().Add(time.Hour).UnixMilli()

	log.Info("read openai-codex credentials from codex cli file")
	return &CodexCliCredential{
		Type:      "oauth",
		Provider:  "openai-codex",
		Access:    fileData.Tokens.AccessToken,
		Refresh:   fileData.Tokens.RefreshToken,
		Expires:   expires,
		AccountID: fileData.Tokens.AccountID,
	}, nil
}

// CheckClaudeCliAvailable checks if Claude Code CLI credentials are available
func (m *SyncManager) CheckClaudeCliAvailable() bool {
	_, err := m.ReadClaudeCliCredentials()
	return err == nil
}

// CheckCodexCliAvailable checks if Codex CLI credentials are available
func (m *SyncManager) CheckCodexCliAvailable() bool {
	_, err := m.ReadCodexCliCredentials()
	return err == nil
}
