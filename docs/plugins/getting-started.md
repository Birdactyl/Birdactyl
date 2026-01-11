# Getting Started

This guide walks you through creating your first Birdactyl plugin.

## Prerequisites

For Go plugins:
- Go 1.24 or later

For Java plugins:
- Java 17 or later
- Maven

## Creating a Go Plugin

Create a new directory for your plugin and initialize a Go module:

```
mkdir my-plugin
cd my-plugin
go mod init my-plugin
go get github.com/Birdactyl/Birdactyl-Go-SDK
```

Create `main.go`:

```go
package main

import (
    "log"
    birdactyl "github.com/Birdactyl/Birdactyl-Go-SDK"
)

func main() {
    plugin := birdactyl.New("my-plugin", "1.0.0").
        SetName("My First Plugin")

    plugin.OnEvent("server.start", func(e birdactyl.Event) birdactyl.EventResult {
        plugin.Log("A server just started!")
        return birdactyl.Allow()
    })

    plugin.Route("GET", "/api/plugins/my-plugin/hello", func(r birdactyl.Request) birdactyl.Response {
        return birdactyl.JSON(map[string]string{"message": "Hello from my plugin!"})
    })

    if err := plugin.Start("localhost:50050"); err != nil {
        log.Fatal(err)
    }
}
```

Build and run:

```
go build -o my-plugin
./my-plugin
```

## Creating a Java Plugin

Create a Maven project with this `pom.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>my-plugin</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>17</maven.compiler.source>
        <maven.compiler.target>17</maven.compiler.target>
    </properties>

    <dependencies>
        <dependency>
            <groupId>io.birdactyl</groupId>
            <artifactId>birdactyl-sdk</artifactId>
            <version>0.2.0</version>
        </dependency>
    </dependencies>
</project>
```

Create `src/main/java/com/example/MyPlugin.java`:

```java
package com.example;

import io.birdactyl.sdk.*;

public class MyPlugin extends BirdactylPlugin {
    public MyPlugin() {
        super("my-plugin", "1.0.0");
        setName("My First Plugin");

        onEvent("server.start", event -> {
            api().log("info", "A server just started!");
            return EventResult.allow();
        });

        route("GET", "/api/plugins/my-plugin/hello", request -> {
            return Response.json(java.util.Map.of("message", "Hello from my plugin!"));
        });
    }

    public static void main(String[] args) throws Exception {
        new MyPlugin().start("localhost:50050");
    }
}
```

Build and run:

```
mvn package
java -jar target/my-plugin-1.0.0.jar
```

## Plugin Lifecycle

When your plugin starts:

1. It connects to the panel's gRPC server at the specified address
2. Sends a registration message with its ID, name, version, and capabilities
3. The panel acknowledges the registration
4. Your `onStart` callback runs (if you set one)
5. The plugin enters a loop, handling messages from the panel

The panel can send:
- Event notifications
- HTTP requests for your routes
- Schedule triggers
- Mixin requests
- Shutdown signals

## Data Directory

Plugins can store persistent data in their data directory. Enable it with `UseDataDir()` (Go) or `useDataDir()` (Java):

```go
plugin := birdactyl.New("my-plugin", "1.0.0").
    UseDataDir()

configPath := plugin.DataPath("config.json")
```

```java
public MyPlugin() {
    super("my-plugin", "1.0.0");
    useDataDir();
}

File configFile = dataPath("config.json");
```

The data directory is created at `{plugins_dir}/{plugin_id}_data/`.

## Logging

Use the plugin's log method to send messages to the panel's log:

```go
plugin.Log("Something happened")
```

```java
api().log("info", "Something happened");
```

## Next Steps

- [Events](events.md) - React to server starts, user logins, and more
- [Routes](routes.md) - Add your own API endpoints
- [UI](ui.md) - Add custom pages, tabs, and sidebar items
- [Mixins](mixins.md) - Intercept and modify panel operations
- [Panel API](panel-api.md) - Manage servers, users, and files
