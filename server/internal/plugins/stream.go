package plugins

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	pb "birdactyl-panel-backend/internal/plugins/proto"

	"github.com/google/uuid"
)

type PluginStream struct {
	ID       string
	Info     *pb.PluginInfo
	stream   pb.PanelService_ConnectServer
	pending  map[string]chan *pb.PluginMessage
	mu       sync.RWMutex
	closed   bool
}

type StreamRegistry struct {
	mu      sync.RWMutex
	streams map[string]*PluginStream
}

var streamRegistry = &StreamRegistry{
	streams: make(map[string]*PluginStream),
}

func GetStreamRegistry() *StreamRegistry {
	return streamRegistry
}

func (r *StreamRegistry) Add(ps *PluginStream) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.streams[ps.ID] = ps
}

func (r *StreamRegistry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.streams, id)
}

func (r *StreamRegistry) Get(id string) *PluginStream {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.streams[id]
}

func (r *StreamRegistry) All() []*PluginStream {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*PluginStream, 0, len(r.streams))
	for _, ps := range r.streams {
		result = append(result, ps)
	}
	return result
}

func (ps *PluginStream) Send(msg *pb.PanelMessage) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.closed {
		return nil
	}
	return ps.stream.Send(msg)
}

func (ps *PluginStream) SendRequest(msg *pb.PanelMessage, timeout time.Duration) (*pb.PluginMessage, error) {
	reqID := uuid.New().String()
	msg.RequestId = reqID

	ch := make(chan *pb.PluginMessage, 1)
	ps.mu.Lock()
	ps.pending[reqID] = ch
	ps.mu.Unlock()

	defer func() {
		ps.mu.Lock()
		delete(ps.pending, reqID)
		ps.mu.Unlock()
	}()

	if err := ps.Send(msg); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (ps *PluginStream) HandleResponse(msg *pb.PluginMessage) {
	ps.mu.RLock()
	ch, ok := ps.pending[msg.RequestId]
	ps.mu.RUnlock()
	if ok {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (ps *PluginStream) Close() {
	ps.mu.Lock()
	ps.closed = true
	for _, ch := range ps.pending {
		close(ch)
	}
	ps.pending = make(map[string]chan *pb.PluginMessage)
	ps.mu.Unlock()
}

func (ps *PluginStream) SendEvent(ev *pb.Event) (*pb.EventResponse, error) {
	resp, err := ps.SendRequest(&pb.PanelMessage{
		Payload: &pb.PanelMessage_Event{Event: ev},
	}, 10*time.Second)
	if err != nil {
		return nil, err
	}
	if r := resp.GetEventResponse(); r != nil {
		return r, nil
	}
	return &pb.EventResponse{Allow: true}, nil
}

func (ps *PluginStream) SendHTTP(req *pb.HTTPRequest) (*pb.HTTPResponse, error) {
	resp, err := ps.SendRequest(&pb.PanelMessage{
		Payload: &pb.PanelMessage_Http{Http: req},
	}, 30*time.Second)
	if err != nil {
		return nil, err
	}
	if r := resp.GetHttpResponse(); r != nil {
		return r, nil
	}
	return nil, nil
}

func (ps *PluginStream) SendSchedule(req *pb.ScheduleRequest) error {
	_, err := ps.SendRequest(&pb.PanelMessage{
		Payload: &pb.PanelMessage_Schedule{Schedule: req},
	}, 60*time.Second)
	return err
}

func (ps *PluginStream) SendMixin(req *pb.MixinRequest) (*pb.MixinResponse, error) {
	resp, err := ps.SendRequest(&pb.PanelMessage{
		Payload: &pb.PanelMessage_Mixin{Mixin: req},
	}, 30*time.Second)
	if err != nil {
		return nil, err
	}
	if r := resp.GetMixinResponse(); r != nil {
		return r, nil
	}
	return &pb.MixinResponse{Action: pb.MixinResponse_NEXT}, nil
}

func (ps *PluginStream) Shutdown() {
	ps.Send(&pb.PanelMessage{
		Payload: &pb.PanelMessage_Shutdown{Shutdown: &pb.Empty{}},
	})
}

func (s *PanelServer) Connect(stream pb.PanelService_ConnectServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	info := msg.GetRegister()
	if info == nil {
		return nil
	}

	ps := &PluginStream{
		ID:      info.Id,
		Info:    info,
		stream:  stream,
		pending: make(map[string]chan *pb.PluginMessage),
	}

	GetStreamRegistry().Add(ps)
	defer func() {
		ps.Close()
		GetStreamRegistry().Remove(ps.ID)
		GetMixinRegistry().Unregister(ps.ID)
		GetAddonTypeRegistry().Unregister(ps.ID)
		GetUIRegistry().Unregister(ps.ID)
		UnregisterSchedules(ps.ID)
		log.Printf("[plugins] %s disconnected", info.Name)
	}()

	for _, sched := range info.Schedules {
		if err := RegisterSchedule(ps.ID, sched.Id, sched.Cron); err != nil {
			log.Printf("[plugins] failed to register schedule %s for %s: %v", sched.Id, ps.ID, err)
		}
	}

	for _, mixin := range info.Mixins {
		GetMixinRegistry().Register(ps.ID, mixin.Target, int(mixin.Priority))
	}

	for _, addonType := range info.AddonTypes {
		GetAddonTypeRegistry().Register(ps.ID, addonType.TypeId)
		log.Printf("[plugins] registered addon type %s for %s", addonType.TypeId, ps.ID)
	}

	if info.Ui != nil {
		GetUIRegistry().Register(ps.ID, info.Name, info.Version, info.Ui)
		log.Printf("[plugins] registered UI for %s (%d pages, %d tabs)",
			ps.ID, len(info.Ui.Pages), len(info.Ui.Tabs))
	}

	if err := stream.Send(&pb.PanelMessage{
		Payload: &pb.PanelMessage_Registered{Registered: &pb.Empty{}},
	}); err != nil {
		return err
	}

	log.Printf("[plugins] %s v%s connected (%d events, %d routes, %d schedules, %d mixins, %d addon types)",
		info.Name, info.Version, len(info.Events), len(info.Routes), len(info.Schedules), len(info.Mixins), len(info.AddonTypes))

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		ps.HandleResponse(msg)
	}
}

func StreamDispatchEvent(pluginID string, ev *pb.Event) (*pb.EventResponse, error) {
	ps := GetStreamRegistry().Get(pluginID)
	if ps == nil {
		return &pb.EventResponse{Allow: true}, nil
	}
	return ps.SendEvent(ev)
}

func StreamDispatchHTTP(pluginID string, req *pb.HTTPRequest) (*pb.HTTPResponse, error) {
	ps := GetStreamRegistry().Get(pluginID)
	if ps == nil {
		return nil, nil
	}
	return ps.SendHTTP(req)
}

func StreamDispatchSchedule(pluginID, scheduleID string) error {
	ps := GetStreamRegistry().Get(pluginID)
	if ps == nil {
		return nil
	}
	return ps.SendSchedule(&pb.ScheduleRequest{ScheduleId: scheduleID})
}

func StreamDispatchMixin(pluginID string, req *pb.MixinRequest) (*pb.MixinResponse, error) {
	ps := GetStreamRegistry().Get(pluginID)
	if ps == nil {
		return &pb.MixinResponse{Action: pb.MixinResponse_NEXT}, nil
	}
	return ps.SendMixin(req)
}

func GetAllStreamPlugins() []*PluginStream {
	return GetStreamRegistry().All()
}

func GetStreamPluginInfo(id string) *pb.PluginInfo {
	ps := GetStreamRegistry().Get(id)
	if ps == nil {
		return nil
	}
	return ps.Info
}

func StreamPluginHasRoute(pluginID, method, path string) bool {
	ps := GetStreamRegistry().Get(pluginID)
	if ps == nil || ps.Info == nil {
		return false
	}
	for _, r := range ps.Info.Routes {
		if (r.Method == "*" || r.Method == method) && matchRoutePath(r.Path, path) {
			return true
		}
	}
	return false
}

func matchRoutePath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		return len(path) >= len(pattern)-1 && path[:len(pattern)-1] == pattern[:len(pattern)-1]
	}
	return false
}

func StreamPluginHasEvent(pluginID, eventType string) bool {
	ps := GetStreamRegistry().Get(pluginID)
	if ps == nil || ps.Info == nil {
		return false
	}
	for _, e := range ps.Info.Events {
		if e == eventType {
			return true
		}
	}
	return false
}

func MarshalJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
