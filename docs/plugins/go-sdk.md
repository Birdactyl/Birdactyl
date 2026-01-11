# Go SDK

The Go SDK provides a clean, chainable API for building Birdactyl plugins.

## Installation

```
go get github.com/Birdactyl/Birdactyl-Go-SDK
```

## Creating a Plugin

```go
package main

import birdactyl "github.com/Birdactyl/Birdactyl-Go-SDK"

func main() {
    plugin := birdactyl.New("plugin-id", "1.0.0")
    plugin.Start("localhost:50050")
}
```

## Plugin Options

Chain these methods to configure your plugin:

```go
plugin := birdactyl.New("my-plugin", "1.0.0").
    SetName("My Plugin").
    UseDataDir().
    OnStart(func() {
        plugin.Log("Plugin started!")
    })
```

## Events

Register event handlers with `OnEvent`:

```go
plugin.OnEvent("server.start", func(e birdactyl.Event) birdactyl.EventResult {
    serverID := e.Data["server_id"]
    return birdactyl.Allow()
})

plugin.OnEvent("user.create", func(e birdactyl.Event) birdactyl.EventResult {
    if e.Data["email"] == "[email]" {
        return birdactyl.Block("This email is not allowed")
    }
    return birdactyl.Allow()
})
```

The `Event` struct:
- `Type` - Event type string
- `Data` - Map of event data
- `Sync` - Whether the panel waits for your response

Return `birdactyl.Allow()` to let the action proceed, or `birdactyl.Block("reason")` to stop it.

## Routes

Add HTTP endpoints with `Route`:

```go
plugin.Route("GET", "/api/plugins/my-plugin/status", func(r birdactyl.Request) birdactyl.Response {
    return birdactyl.JSON(map[string]string{"status": "ok"})
})

plugin.Route("POST", "/api/plugins/my-plugin/action", func(r birdactyl.Request) birdactyl.Response {
    name := r.Body["name"].(string)
    return birdactyl.JSON(map[string]interface{}{"received": name})
})
```

The `Request` struct:
- `Method` - HTTP method
- `Path` - Request path
- `Headers` - Request headers
- `Query` - Query parameters
- `Body` - Parsed JSON body
- `RawBody` - Raw body bytes
- `UserID` - Authenticated user's ID

Response helpers:
- `birdactyl.JSON(data)` - JSON response with `{"success": true, "data": ...}`
- `birdactyl.Error(status, message)` - Error response
- `birdactyl.Text(string)` - Plain text response

## Mixins

Intercept panel operations with `Mixin`:

```go
plugin.Mixin(birdactyl.MixinServerCreate, func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
    name := ctx.GetString("name")
    ctx.Set("name", "[Server] " + name)
    return ctx.Next()
})
```

Use `MixinWithPriority` to control execution order (lower runs first):

```go
plugin.MixinWithPriority(birdactyl.MixinServerCreate, -10, handler)
```

MixinContext methods:
- `Get(key)` - Get input value
- `GetString(key)`, `GetInt(key)`, `GetBool(key)` - Typed getters
- `Input()` - Full input map
- `Set(key, value)` - Modify input for next handler
- `ChainData()` - Data shared between mixin handlers
- `Notify(title, message, type)` - Send notification to user
- `Next()` - Continue to next handler
- `Return(data)` - Return early with custom response
- `Error(message)` - Return an error

## Schedules

Run tasks on a cron schedule:

```go
plugin.Schedule("cleanup", "0 0 * * *", func() {
    plugin.Log("Running daily cleanup")
})

plugin.Schedule("heartbeat", "*/5 * * * *", func() {
    plugin.Log("Heartbeat every 5 minutes")
})
```

## Addon Types

Define custom addon installation handlers:

```go
plugin.AddonType("my-addon-type", "My Addon", "Custom addon handler", 
    func(req birdactyl.AddonTypeRequest) birdactyl.AddonTypeResponse {
        return birdactyl.AddonSuccess("Installed successfully",
            birdactyl.DownloadFile(req.DownloadURL, req.InstallPath, nil),
            birdactyl.ExtractArchive(req.InstallPath),
        )
    })
```

Available actions:
- `DownloadFile(url, path, headers)`
- `ExtractArchive(path)`
- `DeleteFile(path)`
- `CreateFolder(path)`
- `WriteFile(path, content)`
- `ProxyToNode(endpoint, payload)`

## Panel API

Access the panel API through `plugin.API()`:

```go
api := plugin.API()

server, _ := api.GetServer("server-id")
servers := api.ListServers()
api.StartServer("server-id")
api.SendCommand("server-id", "say Hello")

user, _ := api.GetUser("user-id")
users := api.ListUsers()

files := api.ListFiles("server-id", "/")
content, _ := api.ReadFile("server-id", "/server.properties")
api.WriteFile("server-id", "/motd.txt", []byte("Welcome!"))
```

See [Panel API](panel-api.md) for the full reference.

## Async API

For non-blocking operations, use `plugin.Async()`:

```go
async := plugin.Async()

future := async.GetServer("server-id")
server, err := future.Get()

async.GetServer("server-id").Then(func(s *birdactyl.Server) {
    plugin.Log("Got server: " + s.Name)
}).Catch(func(err error) {
    plugin.Log("Error: " + err.Error())
})

results, _ := birdactyl.All(
    async.GetServer("server-1"),
    async.GetServer("server-2"),
).Get()
```

## Configuration

Use `HotConfig` for auto-reloading config files:

```go
type Config struct {
    Enabled bool   `yaml:"enabled"`
    Message string `yaml:"message"`
}

config := birdactyl.NewHotConfig(plugin.DataPath("config.yaml"), Config{
    Enabled: true,
    Message: "Hello",
})

config.DynamicConfig()

config.OnChange(func(c Config) {
    plugin.Log("Config reloaded!")
})

current := config.Get()
```

## Key-Value Storage

Store simple data in the panel's KV store:

```go
api.SetKV("my-plugin:counter", "42")
value, found := api.GetKV("my-plugin:counter")
api.DeleteKV("my-plugin:counter")
```

## HTTP Client

Make external HTTP requests through the panel:

```go
resp := api.HTTPGet("https://api.example.com/data", map[string]string{
    "Authorization": "Bearer token",
})

if resp.Error == "" && resp.Status == 200 {
    var data map[string]interface{}
    json.Unmarshal(resp.Body, &data)
}
```

## Inter-Plugin Communication

Call methods on other plugins:

```go
response, err := api.CallPlugin("other-plugin", "getData", []byte(`{"id": "123"}`))
```

## Console Streaming

Stream server console output in real-time:

```go
stream, _ := api.StreamConsole("server-id", true, 100)
defer stream.Close()

for {
    line, err := stream.Recv()
    if err != nil {
        break
    }
    plugin.Log("Console: " + line)
}
```

## User Interface

Add custom pages, tabs, and sidebar items to the panel. See [UI](ui.md) for the full guide.

```go
package main

import (
    "embed"
    birdactyl "github.com/Birdactyl/Birdactyl-Go-SDK"
)

//go:embed dist/bundle.js
var uiBundle embed.FS

func main() {
    plugin := birdactyl.New("my-plugin", "1.0.0")

    plugin.UI().
        HasBundle().
        EmbedBundle(uiBundle, "dist/bundle.js").
        Page("/", "DashboardPage").Title("Dashboard").Icon("home").Done().
        Page("/settings", "SettingsPage").Title("Settings").Icon("settings").Done().
        Tab("server-tab", "ServerTab", birdactyl.TabTargetServer, "My Plugin").
            Icon("puzzle").Order(100).Done().
        Tab("settings-tab", "SettingsTab", birdactyl.TabTargetUserSettings, "My Plugin").
            Icon("puzzle").Order(10).Done().
        SidebarItem("my-plugin", "My Plugin", "/console/plugins/my-plugin", birdactyl.SidebarSectionPlatform).
            Icon("puzzle").Order(50).
            Child("Dashboard", "/console/plugins/my-plugin").
            Child("Settings", "/console/plugins/my-plugin/settings").
            Done()

    plugin.Start("localhost:50050")
}
```

### UI Builder Methods

| Method | Description |
|--------|-------------|
| `HasBundle()` | Indicate the plugin has a UI bundle |
| `EmbedBundle(fs, path)` | Embed bundle from embedded filesystem |
| `BundleBytes(data)` | Set bundle from raw bytes |
| `Page(path, component)` | Add a page |
| `Tab(id, component, target, label)` | Add a tab |
| `SidebarItem(id, label, href, section)` | Add a sidebar item |

### Tab Targets

| Constant | Description |
|----------|-------------|
| `TabTargetServer` | Server detail page |
| `TabTargetUserSettings` | User settings page |

### Sidebar Sections

| Constant | Description |
|----------|-------------|
| `SidebarSectionNav` | Main navigation |
| `SidebarSectionPlatform` | Platform section |
| `SidebarSectionAdmin` | Admin section |
