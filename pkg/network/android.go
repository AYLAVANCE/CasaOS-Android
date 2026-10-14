//go:build android

package network

import (
	"fmt"
	"net"
)

// GetWLANAddress returns the IP address of the active wlan0 interface
func GetWLANAddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	// Try wlan0 first, then any UP non-loopback interface
	for _, name := range []string{"wlan0", "wlan1", "eth0"} {
		iface, err := net.InterfaceByName(name)
		if err != nil || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "0.0.0.0", fmt.Errorf("no active network interface found")
}

// GetDefaultPort returns the default CasaOS port (unprivileged)
func GetDefaultPort() int {
	return 8080
}
