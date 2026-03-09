package tests

import (
	"encoding/json"
	"testing"

	"birdactyl-panel-backend/internal/models"
)

func TestSystemInfoMarshal(t *testing.T) {
	info := models.SystemInfo{
		Hostname: "test-node",
		OS: models.OSInfo{
			Name: "Ubuntu",
		},
		CPU: models.CPUInfo{
			Cores: 8,
		},
	}

	t.Run("Value returns JSON string", func(t *testing.T) {
		val, err := info.Value()
		if err != nil {
			t.Fatalf("Failed to retrieve system info json value: %v", err)
		}
		
		valStr, ok := val.(string)
		if !ok {
			t.Fatalf("Expected value to be a string representing JSON")
		}
		if valStr == "" {
			t.Fatalf("System info serialized string empty")
		}
	})

	t.Run("Scan loads from JSON", func(t *testing.T) {
		data, _ := json.Marshal(info)
		var target models.SystemInfo
		err := target.Scan(data)
		if err != nil {
			t.Fatalf("Failed to scan system info bytes: %v", err)
		}

		if target.Hostname != info.Hostname || target.CPU.Cores != info.CPU.Cores {
			t.Errorf("System info unmarshaled improperly")
		}
	})
}

func TestServerPortsMarshal(t *testing.T) {
	ports := []models.ServerPort{
		{Port: 25565, Primary: true},
	}

	t.Run("Unmarshal back correctly", func(t *testing.T) {
		data, _ := json.Marshal(ports)
		var target []models.ServerPort
		json.Unmarshal(data, &target)
        
        if len(target) == 0 {
            t.Fatalf("Expected ports slice to have at least one element")
        }

		if target[0].Port != 25565 {
			t.Errorf("Expected port to be 25565, got %d", target[0].Port)
		}
		if !target[0].Primary {
			t.Errorf("Expected primary to be true")
		}
	})
}
