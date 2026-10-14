package platform

import (
	"os"
	"runtime"
)

func IsAndroid() bool {
	if runtime.GOOS == "android" {
		return true
	}
	if os.Getenv("TERMUX_VERSION") != "" {
		return true
	}
	if _, err := os.Stat("/data/data/com.termux"); err == nil {
		return true
	}
	if os.Getenv("ANDROID_ROOT") != "" {
		return true
	}
	return false
}
