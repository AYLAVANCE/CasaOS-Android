package container

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type QEMURuntime struct {
	vmDir     string
	sshPort   int
	apiPort   int
	isRunning bool
	binary    string
}

func NewQEMURuntime() *QEMURuntime {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	
	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	
	return &QEMURuntime{
		vmDir:   filepath.Join(home, ".casaos", "vm"),
		sshPort: 2222,
		apiPort: 2375,
		binary:  filepath.Join(prefix, "bin", "qemu-system-aarch64"),
	}
}

func (q *QEMURuntime) Name() string {
	return "qemu-alpine"
}

func (q *QEMURuntime) Version() (string, error) {
	cmd := exec.Command(q.binary, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (q *QEMURuntime) IsAvailable() bool {
	_, err := os.Stat(q.binary)
	return err == nil
}

// To implement the rest of Runtime interface via SSH, we would use the golang.org/x/crypto/ssh
// package or just exec 'ssh' commands to run docker/podman inside the VM.
// Providing stubs below.

func (q *QEMURuntime) runRemoteCommand(ctx context.Context, args ...string) ([]byte, error) {
	// ssh -p 2222 root@127.0.0.1 docker args...
	sshArgs := []string{"-p", fmt.Sprintf("%d", q.sshPort), "root@127.0.0.1", "docker"}
	sshArgs = append(sshArgs, args...)
	
	cmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	return cmd.CombinedOutput()
}

func (q *QEMURuntime) Create(ctx context.Context, req ContainerCreateRequest) (string, error) {
	return "", fmt.Errorf("not implemented for qemu yet")
}

func (q *QEMURuntime) Start(ctx context.Context, id string) error {
	_, err := q.runRemoteCommand(ctx, "start", id)
	return err
}

func (q *QEMURuntime) Stop(ctx context.Context, id string) error {
	_, err := q.runRemoteCommand(ctx, "stop", id)
	return err
}

func (q *QEMURuntime) Remove(ctx context.Context, id string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, id)
	_, err := q.runRemoteCommand(ctx, args...)
	return err
}

func (q *QEMURuntime) List(ctx context.Context) ([]Container, error) {
	return []Container{}, nil
}

func (q *QEMURuntime) Inspect(ctx context.Context, id string) (*Container, error) {
	return nil, fmt.Errorf("not implemented")
}

func (q *QEMURuntime) Logs(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (q *QEMURuntime) PullImage(ctx context.Context, image string) error {
	_, err := q.runRemoteCommand(ctx, "pull", image)
	return err
}

func (q *QEMURuntime) ListImages(ctx context.Context) ([]ImageInfo, error) {
	return []ImageInfo{}, nil
}

func (q *QEMURuntime) RemoveImage(ctx context.Context, id string) error {
	_, err := q.runRemoteCommand(ctx, "rmi", id)
	return err
}
