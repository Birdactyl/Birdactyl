package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                 uuid.UUID      `gorm:"primaryKey" json:"id"`
	Email              string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Username           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"username"`
	PasswordHash       string         `gorm:"type:varchar(255);not null" json:"-"`
	IsAdmin            bool           `gorm:"default:false" json:"is_admin"`
	IsBanned           bool           `gorm:"default:false" json:"is_banned"`
	ForcePasswordReset bool           `gorm:"default:false" json:"force_password_reset"`
	RAMLimit           *int           `gorm:"default:null" json:"ram_limit"`
	CPULimit           *int           `gorm:"default:null" json:"cpu_limit"`
	DiskLimit          *int           `gorm:"default:null" json:"disk_limit"`
	ServerLimit        *int           `gorm:"default:null" json:"server_limit"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	Sessions           []Session      `gorm:"foreignKey:UserID" json:"-"`
	RegisterIP         string         `gorm:"type:varchar(45);not null" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type Session struct {
	ID            uuid.UUID  `gorm:"primaryKey" json:"id"`
	UserID        uuid.UUID  `gorm:"index;not null" json:"user_id"`
	RefreshToken  string     `gorm:"type:varchar(500);uniqueIndex;not null" json:"-"`
	PreviousToken string     `gorm:"type:varchar(500);index" json:"-"`
	UserAgent     string     `gorm:"type:varchar(500)" json:"user_agent"`
	IP            string     `gorm:"type:varchar(45)" json:"ip"`
	ExpiresAt     time.Time  `gorm:"index;not null" json:"expires_at"`
	LastRefreshAt *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type IPRegistration struct {
	ID        uint      `gorm:"primaryKey"`
	IP        string    `gorm:"type:varchar(45);index;not null"`
	UserID    uuid.UUID `gorm:"not null"`
	CreatedAt time.Time
}
