//go:build !android

package platform

type LinuxPaths struct{}

func NewLinuxPaths() *LinuxPaths {
	return &LinuxPaths{}
}

func (l *LinuxPaths) ConfigDir() string   { return "/etc/casaos" }
func (l *LinuxPaths) DataDir() string     { return "/var/lib/casaos" }
func (l *LinuxPaths) RuntimeDir() string  { return "/var/run/casaos" }
func (l *LinuxPaths) LogDir() string      { return "/var/log/casaos" }
func (l *LinuxPaths) DBDir() string       { return "/var/lib/casaos/db" }
func (l *LinuxPaths) WWWDir() string      { return "/var/lib/casaos/www" }
func (l *LinuxPaths) ShellDir() string    { return "/usr/share/casaos/shell" }
func (l *LinuxPaths) StorageRoot() string { return "/DATA" }
func (l *LinuxPaths) UserDataDir() string { return "/var/lib/casaos/conf" }

func Init() {
	Provider = NewLinuxPaths()
	Sys = NewLinuxSystem()
}
