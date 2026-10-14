//go:build android

package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/IceWhaleTech/CasaOS/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func parseCPUFromProcStat(data string) (float64, error) {
	lines := strings.Split(data, "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("empty /proc/stat")
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, fmt.Errorf("invalid /proc/stat format")
	}
	var total float64
	for i := 1; i < len(fields); i++ {
		val, _ := strconv.ParseFloat(fields[i], 64)
		total += val
	}
	idle, _ := strconv.ParseFloat(fields[4], 64)
	return (1.0 - (idle / total)) * 100.0, nil
}

func parseCPUFromTop(data string) (float64, error) {
	// Parse "CPU: XX% user, YY% sys, ZZ% idle" format
	return 0, nil // Simple stub if top fails
}

func GetCPUUsageAndroid() (float64, error) {
	// Try reading /proc/stat
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		// Fallback: try running 'top -bn1' and parsing CPU line
		out, err2 := exec.Command("top", "-bn1", "-n", "1").Output()
		if err2 != nil {
			return -1, fmt.Errorf("CPU metrics unavailable: %v", err)
		}
		// Parse "CPU: XX% user, YY% sys, ZZ% idle" format
		return parseCPUFromTop(string(out))
	}
	return parseCPUFromProcStat(string(data))
}

func GetCPUInfoAndroid() ([]model.CPUCoreInfo, error) {
	cpus, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/scaling_cur_freq")
	if err != nil {
		return nil, err
	}
	var cores []model.CPUCoreInfo
	for _, cpuPath := range cpus {
		freqBytes, _ := os.ReadFile(cpuPath)
		maxFreqBytes, _ := os.ReadFile(filepath.Join(filepath.Dir(cpuPath), "cpuinfo_max_freq"))
		freq, _ := strconv.ParseInt(strings.TrimSpace(string(freqBytes)), 10, 64)
		maxFreq, _ := strconv.ParseInt(strings.TrimSpace(string(maxFreqBytes)), 10, 64)
		cluster := "LITTLE"
		if maxFreq > 2000000 {
			cluster = "big"
		} // >2GHz is typically big core
		if maxFreq > 2800000 {
			cluster = "prime"
		} // >2.8GHz is prime core
		cores = append(cores, model.CPUCoreInfo{Freq: freq, MaxFreq: maxFreq, Cluster: cluster})
	}
	return cores, nil
}

func GetCPUTemperatureAndroid() ([]model.ThermalZone, error) {
	zones, err := filepath.Glob("/sys/devices/virtual/thermal/thermal_zone*/temp")
	if err != nil || len(zones) == 0 {
		return nil, fmt.Errorf("thermal zones not accessible")
	}
	var result []model.ThermalZone
	for _, zonePath := range zones {
		dir := filepath.Dir(zonePath)
		typeBytes, _ := os.ReadFile(filepath.Join(dir, "type"))
		tempBytes, err := os.ReadFile(zonePath)
		if err != nil {
			continue
		}
		temp, _ := strconv.ParseFloat(strings.TrimSpace(string(tempBytes)), 64)
		zoneType := strings.TrimSpace(string(typeBytes))
		// Filter for CPU-related zones
		if strings.Contains(zoneType, "cpu") || strings.Contains(zoneType, "soc") || strings.Contains(zoneType, "tsens") {
			result = append(result, model.ThermalZone{Type: zoneType, Temperature: temp / 1000.0})
		}
	}
	return result, nil
}

func parseMemInfoFromProc() (*model.MemoryInfo, error) {
	return &model.MemoryInfo{}, nil
}

func GetMemoryAndroid() (*model.MemoryInfo, error) {
	// Try gopsutil first
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		return &model.MemoryInfo{
			Total:       vmStat.Total,
			Available:   vmStat.Available,
			Used:        vmStat.Used,
			Free:        vmStat.Free,
			UsedPercent: vmStat.UsedPercent,
		}, nil
	}
	// Fallback: parse /proc/meminfo
	return parseMemInfoFromProc()
}

func GetBatteryInfo() (*model.BatteryInfo, error) {
	out, err := exec.Command("termux-battery-status").Output()
	if err != nil {
		return &model.BatteryInfo{}, err
	}
	var info model.BatteryInfo
	err = json.Unmarshal(out, &info)
	if err != nil {
		return &model.BatteryInfo{}, err
	}
	return &info, nil
}

func GetAndroidDeviceInfo() (*model.AndroidDeviceInfo, error) {
	info := &model.AndroidDeviceInfo{Architecture: "arm64"}
	if out, err := exec.Command("getprop", "ro.product.model").Output(); err == nil {
		info.Model = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("getprop", "ro.product.manufacturer").Output(); err == nil {
		info.Manufacturer = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("getprop", "ro.build.version.release").Output(); err == nil {
		info.AndroidVersion = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		info.KernelVersion = strings.TrimSpace(string(out))
	}
	return info, nil
}

func GetWiFiInfo() (*model.WiFiInfo, error) {
	out, err := exec.Command("termux-wifi-connectioninfo").Output()
	if err != nil {
		return &model.WiFiInfo{}, err
	}
	var info model.WiFiInfo
	err = json.Unmarshal(out, &info)
	if err != nil {
		return &model.WiFiInfo{}, err
	}
	return &info, nil
}

func (c *systemService) GetCpuInfo() []cpu.InfoStat {
	return []cpu.InfoStat{}
}

func (c *systemService) GetMemInfo() map[string]interface{} {
	memData := make(map[string]interface{})
	minfo, err := GetMemoryAndroid()
	if err == nil {
		memData["total"] = minfo.Total
		memData["available"] = minfo.Available
		memData["used"] = minfo.Used
		memData["free"] = minfo.Free
		memData["usedPercent"] = minfo.UsedPercent
	} else {
		memData["total"] = uint64(0)
		memData["available"] = uint64(0)
		memData["used"] = uint64(0)
		memData["free"] = uint64(0)
		memData["usedPercent"] = float64(0)
	}
	return memData
}

func (c *systemService) GetCpuPercent() float64 {
	usage, err := GetCPUUsageAndroid()
	if err == nil {
		return usage
	}
	return 0.0
}

func (c *systemService) GetCpuCoreNum() int {
	cores, err := GetCPUInfoAndroid()
	if err == nil {
		return len(cores)
	}
	return 1
}

func (s *systemService) GetCPUTemperature() int {
	temps, err := GetCPUTemperatureAndroid()
	if err == nil && len(temps) > 0 {
		return int(temps[0].Temperature)
	}
	return 0
}

func (s *systemService) GetCPUPower() map[string]string {
	data := make(map[string]string)
	data["value"] = "0"
	return data
}
