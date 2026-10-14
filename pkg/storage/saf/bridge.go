package saf

import (
    "encoding/json"
    "fmt"
    "io"
    "os/exec"
    "strings"
)

type FileEntry struct {
    Name     string `json:"name"`
    Type     string `json:"type"` // file or directory
    Size     int64  `json:"size,omitempty"`
    URI      string `json:"uri,omitempty"`
}

type SAFBridge struct{}

func NewSAFBridge() *SAFBridge {
    return &SAFBridge{}
}

func (s *SAFBridge) IsAvailable() bool {
    _, err := exec.LookPath("termux-saf-ls")
    return err == nil
}

func (s *SAFBridge) ListFiles(uri string) ([]FileEntry, error) {
    out, err := exec.Command("termux-saf-ls", uri).Output()
    if err != nil {
        return nil, fmt.Errorf("termux-saf-ls failed: %w", err)
    }
    var entries []FileEntry
    if err := json.Unmarshal(out, &entries); err != nil {
        // If not JSON, parse line-by-line
        lines := strings.Split(strings.TrimSpace(string(out)), "\n")
        for _, line := range lines {
            if line != "" {
                entries = append(entries, FileEntry{Name: line})
            }
        }
    }
    return entries, nil
}

func (s *SAFBridge) ReadFile(uri string) (io.ReadCloser, error) {
    cmd := exec.Command("termux-saf-read", uri)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }
    if err := cmd.Start(); err != nil {
        return nil, err
    }
    return &safReader{cmd: cmd, reader: stdout}, nil
}

func (s *SAFBridge) WriteFile(treeURI, displayName string, r io.Reader) error {
    cmd := exec.Command("termux-saf-write", treeURI, displayName)
    cmd.Stdin = r
    return cmd.Run()
}

func (s *SAFBridge) RequestManagedFolder() (string, error) {
    out, err := exec.Command("termux-saf-managed-folder").Output()
    if err != nil {
        return "", fmt.Errorf("failed to request folder: %w", err)
    }
    return strings.TrimSpace(string(out)), nil
}

type safReader struct {
    cmd    *exec.Cmd
    reader io.Reader
}

func (s *safReader) Read(p []byte) (n int, err error) {
    return s.reader.Read(p)
}

func (s *safReader) Close() error {
    return s.cmd.Wait()
}
