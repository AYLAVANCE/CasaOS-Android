package storage

import (
    "os"
    "path/filepath"
    "time"
)

type StorageDevice struct {
    Name       string `json:"name"`
    Path       string `json:"path"`
    Type       string `json:"type"` // internal, sdcard, otg
    Available  bool   `json:"available"`
    TotalSize  int64  `json:"total_size,omitempty"`
    FreeSize   int64  `json:"free_size,omitempty"`
}

type StorageChangeCallback func(devices []StorageDevice)

type StorageDetector struct {
    home     string
    stopCh   chan struct{}
    callback StorageChangeCallback
}

func NewStorageDetector(home string, callback StorageChangeCallback) *StorageDetector {
    return &StorageDetector{home: home, stopCh: make(chan struct{}), callback: callback}
}

func (d *StorageDetector) DetectDevices() []StorageDevice {
    var devices []StorageDevice
    
    // Internal shared storage
    shared := filepath.Join(d.home, "storage", "shared")
    if info, err := os.Stat(shared); err == nil && info.IsDir() {
        devices = append(devices, StorageDevice{Name: "Internal Storage", Path: shared, Type: "internal", Available: true})
    }
    
    // SD Card
    sdcard := filepath.Join(d.home, "storage", "external-1")
    if info, err := os.Lstat(sdcard); err == nil && (info.Mode()&os.ModeSymlink != 0 || info.IsDir()) {
        target, _ := os.Readlink(sdcard)
        if target != "" {
            devices = append(devices, StorageDevice{Name: "SD Card", Path: sdcard, Type: "sdcard", Available: true})
        }
    }
    
    // Check for additional external storage
    storageDir := filepath.Join(d.home, "storage")
    entries, _ := os.ReadDir(storageDir)
    for _, entry := range entries {
        name := entry.Name()
        if name != "shared" && name != "external-1" && name != "downloads" && name != "dcim" && name != "music" && name != "movies" && name != "pictures" {
            fullPath := filepath.Join(storageDir, name)
            if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
                devices = append(devices, StorageDevice{Name: name, Path: fullPath, Type: "otg", Available: true})
            }
        }
    }
    
    return devices
}

func (d *StorageDetector) StartWatching(interval time.Duration) {
    go func() {
        lastDevices := d.DetectDevices()
        if d.callback != nil {
            d.callback(lastDevices)
        }
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for {
            select {
            case <-d.stopCh:
                return
            case <-ticker.C:
                currentDevices := d.DetectDevices()
                if devicesChanged(lastDevices, currentDevices) {
                    lastDevices = currentDevices
                    if d.callback != nil {
                        d.callback(currentDevices)
                    }
                }
            }
        }
    }()
}

func (d *StorageDetector) Stop() {
    close(d.stopCh)
}

func devicesChanged(a, b []StorageDevice) bool {
    if len(a) != len(b) { return true }
    aMap := make(map[string]bool)
    for _, d := range a { aMap[d.Path] = d.Available }
    for _, d := range b {
        if avail, ok := aMap[d.Path]; !ok || avail != d.Available { return true }
    }
    return false
}
