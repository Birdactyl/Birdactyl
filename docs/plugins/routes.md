# Routes

Routes let your plugin add custom HTTP endpoints to the panel. Users and other systems can call these endpoints to interact with your plugin.

## Registering Routes

**Go:**
```go
plugin.Route("GET", "/api/plugins/my-plugin/status", func(r birdactyl.Request) birdactyl.Response {
    return birdactyl.JSON(map[string]string{"status": "ok"})
})
```

**Java:**
```java
route("GET", "/api/plugins/my-plugin/status", request -> {
    return Response.json(Map.of("status", "ok"));
});
```

## HTTP Methods

You can register routes for any HTTP method:

```go
plugin.Route("GET", "/api/plugins/my-plugin/items", getItems)
plugin.Route("POST", "/api/plugins/my-plugin/items", createItem)
plugin.Route("PUT", "/api/plugins/my-plugin/items", updateItem)
plugin.Route("DELETE", "/api/plugins/my-plugin/items", deleteItem)
```

Use `*` to match any method:

```go
plugin.Route("*", "/api/plugins/my-plugin/webhook", handleWebhook)
```

## Path Patterns

Routes support wildcard matching with `*`:

```go
plugin.Route("GET", "/api/plugins/my-plugin/files/*", func(r birdactyl.Request) birdactyl.Response {
    return birdactyl.JSON(map[string]string{"path": r.Path})
})
```


## Request Data

### Headers

**Go:**
```go
plugin.Route("GET", "/api/plugins/my-plugin/auth", func(r birdactyl.Request) birdactyl.Response {
    token := r.Headers["Authorization"]
    return birdactyl.JSON(map[string]string{"token": token})
})
```

**Java:**
```java
route("GET", "/api/plugins/my-plugin/auth", request -> {
    String token = request.header("Authorization");
    return Response.json(Map.of("token", token));
});
```

### Query Parameters

**Go:**
```go
plugin.Route("GET", "/api/plugins/my-plugin/search", func(r birdactyl.Request) birdactyl.Response {
    query := r.Query["q"]
    page := r.Query["page"]
    return birdactyl.JSON(map[string]string{"query": query, "page": page})
})
```

**Java:**
```java
route("GET", "/api/plugins/my-plugin/search", request -> {
    String query = request.query("q");
    int page = request.queryInt("page", 1);
    boolean active = request.queryBool("active", true);
    return Response.json(Map.of("query", query, "page", page));
});
```

### Request Body

**Go:**
```go
plugin.Route("POST", "/api/plugins/my-plugin/items", func(r birdactyl.Request) birdactyl.Response {
    name := r.Body["name"].(string)
    count := int(r.Body["count"].(float64))
    rawJSON := r.RawBody
    return birdactyl.JSON(map[string]interface{}{"name": name, "count": count})
})
```

**Java:**
```java
route("POST", "/api/plugins/my-plugin/items", request -> {
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
plugin.Route("GET", "/api/plugins/my-plugin/me", func(r birdactyl.Request) birdactyl.Response {
    if r.UserID == "" {
        return birdactyl.Error(401, "Not authenticated")
    }
    user, _ := plugin.API().GetUser(r.UserID)
    return birdactyl.JSON(user)
})
```

**Java:**
```java
route("GET", "/api/plugins/my-plugin/me", request -> {
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

## Route Naming Convention

Keep your routes under `/api/plugins/{plugin-id}/` to avoid conflicts:

```
/api/plugins/my-plugin/status
/api/plugins/my-plugin/items
/api/plugins/my-plugin/items/{id}
/api/plugins/my-plugin/config
```

## Best Practices

1. Use meaningful HTTP methods (GET for reads, POST for creates, etc.)
2. Return appropriate status codes (200, 201, 400, 404, 500)
3. Validate input before processing
4. Check authentication when needed
5. Keep routes fast - offload heavy work to background tasks
6. Use consistent response formats
