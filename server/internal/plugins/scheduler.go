package plugins

import (
	"context"
	"log"
	"sync"
	"time"

	pb "birdactyl-panel-backend/internal/plugins/proto"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
	jobs map[string]cron.EntryID
	mu   sync.Mutex
}

var scheduler *Scheduler

func StartScheduler() {
	scheduler = &Scheduler{
		cron: cron.New(),
		jobs: make(map[string]cron.EntryID),
	}
	scheduler.cron.Start()
	log.Println("[plugins] scheduler started")
}

func StopScheduler() {
	if scheduler != nil {
		scheduler.cron.Stop()
	}
}

func RegisterSchedule(pluginID, scheduleID, cronExpr string) error {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	key := pluginID + ":" + scheduleID

	if oldID, exists := scheduler.jobs[key]; exists {
		scheduler.cron.Remove(oldID)
	}

	entryID, err := scheduler.cron.AddFunc(cronExpr, func() {
		if ps := GetStreamRegistry().Get(pluginID); ps != nil {
			ps.SendSchedule(&pb.ScheduleRequest{ScheduleId: scheduleID})
			return
		}

		plugin := GetRegistry().Get(pluginID)
		if plugin == nil || !plugin.Online {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := plugin.Client.OnSchedule(ctx, &pb.ScheduleRequest{ScheduleId: scheduleID})
		if err != nil {
			log.Printf("[plugins] schedule %s failed for %s: %v", scheduleID, pluginID, err)
		}
	})
	if err != nil {
		return err
	}

	scheduler.jobs[key] = entryID
	log.Printf("[plugins] registered schedule %s for %s: %s", scheduleID, pluginID, cronExpr)
	return nil
}

func UnregisterSchedules(pluginID string) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	for key, entryID := range scheduler.jobs {
		if len(key) > len(pluginID) && key[:len(pluginID)+1] == pluginID+":" {
			scheduler.cron.Remove(entryID)
			delete(scheduler.jobs, key)
		}
	}
}
