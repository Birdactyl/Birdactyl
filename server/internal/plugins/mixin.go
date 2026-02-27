package plugins

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"sync"
	"time"

	pb "birdactyl-panel-backend/internal/plugins/proto"

	"github.com/google/uuid"
)

type MixinEntry struct {
	PluginID string
	Target   string
	Priority int
}

type MixinRegistry struct {
	mu     sync.RWMutex
	mixins map[string][]MixinEntry
}

var mixinRegistry = &MixinRegistry{
	mixins: make(map[string][]MixinEntry),
}

func GetMixinRegistry() *MixinRegistry {
	return mixinRegistry
}

func (r *MixinRegistry) Register(pluginID, target string, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mixins[target] = append(r.mixins[target], MixinEntry{
		PluginID: pluginID,
		Target:   target,
		Priority: priority,
	})

	sort.Slice(r.mixins[target], func(i, j int) bool {
		return r.mixins[target][i].Priority > r.mixins[target][j].Priority
	})

	log.Printf("[mixin] registered %s for target %s (priority: %d)", pluginID, target, priority)
}

func (r *MixinRegistry) Unregister(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for target, entries := range r.mixins {
		filtered := make([]MixinEntry, 0)
		for _, e := range entries {
			if e.PluginID != pluginID {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) == 0 {
			delete(r.mixins, target)
		} else {
			r.mixins[target] = filtered
		}
	}
}

func (r *MixinRegistry) Get(target string) []MixinEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := r.mixins[target]
	result := make([]MixinEntry, len(entries))
	copy(result, entries)
	return result
}

func (r *MixinRegistry) Has(target string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.mixins[target]) > 0
}

type MixinNotification struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type MixinError struct {
	Message       string
	Notifications []MixinNotification
}

func (e *MixinError) Error() string {
	return e.Message
}

var pendingNotifications []MixinNotification
var notificationsMu sync.Mutex

func CollectNotifications() []MixinNotification {
	notificationsMu.Lock()
	defer notificationsMu.Unlock()
	n := pendingNotifications
	pendingNotifications = nil
	return n
}

func ExecuteMixin(target string, input map[string]interface{}, originalFn func(map[string]interface{}) (interface{}, error)) (interface{}, error) {
	entries := GetMixinRegistry().Get(target)
	if len(entries) == 0 {
		return originalFn(input)
	}

	requestID := uuid.New().String()
	currentInput := input
	var chainData map[string]interface{}
	var notifications []MixinNotification

	for _, entry := range entries {
		inputBytes, _ := json.Marshal(currentInput)
		chainBytes, _ := json.Marshal(chainData)

		req := &pb.MixinRequest{
			Target:    target,
			RequestId: requestID,
			Input:     inputBytes,
			ChainData: chainBytes,
		}

		var resp *pb.MixinResponse
		var err error

		if ps := GetStreamRegistry().Get(entry.PluginID); ps != nil {
			resp, err = ps.SendMixin(req)
		} else if plugin := GetRegistry().Get(entry.PluginID); plugin != nil && plugin.Online {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			resp, err = plugin.Client.OnMixin(ctx, req)
			cancel()
			if err != nil {
				GetRegistry().SetOnline(entry.PluginID, false)
			}
		} else {
			continue
		}

		if err != nil {
			log.Printf("[mixin] error calling %s for %s: %v", entry.PluginID, target, err)
			continue
		}

		for _, n := range resp.Notifications {
			notifications = append(notifications, MixinNotification{
				Title:   n.Title,
				Message: n.Message,
				Type:    n.Type,
			})
		}

		switch resp.Action {
		case pb.MixinResponse_RETURN:
			var output map[string]interface{}
			json.Unmarshal(resp.Output, &output)
			storeNotifications(notifications)
			return output, nil

		case pb.MixinResponse_ERROR:
			return nil, &MixinError{Message: resp.Error, Notifications: notifications}

		case pb.MixinResponse_NEXT:
			if len(resp.ModifiedInput) > 0 {
				json.Unmarshal(resp.ModifiedInput, &currentInput)
			}
		}
	}

	storeNotifications(notifications)
	return originalFn(currentInput)
}

func storeNotifications(n []MixinNotification) {
	if len(n) == 0 {
		return
	}
	notificationsMu.Lock()
	pendingNotifications = append(pendingNotifications, n...)
	notificationsMu.Unlock()
}
