# Events

Events let your plugin react to things happening in the panel. When something occurs (like a server starting or a user logging in), the panel notifies all plugins that registered for that event.

## How Events Work

1. Your plugin registers interest in specific event types
2. When that event occurs, the panel sends your plugin the event data
3. Your handler runs and returns whether to allow or block the action
4. For sync events, the panel waits for your response before proceeding

## Registering Event Handlers

**Go:**
```go
plugin.OnEvent("server.start", func(e birdactyl.Event) birdactyl.EventResult {
    serverID := e.Data["server_id"]
    return birdactyl.Allow()
})
```

**Java:**
```java
onEvent("server.start", event -> {
    String serverId = event.get("server_id");
    return EventResult.allow();
});
```

## Event Data

Events carry data in a string map. The available keys depend on the event type.

**Go:**
```go
plugin.OnEvent("server.create", func(e birdactyl.Event) birdactyl.EventResult {
    serverID := e.Data["server_id"]
    userID := e.Data["user_id"]
    name := e.Data["name"]
    return birdactyl.Allow()
})
```

**Java:**
```java
onEvent("server.create", event -> {
    String serverId = event.get("server_id");
    String userId = event.get("user_id");
    String name = event.get("name");
    return EventResult.allow();
});
```

## Allowing and Blocking

Return `Allow()` to let the action proceed, or `Block("reason")` to stop it:

**Go:**
```go
plugin.OnEvent("user.create", func(e birdactyl.Event) birdactyl.EventResult {
    email := e.Data["email"]
    if strings.HasSuffix(email, "@blocked.com") {
        return birdactyl.Block("This email domain is not allowed")
    }
    return birdactyl.Allow()
})
```

**Java:**
```java
onEvent("user.create", event -> {
    String email = event.get("email");
    if (email.endsWith("@blocked.com")) {
        return EventResult.block("This email domain is not allowed");
    }
    return EventResult.allow();
});
```

## Sync vs Async Events

Some events are synchronous, meaning the panel waits for your response before continuing. Others are asynchronous and fire-and-forget.

Check if an event is sync:

**Go:**
```go
plugin.OnEvent("server.start", func(e birdactyl.Event) birdactyl.EventResult {
    if e.Sync {
        plugin.Log("Panel is waiting for our response")
    }
    return birdactyl.Allow()
})
```

**Java:**
```java
onEvent("server.start", event -> {
    if (event.isSync()) {
        api().log("info", "Panel is waiting for our response");
    }
    return EventResult.allow();
});
```

## Available Events

### Server Events

| Event | Data | Description |
|-------|------|-------------|
| `server.create` | server_id, user_id, name, node_id | Server created |
| `server.delete` | server_id, user_id | Server deleted |
| `server.start` | server_id | Server starting |
| `server.stop` | server_id | Server stopping |
| `server.restart` | server_id | Server restarting |
| `server.kill` | server_id | Server force killed |
| `server.suspend` | server_id | Server suspended |
| `server.unsuspend` | server_id | Server unsuspended |
| `server.reinstall` | server_id | Server reinstalling |
| `server.transfer` | server_id, target_node_id | Server transferring |

### User Events

| Event | Data | Description |
|-------|------|-------------|
| `user.create` | user_id, email, username | User registered |
| `user.delete` | user_id | User deleted |
| `user.login` | user_id, ip | User logged in |
| `user.logout` | user_id | User logged out |
| `user.ban` | user_id | User banned |
| `user.unban` | user_id | User unbanned |
| `user.password_reset` | user_id | Password reset requested |

### File Events

| Event | Data | Description |
|-------|------|-------------|
| `file.upload` | server_id, path, size | File uploaded |
| `file.delete` | server_id, path | File deleted |
| `file.write` | server_id, path | File written |

### Backup Events

| Event | Data | Description |
|-------|------|-------------|
| `backup.create` | server_id, backup_id | Backup created |
| `backup.delete` | server_id, backup_id | Backup deleted |
| `backup.restore` | server_id, backup_id | Backup restored |

### Database Events

| Event | Data | Description |
|-------|------|-------------|
| `database.create` | server_id, database_id | Database created |
| `database.delete` | database_id | Database deleted |

### Subuser Events

| Event | Data | Description |
|-------|------|-------------|
| `subuser.add` | server_id, user_id, email | Subuser added |
| `subuser.remove` | server_id, subuser_id | Subuser removed |
| `subuser.update` | server_id, subuser_id | Subuser permissions updated |

### Console Events

| Event | Data | Description |
|-------|------|-------------|
| `console.command` | server_id, user_id, command | Command sent to console |

## Multiple Handlers

You can register multiple handlers for the same event. They all run, but if any returns Block, the action is blocked:

**Go:**
```go
plugin.OnEvent("server.start", handler1)
plugin.OnEvent("server.start", handler2)
```

**Java:**
```java
onEvent("server.start", handler1);
onEvent("server.start", handler2);
```

## Best Practices

1. Keep event handlers fast - they can block panel operations
2. Use async operations for slow tasks (database queries, HTTP requests)
3. Only block events when you have a good reason
4. Log important events for debugging
5. Handle errors gracefully - don't let exceptions crash your plugin

## Frontend Events

Plugins with UI can subscribe to frontend events. See [UI](ui.md) for details.

```tsx
import { useEvent, events } from '@birdactyl/plugin-ui';

useEvent('server:start', (data) => {
  console.log('Server started:', data.serverId);
});

useEvent('file:saved', (data) => {
  console.log('File saved:', data.path);
});

events.emit('plugin:my-plugin:custom', { foo: 'bar' });
```

### Available Frontend Events

| Event | Data |
|-------|------|
| `server:status` | `{ serverId, status, previousStatus }` |
| `server:stats` | `{ serverId, memory, memoryLimit, cpu, disk }` |
| `server:log` | `{ serverId, line }` |
| `server:start` | `{ serverId }` |
| `server:stop` | `{ serverId }` |
| `server:restart` | `{ serverId }` |
| `server:kill` | `{ serverId }` |
| `file:created` | `{ serverId, path }` |
| `file:deleted` | `{ serverId, path }` |
| `file:moved` | `{ serverId, from, to }` |
| `file:uploaded` | `{ serverId, path }` |
| `file:saved` | `{ serverId, path }` |
| `navigation` | `{ path, previousPath }` |
| `user:login` | `{ userId, username }` |
| `user:logout` | `{}` |

### Inter-Plugin Communication

Plugins can communicate via custom events using the `plugin:` prefix:

```tsx
events.emit('plugin:my-plugin:data-updated', { id: '123' });

useEvent('plugin:other-plugin:notification', (data) => {
  console.log('Received:', data);
});
```
