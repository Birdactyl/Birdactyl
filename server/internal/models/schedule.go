package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Schedule struct {
	ID             uuid.UUID      `json:"id" gorm:"primaryKey"`
	ServerID       uuid.UUID      `json:"server_id" gorm:"index;not null"`
	Name           string         `json:"name" gorm:"type:varchar(255);not null"`
	CronExpression string         `json:"cron_expression" gorm:"type:varchar(100);not null"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	OnlyWhenOnline bool           `json:"only_when_online" gorm:"default:false"`
	LastRunAt      *time.Time     `json:"last_run_at"`
	NextRunAt      *time.Time     `json:"next_run_at"`
	Tasks          datatypes.JSON `json:"tasks" gorm:"type:json"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type ScheduleTask struct {
	Sequence int    `json:"sequence"`
	Action   string `json:"action"`
	Payload  string `json:"payload"`
}

func (s *Schedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Tasks == nil {
		s.Tasks = []byte("[]")
	}
	return nil
}
