# Java SDK

The Java SDK lets you build Birdactyl plugins using familiar Java patterns and annotations.

## Installation

Add the dependency to your `pom.xml`:

```xml
<dependency>
    <groupId>io.birdactyl</groupId>
    <artifactId>birdactyl-sdk</artifactId>
    <version>0.2.0</version>
</dependency>
```

## Creating a Plugin

Extend `BirdactylPlugin` and call the constructor with your plugin ID and version:

```java
package com.example;

import io.birdactyl.sdk.*;

public class MyPlugin extends BirdactylPlugin {
    public MyPlugin() {
        super("my-plugin", "1.0.0");
        setName("My Plugin");
    }

    public static void main(String[] args) throws Exception {
        new MyPlugin().start("localhost:50050");
    }
}
```

## Plugin Options

Configure your plugin in the constructor:

```java
public MyPlugin() {
    super("my-plugin", "1.0.0");
    setName("My Plugin");
    useDataDir();
    onStart(() -> {
        api().log("info", "Plugin started!");
    });
}
```

## Events

Register event handlers with `onEvent`:

```java
onEvent("server.start", event -> {
    String serverId = event.get("server_id");
    api().log("info", "Server started: " + serverId);
    return EventResult.allow();
});

onEvent("user.create", event -> {
    if ("[email]".equals(event.get("email"))) {
        return EventResult.block("This email is not allowed");
    }
    return EventResult.allow();
});
```

The `Event` class:
- `getType()` - Event type string
- `getData()` - Map of event data
- `get(key)` - Get a specific data value
- `isSync()` - Whether the panel waits for your response

Return `EventResult.allow()` to let the action proceed, or `EventResult.block("reason")` to stop it.

## Routes

Add HTTP endpoints with `route`:

```java
route("GET", "/api/plugins/my-plugin/status", request -> {
    return Response.json(Map.of("status", "ok"));
});

route("POST", "/api/plugins/my-plugin/action", request -> {
    Map<String, Object> body = request.json();
    String name = (String) body.get("name");
    return Response.json(Map.of("received", name));
});
```

The `Request` class:
- `getMethod()` - HTTP method
- `getPath()` - Request path
- `getHeaders()` - Request headers
- `getQuery()` - Query parameters
- `header(name)` - Get a header value
- `query(name)` - Get a query parameter
- `queryInt(name, default)` - Get query param as int
- `queryBool(name, default)` - Get query param as boolean
- `json()` - Parse body as Map
- `json(Class)` - Parse body as typed object
- `bodyString()` - Raw body as string
- `getUserId()` - Authenticated user's ID

Response helpers:
- `Response.json(data)` - JSON response with `{"success": true, "data": ...}`
- `Response.error(status, message)` - Error response
- `Response.text(string)` - Plain text response
- `Response.html(string)` - HTML response
- `Response.ok(bytes)` - Raw byte response

## Mixins

Intercept panel operations with `mixin`:

```java
mixin(MixinTargets.SERVER_CREATE, ctx -> {
    String name = ctx.getString("name");
    ctx.set("name", "[Server] " + name);
    return ctx.next();
});
```

Use the priority parameter to control execution order (lower runs first):

```java
mixin(MixinTargets.SERVER_CREATE, -10, ctx -> {
    return ctx.next();
});
```

MixinContext methods:
- `get(key)` - Get input value
- `getString(key)`, `getInt(key)`, `getBool(key)` - Typed getters
- `getInput()` - Full input map
- `set(key, value)` - Modify input for next handler
- `getChainData()` - Data shared between mixin handlers
- `notify(title, message, type)` - Send notification to user
- `notifySuccess(title, message)` - Success notification
- `notifyError(title, message)` - Error notification
- `notifyInfo(title, message)` - Info notification
- `next()` - Continue to next handler
- `returnValue(data)` - Return early with custom response
- `error(message)` - Return an error

## Annotation-Based Mixins

For cleaner code, use the `@Mixin` annotation:

```java
@Mixin(value = MixinTargets.SERVER_CREATE, priority = 0)
public class ServerCreateMixin extends MixinClass {
    @Override
    public MixinResult handle(MixinContext ctx) {
        String name = ctx.getString("name");
        ctx.set("name", "[Server] " + name);
        return ctx.next();
    }
}
```

Register mixin classes in your plugin:

```java
public MyPlugin() {
    super("my-plugin", "1.0.0");
    registerMixin(ServerCreateMixin.class);
    registerMixins(ServerCreateMixin.class, UserCreateMixin.class);
}
```

## Schedules

Run tasks on a cron schedule:

```java
schedule("cleanup", "0 0 * * *", () -> {
    api().log("info", "Running daily cleanup");
});

schedule("heartbeat", "*/5 * * * *", () -> {
    api().log("info", "Heartbeat every 5 minutes");
});
```

## Addon Types

Define custom addon installation handlers:

```java
addonType("my-addon-type", ctx -> {
    return AddonTypeResult.success("Installed successfully",
        AddonTypeResult.Action.downloadFile(ctx.getDownloadUrl(), ctx.getInstallPath()),
        AddonTypeResult.Action.extractArchive(ctx.getInstallPath())
    );
});
```

AddonTypeContext methods:
- `getTypeId()` - The addon type ID
- `getServerId()` - Target server ID
- `getNodeId()` - Target node ID
- `getDownloadUrl()` - URL to download from
- `getFileName()` - Original file name
- `getInstallPath()` - Where to install
- `getSourceInfo()` - Additional source metadata
- `getServerVariables()` - Server environment variables

Available actions:
- `Action.downloadFile(url, path)` or `Action.downloadFile(url, path, headers)`
- `Action.extractArchive(path)`
- `Action.deleteFile(path)`
- `Action.createFolder(path)`
- `Action.writeFile(path, content)`
- `Action.proxyToNode(endpoint, payload)`

## Panel API

Access the panel API through `api()`:

```java
PanelAPI api = api();

PanelAPI.Server server = api.getServer("server-id");
List<PanelAPI.Server> servers = api.listServers();
api.startServer("server-id");
api.sendCommand("server-id", "say Hello");

PanelAPI.User user = api.getUser("user-id");
List<PanelAPI.User> users = api.listUsers();

List<PanelAPI.File> files = api.listFiles("server-id", "/");
byte[] content = api.readFile("server-id", "/server.properties");
api.writeFile("server-id", "/motd.txt", "Welcome!".getBytes());
```

See [Panel API](panel-api.md) for the full reference.

## Async API

For non-blocking operations, use `async()`:

```java
PanelAPIAsync async = async();

CompletableFuture<PanelAPI.Server> future = async.getServer("server-id");
PanelAPI.Server server = future.get();

async.getServer("server-id")
    .thenAccept(s -> api().log("info", "Got server: " + s.name))
    .exceptionally(e -> {
        api().log("error", "Error: " + e.getMessage());
        return null;
    });

CompletableFuture<List<PanelAPI.Server>> all = Futures.all(
    async.getServer("server-1"),
    async.getServer("server-2")
);
```

## Configuration

Use `HotConfig` for auto-reloading config files:

```java
public class Config {
    public boolean enabled = true;
    public String message = "Hello";
}

HotConfig<Config> config = new HotConfig<>(
    dataPath("config.yaml"),
    new Config(),
    data -> {
        Config c = new Config();
        c.enabled = (Boolean) data.getOrDefault("enabled", true);
        c.message = (String) data.getOrDefault("message", "Hello");
        return c;
    },
    c -> Map.of("enabled", c.enabled, "message", c.message)
);

config.dynamicConfig();

config.onChange(c -> {
    api().log("info", "Config reloaded!");
});

Config current = config.get();
```

## Saving and Loading Config

For simpler config needs, use the built-in methods:

```java
public class Config {
    public boolean enabled = true;
    public String message = "Hello";
}

Config config = loadConfigOrDefault(new Config(), "config.json");

config.message = "Updated";
saveConfig(config, "config.json");

Config loaded = loadConfig(Config.class, "config.json");
```

## Key-Value Storage

Store simple data in the panel's KV store:

```java
api().setKV("my-plugin:counter", "42");
String value = api().getKV("my-plugin:counter");
api().deleteKV("my-plugin:counter");
```

## HTTP Client

Make external HTTP requests through the panel:

```java
PanelAPI.HTTPResponse resp = api().httpGet(
    "https://api.example.com/data",
    Map.of("Authorization", "Bearer token")
);

if (resp.error.isEmpty() && resp.status == 200) {
    String body = resp.bodyAsString();
}
```

## Inter-Plugin Communication

Call methods on other plugins:

```java
byte[] response = api().callPlugin("other-plugin", "getData", "{\"id\": \"123\"}".getBytes());
```

## Console Streaming

Stream server console output in real-time:

```java
ConsoleStream stream = streamConsole(
    console("server-id")
        .includeHistory(true)
        .historyLines(100)
        .onLine(line -> {
            api().log("info", "Console: " + line);
        })
        .onError(error -> {
            api().log("error", "Stream error: " + error.getMessage());
        })
        .onComplete(() -> {
            api().log("info", "Stream closed");
        })
);

stream.stop();
```

## Custom Async Executor

By default, async operations use the common ForkJoinPool. You can provide your own executor:

```java
public MyPlugin() {
    super("my-plugin", "1.0.0");
    asyncExecutor(Executors.newFixedThreadPool(4));
}
```

## UI

Add custom pages, tabs, and sidebar items to the panel. The UI is built using React components bundled with your plugin.

### Setup

1. Create a `ui/` folder in your plugin project
2. Initialize a React project with the SDK:

```bash
cd ui
npm init -y
npm install @birdactyl/plugin-ui react react-dom
npm install -D typescript vite @vitejs/plugin-react @types/react
```

3. Create your components and build with Vite
4. Embed the bundle in your plugin

### Building UI

Use the `UIBuilder` to define your UI:

```java
public MyPlugin() {
    super("my-plugin", "1.0.0");
    
    setUI(ui()
        .embedBundle("ui/bundle.js")
        .page("/plugins/my-plugin", "DashboardPage")
            .title("My Plugin")
            .icon("Puzzle")
            .done()
        .tab("my-tab", "ServerTab", UIBuilder.TAB_TARGET_SERVER, "My Tab")
            .icon("Settings")
            .order(100)
            .done()
        .sidebarItem("my-plugin", "My Plugin", "/plugins/my-plugin", UIBuilder.SIDEBAR_SECTION_NAV)
            .icon("Puzzle")
            .order(50)
            .done()
        .build());
}
```

### Pages

Add standalone pages accessible via URL:

```java
ui()
    .page("/plugins/my-plugin/settings", "SettingsPage")
        .title("Settings")
        .icon("Settings")
        .adminOnly()
        .done()
    .build();
```

Page options:
- `title(string)` - Page title shown in browser tab
- `icon(string)` - Lucide icon name
- `adminOnly()` - Restrict to admin users
- `guard(string)` - Custom guard expression

### Tabs

Add tabs to existing panel sections:

```java
ui()
    .tab("stats", "StatsTab", UIBuilder.TAB_TARGET_SERVER, "Statistics")
        .icon("BarChart")
        .order(50)
        .done()
    .build();
```

Tab targets:
- `UIBuilder.TAB_TARGET_SERVER` - Server console page
- `UIBuilder.TAB_TARGET_USER_SETTINGS` - User settings page

Tab options:
- `icon(string)` - Lucide icon name
- `order(int)` - Sort order (lower = earlier)

### Sidebar Items

Add items to the panel sidebar:

```java
ui()
    .sidebarItem("my-plugin", "My Plugin", "/plugins/my-plugin", UIBuilder.SIDEBAR_SECTION_NAV)
        .icon("Puzzle")
        .order(50)
        .adminOnly()
        .child("Settings", "/plugins/my-plugin/settings")
        .child("Logs", "/plugins/my-plugin/logs")
        .done()
    .build();
```

Sidebar sections:
- `UIBuilder.SIDEBAR_SECTION_NAV` - Main navigation
- `UIBuilder.SIDEBAR_SECTION_PLATFORM` - Platform section
- `UIBuilder.SIDEBAR_SECTION_ADMIN` - Admin section

Sidebar options:
- `icon(string)` - Lucide icon name
- `order(int)` - Sort order
- `adminOnly()` - Restrict to admin users
- `guard(string)` - Custom guard expression
- `child(label, href)` - Add dropdown child item

### React Components

Create components using the SDK:

```tsx
import { useContext, useEvent, useAPI } from '@birdactyl/plugin-ui';

export function DashboardPage() {
    const ctx = useContext();
    const api = useAPI();
    
    useEvent('server:status', (data) => {
        console.log('Server status changed:', data);
    });
    
    return (
        <div>
            <h1>Dashboard</h1>
            <p>User: {ctx.user?.username}</p>
        </div>
    );
}
```

### Embedding the Bundle

Place your built bundle in `src/main/resources/ui/bundle.js` and embed it:

```java
setUI(ui()
    .embedBundle("ui/bundle.js")
    .page("/plugins/my-plugin", "DashboardPage")
        .done()
    .build());
```

Or load from bytes:

```java
byte[] bundleData = Files.readAllBytes(Path.of("ui/bundle.js"));
setUI(ui()
    .bundleBytes(bundleData)
    .page("/plugins/my-plugin", "DashboardPage")
        .done()
    .build());
```

See [Plugin UI](ui.md) for the full UI documentation.
