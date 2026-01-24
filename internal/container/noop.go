package container

import (
	"context"
	"fmt"
	"io"
)

var errDockerUnavailable = fmt.Errorf("Docker is not available")

// NoopManager 无操作容器管理器（Docker 不可用时使用）
type NoopManager struct{}

func NewNoopManager() *NoopManager {
	return &NoopManager{}
}

func (m *NoopManager) Create(ctx context.Context, config *CreateConfig) (*Container, error) {
	return nil, errDockerUnavailable
}

func (m *NoopManager) Start(ctx context.Context, containerID string) error {
	return errDockerUnavailable
}

func (m *NoopManager) Stop(ctx context.Context, containerID string) error {
	return errDockerUnavailable
}

func (m *NoopManager) Remove(ctx context.Context, containerID string) error {
	return errDockerUnavailable
}

func (m *NoopManager) Exec(ctx context.Context, containerID string, cmd []string) (*ExecResult, error) {
	return nil, errDockerUnavailable
}

func (m *NoopManager) ExecStream(ctx context.Context, containerID string, cmd []string) (*ExecStream, error) {
	return nil, errDockerUnavailable
}

func (m *NoopManager) Logs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	return nil, errDockerUnavailable
}

func (m *NoopManager) Inspect(ctx context.Context, containerID string) (*Container, error) {
	return nil, errDockerUnavailable
}

func (m *NoopManager) ListContainers(ctx context.Context) ([]*Container, error) {
	return nil, errDockerUnavailable
}

func (m *NoopManager) ListImages(ctx context.Context) ([]*Image, error) {
	return nil, errDockerUnavailable
}

func (m *NoopManager) PullImage(ctx context.Context, imageName string) error {
	return errDockerUnavailable
}

func (m *NoopManager) RemoveImage(ctx context.Context, imageID string) error {
	return errDockerUnavailable
}

func (m *NoopManager) CopyToContainer(ctx context.Context, containerID string, srcPath string, dstPath string) error {
	return errDockerUnavailable
}

func (m *NoopManager) Ping(ctx context.Context) error {
	return errDockerUnavailable
}

func (m *NoopManager) Close() error {
	return nil
}
