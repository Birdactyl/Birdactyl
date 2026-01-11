package plugins

import (
	"encoding/json"
	"sync"
	"time"

	pb "birdactyl-panel-backend/internal/plugins/proto"
)

type AddonTypeRegistry struct {
	mu    sync.RWMutex
	types map[string]string
}

var addonTypeRegistry = &AddonTypeRegistry{
	types: make(map[string]string),
}

func GetAddonTypeRegistry() *AddonTypeRegistry {
	return addonTypeRegistry
}

func (r *AddonTypeRegistry) Register(pluginID, typeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[typeID] = pluginID
}

func (r *AddonTypeRegistry) Unregister(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for typeID, pid := range r.types {
		if pid == pluginID {
			delete(r.types, typeID)
		}
	}
}

func (r *AddonTypeRegistry) GetPlugin(typeID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.types[typeID]
}

func (r *AddonTypeRegistry) Has(typeID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.types[typeID]
	return ok
}

func (r *AddonTypeRegistry) All() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range r.types {
		result[k] = v
	}
	return result
}

type AddonTypeRequest struct {
	TypeID          string            `json:"type_id"`
	ServerID        string            `json:"server_id"`
	NodeID          string            `json:"node_id"`
	DownloadURL     string            `json:"download_url"`
	FileName        string            `json:"file_name"`
	InstallPath     string            `json:"install_path"`
	SourceInfo      map[string]string `json:"source_info"`
	ServerVariables map[string]string `json:"server_variables"`
}

type AddonTypeResponse struct {
	Success bool                 `json:"success"`
	Error   string               `json:"error,omitempty"`
	Message string               `json:"message,omitempty"`
	Actions []AddonInstallAction `json:"actions,omitempty"`
}

type AddonInstallAction struct {
	Type         int32             `json:"type"`
	URL          string            `json:"url,omitempty"`
	Path         string            `json:"path,omitempty"`
	Content      []byte            `json:"content,omitempty"`
	Command      string            `json:"command,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	NodePayload  []byte            `json:"node_payload,omitempty"`
	NodeEndpoint string            `json:"node_endpoint,omitempty"`
}

func DispatchAddonType(req AddonTypeRequest) (*AddonTypeResponse, error) {
	pluginID := GetAddonTypeRegistry().GetPlugin(req.TypeID)
	if pluginID == "" {
		return nil, nil
	}

	ps := GetStreamRegistry().Get(pluginID)
	if ps == nil {
		return nil, nil
	}

	pbReq := &pb.AddonTypeRequest{
		TypeId:          req.TypeID,
		ServerId:        req.ServerID,
		NodeId:          req.NodeID,
		DownloadUrl:     req.DownloadURL,
		FileName:        req.FileName,
		InstallPath:     req.InstallPath,
		SourceInfo:      req.SourceInfo,
		ServerVariables: req.ServerVariables,
	}

	resp, err := ps.SendAddonType(pbReq)
	if err != nil {
		return nil, err
	}

	result := &AddonTypeResponse{
		Success: resp.Success,
		Error:   resp.Error,
		Message: resp.Message,
	}

	for _, action := range resp.Actions {
		result.Actions = append(result.Actions, AddonInstallAction{
			Type:         int32(action.Type),
			URL:          action.Url,
			Path:         action.Path,
			Content:      action.Content,
			Command:      action.Command,
			Headers:      action.Headers,
			NodePayload:  action.NodePayload,
			NodeEndpoint: action.NodeEndpoint,
		})
	}

	return result, nil
}

func (ps *PluginStream) SendAddonType(req *pb.AddonTypeRequest) (*pb.AddonTypeResponse, error) {
	resp, err := ps.SendRequest(&pb.PanelMessage{
		Payload: &pb.PanelMessage_AddonType{AddonType: req},
	}, 60*time.Second)
	if err != nil {
		return nil, err
	}
	if r := resp.GetAddonTypeResponse(); r != nil {
		return r, nil
	}
	return &pb.AddonTypeResponse{Success: false, Error: "no response"}, nil
}

func MarshalAddonTypeRequest(req AddonTypeRequest) []byte {
	b, _ := json.Marshal(req)
	return b
}
