package platform

type SystemOps interface {
	Shutdown() error
	Reboot() error
	NotifyReady() error
	AcquireWakeLock() error
	ReleaseWakeLock() error
}

var Sys SystemOps
