package benchmarks

import (
	"encoding/json"
	"testing"

	"birdactyl-panel-backend/internal/models"
)

func BenchmarkSystemInfoMarshal(b *testing.B) {
	info := models.SystemInfo{
		Hostname: "benchmark-node-01",
		OS: models.OSInfo{
			Name:    "Ubuntu",
			Version: "22.04",
			Kernel:  "5.15.0-91-generic",
			Arch:    "x86_64",
		},
		CPU: models.CPUInfo{
			Cores: 8,
			Usage: 42.5,
		},
		Memory: models.MemoryInfo{
			Total:     17179869184,
			Used:      8589934592,
			Available: 8589934592,
			Usage:     50.0,
		},
		Disk: models.DiskInfo{
			Total:     107374182400,
			Used:      53687091200,
			Available: 53687091200,
			Usage:     50.0,
		},
		Uptime: 864000,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		info.Value()
	}
}

func BenchmarkSystemInfoUnmarshal(b *testing.B) {
	info := models.SystemInfo{
		Hostname: "benchmark-node-01",
		OS: models.OSInfo{
			Name:    "Ubuntu",
			Version: "22.04",
			Kernel:  "5.15.0-91-generic",
			Arch:    "x86_64",
		},
		CPU: models.CPUInfo{
			Cores: 8,
			Usage: 42.5,
		},
		Memory: models.MemoryInfo{
			Total:     17179869184,
			Used:      8589934592,
			Available: 8589934592,
			Usage:     50.0,
		},
		Disk: models.DiskInfo{
			Total:     107374182400,
			Used:      53687091200,
			Available: 53687091200,
			Usage:     50.0,
		},
		Uptime: 864000,
	}
	data, _ := json.Marshal(info)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var target models.SystemInfo
		target.Scan(data)
	}
}

func BenchmarkServerPortsMarshal(b *testing.B) {
	ports := []models.ServerPort{
		{Port: 25565, Primary: true},
		{Port: 25566, Primary: false},
		{Port: 25567, Primary: false},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(ports)
	}
}

func BenchmarkServerPortsUnmarshal(b *testing.B) {
	ports := []models.ServerPort{
		{Port: 25565, Primary: true},
		{Port: 25566, Primary: false},
		{Port: 25567, Primary: false},
	}
	data, _ := json.Marshal(ports)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var target []models.ServerPort
		json.Unmarshal(data, &target)
	}
}

func BenchmarkPermissionsJSONRoundTrip(b *testing.B) {
	perms := models.AllPermissions
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(perms)
		var target []string
		json.Unmarshal(data, &target)
	}
}

func BenchmarkPackageVariablesMarshal(b *testing.B) {
	vars := []models.PackageVariable{
		{Name: "SERVER_JAR", Description: "Server jar file", Default: "server.jar", UserEditable: true, Rules: "required|string"},
		{Name: "SERVER_MEMORY", Description: "Memory allocation", Default: "1024", UserEditable: true, Rules: "required|integer"},
		{Name: "SERVER_PORT", Description: "Server port", Default: "25565", UserEditable: false, Rules: "required|integer"},
		{Name: "JAVA_VERSION", Description: "Java version", Default: "17", UserEditable: true, Rules: "required|in:8,11,17,21"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(vars)
	}
}

func BenchmarkPackageVariablesUnmarshal(b *testing.B) {
	vars := []models.PackageVariable{
		{Name: "SERVER_JAR", Description: "Server jar file", Default: "server.jar", UserEditable: true, Rules: "required|string"},
		{Name: "SERVER_MEMORY", Description: "Memory allocation", Default: "1024", UserEditable: true, Rules: "required|integer"},
		{Name: "SERVER_PORT", Description: "Server port", Default: "25565", UserEditable: false, Rules: "required|integer"},
		{Name: "JAVA_VERSION", Description: "Java version", Default: "17", UserEditable: true, Rules: "required|in:8,11,17,21"},
	}
	data, _ := json.Marshal(vars)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var target []models.PackageVariable
		json.Unmarshal(data, &target)
	}
}

func BenchmarkAddonSourceMarshal(b *testing.B) {
	source := models.AddonSource{
		ID:          "modrinth",
		Name:        "Modrinth",
		Icon:        "https://modrinth.com/favicon.ico",
		Type:        "modrinth",
		SearchURL:   "https://api.modrinth.com/v2/search",
		VersionsURL: "https://api.modrinth.com/v2/project/{id}/version",
		DownloadURL: "https://cdn.modrinth.com/data/{id}/versions/{version_id}/{file_name}",
		InstallPath: "/mods",
		FileFilter:  "*.jar",
		Headers:     map[string]string{"User-Agent": "Birdactyl/1.0"},
		Mapping: models.AddonSourceMapping{
			Results:     "hits",
			ID:          "project_id",
			Name:        "title",
			Description: "description",
			Icon:        "icon_url",
			Author:      "author",
			Downloads:   "downloads",
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(source)
	}
}

func BenchmarkAddonSourceUnmarshal(b *testing.B) {
	source := models.AddonSource{
		ID:          "modrinth",
		Name:        "Modrinth",
		Icon:        "https://modrinth.com/favicon.ico",
		Type:        "modrinth",
		SearchURL:   "https://api.modrinth.com/v2/search",
		VersionsURL: "https://api.modrinth.com/v2/project/{id}/version",
		DownloadURL: "https://cdn.modrinth.com/data/{id}/versions/{version_id}/{file_name}",
		InstallPath: "/mods",
		FileFilter:  "*.jar",
		Headers:     map[string]string{"User-Agent": "Birdactyl/1.0"},
		Mapping: models.AddonSourceMapping{
			Results:     "hits",
			ID:          "project_id",
			Name:        "title",
			Description: "description",
			Icon:        "icon_url",
			Author:      "author",
			Downloads:   "downloads",
		},
	}
	data, _ := json.Marshal(source)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var target models.AddonSource
		json.Unmarshal(data, &target)
	}
}

func BenchmarkUserModelMarshal(b *testing.B) {
	user := models.User{
		Email:        "bench@test.com",
		Username:     "benchuser",
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		IsAdmin:      false,
		RegisterIP:   "127.0.0.1",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(user)
	}
}

func BenchmarkServerModelMarshal(b *testing.B) {
	server := models.Server{
		Name:        "Benchmark Server",
		Description: "A server used for benchmarking",
		Status:      models.ServerStatusRunning,
		Memory:      4096,
		CPU:         200,
		Disk:        10240,
		Startup:     "java -Xms128M -Xmx{{SERVER_MEMORY}}M -jar {{SERVER_JAR}}",
		DockerImage: "ghcr.io/pterodactyl/yolks:java_17",
		Ports:       []byte(`[{"port":25565,"primary":true},{"port":25566}]`),
		Variables:   []byte(`{"SERVER_JAR":"server.jar","SERVER_MEMORY":"4096"}`),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(server)
	}
}
