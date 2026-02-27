package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig          `yaml:"server"`
	Database   DatabaseConfig        `yaml:"database"`
	Auth       AuthConfig            `yaml:"auth"`
	Resources  ResourcesConfig       `yaml:"resources"`
	Logging    LoggingConfig         `yaml:"logging"`
	Plugins    PluginsConfig         `yaml:"plugins"`
	RootAdmins []string              `yaml:"root_admins"`
	APIKeys    map[string]APIKeyConfig `yaml:"api_keys"`
}

type APIKeyConfig struct {
	Key     string            `yaml:"key"`
	Headers map[string]string `yaml:"headers"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LoggingConfig struct {
	File string `yaml:"file"`
}

type PluginLoadMode string

const (
	PluginLoadManual  PluginLoadMode = "manual"
	PluginLoadManaged PluginLoadMode = "managed"
)

type PluginsConfig struct {
	Address      string           `yaml:"address"`
	Directory    string           `yaml:"directory"`
	LoadMode     PluginLoadMode   `yaml:"load_mode"`
	AllowDynamic bool             `yaml:"allow_dynamic"`
	Container    ContainerConfig  `yaml:"container"`
}

type ContainerConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Image       string `yaml:"image"`
	NetworkMode string `yaml:"network_mode"`
	MemoryLimit string `yaml:"memory_limit"`
	CPULimit    string `yaml:"cpu_limit"`
}

type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type AuthConfig struct {
	AccountsPerIP         int    `yaml:"accounts_per_ip"`
	AccessTokenExpiry     int    `yaml:"access_token_expiry"`
	RefreshTokenExpiry    int    `yaml:"refresh_token_expiry"`
	TokenRefreshThreshold int    `yaml:"token_refresh_threshold"`
	MaxSessionsPerUser    int    `yaml:"max_sessions_per_user"`
	BcryptCost            int    `yaml:"bcrypt_cost"`
	JWTSecret             string `yaml:"jwt_secret"`
}

type ResourcesConfig struct {
	Enabled      bool `yaml:"enabled"`
	DefaultRAM   int  `yaml:"default_ram"`
	DefaultCPU   int  `yaml:"default_cpu"`
	DefaultDisk  int  `yaml:"default_disk"`
	MaxServers   int  `yaml:"max_servers"`
}

var (
	cfg  *Config
	once sync.Once
	configPath string
	ErrConfigGenerated = fmt.Errorf("config file generated")
)

func Load(path string) (*Config, error) {
	var loadErr error

	once.Do(func() {
		configPath = path
		
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := generateDefaultConfig(path); err != nil {
				loadErr = fmt.Errorf("failed to generate config: %w", err)
				return
			}
			loadErr = ErrConfigGenerated
			return
		}

		data, err := os.ReadFile(path)
		if err != nil {
			loadErr = fmt.Errorf("failed to read config: %w", err)
			return
		}

		cfg = &Config{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			loadErr = fmt.Errorf("failed to parse config: %w", err)
			cfg = nil
			return
		}

		needsSave := cfg.Auth.JWTSecret == ""
		cfg.setDefaults()
		cfg.loadEnvOverrides()
		
		if needsSave {
			saveJWTSecret(path, cfg.Auth.JWTSecret)
		}
	})

	return cfg, loadErr
}

func saveJWTSecret(path, secret string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	
	content := string(data)
	if !strings.Contains(content, "jwt_secret:") {
		content = strings.Replace(content, "bcrypt_cost:", "jwt_secret: \""+secret+"\"\n  bcrypt_cost:", 1)
		os.WriteFile(path, []byte(content), 0644)
	}
}

func Get() *Config {
	return cfg
}

func generateDefaultConfig(path string) error {
	defaultConfig := `server:
  host: "0.0.0.0"
  port: 3000

logging:
  file: "logs/panel.log"

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: ""
  name: "birdactyl"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300

auth:
  accounts_per_ip: 3
  access_token_expiry: 15
  refresh_token_expiry: 43200
  token_refresh_threshold: 1
  max_sessions_per_user: 5
  bcrypt_cost: 12

root_admins: []

resources:
  enabled: true
  default_ram: 4096
  default_cpu: 200
  default_disk: 10240
  max_servers: 3

plugins:
  address: "localhost:50050"
  directory: "plugins"
  allow_dynamic: true
  container:
    enabled: false
    image: "birdactyl/plugin-runtime:latest"
    network_mode: "host"
    memory_limit: "512m"
    cpu_limit: "1.0"
`

	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

func (c *Config) setDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 3000
	}
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Database.Driver == "" {
		c.Database.Driver = "postgres"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 300
	}
	if c.Auth.AccountsPerIP == 0 {
		c.Auth.AccountsPerIP = 3
	}
	if c.Auth.AccessTokenExpiry == 0 {
		c.Auth.AccessTokenExpiry = 15
	}
	if c.Auth.RefreshTokenExpiry == 0 {
		c.Auth.RefreshTokenExpiry = 43200
	}
	if c.Auth.TokenRefreshThreshold == 0 {
		c.Auth.TokenRefreshThreshold = 1
	}
	if c.Auth.MaxSessionsPerUser == 0 {
		c.Auth.MaxSessionsPerUser = 5
	}
	if c.Auth.BcryptCost == 0 {
		c.Auth.BcryptCost = 12
	}
	if c.Auth.JWTSecret == "" {
		secret := make([]byte, 64)
		rand.Read(secret)
		c.Auth.JWTSecret = hex.EncodeToString(secret)
	}
	if c.Resources.DefaultRAM == 0 {
		c.Resources.DefaultRAM = 4096
	}
	if c.Resources.DefaultCPU == 0 {
		c.Resources.DefaultCPU = 200
	}
	if c.Resources.DefaultDisk == 0 {
		c.Resources.DefaultDisk = 10240
	}
	if c.Resources.MaxServers == 0 {
		c.Resources.MaxServers = 3
	}
	if c.Plugins.Address == "" {
		c.Plugins.Address = "localhost:50050"
	}
	if c.Plugins.Directory == "" {
		c.Plugins.Directory = "plugins"
	}
	if c.Plugins.LoadMode == "" {
		c.Plugins.LoadMode = PluginLoadManual
	}
}

func (c *Config) loadEnvOverrides() {
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		c.Database.Name = v
	}
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func IsRootAdmin(userID string) bool {
	if cfg == nil {
		return false
	}
	for _, id := range cfg.RootAdmins {
		if id == userID {
			return true
		}
	}
	return false
}

func AddRootAdmin(userID string) error {
	if cfg == nil || configPath == "" {
		return fmt.Errorf("config not loaded")
	}

	for _, id := range cfg.RootAdmins {
		if id == userID {
			return nil
		}
	}

	cfg.RootAdmins = append(cfg.RootAdmins, userID)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	content := string(data)
	if strings.Contains(content, "root_admins: []") {
		content = strings.Replace(content, "root_admins: []", "root_admins:\n  - \""+userID+"\"", 1)
	} else if strings.Contains(content, "root_admins:") {
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) == "root_admins:" {
				lines = append(lines[:i+1], append([]string{"  - \"" + userID + "\""}, lines[i+1:]...)...)
				break
			}
		}
		content = strings.Join(lines, "\n")
	}

	return os.WriteFile(configPath, []byte(content), 0644)
}
