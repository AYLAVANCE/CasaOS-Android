package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type DokiRuntime struct {
	binaryPath string
	dataDir    string
}

func NewDokiRuntime() *DokiRuntime {
	// Simple lookup for doki binary
	binPath := "doki"
	if path, err := exec.LookPath("doki"); err == nil {
		binPath = path
	} else {
		// Termux default prefix fallback
		prefix := os.Getenv("PREFIX")
		if prefix == "" {
			prefix = "/data/data/com.termux/files/usr"
		}
		fallback := filepath.Join(prefix, "bin", "doki")
		if _, err := os.Stat(fallback); err == nil {
			binPath = fallback
		}
	}

	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	
	return &DokiRuntime{
		binaryPath: binPath,
		dataDir:    filepath.Join(home, ".casaos", "containers"),
	}
}

func (d *DokiRuntime) Name() string {
	return "doki"
}

func (d *DokiRuntime) Version() (string, error) {
	cmd := exec.Command(d.binaryPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (d *DokiRuntime) IsAvailable() bool {
	_, err := exec.LookPath(d.binaryPath)
	return err == nil
}

func (d *DokiRuntime) Create(ctx context.Context, req ContainerCreateRequest) (string, error) {
	// Map to doki run/create args
	args := []string{"create"}
	if req.Name != "" {
		args = append(args, "--name", req.Name)
	}
	for _, p := range req.Ports {
		hostPort := p.HostPort
		if hostPort < 1024 {
			hostPort += 8000
		}
		args = append(args, "-p", fmt.Sprintf("%d:%d", hostPort, p.ContainerPort))
	}
	for _, v := range req.Volumes {
		ro := ""
		if v.ReadOnly {
			ro = ":ro"
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s%s", v.HostPath, v.ContainerPath, ro))
	}
	for k, v := range req.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	
	args = append(args, req.Image)

	cmd := exec.CommandContext(ctx, d.binaryPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("doki create failed: %v, output: %s", err, string(out))
	}
	
	// Doki create should output container ID or name
	return string(out), nil // Simplified for brevity
}

func (d *DokiRuntime) Start(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, d.binaryPath, "start", id)
	return cmd.Run()
}

func (d *DokiRuntime) Stop(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, d.binaryPath, "stop", id)
	return cmd.Run()
}

func (d *DokiRuntime) Remove(ctx context.Context, id string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, id)
	cmd := exec.CommandContext(ctx, d.binaryPath, args...)
	return cmd.Run()
}

func (d *DokiRuntime) List(ctx context.Context) ([]Container, error) {
	cmd := exec.CommandContext(ctx, d.binaryPath, "ps", "--all", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var containers []Container
	if len(out) > 0 {
		if err := json.Unmarshal(out, &containers); err != nil {
			// fallback parse if needed, skip for now
			return nil, nil 
		}
	}
	return containers, nil
}

func (d *DokiRuntime) Inspect(ctx context.Context, id string) (*Container, error) {
	cmd := exec.CommandContext(ctx, d.binaryPath, "inspect", id)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var c Container
	if err := json.Unmarshal(out, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (d *DokiRuntime) Logs(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, id)
	
	cmd := exec.CommandContext(ctx, d.binaryPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	// Return the pipe. Caller is responsible for closing, but since it's a pipe, 
	// closing it will also stop reading. For a full implementation, we'd wrap this to kill the cmd.
	return stdout, nil
}

func (d *DokiRuntime) PullImage(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, d.binaryPath, "pull", image)
	return cmd.Run()
}

func (d *DokiRuntime) ListImages(ctx context.Context) ([]ImageInfo, error) {
	cmd := exec.CommandContext(ctx, d.binaryPath, "images", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var images []ImageInfo
	if len(out) > 0 {
		if err := json.Unmarshal(out, &images); err != nil {
			return nil, nil
		}
	}
	return images, nil
}

func (d *DokiRuntime) RemoveImage(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, d.binaryPath, "rmi", id)
	return cmd.Run()
}
