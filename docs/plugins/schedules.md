# Schedules

Schedules let your plugin run tasks automatically on a cron-based schedule. Use them for cleanup jobs, periodic checks, data synchronization, or any recurring work.

## Registering Schedules

**Go:**
```go
plugin.Schedule("cleanup", "0 0 * * *", func() {
    plugin.Log("Running daily cleanup")
})
```

**Java:**
```java
schedule("cleanup", "0 0 * * *", () -> {
    api().log("info", "Running daily cleanup");
});
```

The first argument is a unique ID for the schedule, the second is a cron expression, and the third is the function to run.

## Cron Expression Format

Cron expressions use 5 fields:

```
minute (0-59)
hour (0-23)
day of month (1-31)
month (1-12)
day of week (0-6, Sunday=0)
```

## Common Patterns

| Expression | Description |
|------------|-------------|
| `* * * * *` | Every minute |
| `*/5 * * * *` | Every 5 minutes |
| `*/15 * * * *` | Every 15 minutes |
| `0 * * * *` | Every hour |
| `0 */2 * * *` | Every 2 hours |
| `0 0 * * *` | Daily at midnight |
| `0 6 * * *` | Daily at 6 AM |
| `0 0 * * 0` | Weekly on Sunday at midnight |
| `0 0 1 * *` | Monthly on the 1st at midnight |
| `0 0 1 1 *` | Yearly on January 1st at midnight |

## Examples

### Health Check

**Go:**
```go
plugin.Schedule("health-check", "*/5 * * * *", func() {
    nodes := plugin.API().ListNodes()
    for _, node := range nodes {
        if !node.IsOnline {
            plugin.Log("Node offline: " + node.Name)
            notifyAdmins(node)
        }
    }
})
```

**Java:**
```java
schedule("health-check", "*/5 * * * *", () -> {
    List<PanelAPI.Node> nodes = api().listNodes();
    for (PanelAPI.Node node : nodes) {
        if (!node.isOnline) {
            api().log("warn", "Node offline: " + node.name);
            notifyAdmins(node);
        }
    }
});
```

### Suspend Expired Servers

**Go:**
```go
plugin.Schedule("check-expiry", "0 0 * * *", func() {
    servers := plugin.API().ListServers()
    for _, server := range servers {
        expiry, found := plugin.API().GetKV("expiry:" + server.ID)
        if found && isExpired(expiry) {
            plugin.API().SuspendServer(server.ID)
            plugin.Log("Suspended expired server: " + server.Name)
        }
    }
})
```

**Java:**
```java
schedule("check-expiry", "0 0 * * *", () -> {
    List<PanelAPI.Server> servers = api().listServers();
    for (PanelAPI.Server server : servers) {
        String expiry = api().getKV("expiry:" + server.id);
        if (expiry != null && isExpired(expiry)) {
            api().suspendServer(server.id);
            api().log("info", "Suspended expired server: " + server.name);
        }
    }
});
```

## Multiple Schedules

Register as many schedules as you need:

```go
plugin.Schedule("task-1", "*/5 * * * *", task1)
plugin.Schedule("task-2", "0 * * * *", task2)
plugin.Schedule("task-3", "0 0 * * *", task3)
```

Each schedule needs a unique ID within your plugin.

## Error Handling

Wrap your schedule handlers in error handling to prevent crashes:

**Go:**
```go
plugin.Schedule("risky-task", "0 * * * *", func() {
    defer func() {
        if r := recover(); r != nil {
            plugin.Log("Schedule panicked: " + fmt.Sprint(r))
        }
    }()
    doRiskyWork()
})
```

**Java:**
```java
schedule("risky-task", "0 * * * *", () -> {
    try {
        doRiskyWork();
    } catch (Exception e) {
        api().log("error", "Schedule failed: " + e.getMessage());
    }
});
```

## Best Practices

1. Use descriptive schedule IDs that explain what the task does
2. Log when schedules run and complete for debugging
3. Handle errors gracefully - don't let one failure break future runs
4. Avoid overlapping schedules that might conflict
5. Keep scheduled tasks reasonably fast - long tasks can pile up
6. Use appropriate intervals - don't poll every minute if hourly is fine
7. Consider time zones when scheduling time-sensitive tasks
