package network

import "runtime"

// GetListenAddress returns the platform-specific default listen address
func GetListenAddress() string {
	if runtime.GOOS == "android" {
		return "0.0.0.0:8080"
	}
	return "127.0.0.1:0"
}
