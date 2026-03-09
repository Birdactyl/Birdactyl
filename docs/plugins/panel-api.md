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

### Create Server

**Go:**
```go
server, err := api.CreateServer("Server Name", "user-id", "node-id", "package-id", 2048, 100, 10240)
```

**Java:**
```java
PanelAPI.Server server = api.createServer("Server Name", "user-id", "node-id", "package-id", 2048, 100, 10240);
```

### Update Server

**Go:**
```go
server, err := api.UpdateServer("server-id", "New Name", 4096, 200, 20480)
```

**Java:**
```java
PanelAPI.Server server = api.updateServer("server-id", "New Name", 4096, 200, 20480);
```

### Delete Server

**Go:**
```go
err := api.DeleteServer("server-id")
```

**Java:**
```java
api.deleteServer("server-id");
```

### Reinstall Server

**Go:**
```go
err := api.ReinstallServer("server-id")
```

**Java:**
```java
api.reinstallServer("server-id");
```

### Transfer Server

**Go:**
```go
err := api.TransferServer("server-id", "target-node-id")
```

**Java:**
```java
api.transferServer("server-id", "target-node-id");
```

### Get Server Stats

**Go:**
```go
stats, err := api.GetServerStats("server-id")
// Fields: MemoryBytes, MemoryLimit, CpuPercent, DiskBytes, NetworkRx, NetworkTx, State
```

**Java:**
```java
PanelAPI.ServerStats stats = api.getServerStats("server-id");
// Fields: memoryBytes, memoryLimit, cpuPercent, diskBytes, networkRx, networkTx, state
```


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


## Allocations

### Add Allocation

**Go:**
```go
err := api.AddAllocation("server-id", 25565)
```

**Java:**
```java
api.addAllocation("server-id", 25565);
```

### Delete Allocation

**Go:**
```go
err := api.DeleteAllocation("server-id", 25565)
```

**Java:**
```java
api.deleteAllocation("server-id", 25565);
```

### Set Primary Allocation

**Go:**
```go
err := api.SetPrimaryAllocation("server-id", 25565)
```

**Java:**
```java
api.setPrimaryAllocation("server-id", 25565);
```


## Variables

### Update Variables

**Go:**
```go
err := api.UpdateServerVariables("server-id", map[string]string{
    "MAX_PLAYERS": "20",
})
```

**Java:**
```java
api.updateServerVariables("server-id", Map.of("MAX_PLAYERS", "20"));
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

### Get Full Log

**Go:**
```go
content, err := api.GetFullLog("server-id")
```

**Java:**
```java
byte[] content = api.getFullLog("server-id");
```

### Search Logs

**Go:**
```go
matches, err := api.SearchLogs("server-id", "error", false, 10)
// Match fields: Line, LineNumber, Timestamp
```

**Java:**
```java
List<PanelAPI.LogMatch> matches = api.searchLogs("server-id", "error", false, 10);
// Match fields: line, lineNumber, timestamp
```

### List Log Files

**Go:**
```go
files, err := api.ListLogFiles("server-id")
// File fields: Name, Size, Modified
```

**Java:**
```java
List<PanelAPI.LogFile> files = api.listLogFiles("server-id");
// File fields: name, size, modified
```

### Read Log File

**Go:**
```go
content, err := api.ReadLogFile("server-id", "latest.log")
```

**Java:**
```java
byte[] content = api.readLogFile("server-id", "latest.log");
```

### Stream Console

**Go:**
```go
stream, err := api.StreamConsole("server-id", true, 100)
defer stream.Close()

for {
    line, err := stream.Recv()
    if err != nil {
        break
    }
    plugin.Log(line)
}
```

**Java:**
```java
ConsoleStream stream = streamConsole(
    console("server-id")
        .includeHistory(true)
        .historyLines(100)
        .onLine(line -> {
            api().log("info", "Console: " + line);
        })
);
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

### Create User

**Go:**
```go
user, err := api.CreateUser("[email]", "username", "password")
```

**Java:**
```java
PanelAPI.User user = api.createUser("[email]", "username", "password");
```

### Update User

**Go:**
```go
user, err := api.UpdateUser("user-id", "new_username", "new_email")
```

**Java:**
```java
PanelAPI.User user = api.updateUser("user-id", "new_username", "new_email");
```

### Delete User

**Go:**
```go
err := api.DeleteUser("user-id")
```

**Java:**
```java
api.deleteUser("user-id");
```

### Admin Status

**Go:**
```go
api.SetAdmin("user-id")
api.RevokeAdmin("user-id")
```

**Java:**
```java
api.setAdmin("user-id");
api.revokeAdmin("user-id");
```

### Set Resources

**Go:**
```go
api.SetUserResources("user-id", 4096, 200, 20480, 5)
```

**Java:**
```java
api.setUserResources("user-id", 4096, 200, 20480, 5);
```

### Password Reset

**Go:**
```go
api.ForcePasswordReset("user-id")
```

**Java:**
```java
api.forcePasswordReset("user-id");
```


## Subuser Management

### List Subusers

**Go:**
```go
subusers := api.ListSubusers("server-id")
// Fields: ID, UserID, Username, Email, Permissions
```

**Java:**
```java
List<PanelAPI.Subuser> subusers = api.listSubusers("server-id");
```

### Add Subuser

**Go:**
```go
subuser, err := api.AddSubuser("server-id", "[email]", []string{"control.start", "control.stop"})
```

**Java:**
```java
PanelAPI.Subuser subuser = api.addSubuser("server-id", "[email]", List.of("control.start", "control.stop"));
```

### Update Subuser

**Go:**
```go
api.UpdateSubuser("server-id", "subuser-id", []string{"control.start"})
```

**Java:**
```java
api.updateSubuser("server-id", "subuser-id", List.of("control.start"));
```

### Remove Subuser

**Go:**
```go
api.RemoveSubuser("server-id", "subuser-id")
```

**Java:**
```java
api.removeSubuser("server-id", "subuser-id");
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

### Create Folder

**Go:**
```go
api.CreateFolder("server-id", "/new-folder")
```

**Java:**
```java
api.createFolder("server-id", "/new-folder");
```

### Move/Rename File

**Go:**
```go
api.MoveFile("server-id", "/old-path.txt", "/new-path.txt")
```

**Java:**
```java
api.moveFile("server-id", "/old-path.txt", "/new-path.txt");
```

### Copy File

**Go:**
```go
api.CopyFile("server-id", "/original.txt", "/copy.txt")
```

**Java:**
```java
api.copyFile("server-id", "/original.txt", "/copy.txt");
```

### Compression

**Go:**
```go
// Compress
api.CompressFiles("server-id", []string{"/plugin.jar", "/config.yml"}, "backup.zip")

// Decompress
api.DecompressFile("server-id", "backup.zip")
```

**Java:**
```java
// Compress
api.compressFiles("server-id", List.of("/plugin.jar", "/config.yml"), "backup.zip");

// Decompress
api.decompressFile("server-id", "backup.zip");
```


## Backups Management

### List Backups

**Go:**
```go
backups := api.ListBackups("server-id")
// Fields: ID, Name, Size, CreatedAt
```

**Java:**
```java
List<PanelAPI.Backup> backups = api.listBackups("server-id");
```

### Create Backup

**Go:**
```go
api.CreateBackup("server-id", "Manual Backup")
```

**Java:**
```java
api.createBackup("server-id", "Manual Backup");
```

### Delete Backup

**Go:**
```go
api.DeleteBackup("server-id", "backup-id")
```

**Java:**
```java
api.deleteBackup("server-id", "backup-id");
```


## Mount Management

### List Mounts

**Go:**
```go
mounts := api.ListMounts()
```

**Java:**
```java
List<PanelAPI.Mount> mounts = api.listMounts();
```

Mount fields: `ID`, `Name`, `Description`, `Source`, `Target`, `ReadOnly`, `UserMountable`, `Navigable`

### Get Mount

**Go:**
```go
mount, err := api.GetMount("mount-id")
```

**Java:**
```java
PanelAPI.Mount mount = api.getMount("mount-id");
```

### Create Mount

**Go:**
```go
mount, err := api.CreateMount("Name", "Desc", "/source", "/target", false, true, true)
```

**Java:**
```java
PanelAPI.Mount mount = api.createMount("Name", "Desc", "/source", "/target", false, true, true);
```

### Update Mount

**Go:**
```go
mount, err := api.UpdateMount("mount-id", "New Name", "New Desc", "/new-source", "/new-target", nil, nil, nil)
```

**Java:**
```java
PanelAPI.Mount mount = api.updateMount("mount-id", "New Name", "New Desc", "/new-source", "/new-target", null, null, null);
```

### Delete Mount

**Go:**
```go
api.DeleteMount("mount-id")
```

**Java:**
```java
api.deleteMount("mount-id");
```

### Server Mount Associations

**Go:**
```go
// Add mount to server
api.AddMountToServer("mount-id", "server-id")

// Remove mount from server
api.RemoveMountFromServer("mount-id", "server-id")

// List mounts assigned to a server
serverMounts := api.GetServerMounts("server-id")
// ServerMountInfo fields: ID, Name, Description, Source, Target, ReadOnly, IsMounted, Navigable

// Mount execution
api.MountServerMount("mount-id", "server-id")
api.UnmountServerMount("mount-id", "server-id")
```

**Java:**
```java
// Add mount to server
api.addMountToServer("mount-id", "server-id");

// Remove mount from server
api.removeMountFromServer("mount-id", "server-id");

// List mounts assigned to a server
List<PanelAPI.ServerMountInfo> serverMounts = api.getServerMounts("server-id");

// Mount execution
api.mountServerMount("mount-id", "server-id");
api.unmountServerMount("mount-id", "server-id");
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

### Delete Database

**Go:**
```go
api.DeleteDatabase("database-id")
```

**Java:**
```java
api.deleteDatabase("database-id");
```

### Rotate Password

**Go:**
```go
db, err := api.RotateDatabasePassword("database-id")
```

**Java:**
```java
PanelAPI.Database db = api.rotateDatabasePassword("database-id");
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

### Get Node

**Go:**
```go
node, err := api.GetNode("node-id")
```

**Java:**
```java
PanelAPI.Node node = api.getNode("node-id");
```

### Create Node

**Go:**
```go
nodeWithToken, err := api.CreateNode("Node Name", "fqdn.example.com", 8443)
// Fields: Node, Token
```

**Java:**
```java
PanelAPI.NodeWithToken nodeWithToken = api.createNode("Node Name", "fqdn.example.com", 8443);
```

### Delete Node

**Go:**
```go
api.DeleteNode("node-id")
```

**Java:**
```java
api.deleteNode("node-id");
```

### Reset Token

**Go:**
```go
newToken := api.ResetNodeToken("node-id")
```

**Java:**
```java
String newToken = api.resetNodeToken("node-id");
```


## Package Management

### List Packages

**Go:**
```go
packages := api.ListPackages()
// Fields: ID, Name, Description, DockerImage, StartupCommand, StopCommand, DefaultMemory, DefaultCpu, DefaultDisk, IsPublic
```

**Java:**
```java
List<PanelAPI.Package> packages = api.listPackages();
```

### Get Package

**Go:**
```go
pkg, err := api.GetPackage("package-id")
```

**Java:**
```java
PanelAPI.Package pkg = api.getPackage("package-id");
```

### Create Package

**Go:**
```go
api.CreatePackage("Name", "Desc", "image", "startup", "stop", "config", 1024, 100, 5120, true)
```

**Java:**
```java
api.createPackage("Name", "Desc", "image", "startup", "stop", "config", 1024, 100, 5120, true);
```

### Update Package

**Go:**
```go
api.UpdatePackage("package-id", "New Name", "New Desc", 2048, 200, 10240)
```

**Java:**
```java
api.updatePackage("package-id", "New Name", "New Desc", 2048, 200, 10240);
```

### Delete Package

**Go:**
```go
api.DeletePackage("package-id")
```

**Java:**
```java
api.deletePackage("package-id");
```


## IP Ban Management

### List IP Bans

**Go:**
```go
bans := api.ListIPBans()
// Fields: ID, IP, Reason, CreatedAt
```

**Java:**
```java
List<PanelAPI.IPBan> bans = api.listIPBans();
```

### Create IP Ban

**Go:**
```go
api.CreateIPBan("1.2.3.4", "Abuse")
```

**Java:**
```java
api.createIPBan("1.2.3.4", "Abuse");
```

### Delete IP Ban

**Go:**
```go
api.DeleteIPBan("ban-id")
```

**Java:**
```java
api.deleteIPBan("ban-id");
```


## Settings

### Get Settings

**Go:**
```go
settings := api.GetSettings()
// Fields: RegistrationEnabled, ServerCreationEnabled
```

**Java:**
```java
PanelAPI.Settings settings = api.getSettings();
```

### Update Settings

**Go:**
```go
api.SetRegistrationEnabled(true)
api.SetServerCreationEnabled(false)
```

**Java:**
```java
api.setRegistrationEnabled(true);
api.setServerCreationEnabled(false);
```


## Activity Logs

### Get Activity Logs

**Go:**
```go
logs := api.GetActivityLogs(50)
// Fields: ID, UserID, Username, Action, Description, IP, IsAdmin, CreatedAt
```

**Java:**
```java
List<PanelAPI.ActivityLog> logs = api.getActivityLogs(50);
```


## Utility

### Query Database

Allows direct SQL queries to the panel's internal database (if permitted).

**Go:**
```go
rows, err := api.QueryDB("SELECT * FROM servers WHERE suspended = ?", "1")
for _, row := range rows {
    plugin.Log(row["name"].(string))
}
```

**Java:**
```java
List<Map<String, Object>> rows = api.queryDB("SELECT * FROM servers WHERE suspended = ?", "1");
```

### Broadcast Event

Send an event to all connected plugins and the UI.

**Go:**
```go
api.BroadcastEvent("my-plugin:custom-event", map[string]string{"foo": "bar"})
```

**Java:**
```java
api.broadcastEvent("my-plugin:custom-event", Map.of("foo", "bar"));
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

resp := api.HTTPPut("https://api.example.com/data", nil, []byte(`{"key": "update"}`))
api.HTTPDelete("https://api.example.com/data", nil)
```

**Java:**
```java
PanelAPI.HTTPResponse resp = api.httpGet("https://api.example.com/data", 
    Map.of("Authorization", "Bearer token"));

PanelAPI.HTTPResponse resp = api.httpPost("https://api.example.com/data",
    Map.of("Content-Type", "application/json"),
    "{\"key\": \"value\"}".getBytes());

PanelAPI.HTTPResponse resp = api.httpPut("https://api.example.com/data", null, "{}".getBytes());
api.httpDelete("https://api.example.com/data", null);
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
