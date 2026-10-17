package platform

import (
	"path/filepath"
	"strings"
)

// PathProvider defines platform-specific directory paths
type PathProvider interface {
	ConfigDir() string   // Where casaos.conf lives
	DataDir() string     // Persistent data storage
	RuntimeDir() string  // PID files, sockets, URLs
	LogDir() string      // Log files
	DBDir() string       // SQLite database
	WWWDir() string      // Static web assets
	ShellDir() string    // Shell scripts/helpers
	StorageRoot() string // User data storage root
	UserDataDir() string // User configuration data
	MountDir() string    // Storage mounts root
}

var Provider PathProvider

func SanitizeMountPath(mountPath string) string {
	if Provider == nil {
		Init()
	}
	mntRoot := Provider.MountDir()
	if strings.HasPrefix(mountPath, "/mnt/") {
		return filepath.Join(mntRoot, strings.TrimPrefix(mountPath, "/mnt/"))
	}
	if mountPath == "/mnt" {
		return mntRoot
	}
	if !filepath.IsAbs(mountPath) {
		return filepath.Join(mntRoot, mountPath)
	}
	return mountPath
}
