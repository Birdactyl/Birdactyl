package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ServerStatus string

const (
	ServerStatusInstalling ServerStatus = "installing"
	ServerStatusRunning    ServerStatus = "running"
	ServerStatusStopped    ServerStatus = "stopped"
	ServerStatusSuspended  ServerStatus = "suspended"
	ServerStatusFailed     ServerStatus = "failed"
)

type Server struct {
	ID          uuid.UUID      `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Description string         `json:"description" gorm:"type:varchar(500)"`
	UserID      uuid.UUID      `json:"user_id" gorm:"not null;index"`
	NodeID      uuid.UUID      `json:"node_id" gorm:"not null;index"`
	PackageID   uuid.UUID      `json:"package_id" gorm:"not null"`
	Status      ServerStatus   `json:"status" gorm:"type:varchar(20);default:'installing'"`
	IsSuspended bool           `json:"is_suspended" gorm:"default:false"`
	ContainerID string         `json:"container_id,omitempty" gorm:"type:varchar(64)"`
	Memory      int            `json:"memory" gorm:"not null"`
	CPU         int            `json:"cpu" gorm:"not null"`
	Disk        int            `json:"disk" gorm:"not null"`
	Startup     string         `json:"startup" gorm:"type:text"`
	DockerImage string         `json:"docker_image" gorm:"type:varchar(500)"`
	Ports       datatypes.JSON `json:"ports" gorm:"type:json"`
	Variables   datatypes.JSON `json:"variables" gorm:"type:json"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Node    *Node    `json:"node,omitempty" gorm:"foreignKey:NodeID"`
	Package *Package `json:"package,omitempty" gorm:"foreignKey:PackageID"`
}

type ServerPort struct {
	Port    int  `json:"port"`
	Primary bool `json:"primary,omitempty"`
}

func (s *Server) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Ports == nil {
		s.Ports = []byte("[]")
	}
	if s.Variables == nil {
		s.Variables = []byte("{}")
	}
	return nil
}
