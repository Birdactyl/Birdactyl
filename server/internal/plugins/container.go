package plugins

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/resources"
)

type ContainerManager struct {
	mu            sync.RWMutex
	containerID   string
	containerName string
	config        config.ContainerConfig
	pluginsDir    string
	panelAddr     string
	processes     map[string]*containerProcess
	running       bool
}

type containerProcess struct {
	id       string
	binary   string
	execID   string
	doneChan chan struct{}
}

var containerManager *ContainerManager
var containerOnce sync.Once

func GetContainerManager() *ContainerManager {
	containerOnce.Do(func() {
		containerManager = &ContainerManager{
			processes: make(map[string]*containerProcess),
		}
	})
	return containerManager
}

func (cm *ContainerManager) Init(cfg config.ContainerConfig, pluginsDir, panelAddr string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config = cfg
	cm.pluginsDir = pluginsDir
	cm.panelAddr = panelAddr
	cm.containerName = "birdactyl-plugins"

	if !cfg.Enabled {
		log.Println("[container] container mode disabled, using host execution")
		return nil
	}

	if err := cm.ensureImage(); err != nil {
		return fmt.Errorf("failed to ensure container image: %w", err)
	}

	if err := cm.startContainer(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	cm.running = true
	log.Printf("[container] plugin container started: %s", cm.containerID[:12])
	return nil
}

func (cm *ContainerManager) ensureImage() error {
	image := cm.config.Image
	if image == "" {
		image = "birdactyl/plugin-runtime:latest"
	}

	cmd := exec.Command("docker", "image", "inspect", image)
	if err := cmd.Run(); err == nil {
		return nil
	}

	log.Printf("[container] building image %s...", image)

	tmpDir, err := os.MkdirTemp("", "birdactyl-build-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, resources.PluginRuntimeDockerfile, 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	cmd = exec.Command("docker", "build", "-t", image, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	return nil
}

func (cm *ContainerManager) startContainer() error {
	cm.stopExistingContainer()

	absPluginsDir, err := filepath.Abs(cm.pluginsDir)
	if err != nil {
		return err
	}

	image := cm.config.Image
	if image == "" {
		image = "birdactyl/plugin-runtime:latest"
	}

	args := []string{
		"run", "-d",
		"--name", cm.containerName,
		"--network", cm.getNetworkMode(),
		"-v", fmt.Sprintf("%s:/plugins:ro", absPluginsDir),
		"-v", fmt.Sprintf("%s_data:/data", cm.containerName),
	}

	if cm.config.MemoryLimit != "" {
		args = append(args, "--memory", cm.config.MemoryLimit)
	}
	if cm.config.CPULimit != "" {
		args = append(args, "--cpus", cm.config.CPULimit)
	}

	args = append(args, image, "sleep", "infinity")

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %w, output: %s", err, string(output))
	}

	cm.containerID = strings.TrimSpace(string(output))
	return nil
}

func (cm *ContainerManager) stopExistingContainer() {
	exec.Command("docker", "stop", cm.containerName).Run()
	exec.Command("docker", "rm", cm.containerName).Run()
}

func (cm *ContainerManager) getNetworkMode() string {
	if cm.config.NetworkMode != "" {
		return cm.config.NetworkMode
	}
	return "host"
}

func (cm *ContainerManager) IsEnabled() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Enabled && cm.running
}

func (cm *ContainerManager) ExecPlugin(binary, panelAddr, dataDir string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return fmt.Errorf("container not running")
	}

	binaryName := filepath.Base(binary)
	containerBinary := "/plugins/" + binaryName
	containerDataDir := "/data"

	execArgs := []string{
		"exec", "-d", cm.containerName,
		containerBinary, panelAddr, containerDataDir,
	}

	cmd := exec.Command("docker", execArgs...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to exec plugin in container: %w", err)
	}

	proc := &containerProcess{
		binary:   binary,
		doneChan: make(chan struct{}),
	}
	cm.processes[binaryName] = proc

	go func() {
		cmd.Wait()
		cm.mu.Lock()
		delete(cm.processes, binaryName)
		cm.mu.Unlock()
		close(proc.doneChan)
		log.Printf("[container] plugin %s exited", binaryName)
	}()

	log.Printf("[container] started plugin %s in container", binaryName)
	return nil
}

func (cm *ContainerManager) ExecJar(jarPath, panelAddr, dataDir string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return fmt.Errorf("container not running")
	}

	jarName := filepath.Base(jarPath)
	containerJar := "/plugins/" + jarName
	containerDataDir := "/data"

	execArgs := []string{
		"exec", "-d", cm.containerName,
		"java", "-jar", containerJar, panelAddr, containerDataDir,
	}

	cmd := exec.Command("docker", execArgs...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to exec jar in container: %w", err)
	}

	proc := &containerProcess{
		binary:   jarPath,
		doneChan: make(chan struct{}),
	}
	cm.processes[jarName] = proc

	go func() {
		cmd.Wait()
		cm.mu.Lock()
		delete(cm.processes, jarName)
		cm.mu.Unlock()
		close(proc.doneChan)
		log.Printf("[container] jar %s exited", jarName)
	}()

	log.Printf("[container] started jar %s in container", jarName)
	return nil
}

func (cm *ContainerManager) ExecPluginWithLogs(binary, panelAddr, dataDir string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return fmt.Errorf("container not running")
	}

	binaryName := filepath.Base(binary)
	containerBinary := "/plugins/" + binaryName
	containerDataDir := "/data"

	execArgs := []string{
		"exec", cm.containerName,
		containerBinary, panelAddr, containerDataDir,
	}

	cmd := exec.Command("docker", execArgs...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to exec plugin in container: %w", err)
	}

	proc := &containerProcess{
		binary:   binary,
		doneChan: make(chan struct{}),
	}
	cm.processes[binaryName] = proc

	go cm.pipeOutput(binaryName, stdout)
	go cm.pipeOutput(binaryName, stderr)

	go func() {
		cmd.Wait()
		cm.mu.Lock()
		if p, ok := cm.processes[binaryName]; ok && p.id != "" {
			GetStreamRegistry().Remove(p.id)
			GetMixinRegistry().Unregister(p.id)
			UnregisterSchedules(p.id)
		}
		delete(cm.processes, binaryName)
		cm.mu.Unlock()
		close(proc.doneChan)
		log.Printf("[container] plugin %s exited", binaryName)
	}()

	log.Printf("[container] started plugin %s in container", binaryName)
	return nil
}

func (cm *ContainerManager) pipeOutput(name string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		cm.mu.RLock()
		proc := cm.processes[name]
		cm.mu.RUnlock()

		prefix := name
		if proc != nil && proc.id != "" {
			prefix = proc.id
		}
		log.Printf("[plugin:%s] %s", prefix, scanner.Text())
	}
}

func (cm *ContainerManager) SetPluginID(binary, id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	binaryName := filepath.Base(binary)
	if proc, ok := cm.processes[binaryName]; ok {
		proc.id = id
	}
}

func (cm *ContainerManager) StopPlugin(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for name, proc := range cm.processes {
		if proc.id == id {
			cmd := exec.Command("docker", "exec", cm.containerName, "pkill", "-f", name)
			cmd.Run()
			return nil
		}
	}
	return nil
}

func (cm *ContainerManager) HasPlugin(binary string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	binaryName := filepath.Base(binary)
	_, exists := cm.processes[binaryName]
	return exists
}

func (cm *ContainerManager) GetPluginIDByPath(binary string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	binaryName := filepath.Base(binary)
	if proc, ok := cm.processes[binaryName]; ok {
		return proc.id
	}
	return ""
}

func (cm *ContainerManager) Shutdown() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return nil
	}

	log.Println("[container] shutting down plugin container...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "stop", cm.containerName)
	cmd.Run()

	cmd = exec.Command("docker", "rm", cm.containerName)
	cmd.Run()

	cm.running = false
	cm.containerID = ""
	cm.processes = make(map[string]*containerProcess)

	log.Println("[container] plugin container stopped")
	return nil
}

func (cm *ContainerManager) IsRunning() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.running
}

func (cm *ContainerManager) GetContainerID() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.containerID
}
