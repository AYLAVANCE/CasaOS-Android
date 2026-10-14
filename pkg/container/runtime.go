package container

import (
	"context"
	"io"
)

// ContainerState represents a container's lifecycle state
type ContainerState string

const (
	StateCreated ContainerState = "created"
	StateRunning ContainerState = "running"
	StateStopped ContainerState = "stopped"
	StateError   ContainerState = "error"
	StatePulling ContainerState = "pulling"
)

// Container represents a running or stopped container
type Container struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	State   ContainerState    `json:"state"`
	Ports   []PortMapping     `json:"ports"`
	Volumes []VolumeMount     `json:"volumes"`
	Env     map[string]string `json:"env"`
	Created int64             `json:"created"`
	Labels  map[string]string `json:"labels,omitempty"`
}

// PortMapping defines a host:container port mapping
type PortMapping struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"` // tcp, udp
}

// VolumeMount defines a bind mount
type VolumeMount struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly      bool   `json:"read_only"`
}

// ContainerCreateRequest defines container creation parameters
type ContainerCreateRequest struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Ports   []PortMapping     `json:"ports"`
	Volumes []VolumeMount     `json:"volumes"`
	Env     map[string]string `json:"env"`
	Restart string            `json:"restart"` // always, on-failure, no
	Labels  map[string]string `json:"labels,omitempty"`
}

// ImageInfo represents a container image
type ImageInfo struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Size    int64    `json:"size"`
	Created int64    `json:"created"`
	Arch    string   `json:"arch"`
}

// Runtime is the abstraction over container engines
type Runtime interface {
	// Engine info
	Name() string
	Version() (string, error)
	IsAvailable() bool

	// Container lifecycle
	Create(ctx context.Context, req ContainerCreateRequest) (string, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Remove(ctx context.Context, id string, force bool) error
	List(ctx context.Context) ([]Container, error)
	Inspect(ctx context.Context, id string) (*Container, error)
	Logs(ctx context.Context, id string, follow bool) (io.ReadCloser, error)

	// Images
	PullImage(ctx context.Context, image string) error
	ListImages(ctx context.Context) ([]ImageInfo, error)
	RemoveImage(ctx context.Context, id string) error
}
