//go:build android

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type LocalStorageAndroid struct {
}

type StorageVolume struct {
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	Type       string  `json:"type"` // internal, sdcard, otg, app_data
	TotalBytes uint64  `json:"total_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	UsedPct    float64 `json:"used_percent"`
	FSType     string  `json:"fs_type"`
	Mounted    bool    `json:"mounted"`
	Writable   bool    `json:"writable"`
}

func NewLocalStorageAndroid() *LocalStorageAndroid {
	return &LocalStorageAndroid{}
}

func (l *LocalStorageAndroid) GetVolumes() ([]StorageVolume, error) {
	var volumes []StorageVolume

	// 1. App internal storage ($HOME)
	home := os.Getenv("HOME")
	if home != "" {
		stat := getStatFS(home)
		volumes = append(volumes, StorageVolume{
			Name:       "Termux Internal",
			Path:       home,
			Type:       "app_data",
			TotalBytes: stat.Total,
			FreeBytes:  stat.Free,
			UsedBytes:  stat.Total - stat.Free,
			UsedPct:    calculateUsedPct(stat.Total, stat.Free),
			Mounted:    true,
			Writable:   true,
		})
	}

	// 2. Shared internal storage (~/storage/shared)
	shared := filepath.Join(home, "storage", "shared")
	if _, err := os.Stat(shared); err == nil {
		stat := getStatFS(shared)
		volumes = append(volumes, StorageVolume{
			Name:       "Internal Storage",
			Path:       shared,
			Type:       "internal",
			TotalBytes: stat.Total,
			FreeBytes:  stat.Free,
			UsedBytes:  stat.Total - stat.Free,
			UsedPct:    calculateUsedPct(stat.Total, stat.Free),
			Mounted:    true,
			Writable:   true,
		})
	}

	// 3. SD Card (~/storage/external-1)
	sdcard := filepath.Join(home, "storage", "external-1")
	if target, err := os.Readlink(sdcard); err == nil && target != "" {
		stat := getStatFS(sdcard)
		volumes = append(volumes, StorageVolume{
			Name:       "SD Card",
			Path:       sdcard,
			Type:       "sdcard",
			TotalBytes: stat.Total,
			FreeBytes:  stat.Free,
			UsedBytes:  stat.Total - stat.Free,
			UsedPct:    calculateUsedPct(stat.Total, stat.Free),
			Mounted:    true,
			Writable:   true,
		})
	}

	// 4. CasaOS data directory (fallback if platform.Provider not used)
	casaosData := filepath.Join(home, ".casaos", "data")
	stat := getStatFS(casaosData)
	volumes = append(volumes, StorageVolume{
		Name:       "CasaOS Data",
		Path:       casaosData,
		Type:       "app_data",
		TotalBytes: stat.Total,
		FreeBytes:  stat.Free,
		UsedBytes:  stat.Total - stat.Free,
		UsedPct:    calculateUsedPct(stat.Total, stat.Free),
		Mounted:    true,
		Writable:   true,
	})

	return volumes, nil
}

func calculateUsedPct(total, free uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(total-free) / float64(total) * 100
}

func getStatFS(path string) struct{ Total, Free, Avail uint64 } {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return struct{ Total, Free, Avail uint64 }{}
	}
	return struct{ Total, Free, Avail uint64 }{
		Total: stat.Blocks * uint64(stat.Bsize),
		Free:  stat.Bfree * uint64(stat.Bsize),
		Avail: stat.Bavail * uint64(stat.Bsize),
	}
}

func (l *LocalStorageAndroid) GetStorageHealth() map[string]interface{} {
	return map[string]interface{}{
		"status":  "unsupported",
		"message": "S.M.A.R.T. monitoring requires root access on Android",
	}
}

func (l *LocalStorageAndroid) FormatDisk(path string, fstype string) error {
	return fmt.Errorf("disk formatting is not supported on Android")
}

func (l *LocalStorageAndroid) MountDisk(source, target, fstype string) error {
	return fmt.Errorf("disk mounting requires root access on Android")
}
