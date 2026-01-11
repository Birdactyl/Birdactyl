package plugins

import (
	"os"
	"path/filepath"
	"sync"

	pb "birdactyl-panel-backend/internal/plugins/proto"
)

type UIManifest struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	HasBundle    bool            `json:"hasBundle"`
	Pages        []UIPage        `json:"pages"`
	Tabs         []UITab         `json:"tabs"`
	SidebarItems []UISidebarItem `json:"sidebarItems"`
	BundleData   []byte          `json:"-"`
}

type UIPage struct {
	Path      string `json:"path"`
	Component string `json:"component"`
	Title     string `json:"title,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Guard     string `json:"guard,omitempty"`
}

type UITab struct {
	ID        string `json:"id"`
	Component string `json:"component"`
	Target    string `json:"target"`
	Label     string `json:"label"`
	Icon      string `json:"icon,omitempty"`
	Order     int    `json:"order,omitempty"`
}

type UISidebarItem struct {
	ID       string           `json:"id"`
	Label    string           `json:"label"`
	Icon     string           `json:"icon,omitempty"`
	Href     string           `json:"href"`
	Section  string           `json:"section"`
	Order    int              `json:"order,omitempty"`
	Guard    string           `json:"guard,omitempty"`
	Children []UISidebarChild `json:"children,omitempty"`
}

type UISidebarChild struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type UIRegistry struct {
	mu        sync.RWMutex
	manifests map[string]*UIManifest
}

var uiRegistry = &UIRegistry{
	manifests: make(map[string]*UIManifest),
}

func GetUIRegistry() *UIRegistry {
	return uiRegistry
}

func (r *UIRegistry) Register(pluginID, name, version string, ui *pb.PluginUIInfo) {
	if ui == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	manifest := &UIManifest{
		ID:           pluginID,
		Name:         name,
		Version:      version,
		HasBundle:    ui.HasBundle,
		Pages:        make([]UIPage, 0, len(ui.Pages)),
		Tabs:         make([]UITab, 0, len(ui.Tabs)),
		SidebarItems: make([]UISidebarItem, 0, len(ui.SidebarItems)),
		BundleData:   ui.BundleData,
	}

	for _, p := range ui.Pages {
		manifest.Pages = append(manifest.Pages, UIPage{
			Path:      p.Path,
			Component: p.Component,
			Title:     p.Title,
			Icon:      p.Icon,
			Guard:     p.Guard,
		})
	}

	for _, t := range ui.Tabs {
		manifest.Tabs = append(manifest.Tabs, UITab{
			ID:        t.Id,
			Component: t.Component,
			Target:    t.Target,
			Label:     t.Label,
			Icon:      t.Icon,
			Order:     int(t.Order),
		})
	}

	for _, s := range ui.SidebarItems {
		item := UISidebarItem{
			ID:       s.Id,
			Label:    s.Label,
			Icon:     s.Icon,
			Href:     s.Href,
			Section:  s.Section,
			Order:    int(s.Order),
			Guard:    s.Guard,
			Children: make([]UISidebarChild, 0, len(s.Children)),
		}
		for _, c := range s.Children {
			item.Children = append(item.Children, UISidebarChild{
				Label: c.Label,
				Href:  c.Href,
			})
		}
		manifest.SidebarItems = append(manifest.SidebarItems, item)
	}

	r.manifests[pluginID] = manifest
}

func (r *UIRegistry) Unregister(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.manifests, pluginID)
}

func (r *UIRegistry) Get(pluginID string) *UIManifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.manifests[pluginID]
}

func (r *UIRegistry) All() []*UIManifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*UIManifest, 0, len(r.manifests))
	for _, m := range r.manifests {
		result = append(result, m)
	}
	return result
}

func GetPluginBundlePath(pluginID, pluginsDir string) string {
	patterns := []string{
		filepath.Join(pluginsDir, pluginID, "ui", "dist", "bundle.js"),
		filepath.Join(pluginsDir, pluginID, "ui", "bundle.js"),
		filepath.Join(pluginsDir, pluginID, "bundle.js"),
		filepath.Join(pluginsDir, pluginID+"_ui", "bundle.js"),
	}

	for _, path := range patterns {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
