package plugins_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/plugins"
	
	birdactyl "github.com/Birdactyl/Birdactyl-Go-SDK"
)

var (
	dbReady       bool
	pluginAPI     *birdactyl.API
	pluginAddr    = "127.0.0.1:50051"
	runningPlugin *birdactyl.Plugin
)

func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	serverDir := filepath.Dir(filepath.Dir(filepath.Dir(thisFile))) 
	cfgPath := filepath.Join(serverDir, "config.yaml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "config.yaml not found at %s\n", cfgPath)
	} else if _, err := config.Load(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "config load failed: %v\n", err)
	} else if err := database.Connect(&config.Get().Database); err != nil {
		fmt.Fprintf(os.Stderr, "database connect failed: %v\n", err)
	} else {
		dbReady = true
	}

	if config.Get() != nil {
		config.Get().Auth.BcryptCost = 4
	}

	if !dbReady {
		fmt.Println("DB not ready, skipping plugin tests")
		os.Exit(0)
	}

	if err := plugins.StartServer(pluginAddr); err != nil {
		fmt.Fprintf(os.Stderr, "start plugin panel server failed: %v\n", err)
		os.Exit(1)
	}

	oldArgs := os.Args
	os.Args = []string{"plugin_test"}

	runningPlugin = birdactyl.New("test-integration-plugin", "1.0.0").SetName("Integration Test Plugin")
	
	runningPlugin.Mixin("test.mixin.modify", func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
		original := ctx.GetString("value")
		ctx.Set("value", original+"_modified")
		return ctx.Next()
	})

	runningPlugin.Mixin("test.mixin.return", func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
		return ctx.Return(map[string]interface{}{"halted": true, "reason": "plugin returned early"})
	})

	runningPlugin.Mixin("test.mixin.error", func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
		return ctx.Error("plugin manually threw an error")
	})

	go func() {
		if err := runningPlugin.Start(pluginAddr); err != nil {
			log.Printf("plugin start error: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)
	os.Args = oldArgs

	pluginAPI = runningPlugin.API()
	if pluginAPI == nil {
		fmt.Fprintln(os.Stderr, "plugin API is nil - connection likely failed")
		os.Exit(1)
	}

	code := m.Run()

	database.Close()
	os.Exit(code)
}

func requireDB(t *testing.T) {
	t.Helper()
	if !dbReady {
		t.Skip("database not available")
	}
	if pluginAPI == nil {
		t.Skip("Plugin SDK not initialized")
	}
}
