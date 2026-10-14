//go:build !android
package platform

import (
	"github.com/coreos/go-systemd/daemon"
)

// NotifyReadiness notifies the supervisor or systemd that the service is ready.
func NotifyReadiness() (bool, error) {
	return daemon.SdNotify(false, daemon.SdNotifyReady)
}
