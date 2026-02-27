package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	scheduler     *cron.Cron
	schedulerOnce sync.Once
	entryMap      = make(map[uuid.UUID]cron.EntryID)
	entryMapMu    sync.RWMutex
)

func InitScheduler() {
	schedulerOnce.Do(func() {
		scheduler = cron.New(cron.WithSeconds())
		scheduler.Start()
		log.Println("[scheduler] started")
		go loadExistingSchedules()
	})
}

func StopScheduler() {
	if scheduler != nil {
		scheduler.Stop()
	}
}

func loadExistingSchedules() {
	var schedules []models.Schedule
	database.DB.Where("is_active = ?", true).Find(&schedules)
	for _, s := range schedules {
		RegisterSchedule(&s)
	}
}

func RegisterSchedule(s *models.Schedule) error {
	if scheduler == nil || !s.IsActive {
		return nil
	}

	entryMapMu.Lock()
	if oldID, exists := entryMap[s.ID]; exists {
		scheduler.Remove(oldID)
	}
	entryMapMu.Unlock()

	entryID, err := scheduler.AddFunc(s.CronExpression, func() {
		executeSchedule(s.ID)
	})
	if err != nil {
		return err
	}

	entryMapMu.Lock()
	entryMap[s.ID] = entryID
	entryMapMu.Unlock()

	entry := scheduler.Entry(entryID)
	nextRun := entry.Next
	database.DB.Model(&models.Schedule{}).Where("id = ?", s.ID).Update("next_run_at", nextRun)

	return nil
}

func UnregisterSchedule(id uuid.UUID) {
	entryMapMu.Lock()
	defer entryMapMu.Unlock()
	if entryID, exists := entryMap[id]; exists {
		scheduler.Remove(entryID)
		delete(entryMap, id)
	}
}

func executeSchedule(scheduleID uuid.UUID) {
	var schedule models.Schedule
	if err := database.DB.First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return
	}

	if !schedule.IsActive {
		return
	}

	var server models.Server
	if err := database.DB.First(&server, "id = ?", schedule.ServerID).Error; err != nil {
		return
	}

	if schedule.OnlyWhenOnline {
		stats := GetServerStats(schedule.ServerID)
		if stats == nil || stats.State != "running" {
			updateNextRun(scheduleID)
			return
		}
	}

	var tasks []models.ScheduleTask
	json.Unmarshal(schedule.Tasks, &tasks)

	for _, task := range tasks {
		executeTask(schedule.ServerID, task)
	}

	now := time.Now()
	database.DB.Model(&schedule).Update("last_run_at", now)
	updateNextRun(scheduleID)
}

func executeTask(serverID uuid.UUID, task models.ScheduleTask) error {
	switch task.Action {
	case "command":
		return SendCommand(serverID, task.Payload)
	case "power":
		switch task.Payload {
		case "start":
			return SendStartServer(serverID)
		case "stop":
			return SendStopServer(serverID)
		case "restart":
			return SendRestartServer(serverID)
		case "kill":
			return SendKillServer(serverID)
		}
	case "delay":
		var seconds int
		fmt.Sscanf(task.Payload, "%d", &seconds)
		if seconds > 0 && seconds <= 300 {
			time.Sleep(time.Duration(seconds) * time.Second)
		}
	case "backup":
		return CreateServerArchive(serverID)
	}
	return nil
}

func updateNextRun(scheduleID uuid.UUID) {
	entryMapMu.RLock()
	entryID, exists := entryMap[scheduleID]
	entryMapMu.RUnlock()

	if exists && scheduler != nil {
		entry := scheduler.Entry(entryID)
		database.DB.Model(&models.Schedule{}).Where("id = ?", scheduleID).Update("next_run_at", entry.Next)
	}
}

func RunScheduleNow(scheduleID uuid.UUID) error {
	go executeSchedule(scheduleID)
	return nil
}

func GetSchedulesByServer(serverID uuid.UUID) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := database.DB.Where("server_id = ?", serverID).Order("created_at desc").Find(&schedules).Error
	return schedules, err
}

func GetScheduleByID(id uuid.UUID) (*models.Schedule, error) {
	var schedule models.Schedule
	err := database.DB.First(&schedule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func CreateSchedule(schedule *models.Schedule) error {
	if err := database.DB.Create(schedule).Error; err != nil {
		return err
	}
	if schedule.IsActive {
		RegisterSchedule(schedule)
	}
	return nil
}

func UpdateSchedule(id uuid.UUID, updates map[string]interface{}) (*models.Schedule, error) {
	var schedule models.Schedule
	if err := database.DB.First(&schedule, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&schedule).Updates(updates).Error; err != nil {
		return nil, err
	}

	database.DB.First(&schedule, "id = ?", id)

	UnregisterSchedule(id)
	if schedule.IsActive {
		RegisterSchedule(&schedule)
	}

	return &schedule, nil
}

func DeleteSchedule(id uuid.UUID) error {
	UnregisterSchedule(id)
	return database.DB.Delete(&models.Schedule{}, "id = ?", id).Error
}
