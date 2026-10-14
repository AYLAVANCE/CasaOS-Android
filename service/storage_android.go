//go:build android

package service

import (
    "os"
    "path/filepath"
    "time"
    
    "github.com/IceWhaleTech/CasaOS/pkg/platform"
    "github.com/IceWhaleTech/CasaOS/pkg/storage"
    safpkg "github.com/IceWhaleTech/CasaOS/pkg/storage/saf"
)

type AndroidStorageService struct {
    detector  *storage.StorageDetector
    safBridge *safpkg.SAFBridge
}

func NewAndroidStorageService() *AndroidStorageService {
    home := os.Getenv("HOME")
    if home == "" {
        home = "/data/data/com.termux/files/home"
    }
    
    svc := &AndroidStorageService{
        safBridge: safpkg.NewSAFBridge(),
    }
    
    svc.detector = storage.NewStorageDetector(home, func(devices []storage.StorageDevice) {
        // TODO: emit storage change event to MessageBus
    })
    
    // Start watching for storage changes every 5 seconds
    svc.detector.StartWatching(5 * time.Second)
    
    // Ensure CasaOS storage root exists
    os.MkdirAll(platform.Provider.StorageRoot(), 0755)
    
    return svc
}

func (s *AndroidStorageService) GetStorageDevices() []storage.StorageDevice {
    return s.detector.DetectDevices()
}

func (s *AndroidStorageService) GetSAFBridge() *safpkg.SAFBridge {
    return s.safBridge
}

func (s *AndroidStorageService) Stop() {
    s.detector.Stop()
}

// TranslateAndroidPath converts standard Linux paths to Android-equivalent paths
func TranslateAndroidPath(path string) string {
    home := os.Getenv("HOME")
    if home == "" {
        home = "/data/data/com.termux/files/home"
    }
    
    replacements := map[string]string{
        "/DATA":           filepath.Join(home, "storage", "shared", "CasaOS"),
        "/home":           home,
        "/tmp":            filepath.Join(home, ".casaos", "tmp"),
        "/var/lib/casaos": filepath.Join(home, ".casaos", "data"),
    }
    
    for old, new := range replacements {
        if path == old || len(path) > len(old) && path[:len(old)+1] == old+"/" {
            return new + path[len(old):]
        }
    }
    return path
}
