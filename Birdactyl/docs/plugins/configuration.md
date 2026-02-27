# Configuration

Plugins often need configuration files that users can edit. Both SDKs provide utilities for loading, saving, and hot-reloading configuration files.

## Data Directory

First, enable the data directory for your plugin:

**Go:**
```go
plugin := birdactyl.New("my-plugin", "1.0.0").
    UseDataDir()
```

**Java:**
```java
public MyPlugin() {
    super("my-plugin", "1.0.0");
    useDataDir();
}
```

The data directory is created at `{plugins_dir}/{plugin_id}_data/`.

Access paths within it:

**Go:**
```go
configPath := plugin.DataPath("config.yaml")
dataPath := plugin.DataPath("data/users.json")
dir := plugin.DataDir()
```

**Java:**
```java
File configFile = dataPath("config.yaml");
File dataFile = dataPath("data/users.json");
File dir = dataDir();
```

## Simple Config (JSON)

For basic configuration needs, use the built-in JSON methods:

**Go:**
```go
type Config struct {
    Enabled bool   `json:"enabled"`
    Message string `json:"message"`
    MaxItems int   `json:"max_items"`
}

config := Config{Enabled: true, Message: "Hello", MaxItems: 100}
plugin.SaveConfig(config)

var loaded Config
plugin.LoadConfig(&loaded)
```

**Java:**
```java
public class Config {
    public boolean enabled = true;
    public String message = "Hello";
    public int maxItems = 100;
}

Config config = new Config();
saveConfig(config);

Config loaded = loadConfig(Config.class);
Config config = loadConfigOrDefault(new Config(), "config.json");
```

## Hot-Reloadable Config (YAML)

For configuration that should reload when the file changes, use `HotConfig`:

**Go:**
```go
type Config struct {
    Enabled  bool   `yaml:"enabled"`
    Message  string `yaml:"message"`
    MaxItems int    `yaml:"max_items"`
}

config := birdactyl.NewHotConfig(plugin.DataPath("config.yaml"), Config{
    Enabled:  true,
    Message:  "Hello",
    MaxItems: 100,
})

config.DynamicConfig()

config.OnChange(func(c Config) {
    plugin.Log("Config reloaded!")
    plugin.Log("New message: " + c.Message)
})

current := config.Get()
plugin.Log(current.Message)

current.MaxItems = 200
config.Set(current)
```

**Java:**
```java
public class Config {
    public boolean enabled = true;
    public String message = "Hello";
    public int maxItems = 100;
}

HotConfig<Config> config = new HotConfig<>(
    dataPath("config.yaml"),
    new Config(),
    data -> {
        Config c = new Config();
        c.enabled = (Boolean) data.getOrDefault("enabled", true);
        c.message = (String) data.getOrDefault("message", "Hello");
        c.maxItems = ((Number) data.getOrDefault("max_items", 100)).intValue();
        return c;
    },
    c -> Map.of("enabled", c.enabled, "message", c.message, "max_items", c.maxItems)
);

config.dynamicConfig();

config.onChange(c -> {
    api().log("info", "Config reloaded!");
});

Config current = config.get();
```

## Stopping Hot Reload

When your plugin shuts down, stop the config watcher:

**Go:**
```go
config.StopWatching()
```

**Java:**
```java
config.stopWatching();
```

## Key-Value Storage

For simple key-value data that persists across restarts, use the panel's KV store:

**Go:**
```go
api := plugin.API()
api.SetKV("my-plugin:counter", "42")
value, found := api.GetKV("my-plugin:counter")
if found {
    plugin.Log("Counter: " + value)
}
api.DeleteKV("my-plugin:counter")
```

**Java:**
```java
api().setKV("my-plugin:counter", "42");
String value = api().getKV("my-plugin:counter");
if (value != null) {
    api().log("info", "Counter: " + value);
}
api().deleteKV("my-plugin:counter");
```

Use a prefix like `my-plugin:` to avoid key collisions with other plugins.

## Best Practices

1. Always provide sensible defaults
2. Use YAML for human-editable config, JSON for machine-generated data
3. Validate config values after loading
4. Log when config reloads for debugging
5. Use descriptive key names
6. Document config options for users
7. Prefix KV keys with your plugin ID
8. Handle missing or corrupted config files gracefully
