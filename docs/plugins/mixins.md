# Mixins

Mixins let your plugin intercept and modify panel operations before they execute. You can validate input, transform data, add side effects, or completely override the default behavior.

## How Mixins Work

1. Your plugin registers a mixin for a specific operation (like `server.create`)
2. When that operation runs, the panel calls your mixin handler
3. Your handler can inspect the input, modify it, or return early
4. If you call `Next()`, the operation continues (possibly with your modifications)
5. Multiple plugins can register mixins for the same operation - they run in priority order

## Registering Mixins

**Go:**
```go
plugin.Mixin(birdactyl.MixinServerCreate, func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
    return ctx.Next()
})
```

**Java:**
```java
mixin(MixinTargets.SERVER_CREATE, ctx -> {
    return ctx.next();
});
```

## Mixin Priority

Lower priority values run first. Use priority to control execution order when multiple plugins hook the same operation:

**Go:**
```go
plugin.MixinWithPriority(birdactyl.MixinServerCreate, -10, earlyHandler)
plugin.MixinWithPriority(birdactyl.MixinServerCreate, 0, normalHandler)
plugin.MixinWithPriority(birdactyl.MixinServerCreate, 10, lateHandler)
```

**Java:**
```java
mixin(MixinTargets.SERVER_CREATE, -10, earlyHandler);
mixin(MixinTargets.SERVER_CREATE, 0, normalHandler);
mixin(MixinTargets.SERVER_CREATE, 10, lateHandler);
```


## Reading Input

Access the operation's input data through the context:

**Go:**
```go
plugin.Mixin(birdactyl.MixinServerCreate, func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
    name := ctx.GetString("name")
    userID := ctx.GetString("user_id")
    memory := ctx.GetInt("memory")
    allInput := ctx.Input()
    return ctx.Next()
})
```

**Java:**
```java
mixin(MixinTargets.SERVER_CREATE, ctx -> {
    String name = ctx.getString("name");
    String userId = ctx.getString("user_id");
    int memory = ctx.getInt("memory");
    Map<String, Object> allInput = ctx.getInput();
    return ctx.next();
});
```

## Modifying Input

Change input values before the operation runs:

**Go:**
```go
plugin.Mixin(birdactyl.MixinServerCreate, func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
    name := ctx.GetString("name")
    ctx.Set("name", "[Managed] " + name)
    memory := ctx.GetInt("memory")
    if memory < 512 {
        ctx.Set("memory", 512)
    }
    return ctx.Next()
})
```

**Java:**
```java
mixin(MixinTargets.SERVER_CREATE, ctx -> {
    String name = ctx.getString("name");
    ctx.set("name", "[Managed] " + name);
    int memory = ctx.getInt("memory");
    if (memory < 512) {
        ctx.set("memory", 512);
    }
    return ctx.next();
});
```

## Blocking Operations

Return an error to stop the operation:

**Go:**
```go
plugin.Mixin(birdactyl.MixinServerCreate, func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
    userID := ctx.GetString("user_id")
    user, _ := plugin.API().GetUser(userID)
    if user.IsBanned {
        return ctx.Error("Banned users cannot create servers")
    }
    return ctx.Next()
})
```

**Java:**
```java
mixin(MixinTargets.SERVER_CREATE, ctx -> {
    String userId = ctx.getString("user_id");
    PanelAPI.User user = api().getUser(userId);
    if (user.isBanned) {
        return ctx.error("Banned users cannot create servers");
    }
    return ctx.next();
});
```

## User Notifications

Send notifications to the user performing the action:

**Go:**
```go
plugin.Mixin(birdactyl.MixinServerCreate, func(ctx *birdactyl.MixinContext) birdactyl.MixinResult {
    ctx.NotifyInfo("Server Creation", "Your server is being created...")
    memory := ctx.GetInt("memory")
    if memory > 8192 {
        ctx.NotifySuccess("High Memory", "You're creating a high-memory server!")
    }
    return ctx.Next()
})
```

**Java:**
```java
mixin(MixinTargets.SERVER_CREATE, ctx -> {
    ctx.notifyInfo("Server Creation", "Your server is being created...");
    int memory = ctx.getInt("memory");
    if (memory > 8192) {
        ctx.notifySuccess("High Memory", "You're creating a high-memory server!");
    }
    return ctx.next();
});
```

Notification types:
- `NotifyInfo` / `notifyInfo` - Blue info notification
- `NotifySuccess` / `notifySuccess` - Green success notification
- `NotifyError` / `notifyError` - Red error notification
- `Notify(title, message, type)` / `notify(title, message, type)` - Custom type

## Annotation-Based Mixins (Java)

For cleaner code, use the `@Mixin` annotation:

```java
@Mixin(value = MixinTargets.SERVER_CREATE, priority = 0)
public class ServerCreateMixin extends MixinClass {
    @Override
    public MixinResult handle(MixinContext ctx) {
        String name = ctx.getString("name");
        ctx.set("name", "[Managed] " + name);
        return ctx.next();
    }
}
```

Register in your plugin:

```java
public MyPlugin() {
    super("my-plugin", "1.0.0");
    registerMixin(ServerCreateMixin.class);
}
```

## Available Mixin Targets

### Server Operations

| Target | Input Fields |
|--------|--------------|
| `server.create` | name, user_id, node_id, package_id, memory, cpu, disk |
| `server.update` | server_id, name, memory, cpu, disk |
| `server.delete` | server_id |
| `server.start` | server_id |
| `server.stop` | server_id |
| `server.restart` | server_id |
| `server.kill` | server_id |
| `server.suspend` | server_id |
| `server.unsuspend` | server_id |

### User Operations

| Target | Input Fields |
|--------|--------------|
| `user.create` | email, username, password |
| `user.update` | user_id, email, username |
| `user.delete` | user_id |
| `user.authenticate` | email, password |
| `user.ban` | user_id |
| `user.unban` | user_id |

### File Operations

| Target | Input Fields |
|--------|--------------|
| `file.read` | server_id, path |
| `file.write` | server_id, path, content |
| `file.delete` | server_id, path |
| `file.upload` | server_id, path |
| `file.move` | server_id, from, to |
| `file.copy` | server_id, from, to |

## Best Practices

1. Use appropriate priority values - validation should run early, logging should run late
2. Keep mixin handlers fast - they block the operation
3. Always call `Next()` unless you intend to block or override
4. Use notifications sparingly - too many can annoy users
5. Document what your mixins do for other plugin developers
6. Handle errors gracefully - don't crash on unexpected input
