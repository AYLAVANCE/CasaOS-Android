//go:build !android

package platform

import (
	"errors"
	"github.com/coreos/go-systemd/daemon"
)

type LinuxSystem struct{}

func NewLinuxSystem() *LinuxSystem {
	return &LinuxSystem{}
}

func (l *LinuxSystem) Shutdown() error {
	return errors.New("not implemented")
}

func (l *LinuxSystem) Reboot() error {
	return errors.New("not implemented")
}

func (l *LinuxSystem) NotifyReady() error {
	_, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	return err
}

func (l *LinuxSystem) AcquireWakeLock() error {
	return nil
}

func (l *LinuxSystem) ReleaseWakeLock() error {
	return nil
}
