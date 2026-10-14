package service

import (
	"os"
	"path/filepath"
)

type StorageUsage struct {
	AppDataUsage map[string]int64 `json:"app_data_usage"`
	TotalUsage   int64            `json:"total_usage"`
	Breakdown    []UsageEntry     `json:"breakdown"`
}

type UsageEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
	Type string `json:"type"` // container, appdata, cache, log
}

func CalculateUsage() (*StorageUsage, error) {
	usage := &StorageUsage{
		AppDataUsage: make(map[string]int64),
	}

	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}

	dirsToScan := []struct {
		Name string
		Path string
		Type string
	}{
		{"CasaOS Data", filepath.Join(home, ".casaos", "data"), "appdata"},
		{"CasaOS Logs", filepath.Join(home, ".casaos", "log"), "log"},
		{"Containers", filepath.Join(home, ".casaos", "containers"), "container"},
		{"Shared Storage", filepath.Join(home, "storage", "shared", "CasaOS"), "appdata"},
	}

	for _, d := range dirsToScan {
		size, err := dirSize(d.Path)
		if err == nil && size > 0 {
			usage.Breakdown = append(usage.Breakdown, UsageEntry{
				Name: d.Name,
				Path: d.Path,
				Size: size,
				Type: d.Type,
			})
			usage.TotalUsage += size
			usage.AppDataUsage[d.Name] = size
		}
	}

	return usage, nil
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore errors to continue walking
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
