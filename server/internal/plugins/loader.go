package plugins

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/config"
	pb "birdactyl-panel-backend/internal/plugins/proto"
)

func InitPluginSystem(cfg config.PluginsConfig) error {
	if cfg.Container.Enabled {
		if err := GetContainerManager().Init(cfg.Container, cfg.Directory, cfg.Address); err != nil {
			return fmt.Errorf("failed to initialize container: %w", err)
		}
	}
	return nil
}

var (
	nextPort   = 50100
	nextPortMu sync.Mutex
)

func getNextPort() int {
	nextPortMu.Lock()
	defer nextPortMu.Unlock()
	port := nextPort
	nextPort++
	return port
}

func waitForStreamPlugin(pluginPath string, timeout time.Duration, existingIDs map[string]bool) (*PluginStream, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, ps := range GetStreamRegistry().All() {
			if GetContainerManager().IsEnabled() {
				if GetContainerManager().GetPluginIDByPath(pluginPath) == ps.ID {
					return ps, nil
				}
				if !existingIDs[ps.ID] && GetContainerManager().HasPlugin(pluginPath) {
					return ps, nil
				}
			} else {
				if GetProcessManager().GetPathByID(ps.ID) == pluginPath {
					return ps, nil
				}
				if !existingIDs[ps.ID] && GetProcessManager().HasPath(pluginPath) {
					return ps, nil
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("timeout waiting for plugin to connect")
}

func LoadPlugins(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		path := filepath.Join(dir, name)

		if filepath.Ext(name) == ".jar" {
			wg.Add(1)
			go func(p, n string) {
				defer wg.Done()
				if err := loadJar(p, dir); err != nil {
					log.Printf("[plugins] failed to load %s: %v", n, err)
				}
			}(path, name)
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.Mode()&0111 == 0 {
			continue
		}

		wg.Add(1)
		go func(p, n string) {
			defer wg.Done()
			if err := loadBinary(p, dir); err != nil {
				log.Printf("[plugins] failed to load %s: %v", n, err)
			}
		}(path, name)
	}

	wg.Wait()
	return nil
}

func loadJar(jarPath, pluginsDir string) error {
	cfg := config.Get()

	if GetContainerManager().IsEnabled() {
		if GetContainerManager().HasPlugin(jarPath) {
			return nil
		}

		panelAddr := cfg.Plugins.Address

		existingIDs := make(map[string]bool)
		for _, ps := range GetStreamRegistry().All() {
			existingIDs[ps.ID] = true
		}

		if err := GetContainerManager().ExecJar(jarPath, panelAddr, pluginsDir); err != nil {
			return fmt.Errorf("failed to start jar in container: %w", err)
		}

		time.Sleep(1500 * time.Millisecond)

		ps, err := waitForStreamPlugin(jarPath, 5*time.Second, existingIDs)
		if err != nil {
			GetContainerManager().StopPlugin(filepath.Base(jarPath))
			return fmt.Errorf("jar plugin did not connect: %w", err)
		}

		GetContainerManager().SetPluginID(jarPath, ps.ID)

		log.Printf("[plugins] loaded jar %s v%s (%d events, %d routes, %d schedules, %d mixins) [container]",
			ps.Info.Name, ps.Info.Version, len(ps.Info.Events), len(ps.Info.Routes), len(ps.Info.Schedules), len(ps.Info.Mixins))

		return nil
	}

	port := getNextPort()

	pluginCfg := PluginConfig{
		Binary:  jarPath,
		Address: fmt.Sprintf("localhost:%d", port),
	}

	if err := GetProcessManager().StartJar(pluginCfg, port, pluginsDir); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	time.Sleep(1500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, client, err := connectToPlugin(pluginCfg.Address)
	if err != nil {
		GetProcessManager().StopByPath(jarPath)
		return fmt.Errorf("failed to connect: %w", err)
	}

	info, err := client.GetInfo(ctx, &pb.Empty{})
	if err != nil {
		conn.Close()
		GetProcessManager().StopByPath(jarPath)
		return fmt.Errorf("failed to get info: %w", err)
	}

	pluginCfg.ID = info.Id
	pluginCfg.Name = info.Name

	if err := GetRegistry().RegisterWithConn(pluginCfg, conn, client, info); err != nil {
		conn.Close()
		GetProcessManager().StopByPath(jarPath)
		return err
	}

	GetProcessManager().SetID(jarPath, pluginCfg.ID)

	for _, sched := range info.Schedules {
		if err := RegisterSchedule(pluginCfg.ID, sched.Id, sched.Cron); err != nil {
			log.Printf("[plugins] failed to register schedule %s for %s: %v", sched.Id, pluginCfg.ID, err)
		}
	}

	for _, mixin := range info.Mixins {
		GetMixinRegistry().Register(pluginCfg.ID, mixin.Target, int(mixin.Priority))
	}

	log.Printf("[plugins] loaded %s v%s (%d events, %d routes, %d schedules, %d mixins)",
		info.Name, info.Version, len(info.Events), len(info.Routes), len(info.Schedules), len(info.Mixins))

	return nil
}

func loadBinary(binaryPath, pluginsDir string) error {
	cfg := config.Get()
	panelAddr := cfg.Plugins.Address

	existingIDs := make(map[string]bool)
	for _, ps := range GetStreamRegistry().All() {
		existingIDs[ps.ID] = true
	}

	if GetContainerManager().IsEnabled() {
		if GetContainerManager().HasPlugin(binaryPath) {
			return nil
		}

		if err := GetContainerManager().ExecPluginWithLogs(binaryPath, panelAddr, pluginsDir); err != nil {
			return fmt.Errorf("failed to start in container: %w", err)
		}

		ps, err := waitForStreamPlugin(binaryPath, 5*time.Second, existingIDs)
		if err != nil {
			GetContainerManager().StopPlugin(filepath.Base(binaryPath))
			return fmt.Errorf("plugin did not connect: %w", err)
		}

		GetContainerManager().SetPluginID(binaryPath, ps.ID)

		log.Printf("[plugins] loaded %s v%s (%d events, %d routes, %d schedules, %d mixins, %d addon types) [container]",
			ps.Info.Name, ps.Info.Version, len(ps.Info.Events), len(ps.Info.Routes), len(ps.Info.Schedules), len(ps.Info.Mixins), len(ps.Info.AddonTypes))

		return nil
	}

	pluginCfg := PluginConfig{
		Binary: binaryPath,
	}

	if err := GetProcessManager().StartStreaming(pluginCfg, panelAddr, pluginsDir); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	ps, err := waitForStreamPlugin(binaryPath, 5*time.Second, existingIDs)
	if err != nil {
		GetProcessManager().StopByPath(binaryPath)
		return fmt.Errorf("plugin did not connect: %w", err)
	}

	GetProcessManager().SetID(binaryPath, ps.ID)

	log.Printf("[plugins] loaded %s v%s (%d events, %d routes, %d schedules, %d mixins, %d addon types)",
		ps.Info.Name, ps.Info.Version, len(ps.Info.Events), len(ps.Info.Routes), len(ps.Info.Schedules), len(ps.Info.Mixins), len(ps.Info.AddonTypes))

	return nil
}

func LoadPlugin(pluginCfg PluginConfig) error {
	cfg := config.Get()
	if !cfg.Plugins.AllowDynamic {
		return ErrDynamicDisabled
	}

	if pluginCfg.Binary != "" {
		if filepath.Ext(pluginCfg.Binary) == ".jar" {
			return loadJar(pluginCfg.Binary, cfg.Plugins.Directory)
		}
		return loadBinary(pluginCfg.Binary, cfg.Plugins.Directory)
	}

	if pluginCfg.ID == "" || pluginCfg.Address == "" {
		return ErrInvalidConfig
	}

	if err := GetRegistry().Register(pluginCfg); err != nil {
		return err
	}

	go initPlugin(pluginCfg.ID)
	return nil
}

func UnloadPlugin(id string) error {
	cfg := config.Get()
	if !cfg.Plugins.AllowDynamic {
		return ErrDynamicDisabled
	}

	streamPlugin := GetStreamRegistry().Get(id)
	legacyPlugin := GetRegistry().Get(id)

	if streamPlugin == nil && legacyPlugin == nil {
		return ErrPluginNotFound
	}

	UnregisterSchedules(id)
	GetMixinRegistry().Unregister(id)

	if GetContainerManager().IsEnabled() {
		GetContainerManager().StopPlugin(id)
	} else {
		GetProcessManager().Stop(id)
	}

	GetStreamRegistry().Remove(id)
	GetRegistry().Unregister(id)

	log.Printf("[plugins] unloaded plugin: %s", id)
	return nil
}

func initPlugin(pluginID string) {
	plugin := GetRegistry().Get(pluginID)
	if plugin == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := plugin.Client.GetInfo(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("[plugins] failed to get info from %s: %v", pluginID, err)
		GetRegistry().SetOnline(pluginID, false)
		return
	}

	for _, sched := range info.Schedules {
		if err := RegisterSchedule(pluginID, sched.Id, sched.Cron); err != nil {
			log.Printf("[plugins] failed to register schedule %s for %s: %v", sched.Id, pluginID, err)
		}
	}

	for _, mixin := range info.Mixins {
		GetMixinRegistry().Register(pluginID, mixin.Target, int(mixin.Priority))
	}

	log.Printf("[plugins] initialized %s v%s with %d routes, %d schedules, %d mixins",
		info.Name, info.Version, len(info.Routes), len(info.Schedules), len(info.Mixins))
}
