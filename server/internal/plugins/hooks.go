package plugins

import (
	"context"
	"log"
	"time"

	pb "birdactyl-panel-backend/internal/plugins/proto"
)

func Emit(event EventType, data map[string]string) (bool, string) {
	ev := &pb.Event{
		Type:      string(event),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
		Sync:      SyncEvents[event],
	}

	legacyPlugins := GetRegistry().GetByEvent(event)
	var streamPlugins []*PluginStream
	for _, ps := range GetStreamRegistry().All() {
		if StreamPluginHasEvent(ps.ID, string(event)) {
			streamPlugins = append(streamPlugins, ps)
		}
	}

	if len(legacyPlugins) == 0 && len(streamPlugins) == 0 {
		return true, ""
	}

	if ev.Sync {
		return emitSync(legacyPlugins, streamPlugins, ev)
	}

	go emitAsync(legacyPlugins, streamPlugins, ev)
	return true, ""
}

func emitSync(legacyPlugins []*Plugin, streamPlugins []*PluginStream, ev *pb.Event) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for _, p := range legacyPlugins {
		resp, err := p.Client.OnEvent(ctx, ev)
		if err != nil {
			log.Printf("[plugins] sync event %s to %s failed: %v", ev.Type, p.Config.ID, err)
			GetRegistry().SetOnline(p.Config.ID, false)
			continue
		}
		if !resp.Allow {
			return false, resp.Message
		}
	}

	for _, ps := range streamPlugins {
		resp, err := ps.SendEvent(ev)
		if err != nil {
			log.Printf("[plugins] sync event %s to %s failed: %v", ev.Type, ps.ID, err)
			continue
		}
		if !resp.Allow {
			return false, resp.Message
		}
	}

	return true, ""
}

func emitAsync(legacyPlugins []*Plugin, streamPlugins []*PluginStream, ev *pb.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, p := range legacyPlugins {
		_, err := p.Client.OnEvent(ctx, ev)
		if err != nil {
			log.Printf("[plugins] async event %s to %s failed: %v", ev.Type, p.Config.ID, err)
			GetRegistry().SetOnline(p.Config.ID, false)
		}
	}

	for _, ps := range streamPlugins {
		_, err := ps.SendEvent(ev)
		if err != nil {
			log.Printf("[plugins] async event %s to %s failed: %v", ev.Type, ps.ID, err)
		}
	}
}
