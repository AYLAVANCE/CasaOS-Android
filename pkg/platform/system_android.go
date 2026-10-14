//go:build android

package platform

import (
	"errors"
	"os/exec"
)

type AndroidSystem struct{}

func NewAndroidSystem() *AndroidSystem {
	return &AndroidSystem{}
}

func (a *AndroidSystem) Shutdown() error {
	return errors.New("not supported on Android")
}

func (a *AndroidSystem) Reboot() error {
	return errors.New("not supported on Android")
}

func (a *AndroidSystem) NotifyReady() error {
	// Write ready status to a file or Unix socket
	return nil
}

func (a *AndroidSystem) AcquireWakeLock() error {
	cmd := exec.Command("termux-wake-lock")
	return cmd.Run()
}

func (a *AndroidSystem) ReleaseWakeLock() error {
	cmd := exec.Command("termux-wake-unlock")
	return cmd.Run()
}
