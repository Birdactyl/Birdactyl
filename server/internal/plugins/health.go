package plugins

import (
	"context"
	"log"
	"time"

	pb "birdactyl-panel-backend/internal/plugins/proto"
)

var healthStop chan struct{}

func StartHealthCheck() {
	healthStop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				checkPlugins()
			case <-healthStop:
				return
			}
		}
	}()
	log.Println("[plugins] health check started")
}

func StopHealthCheck() {
	if healthStop != nil {
		close(healthStop)
	}
}

func checkPlugins() {
	for _, p := range GetRegistry().All() {
		go checkPlugin(p.Config.ID, p.Online)
	}
}

func checkPlugin(pluginID string, wasOnline bool) {
	plugin := GetRegistry().Get(pluginID)
	if plugin == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	info, err := plugin.Client.GetInfo(ctx, &pb.Empty{})
	if err != nil {
		if wasOnline {
			GetRegistry().SetOnline(pluginID, false)
			log.Printf("[plugins] %s went offline", pluginID)
		}
		return
	}

	if !wasOnline {
		GetRegistry().SetOnline(pluginID, true)
		for _, sched := range info.Schedules {
			RegisterSchedule(pluginID, sched.Id, sched.Cron)
		}
		log.Printf("[plugins] reconnected to %s v%s", info.Name, info.Version)
	}
}
