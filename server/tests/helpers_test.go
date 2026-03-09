package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
)

var dbReady bool

func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	serverDir := filepath.Dir(filepath.Dir(thisFile))
	cfgPath := filepath.Join(serverDir, "config.yaml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "config.yaml not found at %s (DB tests will be skipped)\n", cfgPath)
	} else if _, err := config.Load(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "config load failed: %v (DB tests will be skipped)\n", err)
	} else if err := database.Connect(&config.Get().Database); err != nil {
		fmt.Fprintf(os.Stderr, "database connect failed: %v (DB tests will be skipped)\n", err)
	} else {
		dbReady = true
	}

	code := m.Run()

	if dbReady {
		database.Close()
	}
	os.Exit(code)
}

func requireDB(t *testing.T) {
	t.Helper()
	if !dbReady {
		t.Skip("database not available")
	}
}

func parseJSONResponse(resp *http.Response) map[string]interface{} {
	var body map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &body)
	return body
}

func toJSONBody(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}
