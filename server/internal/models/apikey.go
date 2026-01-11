package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKey struct {
	ID         uuid.UUID  `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"index;not null" json:"user_id"`
	Name       string     `gorm:"type:varchar(255);not null" json:"name"`
	KeyHash    string     `gorm:"type:varchar(255);not null" json:"-"`
	KeyPrefix  string     `gorm:"type:varchar(32);not null" json:"key_prefix"`
	ExpiresAt  *time.Time `gorm:"index" json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (k *APIKey) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}

func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}
