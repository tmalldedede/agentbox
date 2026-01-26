package skill

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/tmalldedede/agentbox/internal/container"
)

// Resolver 依赖检查器
// 检查 Skill 运行所需的依赖是否满足
type Resolver struct {
	containerMgr container.Manager
}

// NewResolver 创建依赖检查器
func NewResolver(containerMgr container.Manager) *Resolver {
	return &Resolver{
		containerMgr: containerMgr,
	}
}

// MissingDeps 缺失的依赖
type MissingDeps struct {
	Bins    []string `json:"bins,omitempty"`     // 缺失的二进制
	AnyBins []string `json:"any_bins,omitempty"` // 缺失的任选二进制
	Env     []string `json:"env,omitempty"`      // 缺失的环境变量
	Config  []string `json:"config,omitempty"`   // 缺失的配置项
	Pip     []string `json:"pip,omitempty"`      // 缺失的 Python 包
	Npm     []string `json:"npm,omitempty"`      // 缺失的 Node.js 包
	OS      []string `json:"os,omitempty"`       // 不支持的操作系统
}

// IsEmpty 检查是否没有缺失依赖
func (m *MissingDeps) IsEmpty() bool {
	return len(m.Bins) == 0 && len(m.AnyBins) == 0 && len(m.Env) == 0 &&
		len(m.Config) == 0 && len(m.Pip) == 0 && len(m.Npm) == 0 && len(m.OS) == 0
}

// ConfigCheck 配置检查结果
type ConfigCheck struct {
	Path      string      `json:"path"`
	Value     interface{} `json:"value"`
	Satisfied bool        `json:"satisfied"`
}

// InstallOption 可用的安装选项
type InstallOption struct {
	ID    string      `json:"id"`
	Kind  InstallKind `json:"kind"`
	Label string      `json:"label"`
	Bins  []string    `json:"bins,omitempty"`
}

// SkillStatusEntry 完整的 Skill 状态报告（借鉴 Clawdbot）
type SkillStatusEntry struct {
	// 基本信息
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	FilePath    string `json:"file_path"`
	SkillKey    string `json:"skill_key"`
	PrimaryEnv  string `json:"primary_env,omitempty"`
	Emoji       string `json:"emoji,omitempty"`
	Homepage    string `json:"homepage,omitempty"`

	// 状态标志
	Always            bool `json:"always"`              // 始终包含
	Disabled          bool `json:"disabled"`            // 被禁用
	BlockedByAllowlist bool `json:"blocked_by_allowlist"` // 被白名单阻止
	Eligible          bool `json:"eligible"`            // 是否可用

	// 依赖要求
	Requirements struct {
		Bins    []string `json:"bins,omitempty"`
		AnyBins []string `json:"any_bins,omitempty"`
		Env     []string `json:"env,omitempty"`
		Config  []string `json:"config,omitempty"`
		OS      []string `json:"os,omitempty"`
	} `json:"requirements"`

	// 缺失依赖
	Missing struct {
		Bins    []string `json:"bins,omitempty"`
		AnyBins []string `json:"any_bins,omitempty"`
		Env     []string `json:"env,omitempty"`
		Config  []string `json:"config,omitempty"`
		OS      []string `json:"os,omitempty"`
	} `json:"missing"`

	// 配置检查
	ConfigChecks []ConfigCheck `json:"config_checks,omitempty"`

	// 安装选项
	Install []InstallOption `json:"install,omitempty"`
}

// CheckResult 检查结果
type CheckResult struct {
	SkillID   string       `json:"skill_id"`
	Satisfied bool         `json:"satisfied"`
	Missing   *MissingDeps `json:"missing,omitempty"`
	Error     string       `json:"error,omitempty"`
}

// SkillStatusReport 完整状态报告
type SkillStatusReport struct {
	WorkspaceDir    string              `json:"workspace_dir"`
	ManagedSkillsDir string             `json:"managed_skills_dir"`
	Skills          []*SkillStatusEntry `json:"skills"`
}

// Check 检查 Skill 依赖是否满足
// containerID: 目标容器 ID（如果为空，则在主机上检查）
func (r *Resolver) Check(ctx context.Context, skill *Skill, containerID string) (*CheckResult, error) {
	result := &CheckResult{
		SkillID:   skill.ID,
		Satisfied: true,
		Missing:   &MissingDeps{},
	}

	// 如果标记为 always，跳过依赖检查
	if skill.Always {
		return result, nil
	}

	if skill.Requirements == nil || !skill.Requirements.HasRequirements() {
		return result, nil
	}

	reqs := skill.Requirements

	// 检查操作系统
	if len(reqs.OS) > 0 {
		currentOS := getCurrentPlatform()
		if !containsString(reqs.OS, currentOS) {
			result.Missing.OS = reqs.OS
			result.Satisfied = false
		}
	}

	// 检查二进制依赖（全部需要）
	if len(reqs.Bins) > 0 {
		missing, err := r.checkBins(ctx, containerID, reqs.Bins)
		if err != nil {
			result.Error = fmt.Sprintf("failed to check bins: %v", err)
		}
		if len(missing) > 0 {
			result.Missing.Bins = missing
			result.Satisfied = false
		}
	}

	// 检查任选二进制（满足一个即可）
	if len(reqs.AnyBins) > 0 {
		found := false
		for _, bin := range reqs.AnyBins {
			missing, _ := r.checkBins(ctx, containerID, []string{bin})
			if len(missing) == 0 {
				found = true
				break
			}
		}
		if !found {
			result.Missing.AnyBins = reqs.AnyBins
			result.Satisfied = false
		}
	}

	// 检查环境变量
	if len(reqs.Env) > 0 {
		missing, err := r.checkEnv(ctx, containerID, reqs.Env)
		if err != nil {
			result.Error = fmt.Sprintf("failed to check env: %v", err)
		}
		if len(missing) > 0 {
			result.Missing.Env = missing
			result.Satisfied = false
		}
	}

	// 检查 Python 包
	if len(reqs.Pip) > 0 {
		missing, err := r.checkPip(ctx, containerID, reqs.Pip)
		if err != nil {
			log.Debug("pip check error", "error", err)
		}
		if len(missing) > 0 {
			result.Missing.Pip = missing
			result.Satisfied = false
		}
	}

	// 检查 Node.js 包
	if len(reqs.Npm) > 0 {
		missing, err := r.checkNpm(ctx, containerID, reqs.Npm)
		if err != nil {
			log.Debug("npm check error", "error", err)
		}
		if len(missing) > 0 {
			result.Missing.Npm = missing
			result.Satisfied = false
		}
	}

	// Config 检查需要外部 MCP Manager，暂时跳过
	if len(reqs.Config) > 0 {
		result.Missing.Config = reqs.Config // 标记为需要检查
	}

	return result, nil
}

// getCurrentPlatform 获取当前操作系统
func getCurrentPlatform() string {
	return runtime.GOOS // 返回 "darwin", "linux", "windows" 等
}

// containsString 检查字符串数组是否包含指定字符串
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

// CheckMultiple 批量检查多个 Skill
func (r *Resolver) CheckMultiple(ctx context.Context, skills []*Skill, containerID string) (map[string]*CheckResult, error) {
	results := make(map[string]*CheckResult)

	for _, s := range skills {
		result, err := r.Check(ctx, s, containerID)
		if err != nil {
			result = &CheckResult{
				SkillID:   s.ID,
				Satisfied: false,
				Error:     err.Error(),
			}
		}
		results[s.ID] = result
	}

	return results, nil
}

// checkBins 检查二进制是否存在
func (r *Resolver) checkBins(ctx context.Context, containerID string, bins []string) ([]string, error) {
	if containerID == "" {
		return r.checkBinsLocal(bins), nil
	}
	return r.checkBinsInContainer(ctx, containerID, bins)
}

// checkBinsLocal 在本地检查二进制
func (r *Resolver) checkBinsLocal(bins []string) []string {
	var missing []string
	for _, bin := range bins {
		// 使用 exec.LookPath 检查二进制是否存在于 PATH 中
		if _, err := exec.LookPath(bin); err != nil {
			missing = append(missing, bin)
		}
	}
	return missing
}

// checkBinsInContainer 在容器内检查二进制
func (r *Resolver) checkBinsInContainer(ctx context.Context, containerID string, bins []string) ([]string, error) {
	var missing []string

	for _, bin := range bins {
		cmd := []string{"sh", "-c", fmt.Sprintf("which %s > /dev/null 2>&1", bin)}
		result, err := r.containerMgr.Exec(ctx, containerID, cmd)
		if err != nil {
			return nil, err
		}
		if result.ExitCode != 0 {
			missing = append(missing, bin)
		}
	}

	return missing, nil
}

// checkEnv 检查环境变量是否设置
func (r *Resolver) checkEnv(ctx context.Context, containerID string, envVars []string) ([]string, error) {
	if containerID == "" {
		return r.checkEnvLocal(envVars), nil
	}
	return r.checkEnvInContainer(ctx, containerID, envVars)
}

// checkEnvLocal 在本地检查环境变量
func (r *Resolver) checkEnvLocal(envVars []string) []string {
	var missing []string
	for _, env := range envVars {
		// 支持 ENV=value 格式，只检查变量名
		envName := env
		if idx := strings.Index(env, "="); idx > 0 {
			envName = env[:idx]
		}
		if _, ok := lookupEnv(envName); !ok {
			missing = append(missing, envName)
		}
	}
	return missing
}

// lookupEnv 环境变量查找（可被测试 mock）
var lookupEnv = os.LookupEnv

// checkEnvInContainer 在容器内检查环境变量
func (r *Resolver) checkEnvInContainer(ctx context.Context, containerID string, envVars []string) ([]string, error) {
	var missing []string

	for _, env := range envVars {
		// 支持 ENV=value 格式
		envName := env
		if idx := strings.Index(env, "="); idx > 0 {
			envName = env[:idx]
		}

		cmd := []string{"sh", "-c", fmt.Sprintf("test -n \"$%s\"", envName)}
		result, err := r.containerMgr.Exec(ctx, containerID, cmd)
		if err != nil {
			return nil, err
		}
		if result.ExitCode != 0 {
			missing = append(missing, envName)
		}
	}

	return missing, nil
}

// checkPip 检查 Python 包是否安装
func (r *Resolver) checkPip(ctx context.Context, containerID string, packages []string) ([]string, error) {
	if containerID == "" {
		return r.checkPipLocal(packages), nil
	}
	return r.checkPipInContainer(ctx, containerID, packages)
}

// checkPipLocal 在本地检查 Python 包
func (r *Resolver) checkPipLocal(packages []string) []string {
	var missing []string
	for _, pkg := range packages {
		// 提取包名（去除版本约束）
		pkgName := extractPkgName(pkg)
		// 尝试 import 来检查包是否存在
		cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s", pkgName))
		if err := cmd.Run(); err != nil {
			missing = append(missing, pkg)
		}
	}
	return missing
}

// checkPipInContainer 在容器内检查 Python 包
func (r *Resolver) checkPipInContainer(ctx context.Context, containerID string, packages []string) ([]string, error) {
	var missing []string
	for _, pkg := range packages {
		pkgName := extractPkgName(pkg)
		cmd := []string{"python3", "-c", fmt.Sprintf("import %s", pkgName)}
		result, err := r.containerMgr.Exec(ctx, containerID, cmd)
		if err != nil {
			return nil, err
		}
		if result.ExitCode != 0 {
			missing = append(missing, pkg)
		}
	}
	return missing, nil
}

// extractPkgName 从包规范中提取包名（去除版本约束）
func extractPkgName(pkg string) string {
	pkgName := pkg
	for _, sep := range []string{">=", "<=", "==", ">", "<", "~=", "[", ";"} {
		if idx := strings.Index(pkg, sep); idx > 0 {
			pkgName = pkg[:idx]
			break
		}
	}
	// 将连字符转换为下划线（pip 包名规范）
	return strings.ReplaceAll(pkgName, "-", "_")
}

// checkNpm 检查 Node.js 包是否安装
func (r *Resolver) checkNpm(ctx context.Context, containerID string, packages []string) ([]string, error) {
	if containerID == "" {
		return r.checkNpmLocal(packages), nil
	}
	return r.checkNpmInContainer(ctx, containerID, packages)
}

// checkNpmLocal 在本地检查 Node.js 包
func (r *Resolver) checkNpmLocal(packages []string) []string {
	var missing []string
	for _, pkg := range packages {
		pkgName := extractNpmPkgName(pkg)
		// 使用 npm list 检查全局包
		cmd := exec.Command("sh", "-c", fmt.Sprintf("npm list -g %s 2>/dev/null | grep -q %s", pkgName, pkgName))
		if err := cmd.Run(); err != nil {
			missing = append(missing, pkg)
		}
	}
	return missing
}

// checkNpmInContainer 在容器内检查 Node.js 包
func (r *Resolver) checkNpmInContainer(ctx context.Context, containerID string, packages []string) ([]string, error) {
	var missing []string
	for _, pkg := range packages {
		pkgName := extractNpmPkgName(pkg)
		cmd := []string{"sh", "-c", fmt.Sprintf("npm list %s 2>/dev/null | grep -q %s", pkgName, pkgName)}
		result, err := r.containerMgr.Exec(ctx, containerID, cmd)
		if err != nil {
			return nil, err
		}
		if result.ExitCode != 0 {
			missing = append(missing, pkg)
		}
	}
	return missing, nil
}

// extractNpmPkgName 从 npm 包规范中提取包名
func extractNpmPkgName(pkg string) string {
	pkgName := pkg
	if idx := strings.Index(pkg, "@"); idx > 0 && idx < len(pkg)-1 {
		// 处理 @scope/package@version 格式
		if pkg[0] == '@' {
			// scoped package: @scope/pkg@version
			secondAt := strings.Index(pkg[1:], "@")
			if secondAt > 0 {
				pkgName = pkg[:secondAt+1]
			}
		} else {
			// regular package: pkg@version
			pkgName = pkg[:idx]
		}
	}
	return pkgName
}

// InstallMissing 安装缺失的依赖（可选功能）
func (r *Resolver) InstallMissing(ctx context.Context, containerID string, missing *MissingDeps) error {
	if missing == nil || missing.IsEmpty() {
		return nil
	}

	// 安装 pip 包
	if len(missing.Pip) > 0 {
		pkgs := strings.Join(missing.Pip, " ")
		cmd := []string{"pip3", "install", "--quiet"}
		cmd = append(cmd, missing.Pip...)
		log.Info("installing pip packages", "packages", pkgs)
		result, err := r.containerMgr.Exec(ctx, containerID, cmd)
		if err != nil {
			return fmt.Errorf("failed to install pip packages: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("pip install failed: %s", result.Stderr)
		}
	}

	// 安装 npm 包
	if len(missing.Npm) > 0 {
		pkgs := strings.Join(missing.Npm, " ")
		cmd := []string{"npm", "install", "-g", "--quiet"}
		cmd = append(cmd, missing.Npm...)
		log.Info("installing npm packages", "packages", pkgs)
		result, err := r.containerMgr.Exec(ctx, containerID, cmd)
		if err != nil {
			return fmt.Errorf("failed to install npm packages: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("npm install failed: %s", result.Stderr)
		}
	}

	// 二进制和环境变量无法自动安装
	if len(missing.Bins) > 0 {
		log.Warn("missing binaries cannot be auto-installed", "bins", missing.Bins)
	}
	if len(missing.Env) > 0 {
		log.Warn("missing env vars cannot be auto-set", "env", missing.Env)
	}

	return nil
}

// QuickCheck 快速检查（不执行容器命令，仅返回需要检查的项）
func (r *Resolver) QuickCheck(skill *Skill) *CheckResult {
	result := &CheckResult{
		SkillID:   skill.ID,
		Satisfied: true,
	}

	// 如果标记为 always，直接满足
	if skill.Always {
		return result
	}

	if skill.Requirements == nil || !skill.Requirements.HasRequirements() {
		return result
	}

	// 仅标记有依赖，实际检查需要调用 Check
	result.Missing = &MissingDeps{
		Bins:    skill.Requirements.Bins,
		AnyBins: skill.Requirements.AnyBins,
		Env:     skill.Requirements.Env,
		Config:  skill.Requirements.Config,
		Pip:     skill.Requirements.Pip,
		Npm:     skill.Requirements.Npm,
		OS:      skill.Requirements.OS,
	}
	result.Satisfied = false

	return result
}

// BuildStatus 构建完整的 Skill 状态报告（借鉴 Clawdbot）
func (r *Resolver) BuildStatus(ctx context.Context, skill *Skill, config *SkillConfig) *SkillStatusEntry {
	status := &SkillStatusEntry{
		Name:        skill.Name,
		Description: skill.Description,
		Source:      string(skill.Source),
		FilePath:    skill.SourcePath,
		SkillKey:    skill.ID,
		PrimaryEnv:  skill.PrimaryEnv,
		Emoji:       skill.Emoji,
		Homepage:    skill.Homepage,
		Always:      skill.Always,
	}

	// 检查是否被禁用
	if config != nil && !config.Enabled {
		status.Disabled = true
	}

	// 填充依赖要求
	if skill.Requirements != nil {
		status.Requirements.Bins = skill.Requirements.Bins
		status.Requirements.AnyBins = skill.Requirements.AnyBins
		status.Requirements.Env = skill.Requirements.Env
		status.Requirements.Config = skill.Requirements.Config
		status.Requirements.OS = skill.Requirements.OS
	}

	// 如果标记为 always，跳过缺失检查
	if skill.Always {
		status.Eligible = !status.Disabled && !status.BlockedByAllowlist
		return status
	}

	// 检查缺失依赖
	checkResult, _ := r.Check(ctx, skill, "")
	if checkResult != nil && checkResult.Missing != nil {
		status.Missing.Bins = checkResult.Missing.Bins
		status.Missing.AnyBins = checkResult.Missing.AnyBins
		status.Missing.Env = checkResult.Missing.Env
		status.Missing.Config = checkResult.Missing.Config
		status.Missing.OS = checkResult.Missing.OS
	}

	// 检查环境变量（考虑 config 覆盖）
	if config != nil && len(status.Missing.Env) > 0 {
		var stillMissing []string
		for _, envName := range status.Missing.Env {
			// 检查 config.Env 覆盖
			if _, ok := config.Env[envName]; ok {
				continue
			}
			// 检查 config.APIKey 对应 primaryEnv
			if config.APIKey != "" && skill.PrimaryEnv == envName {
				continue
			}
			stillMissing = append(stillMissing, envName)
		}
		status.Missing.Env = stillMissing
	}

	// 构建安装选项
	status.Install = r.buildInstallOptions(skill)

	// 判断是否可用
	status.Eligible = !status.Disabled && !status.BlockedByAllowlist &&
		len(status.Missing.Bins) == 0 &&
		len(status.Missing.AnyBins) == 0 &&
		len(status.Missing.Env) == 0 &&
		len(status.Missing.Config) == 0 &&
		len(status.Missing.OS) == 0

	return status
}

// buildInstallOptions 构建安装选项
func (r *Resolver) buildInstallOptions(skill *Skill) []InstallOption {
	if len(skill.Install) == 0 {
		return nil
	}

	currentOS := getCurrentPlatform()
	var options []InstallOption

	for i, spec := range skill.Install {
		// 过滤不支持当前平台的安装选项
		if len(spec.OS) > 0 && !containsString(spec.OS, currentOS) {
			continue
		}

		option := InstallOption{
			ID:   spec.ID,
			Kind: spec.Kind,
			Bins: spec.Bins,
		}

		// 设置 ID
		if option.ID == "" {
			option.ID = fmt.Sprintf("%s-%d", spec.Kind, i)
		}

		// 设置 Label
		if spec.Label != "" {
			option.Label = spec.Label
		} else {
			switch spec.Kind {
			case InstallKindBrew:
				if spec.Formula != "" {
					option.Label = fmt.Sprintf("Install %s (brew)", spec.Formula)
				}
			case InstallKindNode:
				if spec.Package != "" {
					option.Label = fmt.Sprintf("Install %s (npm)", spec.Package)
				}
			case InstallKindGo:
				if spec.Module != "" {
					option.Label = fmt.Sprintf("Install %s (go)", spec.Module)
				}
			case InstallKindUV, InstallKindPip:
				if spec.Package != "" {
					option.Label = fmt.Sprintf("Install %s (%s)", spec.Package, spec.Kind)
				}
			case InstallKindDownload:
				if spec.URL != "" {
					parts := strings.Split(spec.URL, "/")
					option.Label = fmt.Sprintf("Download %s", parts[len(parts)-1])
				}
			default:
				option.Label = "Run installer"
			}
		}

		options = append(options, option)
	}

	return options
}

// BuildStatusReport 构建完整的状态报告
func (r *Resolver) BuildStatusReport(ctx context.Context, skills []*Skill, configs map[string]*SkillConfig) *SkillStatusReport {
	report := &SkillStatusReport{
		Skills: make([]*SkillStatusEntry, 0, len(skills)),
	}

	for _, skill := range skills {
		var config *SkillConfig
		if configs != nil {
			config = configs[skill.ID]
		}
		status := r.BuildStatus(ctx, skill, config)
		report.Skills = append(report.Skills, status)
	}

	return report
}
