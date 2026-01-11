package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Subuser struct {
	ID          uuid.UUID      `gorm:"primaryKey" json:"id"`
	ServerID    uuid.UUID      `gorm:"index;not null" json:"server_id"`
	UserID      uuid.UUID      `gorm:"index;not null" json:"user_id"`
	Permissions datatypes.JSON `gorm:"type:json" json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	Server *Server `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (s *Subuser) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Permissions == nil {
		s.Permissions = []byte("[]")
	}
	return nil
}

func (s *Subuser) GetPermissions() []string {
	var perms []string
	json.Unmarshal(s.Permissions, &perms)
	return perms
}

func (s *Subuser) HasPermission(perm string) bool {
	return HasPermission(s.GetPermissions(), perm)
}
