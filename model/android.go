package model

type CPUCoreInfo struct {
	Freq      int64  `json:"freq"`
	MaxFreq   int64  `json:"max_freq"`
	Cluster   string `json:"cluster"` // e.g., big, LITTLE
}

type ThermalZone struct {
	Type        string  `json:"type"`
	Temperature float64 `json:"temperature"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type BatteryInfo struct {
	Level       int     `json:"level"`       // 0-100
	Status      string  `json:"status"`      // charging, discharging, full
	Health      string  `json:"health"`      // good, overheat, dead
	Temperature float64 `json:"temperature"` // degrees C
	Plugged     string  `json:"plugged"`     // ac, usb, wireless
}

type AndroidDeviceInfo struct {
	Model          string `json:"model"`
	Manufacturer   string `json:"manufacturer"`
	AndroidVersion string `json:"android_version"`
	SDKVersion     int    `json:"sdk_version"`
	KernelVersion  string `json:"kernel_version"`
	Architecture   string `json:"architecture"`
}

type WiFiInfo struct {
	SSID        string `json:"ssid"`
	SignalLevel int    `json:"signal_level"` // dBm
	LinkSpeed   int    `json:"link_speed"`   // Mbps
	IPAddress   string `json:"ip_address"`
	Frequency   int    `json:"frequency"`    // MHz
}
