//go:build android

package config

import (
	"path/filepath"
	"github.com/IceWhaleTech/CasaOS/pkg/platform"
)

func init() {
	platform.Init() // Ensure provider is initialized

	// Set paths to Termux paths before the main init parses the config
	CasaOSConfigFilePath = filepath.Join(platform.Provider.ConfigDir(), "casaos.conf")

	AppInfo.DBPath = platform.Provider.DataDir()
	AppInfo.LogPath = platform.Provider.LogDir()
	AppInfo.ShellPath = platform.Provider.ShellDir()
	AppInfo.UserDataPath = platform.Provider.UserDataDir()
	
	CommonInfo.RuntimePath = platform.Provider.RuntimeDir()
}
