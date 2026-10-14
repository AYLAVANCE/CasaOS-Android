package storage

import (
	"fmt"
	"os"
	"strings"
)

// TranslatePermissions converts chmod/chown intent to Android-compatible operations
func TranslatePermissions(path string, mode os.FileMode) error {
	if !isWritablePath(path) {
		return fmt.Errorf("cannot change permissions on %s: path is outside app sandbox", path)
	}
	return os.Chmod(path, mode)
}

func isWritablePath(path string) bool {
	home := os.Getenv("HOME")
	prefix := os.Getenv("PREFIX")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}

	cleanPath := path
	return strings.HasPrefix(cleanPath, home) || strings.HasPrefix(cleanPath, prefix)
}

// EnsureContainerWritable makes a path writable for containers
func EnsureContainerWritable(path string) error {
	return os.MkdirAll(path, 0755)
}
