package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

// DockerManager Docker 容器管理器实现
type DockerManager struct {
	client *client.Client
}

// NewDockerManager 创建 Docker 管理器
func NewDockerManager() (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerManager{client: cli}, nil
}

// Create 创建容器
func (m *DockerManager) Create(ctx context.Context, config *CreateConfig) (*Container, error) {
	// 构建环境变量
	env := make([]string, 0, len(config.Env))
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// 构建挂载配置
	mounts := make([]mount.Mount, 0, len(config.Mounts))
	for _, m := range config.Mounts {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// 构建资源限制
	resources := container.Resources{}
	if config.Resources.CPULimit > 0 {
		resources.NanoCPUs = int64(config.Resources.CPULimit * 1e9)
	}
	if config.Resources.MemoryLimit > 0 {
		resources.Memory = config.Resources.MemoryLimit
	}

	// 创建容器
	resp, err := m.client.ContainerCreate(ctx,
		&container.Config{
			Image:  config.Image,
			Cmd:    config.Cmd,
			Env:    env,
			Labels: config.Labels,
			Tty:    true,
		},
		&container.HostConfig{
			Mounts:      mounts,
			Resources:   resources,
			NetworkMode: container.NetworkMode(config.NetworkMode),
		},
		nil,
		nil,
		config.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return &Container{
		ID:     resp.ID,
		Name:   config.Name,
		Image:  config.Image,
		Status: StatusCreated,
		Labels: config.Labels,
	}, nil
}

// Start 启动容器
func (m *DockerManager) Start(ctx context.Context, containerID string) error {
	if err := m.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

// Stop 停止容器
func (m *DockerManager) Stop(ctx context.Context, containerID string) error {
	timeout := 10 // 10 秒超时
	if err := m.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

// Remove 删除容器
func (m *DockerManager) Remove(ctx context.Context, containerID string) error {
	if err := m.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

// Exec 在容器中执行命令
func (m *DockerManager) Exec(ctx context.Context, containerID string, cmd []string) (*ExecResult, error) {
	// 创建 exec 实例
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := m.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	// 执行命令
	attachResp, err := m.client.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attachResp.Close()

	// 读取输出
	var stdout, stderr bytes.Buffer
	_, err = io.Copy(&stdout, attachResp.Reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read exec output: %w", err)
	}

	// 获取退出码
	inspectResp, err := m.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec: %w", err)
	}

	return &ExecResult{
		ExitCode: inspectResp.ExitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

// Logs 获取容器日志
func (m *DockerManager) Logs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	}

	reader, err := m.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return reader, nil
}

// Inspect 获取容器信息
func (m *DockerManager) Inspect(ctx context.Context, containerID string) (*Container, error) {
	info, err := m.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	status := parseStatus(info.State.Status)

	return &Container{
		ID:      info.ID,
		Name:    strings.TrimPrefix(info.Name, "/"),
		Image:   info.Config.Image,
		Status:  status,
		Created: 0, // info.Created is string, needs parsing
		Labels:  info.Config.Labels,
	}, nil
}

// Close 关闭客户端
func (m *DockerManager) Close() error {
	return m.client.Close()
}

// parseStatus 解析容器状态
func parseStatus(status string) ContainerStatus {
	switch status {
	case "created":
		return StatusCreated
	case "running":
		return StatusRunning
	case "paused":
		return StatusPaused
	case "restarting":
		return StatusRestarting
	case "exited":
		return StatusExited
	case "removing":
		return StatusRemoving
	case "dead":
		return StatusDead
	default:
		return ContainerStatus(status)
	}
}

// Ping 测试 Docker 连接
func (m *DockerManager) Ping(ctx context.Context) error {
	_, err := m.client.Ping(ctx)
	return err
}

// ListContainers 列出所有 AgentBox 管理的容器
func (m *DockerManager) ListContainers(ctx context.Context) ([]*Container, error) {
	containers, err := m.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]*Container, 0)
	for _, c := range containers {
		// 只返回带有 agentbox 标签的容器
		if _, ok := c.Labels["agentbox.managed"]; ok {
			result = append(result, &Container{
				ID:      c.ID,
				Name:    strings.TrimPrefix(c.Names[0], "/"),
				Image:   c.Image,
				Status:  parseStatus(c.State),
				Created: c.Created,
				Labels:  c.Labels,
			})
		}
	}

	return result, nil
}
