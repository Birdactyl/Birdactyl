package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityLog struct {
	ID          uuid.UUID `gorm:"primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"index;not null" json:"user_id"`
	Username    string    `gorm:"type:varchar(255);not null" json:"username"`
	Action      string    `gorm:"type:varchar(100);not null;index" json:"action"`
	Description string    `gorm:"type:varchar(500)" json:"description"`
	IP          string    `gorm:"type:varchar(45)" json:"ip"`
	UserAgent   string    `gorm:"type:varchar(500)" json:"user_agent"`
	IsAdmin     bool      `gorm:"index" json:"is_admin"`
	Metadata    string    `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
}

func (a *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
