package plugins

import (
	"sync"

	pb "birdactyl-panel-backend/internal/plugins/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Plugin struct {
	Config PluginConfig
	Conn   *grpc.ClientConn
	Client pb.PluginServiceClient
	Online bool
	Info   *pb.PluginInfo
}

type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
}

var registry = &Registry{
	plugins: make(map[string]*Plugin),
}

func GetRegistry() *Registry {
	return registry
}

func connectToPlugin(address string) (*grpc.ClientConn, pb.PluginServiceClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return conn, pb.NewPluginServiceClient(conn), nil
}

func (r *Registry) Register(cfg PluginConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	r.plugins[cfg.ID] = &Plugin{
		Config: cfg,
		Conn:   conn,
		Client: pb.NewPluginServiceClient(conn),
		Online: true,
	}
	return nil
}

func (r *Registry) RegisterWithConn(cfg PluginConfig, conn *grpc.ClientConn, client pb.PluginServiceClient, info *pb.PluginInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	events := make([]EventType, len(info.Events))
	for i, e := range info.Events {
		events[i] = EventType(e)
	}
	cfg.Events = events

	r.plugins[cfg.ID] = &Plugin{
		Config: cfg,
		Conn:   conn,
		Client: client,
		Online: true,
		Info:   info,
	}
	return nil
}

func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.plugins[id]; ok {
		if p.Conn != nil {
			p.Conn.Close()
		}
		delete(r.plugins, id)
	}
}

func (r *Registry) Get(id string) *Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.plugins[id]
}

func (r *Registry) GetByEvent(event EventType) []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Plugin
	for _, p := range r.plugins {
		if !p.Online {
			continue
		}
		for _, e := range p.Config.Events {
			if e == event {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

func (r *Registry) All() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result
}

func (r *Registry) SetOnline(id string, online bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.plugins[id]; ok {
		p.Online = online
	}
}
