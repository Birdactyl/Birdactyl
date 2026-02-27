package models

import (
	"time"

	"github.com/google/uuid"
)

type IPBan struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	IP        string    `gorm:"type:varchar(45);uniqueIndex;not null" json:"ip"`
	Reason    string    `gorm:"type:varchar(500)" json:"reason"`
	BannedBy  uuid.UUID `gorm:"type:char(36)" json:"banned_by"`
	CreatedAt time.Time `json:"created_at"`
}
