# Addon Types

Addon types let your plugin define custom handlers for installing addons (mods, plugins, resource packs, etc.) on servers.

## How It Works

1. Your plugin registers an addon type with a unique ID
2. When a user installs an addon of that type, the panel calls your handler
3. Your handler receives information about the addon and target server
4. You return a list of actions (download, extract, write files, etc.)
5. The panel executes those actions on the target node

## Registering an Addon Type

**Go:**
```go
plugin.AddonType("my-addon-type", "My Addon Type", "Handles custom addon installation",
    func(req birdactyl.AddonTypeRequest) birdactyl.AddonTypeResponse {
        return birdactyl.AddonSuccess("Installed successfully",
            birdactyl.DownloadFile(req.DownloadURL, req.InstallPath, nil),
        )
    })
```

**Java:**
```java
addonType("my-addon-type", ctx -> {
    return AddonTypeResult.success("Installed successfully",
        AddonTypeResult.Action.downloadFile(ctx.getDownloadUrl(), ctx.getInstallPath())
    );
});
```

## Request Context

Your handler receives context about the addon being installed:

| Field | Description |
|-------|-------------|
| `TypeID` | Your addon type ID |
| `ServerID` | Target server's ID |
| `NodeID` | Target node's ID |
| `DownloadURL` | URL to download the addon from |
| `FileName` | Original file name |
| `InstallPath` | Suggested installation path |
| `SourceInfo` | Additional metadata from the addon source |
| `ServerVariables` | Server's environment variables |

## Available Actions

### Download File

**Go:**
```go
birdactyl.DownloadFile(req.DownloadURL, "/plugins/myplugin.jar", nil)
birdactyl.DownloadFile(url, path, map[string]string{"Authorization": "Bearer token"})
```

**Java:**
```java
Action.downloadFile(ctx.getDownloadUrl(), "/plugins/myplugin.jar")
Action.downloadFile(url, path, Map.of("Authorization", "Bearer token"))
```

### Extract Archive

**Go:**
```go
birdactyl.ExtractArchive("/temp/addon.zip")
```

**Java:**
```java
Action.extractArchive("/temp/addon.zip")
```

### Delete File

**Go:**
```go
birdactyl.DeleteFile("/temp/addon.zip")
```

**Java:**
```java
Action.deleteFile("/temp/addon.zip")
```

### Create Folder

**Go:**
```go
birdactyl.CreateFolder("/plugins/MyPlugin")
```

**Java:**
```java
Action.createFolder("/plugins/MyPlugin")
```

### Write File

**Go:**
```go
birdactyl.WriteFile("/plugins/MyPlugin/config.yml", []byte("enabled: true\n"))
```

**Java:**
```java
Action.writeFile("/plugins/MyPlugin/config.yml", "enabled: true\n".getBytes())
```

## Response Types

### Success

**Go:**
```go
return birdactyl.AddonSuccess("Installed successfully",
    birdactyl.DownloadFile(url, path, nil),
    birdactyl.ExtractArchive(path),
    birdactyl.DeleteFile(path),
)
```

**Java:**
```java
return AddonTypeResult.success("Installed successfully",
    Action.downloadFile(url, path),
    Action.extractArchive(path),
    Action.deleteFile(path)
);
```

### Error

**Go:**
```go
return birdactyl.AddonError("This addon is not compatible with the server version")
```

**Java:**
```java
return AddonTypeResult.error("This addon is not compatible with the server version");
```

## Example: Minecraft Plugin Handler

**Go:**
```go
plugin.AddonType("minecraft-plugin", "Minecraft Plugin", "Installs Minecraft plugins",
    func(req birdactyl.AddonTypeRequest) birdactyl.AddonTypeResponse {
        serverType := req.ServerVariables["SERVER_TYPE"]
        if serverType != "paper" && serverType != "spigot" {
            return birdactyl.AddonError("This addon requires a Paper or Spigot server")
        }
        installPath := "/plugins/" + req.FileName
        return birdactyl.AddonSuccess("Plugin installed to " + installPath,
            birdactyl.DownloadFile(req.DownloadURL, installPath, nil),
        )
    })
```

**Java:**
```java
addonType("minecraft-plugin", ctx -> {
    String serverType = ctx.getServerVariable("SERVER_TYPE");
    if (!"paper".equals(serverType) && !"spigot".equals(serverType)) {
        return AddonTypeResult.error("This addon requires a Paper or Spigot server");
    }
    String installPath = "/plugins/" + ctx.getFileName();
    return AddonTypeResult.success("Plugin installed to " + installPath,
        Action.downloadFile(ctx.getDownloadUrl(), installPath)
    );
});
```

## Best Practices

1. Validate server compatibility before returning actions
2. Use descriptive success/error messages
3. Clean up temporary files after extraction
4. Handle missing or malformed source info gracefully
5. Use appropriate paths based on server type
6. Consider file conflicts and overwrites
7. Log important operations for debugging
