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
| `server.creating` | name, user_id, node_id | Before server is created |
| `server.updating` | server_id, name, memory | Before server is updated |
| `server.suspending` | server_id | Before server is suspended |
| `server.unsuspending`| server_id | Before server is unsuspended |
| `server.statusUpdate`| server_id, status | After server status change |


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
| `user.registering` | email, username | Before user registers |
| `user.creating` | email, username | Before user is created by admin |
| `user.updating` | user_id, email, username | Before user is updated |
| `user.banning` | user_id | Before user is banned |
| `user.unbanning` | user_id | Before user is unbanned |
| `user.statusUpdate` | user_id, active | After user status change |


### File Events

| Event | Data | Description |
|-------|------|-------------|
| `file.upload` | server_id, path, size | File uploaded |
| `file.delete` | server_id, path | File deleted |
| `file.write` | server_id, path | File written |
| `file.create` | server_id, path | Folder created |
| `file.move` | server_id, from, to | File/Folder moved |
| `file.copy` | server_id, from, to | File/Folder copied |
| `file.compress` | server_id, destination | Files compressed |
| `file.decompress` | server_id, path | Archive decompressed |


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
 
 ### Node Events
 
 | Event | Data | Description |
 |-------|------|-------------|
 | `node.create` | node_id, name, fqdn | Node created |
 | `node.creating` | name, fqdn | Before node is created |
 | `node.update` | node_id, name, fqdn | Node updated |
 | `node.updating` | node_id, name | Before node is updated |
 | `node.delete` | node_id | Node deleted |
 | `node.deleting` | node_id | Before node is deleted |
 | `node.heartbeat` | node_id, is_online | Node heartbeat received |
 | `node.online` | node_id | Node came online |
 | `node.offline` | node_id | Node went offline |
 
 ### Package Events
 
 | Event | Data | Description |
 |-------|------|-------------|
 | `package.create` | package_id, name | Package created |
 | `package.creating` | name, docker_image | Before package is created |
 | `package.update` | package_id, name | Package updated |
 | `package.updating` | package_id, name | Before package is updated |
 | `package.delete` | package_id | Package deleted |
 | `package.deleting` | package_id | Before package is deleted |
 
 ### System Events
 
 | Event | Data | Description |
 |-------|------|-------------|
 | `system.startup` | version | Panel starting up |
 | `system.shutdown` | | Panel shutting down |
 | `system.maintenance` | enabled | Maintenance mode toggled |
 
 ### Plugin Events
 
 | Event | Data | Description |
 |-------|------|-------------|
 | `plugin.loading` | plugin_id | Before a plugin is loaded |
 | `plugin.loaded` | plugin_id | After a plugin is loaded |
 | `plugin.unloading` | plugin_id | Before a plugin is unloaded |
 | `plugin.unloaded` | plugin_id | After a plugin is unloaded |


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
