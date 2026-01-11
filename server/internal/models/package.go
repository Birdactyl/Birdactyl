package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PackagePort struct {
	Name     string `json:"name"`
	Default  int    `json:"default"`
	Protocol string `json:"protocol"`
	Primary  bool   `json:"primary,omitempty"`
}

type PackageVariable struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Default      string `json:"default"`
	UserEditable bool   `json:"user_editable"`
	Rules        string `json:"rules,omitempty"`
}

type PackageConfigFile struct {
	Path     string `json:"path"`
	Template string `json:"template"`
}

type AddonSourceMapping struct {
	Results     string `json:"results"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Author      string `json:"author"`
	Downloads   string `json:"downloads"`
	VersionID   string `json:"version_id"`
	VersionName string `json:"version_name"`
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
}

type AddonSource struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Icon        string             `json:"icon"`
	Type        string             `json:"type"`
	APIKey      string             `json:"api_key,omitempty"`
	SearchURL   string             `json:"search_url"`
	VersionsURL string             `json:"versions_url"`
	DownloadURL string             `json:"download_url"`
	InstallPath string             `json:"install_path"`
	FileFilter  string             `json:"file_filter"`
	Headers     map[string]string  `json:"headers"`
	Mapping     AddonSourceMapping `json:"mapping"`
}

type Package struct {
	ID                  uuid.UUID      `json:"id" gorm:"primaryKey"`
	Name                string         `json:"name" gorm:"type:varchar(255);not null"`
	Version             string         `json:"version" gorm:"type:varchar(50)"`
	Author              string         `json:"author" gorm:"type:varchar(255)"`
	Description         string         `json:"description" gorm:"type:text"`
	Icon                string         `json:"icon" gorm:"type:varchar(500)"`
	DockerImage         string         `json:"docker_image" gorm:"type:varchar(500);not null"`
	InstallImage        string         `json:"install_image" gorm:"type:varchar(500)"`
	Startup             string         `json:"startup" gorm:"type:text;not null"`
	InstallScript       string         `json:"install_script" gorm:"type:text"`
	StopSignal          string         `json:"stop_signal" gorm:"type:varchar(20);default:'SIGTERM'"`
	StopCommand         string         `json:"stop_command" gorm:"type:varchar(255)"`
	StopTimeout         int            `json:"stop_timeout" gorm:"default:30"`
	StartupEditable     bool           `json:"startup_editable" gorm:"default:false"`
	DockerImageEditable bool           `json:"docker_image_editable" gorm:"default:false"`
	Ports               datatypes.JSON `json:"ports" gorm:"type:json"`
	Variables           datatypes.JSON `json:"variables" gorm:"type:json"`
	ConfigFiles         datatypes.JSON `json:"config_files" gorm:"type:json"`
	AddonSources        datatypes.JSON `json:"addon_sources" gorm:"type:json"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

func (p *Package) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Ports == nil {
		p.Ports = []byte("[]")
	}
	if p.Variables == nil {
		p.Variables = []byte("[]")
	}
	if p.ConfigFiles == nil {
		p.ConfigFiles = []byte("[]")
	}
	if p.AddonSources == nil {
		p.AddonSources = []byte("[]")
	}
	return nil
}
