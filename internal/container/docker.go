package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
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
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := m.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	// 执行命令
	attachResp, err := m.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
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

// ExecStream 在容器中执行命令并流式返回输出
func (m *DockerManager) ExecStream(ctx context.Context, containerID string, cmd []string) (*ExecStream, error) {
	// 创建 exec 实例
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execResp, err := m.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	// 执行命令
	attachResp, err := m.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %w", err)
	}

	// 启动执行
	if err := m.client.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		attachResp.Close()
		return nil, fmt.Errorf("failed to start exec: %w", err)
	}

	done := make(chan struct{})

	return &ExecStream{
		ExecID: execResp.ID,
		Reader: attachResp.Conn,
		Done:   done,
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

// ListImages 列出所有镜像
func (m *DockerManager) ListImages(ctx context.Context) ([]*Image, error) {
	images, err := m.client.ImageList(ctx, image.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	// 获取所有容器以检查镜像是否在使用中
	containers, err := m.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// 构建镜像使用映射
	imageInUse := make(map[string]bool)
	for _, c := range containers {
		imageInUse[c.ImageID] = true
	}

	// Agent 镜像前缀列表
	agentImagePrefixes := []string{
		"anthropic/claude-code",
		"ghcr.io/openai/codex",
		"agentbox/",
	}

	result := make([]*Image, 0, len(images))
	for _, img := range images {
		// 检查是否是 Agent 镜像
		isAgentImage := false
		for _, tag := range img.RepoTags {
			for _, prefix := range agentImagePrefixes {
				if strings.HasPrefix(tag, prefix) {
					isAgentImage = true
					break
				}
			}
			if isAgentImage {
				break
			}
		}

		result = append(result, &Image{
			ID:           img.ID,
			Tags:         img.RepoTags,
			Size:         img.Size,
			Created:      img.Created,
			InUse:        imageInUse[img.ID],
			IsAgentImage: isAgentImage,
		})
	}

	return result, nil
}

// PullImage 拉取镜像
func (m *DockerManager) PullImage(ctx context.Context, imageName string) error {
	reader, err := m.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	// 消费输出直到完成
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to complete image pull: %w", err)
	}

	return nil
}

// RemoveImage 删除镜像
func (m *DockerManager) RemoveImage(ctx context.Context, imageID string) error {
	_, err := m.client.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force:         false,
		PruneChildren: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove image: %w", err)
	}
	return nil
}
