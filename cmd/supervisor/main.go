package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Services []ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Name          string   `yaml:"name"`
	Command       string   `yaml:"command"`
	Args          []string `yaml:"args"`
	HealthCheck   string   `yaml:"health_check"`
	DependsOn     []string `yaml:"depends_on"`
	RestartPolicy string   `yaml:"restart_policy"`
	RestartDelay  string   `yaml:"restart_delay"`
	MaxRestarts   int      `yaml:"max_restarts"`
}

type Supervisor struct {
	config    Config
	processes map[string]*exec.Cmd
	pids      map[string]int
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewSupervisor(cfg Config) *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Supervisor{
		config:    cfg,
		processes: make(map[string]*exec.Cmd),
		pids:      make(map[string]int),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func acquireWakeLock() {
	exec.Command("termux-wake-lock").Run()
}

func releaseWakeLock() {
	exec.Command("termux-wake-unlock").Run()
}

func (s *Supervisor) StartAll() {
	acquireWakeLock()
	defer releaseWakeLock()

	for _, svc := range s.config.Services {
		s.startService(svc.Name)
	}

	// Health monitoring loop
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.monitor()
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	for sig := range c {
		if sig == syscall.SIGHUP {
			log.Println("Reloading config not fully implemented")
		} else {
			log.Println("Shutting down gracefully...")
			s.StopAll()
			break
		}
	}
}

func (s *Supervisor) startService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.processes[name]; ok {
		return
	}

	var svcConfig *ServiceConfig
	for _, sc := range s.config.Services {
		if sc.Name == name {
			svcConfig = &sc
			break
		}
	}
	if svcConfig == nil {
		return
	}

	// Simplistic replace for config dir
	configDir := "/data/data/com.termux/files/usr/etc/casaos"
	var args []string
	for _, arg := range svcConfig.Args {
		args = append(args, strings.ReplaceAll(arg, "{{CONFIG_DIR}}", configDir))
	}

	cmd := exec.CommandContext(s.ctx, svcConfig.Command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start %s: %v", name, err)
		return
	}

	s.processes[name] = cmd
	if cmd.Process != nil {
		s.pids[name] = cmd.Process.Pid
		s.writePidFile(name, cmd.Process.Pid)
	}
	
	log.Printf("Started %s (PID %d)", name, s.pids[name])

	go func() {
		err := cmd.Wait()
		log.Printf("Service %s exited with error: %v", name, err)
		s.mu.Lock()
		delete(s.processes, name)
		s.mu.Unlock()
	}()
}

func (s *Supervisor) StopAll() {
	s.cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	for name, cmd := range s.processes {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		s.removePidFile(name)
	}
}

func (s *Supervisor) monitor() {
	s.mu.Lock()
	defer s.mu.Unlock()

	statusFile := "/data/data/com.termux/files/home/.casaos/run/supervisor.status"
	os.MkdirAll(filepath.Dir(statusFile), 0755)
	
	var status []string
	for _, svc := range s.config.Services {
		pid, ok := s.pids[svc.Name]
		if !ok || s.processes[svc.Name] == nil {
			status = append(status, fmt.Sprintf("%s: DOWN", svc.Name))
			// Restart logic could go here
		} else {
			if svc.HealthCheck != "" {
				url := strings.ReplaceAll(svc.HealthCheck, "{{PORT}}", "8080")
				resp, err := http.Get(url)
				if err == nil && resp.StatusCode == 200 {
					status = append(status, fmt.Sprintf("%s: UP (PID %d, HEALTHY)", svc.Name, pid))
				} else {
					status = append(status, fmt.Sprintf("%s: UP (PID %d, UNHEALTHY)", svc.Name, pid))
				}
				if resp != nil {
					resp.Body.Close()
				}
			} else {
				status = append(status, fmt.Sprintf("%s: UP (PID %d)", svc.Name, pid))
			}
		}
	}
	os.WriteFile(statusFile, []byte(strings.Join(status, "\n")), 0644)
}

func (s *Supervisor) writePidFile(name string, pid int) {
	home := "/data/data/com.termux/files/home"
	pidPath := filepath.Join(home, ".casaos", "run", fmt.Sprintf("%s.pid", name))
	os.MkdirAll(filepath.Dir(pidPath), 0755)
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func (s *Supervisor) removePidFile(name string) {
	home := "/data/data/com.termux/files/home"
	pidPath := filepath.Join(home, ".casaos", "run", fmt.Sprintf("%s.pid", name))
	os.Remove(pidPath)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: casaos-supervisor [start|stop|restart|status|log]")
		return
	}

	cfgPath := "/data/data/com.termux/files/usr/etc/casaos/supervisor.yaml"
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		// Fallback for testing
		data, err = os.ReadFile("build/sysroot/etc/casaos/supervisor.yaml")
		if err != nil {
			log.Fatalf("Cannot read config: %v", err)
		}
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Cannot parse config: %v", err)
	}

	cmd := os.Args[1]
	switch cmd {
	case "start":
		sup := NewSupervisor(cfg)
		sup.StartAll()
	case "stop":
		log.Println("Send SIGTERM to running supervisor")
	case "restart":
		log.Println("Not fully implemented: restarting all")
	case "status":
		statusData, _ := os.ReadFile("/data/data/com.termux/files/home/.casaos/run/supervisor.status")
		fmt.Println(string(statusData))
	case "log":
		log.Println("Log tailing not fully implemented")
	default:
		fmt.Println("Unknown command")
	}
}
