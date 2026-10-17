//go:build !android

package platform

import (
	"os"
	"path/filepath"
)

type LinuxPaths struct {
	isAndroid bool
	prefix    string
	home      string
}

func NewLinuxPaths() *LinuxPaths {
	isAndroid := IsAndroid()
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
	return &LinuxPaths{
		isAndroid: isAndroid,
		prefix:    prefix,
		home:      home,
	}
}

func (l *LinuxPaths) ConfigDir() string {
	if l.isAndroid {
		prefixConf := filepath.Join(l.prefix, "etc", "casaos")
		if _, err := os.Stat(filepath.Join(prefixConf, "casaos.conf")); err == nil {
			return prefixConf
		}
		homeConf := filepath.Join(l.home, ".casaos", "conf")
		if _, err := os.Stat(filepath.Join(homeConf, "casaos.conf")); err == nil {
			return homeConf
		}
		if err := os.MkdirAll(prefixConf, 0755); err == nil {
			return prefixConf
		}
		_ = os.MkdirAll(homeConf, 0755)
		return homeConf
	}
	return "/etc/casaos"
}

func (l *LinuxPaths) DataDir() string {
	if l.isAndroid {
		return filepath.Join(l.home, ".casaos")
	}
	return "/var/lib/casaos"
}

func (l *LinuxPaths) RuntimeDir() string {
	if l.isAndroid {
		return filepath.Join(l.home, ".casaos", "run")
	}
	return "/var/run/casaos"
}

func (l *LinuxPaths) LogDir() string {
	if l.isAndroid {
		return filepath.Join(l.home, ".casaos", "log")
	}
	return "/var/log/casaos"
}

func (l *LinuxPaths) DBDir() string {
	if l.isAndroid {
		return filepath.Join(l.home, ".casaos", "db")
	}
	return "/var/lib/casaos/db"
}

func (l *LinuxPaths) WWWDir() string {
	if l.isAndroid {
		return filepath.Join(l.home, ".casaos", "www")
	}
	return "/var/lib/casaos/www"
}

func (l *LinuxPaths) ShellDir() string {
	if l.isAndroid {
		return filepath.Join(l.prefix, "share", "casaos", "shell")
	}
	return "/usr/share/casaos/shell"
}

func (l *LinuxPaths) StorageRoot() string {
	if l.isAndroid {
		return filepath.Join(l.home, "storage", "shared", "CasaOS")
	}
	return "/DATA"
}

func (l *LinuxPaths) UserDataDir() string {
	if l.isAndroid {
		return filepath.Join(l.home, ".casaos", "conf")
	}
	return "/var/lib/casaos/conf"
}

func (l *LinuxPaths) MountDir() string {
	if l.isAndroid {
		mntDir := filepath.Join(l.home, ".casaos", "mnt")
		_ = os.MkdirAll(mntDir, 0755)
		return mntDir
	}
	return "/mnt"
}

func Init() {
	Provider = NewLinuxPaths()
	Sys = NewLinuxSystem()
}
