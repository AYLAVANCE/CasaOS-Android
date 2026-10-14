package container

import (
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

type ComposeService struct {
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports"`
	Volumes     []string          `yaml:"volumes"`
	Environment map[string]string `yaml:"environment"`
	Restart     string            `yaml:"restart"`
}

type ComposeFile struct {
	Version  string                    `yaml:"version"`
	Services map[string]ComposeService `yaml:"services"`
}

func ParseCompose(data []byte) (*ComposeFile, error) {
	var c ComposeFile
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *ComposeFile) ToCreateRequests() ([]ContainerCreateRequest, error) {
	var reqs []ContainerCreateRequest

	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	casaData := home + "/.casaos/data"

	for name, svc := range c.Services {
		req := ContainerCreateRequest{
			Name:    name,
			Image:   svc.Image,
			Env:     svc.Environment,
			Restart: svc.Restart,
			Labels:  make(map[string]string),
		}

		// Translate Ports
		for _, p := range svc.Ports {
			parts := strings.Split(p, ":")
			if len(parts) == 2 {
				hostP, _ := strconv.Atoi(parts[0])
				contP, _ := strconv.Atoi(parts[1])

				// Offset privileged ports
				if hostP < 1024 {
					hostP += 8000
				}

				req.Ports = append(req.Ports, PortMapping{
					HostPort:      hostP,
					ContainerPort: contP,
					Protocol:      "tcp",
				})
			}
		}

		// Translate Volumes
		for _, v := range svc.Volumes {
			parts := strings.Split(v, ":")
			if len(parts) >= 2 {
				hostPath := parts[0]
				contPath := parts[1]
				
				// Translate standard Linux paths to Termux paths
				if strings.HasPrefix(hostPath, "/var/lib/") || strings.HasPrefix(hostPath, "/DATA") {
					hostPath = casaData + "/" + strings.TrimPrefix(hostPath, "/var/lib/")
					hostPath = strings.ReplaceAll(hostPath, "//", "/")
				}
				
				req.Volumes = append(req.Volumes, VolumeMount{
					HostPath:      hostPath,
					ContainerPath: contPath,
					ReadOnly:      len(parts) == 3 && parts[2] == "ro",
				})
			}
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}
