package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"cauthon-axis/internal/config"
	"cauthon-axis/internal/docker"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
)

var serverUID = "1000"
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]|\x1b[=]`)
var serverIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var (
	statsCache      = make(map[string]*cachedStats)
	statsCacheMu    sync.RWMutex
	serverConfigs   = make(map[string]*ServerConfig)
	serverConfigsMu sync.RWMutex
)

type cachedStats struct {
	stats     *ServerStats
	updatedAt time.Time
}

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func ValidateServerID(id string) error {
	if id == "" || len(id) > 64 || !serverIDRegex.MatchString(id) {
		return fmt.Errorf("invalid server id")
	}
	return nil
}

func init() {
	serverUID = strconv.Itoa(os.Getuid())
	go statsRefresher()
}

func statsRefresher() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		statsCacheMu.RLock()
		serverIDs := make([]string, 0, len(statsCache))
		for id := range statsCache {
			serverIDs = append(serverIDs, id)
		}
		statsCacheMu.RUnlock()

		for _, id := range serverIDs {
			if stats, err := fetchStats(id); err == nil {
				statsCacheMu.Lock()
				statsCache[id] = &cachedStats{stats: stats, updatedAt: time.Now()}
				statsCacheMu.Unlock()
			}
		}
	}
}

func GetServerUID() string {
	return serverUID
}

type ServerConfig struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	DockerImage   string            `json:"docker_image"`
	InstallImage  string            `json:"install_image"`
	InstallScript string            `json:"install_script"`
	Startup       string            `json:"startup"`
	Memory        int               `json:"memory"`
	CPU           int               `json:"cpu"`
	Disk          int               `json:"disk"`
	Ports         []PortConfig      `json:"ports"`
	Variables     map[string]string `json:"variables"`
	StopSignal    string            `json:"stop_signal"`
	StopCommand   string            `json:"stop_command"`
	StopTimeout   int               `json:"stop_timeout"`
}

type PortConfig struct {
	Host      int    `json:"host"`
	Container int    `json:"container"`
	Protocol  string `json:"protocol"`
}

type ServerStats struct {
	MemoryUsage int64   `json:"memory_usage"`
	MemoryLimit int64   `json:"memory_limit"`
	CPUPercent  float64 `json:"cpu_percent"`
	DiskUsage   int64   `json:"disk_usage"`
	NetRx       int64   `json:"net_rx"`
	NetTx       int64   `json:"net_tx"`
}

func containerName(serverID string) string {
	return fmt.Sprintf("birdactyl-%s", serverID)
}

func installContainerName(serverID string) string {
	return fmt.Sprintf("birdactyl-install-%s", serverID)
}

func serverDataDir(serverID string) string {
	cfg := config.Get()
	return filepath.Join(cfg.Node.DataDir, serverID)
}

func Create(cfg ServerConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	BroadcastLog(cfg.ID, "Starting server installation...")

	dataDir := serverDataDir(cfg.ID)
	if err := os.MkdirAll(dataDir, 0777); err != nil {
		BroadcastLog(cfg.ID, fmt.Sprintf("Failed to create data directory: %v", err))
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	os.Chmod(dataDir, 0777)

	chownRecursive(dataDir)

	BroadcastLog(cfg.ID, "Created server data directory")

	if cfg.InstallScript != "" {
		BroadcastLog(cfg.ID, "Running install script...")
		if err := runInstall(ctx, cfg, dataDir); err != nil {
			BroadcastLog(cfg.ID, fmt.Sprintf("Installation failed: %v", err))
			return fmt.Errorf("install failed: %w", err)
		}
		BroadcastLog(cfg.ID, "Install script completed successfully")
	}

	if !docker.ImageExists(ctx, cfg.DockerImage) {
		BroadcastLog(cfg.ID, fmt.Sprintf("Pulling Docker image: %s", cfg.DockerImage))
		if err := docker.PullImage(ctx, cfg.DockerImage); err != nil {
			BroadcastLog(cfg.ID, fmt.Sprintf("Failed to pull image: %v", err))
			return fmt.Errorf("failed to pull image: %w", err)
		}
		BroadcastLog(cfg.ID, "Docker image pulled successfully")
	}

	startup := cfg.Startup
	for k, v := range cfg.Variables {
		startup = strings.ReplaceAll(startup, "{{"+k+"}}", v)
	}

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, p := range cfg.Ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		containerPort := nat.Port(fmt.Sprintf("%d/%s", p.Container, proto))
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: strconv.Itoa(p.Host),
		}}
	}

	containerCfg := &container.Config{
		Image:        cfg.DockerImage,
		Cmd:          []string{"/bin/sh", "-c", startup},
		ExposedPorts: exposedPorts,
		Env:          buildEnv(cfg.Variables),
		WorkingDir:   "/home/container",
		User:         serverUID + ":" + serverUID,
		StopSignal:   cfg.StopSignal,
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}

	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
		Mounts: []mount.Mount{{
			Type:   mount.TypeBind,
			Source: dataDir,
			Target: "/home/container",
		}},
		Resources: container.Resources{
			Memory:   int64(cfg.Memory) * 1024 * 1024,
			NanoCPUs: int64(cfg.CPU) * 10000000,
		},
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}

	name := containerName(cfg.ID)

	if docker.ContainerExists(ctx, name) {
		if id, err := docker.GetContainerID(ctx, name); err == nil {
			docker.StopContainer(ctx, id, 10)
			docker.RemoveContainer(ctx, id, true)
		}
	}

	_, err := docker.CreateContainer(ctx, name, containerCfg, hostCfg)
	if err == nil {
		serverConfigsMu.Lock()
		serverConfigs[cfg.ID] = &cfg
		serverConfigsMu.Unlock()
	}
	return err
}

func runInstall(ctx context.Context, cfg ServerConfig, dataDir string) error {
	installImage := cfg.InstallImage
	if installImage == "" {
		installImage = "alpine:latest"
	}

	if !docker.ImageExists(ctx, installImage) {
		BroadcastLog(cfg.ID, fmt.Sprintf("Pulling install image: %s", installImage))
		if err := docker.PullImage(ctx, installImage); err != nil {
			return fmt.Errorf("failed to pull install image: %w", err)
		}
	}

	script := cfg.InstallScript
	for k, v := range cfg.Variables {
		script = strings.ReplaceAll(script, "{{"+k+"}}", v)
	}

	var fullScript string
	if strings.Contains(installImage, "alpine") {
		fullScript = fmt.Sprintf("apk add --no-cache curl jq bash && cd /home/container && %s && chmod -R 777 /home/container", script)
	} else {
		fullScript = script
	}

	containerCfg := &container.Config{
		Image:      installImage,
		Cmd:        []string{"/bin/sh", "-c", fullScript},
		Env:        buildEnv(cfg.Variables),
		WorkingDir: "/home/container",
		Tty:        true,
	}

	hostCfg := &container.HostConfig{
		Mounts: []mount.Mount{{
			Type:   mount.TypeBind,
			Source: dataDir,
			Target: "/home/container",
		}},
	}

	name := installContainerName(cfg.ID)

	if docker.ContainerExists(ctx, name) {
		if id, err := docker.GetContainerID(ctx, name); err == nil {
			docker.RemoveContainer(ctx, id, true)
		}
	}

	id, err := docker.CreateContainer(ctx, name, containerCfg, hostCfg)
	if err != nil {
		return err
	}

	if err := docker.StartContainer(ctx, id); err != nil {
		docker.RemoveContainer(ctx, id, true)
		return err
	}

	go streamInstallLogs(ctx, cfg.ID, id)

	statusCh, errCh := docker.Client.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		docker.RemoveContainer(ctx, id, true)
		return err
	case status := <-statusCh:
		docker.RemoveContainer(ctx, id, true)
		if status.StatusCode != 0 {
			return fmt.Errorf("install script exited with code %d", status.StatusCode)
		}
	}

	chownRecursive(dataDir)

	return nil
}

func CreateContainer(cfg ServerConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dataDir := serverDataDir(cfg.ID)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	chownRecursive(dataDir)

	if !docker.ImageExists(ctx, cfg.DockerImage) {
		BroadcastLog(cfg.ID, fmt.Sprintf("Pulling Docker image: %s", cfg.DockerImage))
		if err := docker.PullImage(ctx, cfg.DockerImage); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
	}

	startup := cfg.Startup
	for k, v := range cfg.Variables {
		startup = strings.ReplaceAll(startup, "{{"+k+"}}", v)
	}

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, p := range cfg.Ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		containerPort := nat.Port(fmt.Sprintf("%d/%s", p.Container, proto))
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: strconv.Itoa(p.Host),
		}}
	}

	containerCfg := &container.Config{
		Image:        cfg.DockerImage,
		Cmd:          []string{"/bin/sh", "-c", startup},
		ExposedPorts: exposedPorts,
		Env:          buildEnv(cfg.Variables),
		WorkingDir:   "/home/container",
		User:         serverUID + ":" + serverUID,
		StopSignal:   cfg.StopSignal,
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}

	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
		Mounts: []mount.Mount{{
			Type:   mount.TypeBind,
			Source: dataDir,
			Target: "/home/container",
		}},
		Resources: container.Resources{
			Memory:   int64(cfg.Memory) * 1024 * 1024,
			NanoCPUs: int64(cfg.CPU) * 10000000,
		},
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}

	name := containerName(cfg.ID)
	_, err := docker.CreateContainer(ctx, name, containerCfg, hostCfg)
	if err == nil {
		serverConfigsMu.Lock()
		serverConfigs[cfg.ID] = &cfg
		serverConfigsMu.Unlock()
	}
	return err
}

func Start(serverID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id, err := docker.GetContainerID(ctx, containerName(serverID))
	if err != nil {
		return err
	}
	return docker.StartContainer(ctx, id)
}

func Stop(serverID string, timeout int) error {
	return StopWithConfig(serverID, timeout, "", "")
}

func StopWithConfig(serverID string, timeout int, stopCommand string, stopSignal string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+10)*time.Second)
	defer cancel()

	id, err := docker.GetContainerID(ctx, containerName(serverID))
	if err != nil {
		return err
	}

	docker.Client.ContainerUpdate(ctx, id, container.UpdateConfig{
		RestartPolicy: container.RestartPolicy{Name: "no"},
	})

	if stopCommand == "" {
		serverConfigsMu.RLock()
		cfg := serverConfigs[serverID]
		serverConfigsMu.RUnlock()
		if cfg != nil {
			stopCommand = cfg.StopCommand
		}
	}

	if stopCommand != "" {
		BroadcastLog(serverID, fmt.Sprintf("Sending stop command: %s", stopCommand))
		if cmdErr := SendCommand(serverID, stopCommand); cmdErr != nil {
			BroadcastLog(serverID, fmt.Sprintf("Stop command failed: %v, falling back to signal", cmdErr))
		} else {
			waitCtx, waitCancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
			defer waitCancel()

			statusCh, errCh := docker.Client.ContainerWait(waitCtx, id, container.WaitConditionNotRunning)
			select {
			case <-statusCh:
				BroadcastLog(serverID, "Server stopped gracefully")
				return nil
			case err := <-errCh:
				if err != nil {
					BroadcastLog(serverID, fmt.Sprintf("Wait failed: %v, sending stop signal", err))
				}
			case <-waitCtx.Done():
				BroadcastLog(serverID, "Graceful shutdown timed out, sending stop signal")
			}
		}
	}

	return docker.StopContainer(ctx, id, timeout)
}

func Kill(serverID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := docker.GetContainerID(ctx, containerName(serverID))
	if err != nil {
		return err
	}
	return docker.KillContainer(ctx, id)
}

func Restart(serverID string, timeout int) error {
	return RestartWithConfig(serverID, timeout, "", "")
}

func RestartWithConfig(serverID string, timeout int, stopCommand string, stopSignal string) error {
	BroadcastLog(serverID, "Restarting server...")
	
	if err := StopWithConfig(serverID, timeout, stopCommand, stopSignal); err != nil {
		BroadcastLog(serverID, fmt.Sprintf("Stop failed: %v, attempting force restart", err))
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+30)*time.Second)
		defer cancel()
		id, err := docker.GetContainerID(ctx, containerName(serverID))
		if err != nil {
			return err
		}
		return docker.RestartContainer(ctx, id, timeout)
	}
	
	if err := Start(serverID); err != nil {
		return fmt.Errorf("restart failed during start: %w", err)
	}
	
	BroadcastLog(serverID, "Server restarted")
	return nil
}

func RemoveContainer(serverID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	name := containerName(serverID)
	if id, err := docker.GetContainerID(ctx, name); err == nil {
		docker.StopContainer(ctx, id, 5)
		docker.RemoveContainer(ctx, id, true)
	}
}

func Reinstall(cfg ServerConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	name := containerName(cfg.ID)
	if docker.ContainerExists(ctx, name) {
		if id, err := docker.GetContainerID(ctx, name); err == nil {
			BroadcastLog(cfg.ID, "Stopping server for reinstall...")
			docker.StopContainer(ctx, id, 10)
			docker.RemoveContainer(ctx, id, true)
		}
	}

	BroadcastLog(cfg.ID, "Starting reinstallation...")

	dataDir := serverDataDir(cfg.ID)
	os.Chmod(dataDir, 0777)
	chownRecursive(dataDir)

	if cfg.InstallScript != "" {
		BroadcastLog(cfg.ID, "Running install script...")
		if err := runInstall(ctx, cfg, dataDir); err != nil {
			BroadcastLog(cfg.ID, fmt.Sprintf("Installation failed: %v", err))
			return fmt.Errorf("install failed: %w", err)
		}
		BroadcastLog(cfg.ID, "Install script completed successfully")
	}

	if !docker.ImageExists(ctx, cfg.DockerImage) {
		BroadcastLog(cfg.ID, fmt.Sprintf("Pulling Docker image: %s", cfg.DockerImage))
		if err := docker.PullImage(ctx, cfg.DockerImage); err != nil {
			BroadcastLog(cfg.ID, fmt.Sprintf("Failed to pull image: %v", err))
			return fmt.Errorf("failed to pull image: %w", err)
		}
		BroadcastLog(cfg.ID, "Docker image pulled successfully")
	}

	startup := cfg.Startup
	for k, v := range cfg.Variables {
		startup = strings.ReplaceAll(startup, "{{"+k+"}}", v)
	}

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, p := range cfg.Ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		containerPort := nat.Port(fmt.Sprintf("%d/%s", p.Container, proto))
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: strconv.Itoa(p.Host),
		}}
	}

	containerCfg := &container.Config{
		Image:        cfg.DockerImage,
		Cmd:          []string{"/bin/sh", "-c", startup},
		ExposedPorts: exposedPorts,
		Env:          buildEnv(cfg.Variables),
		WorkingDir:   "/home/container",
		User:         serverUID + ":" + serverUID,
		StopSignal:   cfg.StopSignal,
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}

	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
		Mounts: []mount.Mount{{
			Type:   mount.TypeBind,
			Source: dataDir,
			Target: "/home/container",
		}},
		Resources: container.Resources{
			Memory:   int64(cfg.Memory) * 1024 * 1024,
			NanoCPUs: int64(cfg.CPU) * 10000000,
		},
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}

	_, err := docker.CreateContainer(ctx, name, containerCfg, hostCfg)
	if err != nil {
		BroadcastLog(cfg.ID, fmt.Sprintf("Failed to create container: %v", err))
		return err
	}

	BroadcastLog(cfg.ID, "Reinstallation complete")
	return nil
}

func Delete(serverID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	name := containerName(serverID)
	if docker.ContainerExists(ctx, name) {
		id, err := docker.GetContainerID(ctx, name)
		if err == nil {
			docker.KillContainer(ctx, id)
			docker.RemoveContainer(ctx, id, true)
		}
	}

	go func() {
		dataDir := serverDataDir(serverID)
		os.RemoveAll(dataDir)
	}()

	return nil
}

func GetStatus(serverID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := containerName(serverID)
	if !docker.ContainerExists(ctx, name) {
		return "offline", nil
	}

	id, err := docker.GetContainerID(ctx, name)
	if err != nil {
		return "offline", nil
	}

	info, err := docker.Client.ContainerInspect(ctx, id)
	if err != nil {
		return "offline", nil
	}

	if info.State.Running {
		return "running", nil
	}
	if info.State.Restarting {
		return "running", nil
	}
	return "stopped", nil
}

func GetStats(serverID string) (*ServerStats, error) {
	statsCacheMu.Lock()
	cached, exists := statsCache[serverID]
	if !exists {
		statsCache[serverID] = &cachedStats{}
	}
	statsCacheMu.Unlock()

	if exists && cached.stats != nil && time.Since(cached.updatedAt) < 2*time.Second {
		return cached.stats, nil
	}

	return fetchStats(serverID)
}

func fetchStats(serverID string) (*ServerStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := docker.GetContainerID(ctx, containerName(serverID))
	if err != nil {
		return nil, err
	}

	stats, err := docker.GetContainerStats(ctx, id)
	if err != nil {
		return nil, err
	}

	cpuPercent := 0.0
	if stats.PreCPUStats.CPUUsage.TotalUsage > 0 && stats.PreCPUStats.SystemUsage > 0 {
		if stats.CPUStats.CPUUsage.TotalUsage >= stats.PreCPUStats.CPUUsage.TotalUsage &&
			stats.CPUStats.SystemUsage >= stats.PreCPUStats.SystemUsage {
			cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
			systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
			if systemDelta > 0 && cpuDelta > 0 {
				numCPUs := stats.CPUStats.OnlineCPUs
				if numCPUs == 0 {
					numCPUs = uint32(len(stats.CPUStats.CPUUsage.PercpuUsage))
				}
				if numCPUs == 0 {
					numCPUs = 1
				}
				cpuPercent = (cpuDelta / systemDelta) * float64(numCPUs) * 100.0
				if cpuPercent > 10000 {
					cpuPercent = 0.0
				}
			}
		}
	}

	var netRx, netTx int64
	for _, net := range stats.Networks {
		netRx += int64(net.RxBytes)
		netTx += int64(net.TxBytes)
	}

	diskUsage := getDirSize(serverDataDir(serverID))

	return &ServerStats{
		MemoryUsage: int64(stats.MemoryStats.Usage),
		MemoryLimit: int64(stats.MemoryStats.Limit),
		CPUPercent:  cpuPercent,
		DiskUsage:   diskUsage,
		NetRx:       netRx,
		NetTx:       netTx,
	}, nil
}

func buildEnv(vars map[string]string) []string {
	env := make([]string, 0, len(vars))
	for k, v := range vars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func chownRecursive(path string) {
	uid, _ := strconv.Atoi(serverUID)
	filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			os.Chown(name, uid, uid)
		}
		return nil
	})
}

func GetDiskUsage(serverID string) int64 {
	return getDirSize(serverDataDir(serverID))
}


func IsDataDirEmpty(serverID string) bool {
	dataDir := serverDataDir(serverID)
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return true
	}
	return len(entries) == 0
}
