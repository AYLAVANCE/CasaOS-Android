//go:build android

package platform

import (
	"os"
	"path/filepath"
)

type AndroidPaths struct {
	prefix string
	home   string
}

func NewAndroidPaths() *AndroidPaths {
	prefix := os.Getenv("TERMUX_PREFIX")
	if prefix == "" {
		prefix = os.Getenv("PREFIX")
		if prefix == "" {
			prefix = "/data/data/com.termux/files/usr"
		}
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	return &AndroidPaths{
		prefix: prefix,
		home:   home,
	}
}

func (a *AndroidPaths) ConfigDir() string   { return filepath.Join(a.prefix, "etc", "casaos") }
func (a *AndroidPaths) DataDir() string     { return filepath.Join(a.home, ".casaos", "data") }
func (a *AndroidPaths) RuntimeDir() string  { return filepath.Join(a.home, ".casaos", "run") }
func (a *AndroidPaths) LogDir() string      { return filepath.Join(a.home, ".casaos", "log") }
func (a *AndroidPaths) DBDir() string       { return filepath.Join(a.home, ".casaos", "db") }
func (a *AndroidPaths) WWWDir() string      { return filepath.Join(a.home, ".casaos", "www") }
func (a *AndroidPaths) ShellDir() string    { return filepath.Join(a.prefix, "share", "casaos", "shell") }
func (a *AndroidPaths) StorageRoot() string { return filepath.Join(a.home, "storage", "shared", "CasaOS") }
func (a *AndroidPaths) UserDataDir() string { return filepath.Join(a.home, ".casaos", "conf") }

func Init() {
	Provider = NewAndroidPaths()
	Sys = NewAndroidSystem()
}
