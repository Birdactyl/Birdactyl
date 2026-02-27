package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DatabaseHost struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"type:varchar(255);not null"`
	Host         string    `json:"host" gorm:"type:varchar(255);not null"`
	Port         int       `json:"port" gorm:"not null;default:3306"`
	Username     string    `json:"username" gorm:"type:varchar(255);not null"`
	Password     string    `json:"-" gorm:"type:varchar(255);not null"`
	MaxDatabases int       `json:"max_databases" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Databases []ServerDatabase `json:"databases,omitempty" gorm:"foreignKey:HostID"`
}

func (h *DatabaseHost) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

type ServerDatabase struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey"`
	ServerID     uuid.UUID `json:"server_id" gorm:"not null;index"`
	HostID       uuid.UUID `json:"host_id" gorm:"not null;index"`
	DatabaseName string    `json:"database_name" gorm:"type:varchar(64);not null"`
	Username     string    `json:"username" gorm:"type:varchar(32);not null"`
	Password     string    `json:"password" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at"`

	Server *Server       `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	Host   *DatabaseHost `json:"host,omitempty" gorm:"foreignKey:HostID"`
}

func (d *ServerDatabase) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
