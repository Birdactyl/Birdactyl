package plugins

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type processInfo struct {
	cmd  *exec.Cmd
	id   string
	path string
}

type ProcessManager struct {
	mu        sync.RWMutex
	processes map[string]*processInfo
	byPath    map[string]string
}

var processManager = &ProcessManager{
	processes: make(map[string]*processInfo),
	byPath:    make(map[string]string),
}

func GetProcessManager() *ProcessManager {
	return processManager
}

func (pm *ProcessManager) StartWithPort(cfg PluginConfig, port int, dataDir string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.byPath[cfg.Binary]; exists {
		return nil
	}

	cmd := exec.Command(cfg.Binary, fmt.Sprintf("%d", port), dataDir)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	key := cfg.Binary
	pm.processes[key] = &processInfo{cmd: cmd, path: cfg.Binary}
	pm.byPath[cfg.Binary] = key

	go pm.pipeOutput(key, stdout)
	go pm.pipeOutput(key, stderr)

	log.Printf("[plugins] started %s on port %d (pid: %d)", cfg.Binary, port, cmd.Process.Pid)

	go func() {
		cmd.Wait()
		pm.mu.Lock()
		if info, ok := pm.processes[key]; ok && info.id != "" {
			GetRegistry().SetOnline(info.id, false)
		}
		delete(pm.processes, key)
		delete(pm.byPath, cfg.Binary)
		pm.mu.Unlock()
		log.Printf("[plugins] process %s exited", cfg.Binary)
	}()

	return nil
}

func (pm *ProcessManager) StartStreaming(cfg PluginConfig, panelAddr string, dataDir string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.byPath[cfg.Binary]; exists {
		return nil
	}

	cmd := exec.Command(cfg.Binary, panelAddr, dataDir)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	key := cfg.Binary
	pm.processes[key] = &processInfo{cmd: cmd, path: cfg.Binary}
	pm.byPath[cfg.Binary] = key

	go pm.pipeOutput(key, stdout)
	go pm.pipeOutput(key, stderr)

	log.Printf("[plugins] started %s (streaming, pid: %d)", cfg.Binary, cmd.Process.Pid)

	go func() {
		cmd.Wait()
		pm.mu.Lock()
		if info, ok := pm.processes[key]; ok && info.id != "" {
			GetStreamRegistry().Remove(info.id)
			GetMixinRegistry().Unregister(info.id)
			UnregisterSchedules(info.id)
		}
		delete(pm.processes, key)
		delete(pm.byPath, cfg.Binary)
		pm.mu.Unlock()
		log.Printf("[plugins] process %s exited", cfg.Binary)
	}()

	return nil
}

func (pm *ProcessManager) StartJar(cfg PluginConfig, port int, dataDir string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.byPath[cfg.Binary]; exists {
		return nil
	}

	cmd := exec.Command("java", "-jar", cfg.Binary, fmt.Sprintf("%d", port), dataDir)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	key := cfg.Binary
	pm.processes[key] = &processInfo{cmd: cmd, path: cfg.Binary}
	pm.byPath[cfg.Binary] = key

	go pm.pipeOutput(key, stdout)
	go pm.pipeOutput(key, stderr)

	log.Printf("[plugins] started jar %s on port %d (pid: %d)", cfg.Binary, port, cmd.Process.Pid)

	go func() {
		cmd.Wait()
		pm.mu.Lock()
		if info, ok := pm.processes[key]; ok && info.id != "" {
			GetRegistry().SetOnline(info.id, false)
		}
		delete(pm.processes, key)
		delete(pm.byPath, cfg.Binary)
		pm.mu.Unlock()
		log.Printf("[plugins] jar %s exited", cfg.Binary)
	}()

	return nil
}

func (pm *ProcessManager) pipeOutput(key string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		pm.mu.RLock()
		info := pm.processes[key]
		pm.mu.RUnlock()

		prefix := key
		if info != nil && info.id != "" {
			prefix = info.id
		}
		log.Printf("[plugin:%s] %s", prefix, scanner.Text())
	}
}

func (pm *ProcessManager) SetID(path, id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if key, ok := pm.byPath[path]; ok {
		if info, ok := pm.processes[key]; ok {
			info.id = id
			pm.processes[id] = info
			if key != id {
				delete(pm.processes, key)
			}
			pm.byPath[path] = id
		}
	}
}

func (pm *ProcessManager) GetPathByID(id string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for path, key := range pm.byPath {
		if key == id {
			return path
		}
		if info, ok := pm.processes[key]; ok && info.id == id {
			return path
		}
	}

	for _, info := range pm.processes {
		if info.id == id {
			return info.path
		}
	}

	return ""
}

func (pm *ProcessManager) Stop(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	info, exists := pm.processes[id]
	if !exists {
		return nil
	}

	if info.cmd.Process != nil {
		info.cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- info.cmd.Wait() }()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			info.cmd.Process.Kill()
		}
	}

	delete(pm.processes, id)
	delete(pm.byPath, info.path)
	return nil
}

func (pm *ProcessManager) StopByPath(path string) error {
	pm.mu.Lock()
	key, exists := pm.byPath[path]
	pm.mu.Unlock()

	if !exists {
		return nil
	}
	return pm.Stop(key)
}

func (pm *ProcessManager) StopAll() {
	pm.mu.RLock()
	ids := make([]string, 0, len(pm.processes))
	for id := range pm.processes {
		ids = append(ids, id)
	}
	pm.mu.RUnlock()

	for _, id := range ids {
		pm.Stop(id)
	}
}

func (pm *ProcessManager) IsRunning(id string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, exists := pm.processes[id]
	return exists
}

func (pm *ProcessManager) HasPath(path string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, exists := pm.byPath[path]
	return exists
}
