package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

var ErrConfigGenerated = errors.New("config file generated")

type Config struct {
	Panel   PanelConfig   `yaml:"panel"`
	Node    NodeConfig    `yaml:"node"`
	Logging LoggingConfig `yaml:"logging"`
}

type PanelConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type LoggingConfig struct {
	File string `yaml:"file"`
}

type NodeConfig struct {
	Listen    string `yaml:"listen"`
	DataDir   string `yaml:"data_dir"`
	BackupDir string `yaml:"backup_dir"`
	DisplayIP string `yaml:"display_ip"`
}

var cfg *Config
var configPath string

func Load(path string) (*Config, error) {
	configPath = path

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := generateDefaultConfig(path); err != nil {
			return nil, err
		}
		return nil, ErrConfigGenerated
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg = &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.Node.Listen == "" {
		cfg.Node.Listen = "0.0.0.0:8443"
	}
	if cfg.Node.DataDir == "" {
		cfg.Node.DataDir = "/var/lib/birdactyl/servers"
	}
	if cfg.Node.BackupDir == "" {
		cfg.Node.BackupDir = "/var/lib/birdactyl/backups"
	}

	return cfg, nil
}

func Get() *Config {
	return cfg
}

func Save() error {
	if configPath == "" {
		return errors.New("config path not set")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func generateDefaultConfig(path string) error {
	defaultConfig := `panel:
  url: "http://localhost:3000"
  token: ""

node:
  listen: "0.0.0.0:8443"
  data_dir: "/var/lib/birdactyl/servers"
  backup_dir: "/var/lib/birdactyl/backups"
  display_ip: ""

logging:
  file: "logs/axis.log"
`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}
