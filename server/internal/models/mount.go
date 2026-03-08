package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Mount struct {
	ID            uuid.UUID `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"type:varchar(255);not null"`
	Description   string    `json:"description" gorm:"type:text"`
	Source        string    `json:"source" gorm:"type:text;not null"`
	Target        string    `json:"target" gorm:"type:text;not null"`
	ReadOnly      bool      `json:"read_only" gorm:"default:false"`
	UserMountable bool      `json:"user_mountable" gorm:"default:false"`
	Navigable     bool      `json:"navigable" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Servers  []Server  `json:"servers,omitempty" gorm:"many2many:server_mounts;"`
	Nodes    []Node    `json:"nodes,omitempty" gorm:"many2many:node_mounts;"`
	Packages []Package `json:"packages,omitempty" gorm:"many2many:package_mounts;"`
}

func (m *Mount) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
