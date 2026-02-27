# Panel API

The Panel API gives your plugin full access to manage servers, users, files, databases, and more. All operations go through the gRPC connection to the panel.

## Accessing the API

**Go:**
```go
api := plugin.API()
server, err := api.GetServer("server-id")
```

**Java:**
```java
PanelAPI api = api();
PanelAPI.Server server = api.getServer("server-id");
```

## Server Management

### Get Server

**Go:**
```go
server, err := api.GetServer("server-id")
if err != nil {
    plugin.Log("Server not found")
    return
}
plugin.Log("Server: " + server.Name)
```

**Java:**
```java
PanelAPI.Server server = api.getServer("server-id");
```

Server fields: `ID`, `Name`, `OwnerID`, `NodeID`, `Status`, `Suspended`, `Memory`, `CPU`, `Disk`, `PackageID`, `PrimaryAllocation`

### List Servers

**Go:**
```go
servers := api.ListServers()
for _, s := range servers {
    plugin.Log(s.Name)
}
userServers := api.ListServersByUser("user-id")
```

**Java:**
```java
List<PanelAPI.Server> servers = api.listServers();
List<PanelAPI.Server> userServers = api.listServersByUser("user-id");
```

### Server Power Actions

**Go:**
```go
api.StartServer("server-id")
api.StopServer("server-id")
api.RestartServer("server-id")
api.KillServer("server-id")
```

**Java:**
```java
api.startServer("server-id");
api.stopServer("server-id");
api.restartServer("server-id");
api.killServer("server-id");
```

### Suspend/Unsuspend

**Go:**
```go
api.SuspendServer("server-id")
api.UnsuspendServer("server-id")
```

**Java:**
```java
api.suspendServer("server-id");
api.unsuspendServer("server-id");
```


## Console

### Get Console Log

**Go:**
```go
lines, err := api.GetConsoleLog("server-id", 100)
for _, line := range lines {
    plugin.Log(line)
}
```

**Java:**
```java
List<String> lines = api.getConsoleLog("server-id", 100);
```

### Send Command

**Go:**
```go
api.SendCommand("server-id", "say Hello from plugin!")
```

**Java:**
```java
api.sendCommand("server-id", "say Hello from plugin!");
```

## User Management

### Get User

**Go:**
```go
user, err := api.GetUser("user-id")
user, err := api.GetUserByEmail("[email]")
user, err := api.GetUserByUsername("username")
```

**Java:**
```java
PanelAPI.User user = api.getUser("user-id");
PanelAPI.User user = api.getUserByEmail("[email]");
PanelAPI.User user = api.getUserByUsername("username");
```

User fields: `ID`, `Username`, `Email`, `IsAdmin`, `IsBanned`, `ForcePasswordReset`, `RamLimit`, `CpuLimit`, `DiskLimit`, `ServerLimit`, `CreatedAt`

### List Users

**Go:**
```go
users := api.ListUsers()
```

**Java:**
```java
List<PanelAPI.User> users = api.listUsers();
```

### Ban/Unban User

**Go:**
```go
api.BanUser("user-id")
api.UnbanUser("user-id")
```

**Java:**
```java
api.banUser("user-id");
api.unbanUser("user-id");
```

## File Management

### List Files

**Go:**
```go
files := api.ListFiles("server-id", "/")
for _, f := range files {
    if f.IsDir {
        plugin.Log("[DIR] " + f.Name)
    } else {
        plugin.Log(fmt.Sprintf("%s (%d bytes)", f.Name, f.Size))
    }
}
```

**Java:**
```java
List<PanelAPI.File> files = api.listFiles("server-id", "/");
```

File fields: `Name`, `Size`, `IsDir`, `ModTime`, `Mime`

### Read File

**Go:**
```go
content, err := api.ReadFile("server-id", "/server.properties")
text := string(content)
```

**Java:**
```java
byte[] content = api.readFile("server-id", "/server.properties");
String text = new String(content);
```

### Write File

**Go:**
```go
api.WriteFile("server-id", "/motd.txt", []byte("Welcome to the server!"))
```

**Java:**
```java
api.writeFile("server-id", "/motd.txt", "Welcome to the server!".getBytes());
```

### Delete File

**Go:**
```go
api.DeleteFile("server-id", "/old-file.txt")
```

**Java:**
```java
api.deleteFile("server-id", "/old-file.txt");
```

## Database Management

### List Databases

**Go:**
```go
databases := api.ListDatabases("server-id")
```

**Java:**
```java
List<PanelAPI.Database> databases = api.listDatabases("server-id");
```

Database fields: `ID`, `Name`, `Username`, `Password`, `Host`, `Port`

### Create Database

**Go:**
```go
db, err := api.CreateDatabase("server-id", "my_database")
plugin.Log("Password: " + db.Password)
```

**Java:**
```java
PanelAPI.Database db = api.createDatabase("server-id", "my_database");
```

## Node Management

### List Nodes

**Go:**
```go
nodes := api.ListNodes()
```

**Java:**
```java
List<PanelAPI.Node> nodes = api.listNodes();
```

Node fields: `ID`, `Name`, `FQDN`, `Port`, `IsOnline`, `LastHeartbeat`

## Key-Value Storage

**Go:**
```go
api.SetKV("my-plugin:config", `{"enabled": true}`)
value, found := api.GetKV("my-plugin:config")
api.DeleteKV("my-plugin:config")
```

**Java:**
```java
api.setKV("my-plugin:config", "{\"enabled\": true}");
String value = api.getKV("my-plugin:config");
api.deleteKV("my-plugin:config");
```

## HTTP Client

Make external HTTP requests through the panel:

**Go:**
```go
resp := api.HTTPGet("https://api.example.com/data", map[string]string{
    "Authorization": "Bearer token",
})

resp := api.HTTPPost("https://api.example.com/data", 
    map[string]string{"Content-Type": "application/json"},
    []byte(`{"key": "value"}`),
)
```

**Java:**
```java
PanelAPI.HTTPResponse resp = api.httpGet("https://api.example.com/data", 
    Map.of("Authorization", "Bearer token"));

PanelAPI.HTTPResponse resp = api.httpPost("https://api.example.com/data",
    Map.of("Content-Type", "application/json"),
    "{\"key\": \"value\"}".getBytes());
```

## Inter-Plugin Communication

Call methods on other plugins:

**Go:**
```go
response, err := api.CallPlugin("other-plugin", "getData", []byte(`{"id": "123"}`))
```

**Java:**
```java
byte[] response = api.callPlugin("other-plugin", "getData", "{\"id\": \"123\"}".getBytes());
```
