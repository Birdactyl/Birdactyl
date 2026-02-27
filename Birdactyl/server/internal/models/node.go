package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Node struct {
	ID            uuid.UUID      `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	Icon          string         `gorm:"type:varchar(500)" json:"icon"`
	FQDN          string         `gorm:"type:varchar(255);not null" json:"fqdn"`
	Port          int            `gorm:"not null;default:8443" json:"port"`
	TokenID       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	TokenHash     string         `gorm:"type:varchar(255);not null" json:"-"`
	DaemonToken   string         `gorm:"type:varchar(255);not null" json:"-"`
	IsOnline      bool           `gorm:"default:false" json:"is_online"`
	AuthError     bool           `gorm:"default:false" json:"auth_error"`
	LastHeartbeat *time.Time     `json:"last_heartbeat"`
	SystemInfo    SystemInfo     `gorm:"type:json" json:"system_info"`
	DisplayIP     string         `gorm:"type:varchar(255)" json:"display_ip"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (n *Node) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

type SystemInfo struct {
	Hostname string     `json:"hostname"`
	OS       OSInfo     `json:"os"`
	CPU      CPUInfo    `json:"cpu"`
	Memory   MemoryInfo `json:"memory"`
	Disk     DiskInfo   `json:"disk"`
	Uptime   uint64     `json:"uptime_seconds"`
}

type OSInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kernel  string `json:"kernel"`
	Arch    string `json:"arch"`
}

type CPUInfo struct {
	Cores int     `json:"cores"`
	Usage float64 `json:"usage_percent"`
}

type MemoryInfo struct {
	Total     uint64  `json:"total_bytes"`
	Used      uint64  `json:"used_bytes"`
	Available uint64  `json:"available_bytes"`
	Usage     float64 `json:"usage_percent"`
}

type DiskInfo struct {
	Total     uint64  `json:"total_bytes"`
	Used      uint64  `json:"used_bytes"`
	Available uint64  `json:"available_bytes"`
	Usage     float64 `json:"usage_percent"`
}

func (s SystemInfo) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SystemInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), s)
}
