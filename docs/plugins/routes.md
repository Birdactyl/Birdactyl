# Routes

Routes let your plugin add custom HTTP endpoints to the panel. Users and other systems can call these endpoints to interact with your plugin.

## Registering Routes

**Go:**
```go
plugin.Route("GET", "/status", func(r birdactyl.Request) birdactyl.Response {
    return birdactyl.JSON(map[string]string{"status": "ok"})
})
```

**Java:**
```java
route("GET", "/status", request -> {
    return Response.json(Map.of("status", "ok"));
});
```

## HTTP Methods

You can register routes for any HTTP method:

```go
plugin.Route("GET", "/items", getItems)
plugin.Route("POST", "/items", createItem)
plugin.Route("PUT", "/items", updateItem)
plugin.Route("DELETE", "/items", deleteItem)
```

Use `*` to match any method:

```go
plugin.Route("*", "/webhook", handleWebhook)
```

## Path Patterns

Routes support wildcard matching with `*`:

```go
plugin.Route("GET", "/files/*", func(r birdactyl.Request) birdactyl.Response {
    return birdactyl.JSON(map[string]string{"path": r.Path})
})
```

## Rate Limiting

You can optionally apply rate limiting to your routes. If no rate limit is specified, the route is unlimited.

### Custom Rate Limits

Specify requests per minute and burst limit:

**Go:**
```go
plugin.Route("POST", "/webhook", handleWebhook).RateLimit(5, 10)

plugin.Route("GET", "/data", getData).RateLimit(30, 40)
```

**Java:**
```java
route("POST", "/webhook", this::handleWebhook).rateLimit(5, 10);

route("GET", "/data", this::getData).rateLimit(30, 40);
```

### Preset Rate Limits

Use panel presets for common scenarios:

| Preset | Requests/Min | Burst |
|--------|-------------|-------|
| `read` | 60 | 80 |
| `write` | 30 | 40 |
| `strict` | 10 | 15 |

**Go:**
```go
plugin.Route("GET", "/status", getStatus).RateLimitPreset(birdactyl.PresetRead)

plugin.Route("POST", "/action", doAction).RateLimitPreset(birdactyl.PresetWrite)

plugin.Route("POST", "/sensitive", sensitive).RateLimitPreset(birdactyl.PresetStrict)
```

**Java:**
```java
route("GET", "/status", this::getStatus).rateLimitPreset(PRESET_READ);

route("POST", "/action", this::doAction).rateLimitPreset(PRESET_WRITE);

route("POST", "/sensitive", this::sensitive).rateLimitPreset(PRESET_STRICT);
```

### Rate Limit Response

When a client exceeds the rate limit, they receive:

```json
{
  "success": false,
  "error": {
    "code": 429,
    "message": "Rate limit exceeded",
    "retry_after": 5
  }
}
```

Response headers are always included:
- `X-RateLimit-Limit` - Requests allowed per minute
- `X-RateLimit-Remaining` - Requests remaining
- `X-RateLimit-Reset` - Seconds until limit resets
- `Retry-After` - Seconds to wait (only on 429)

## Request Data

### Headers

**Go:**
```go
plugin.Route("GET", "/auth", func(r birdactyl.Request) birdactyl.Response {
    token := r.Headers["Authorization"]
    return birdactyl.JSON(map[string]string{"token": token})
})
```

**Java:**
```java
route("GET", "/auth", request -> {
    String token = request.header("Authorization");
    return Response.json(Map.of("token", token));
});
```

### Query Parameters

**Go:**
```go
plugin.Route("GET", "/search", func(r birdactyl.Request) birdactyl.Response {
    query := r.Query["q"]
    page := r.Query["page"]
    return birdactyl.JSON(map[string]string{"query": query, "page": page})
})
```

**Java:**
```java
route("GET", "/search", request -> {
    String query = request.query("q");
    int page = request.queryInt("page", 1);
    boolean active = request.queryBool("active", true);
    return Response.json(Map.of("query", query, "page", page));
});
```

### Request Body

**Go:**
```go
plugin.Route("POST", "/items", func(r birdactyl.Request) birdactyl.Response {
    name := r.Body["name"].(string)
    count := int(r.Body["count"].(float64))
    rawJSON := r.RawBody
    return birdactyl.JSON(map[string]interface{}{"name": name, "count": count})
})
```

**Java:**
```java
route("POST", "/items", request -> {
    Map<String, Object> body = request.json();
    String name = (String) body.get("name");
    MyRequest typed = request.json(MyRequest.class);
    String raw = request.bodyString();
    return Response.json(Map.of("name", name));
});
```

### Authenticated User

The panel passes the authenticated user's ID with each request:

**Go:**
```go
plugin.Route("GET", "/me", func(r birdactyl.Request) birdactyl.Response {
    if r.UserID == "" {
        return birdactyl.Error(401, "Not authenticated")
    }
    user, _ := plugin.API().GetUser(r.UserID)
    return birdactyl.JSON(user)
})
```

**Java:**
```java
route("GET", "/me", request -> {
    if (request.getUserId().isEmpty()) {
        return Response.error(401, "Not authenticated");
    }
    PanelAPI.User user = api().getUser(request.getUserId());
    return Response.json(user);
});
```

## Response Types

### JSON Response

Wraps your data in `{"success": true, "data": ...}`:

**Go:**
```go
return birdactyl.JSON(map[string]interface{}{
    "items": items,
    "total": len(items),
})
```

**Java:**
```java
return Response.json(Map.of(
    "items", items,
    "total", items.size()
));
```

### Error Response

Returns `{"success": false, "error": "message"}`:

**Go:**
```go
return birdactyl.Error(404, "Item not found")
return birdactyl.Error(400, "Invalid request")
return birdactyl.Error(500, "Internal error")
```

**Java:**
```java
return Response.error(404, "Item not found");
return Response.error(400, "Invalid request");
return Response.error(500, "Internal error");
```

### Text Response

**Go:**
```go
return birdactyl.Text("Hello, world!")
```

**Java:**
```java
return Response.text("Hello, world!");
```

### Custom Response

**Go:**
```go
resp := birdactyl.JSON(data).
    WithStatus(201).
    WithHeader("X-Custom", "value")
return resp
```

**Java:**
```java
return Response.ok(bytes)
    .status(201)
    .header("X-Custom", "value");
```

## Route URL Structure

Plugin routes are accessible at `/api/v1/plugins/{plugin-id}/{path}`:

```
/api/v1/plugins/my-plugin/status
/api/v1/plugins/my-plugin/items
/api/v1/plugins/my-plugin/items/123
/api/v1/plugins/my-plugin/config
```

## Best Practices

1. Use meaningful HTTP methods (GET for reads, POST for creates, etc.)
2. Return appropriate status codes (200, 201, 400, 404, 500)
3. Validate input before processing
4. Check authentication when needed
5. Keep routes fast - offload heavy work to background tasks
6. Use consistent response formats
7. Apply rate limits to protect against abuse, especially on write endpoints and webhooks
