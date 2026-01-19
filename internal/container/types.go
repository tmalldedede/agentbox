package container

import (
	"context"
	"io"
)

// Manager 容器管理器接口
type Manager interface {
	// Container operations
	// Create 创建容器
	Create(ctx context.Context, config *CreateConfig) (*Container, error)

	// Start 启动容器
	Start(ctx context.Context, containerID string) error

	// Stop 停止容器
	Stop(ctx context.Context, containerID string) error

	// Remove 删除容器
	Remove(ctx context.Context, containerID string) error

	// Exec 在容器中执行命令
	Exec(ctx context.Context, containerID string, cmd []string) (*ExecResult, error)

	// ExecStream 在容器中执行命令并流式返回输出
	ExecStream(ctx context.Context, containerID string, cmd []string) (*ExecStream, error)

	// Logs 获取容器日志
	Logs(ctx context.Context, containerID string) (io.ReadCloser, error)

	// Inspect 获取容器信息
	Inspect(ctx context.Context, containerID string) (*Container, error)

	// ListContainers 列出所有 AgentBox 管理的容器
	ListContainers(ctx context.Context) ([]*Container, error)

	// Image operations
	// ListImages 列出所有镜像
	ListImages(ctx context.Context) ([]*Image, error)

	// PullImage 拉取镜像
	PullImage(ctx context.Context, imageName string) error

	// RemoveImage 删除镜像
	RemoveImage(ctx context.Context, imageID string) error

	// Ping 检查连接
	Ping(ctx context.Context) error

	// Close 关闭连接
	Close() error
}

// ExecStream 流式执行结果
type ExecStream struct {
	ExecID string         // Exec ID
	Reader io.ReadCloser  // 输出流
	Done   chan struct{}  // 完成信号
}

// CreateConfig 创建容器配置
type CreateConfig struct {
	Name        string            // 容器名称
	Image       string            // 镜像
	Cmd         []string          // 启动命令
	Env         map[string]string // 环境变量
	Mounts      []Mount           // 挂载配置
	Resources   ResourceConfig    // 资源限制
	NetworkMode string            // 网络模式
	Labels      map[string]string // 标签
}

// Mount 挂载配置
type Mount struct {
	Source   string // 宿主机路径
	Target   string // 容器内路径
	ReadOnly bool   // 是否只读
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	CPULimit    float64 // CPU 核心数 (如 2.0 = 2核)
	MemoryLimit int64   // 内存限制 (bytes)
}

// Container 容器信息
type Container struct {
	ID      string            // 容器 ID
	Name    string            // 容器名称
	Image   string            // 镜像
	Status  ContainerStatus   // 状态
	Created int64             // 创建时间 (Unix timestamp)
	Labels  map[string]string // 标签
}

// ContainerStatus 容器状态
type ContainerStatus string

const (
	StatusCreated    ContainerStatus = "created"
	StatusRunning    ContainerStatus = "running"
	StatusPaused     ContainerStatus = "paused"
	StatusRestarting ContainerStatus = "restarting"
	StatusExited     ContainerStatus = "exited"
	StatusRemoving   ContainerStatus = "removing"
	StatusDead       ContainerStatus = "dead"
)

// ExecResult 执行结果
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Image 镜像信息
type Image struct {
	ID        string   `json:"id"`
	Tags      []string `json:"tags"`
	Size      int64    `json:"size"`
	Created   int64    `json:"created"`
	InUse     bool     `json:"in_use"`
	IsAgentImage bool  `json:"is_agent_image"`
}
