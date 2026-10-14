//go:build android
package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// NotifyReadiness notifies the supervisor or systemd that the service is ready.
// On Android we just write a file to indicate readiness.
func NotifyReadiness() (bool, error) {
	// Simplified notification: create a readiness file based on PID
	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	readyPath := filepath.Join(home, ".casaos", "run", fmt.Sprintf("ready_%d.lock", os.Getpid()))
	f, err := os.Create(readyPath)
	if err != nil {
		return false, err
	}
	f.Close()
	return true, nil
}
