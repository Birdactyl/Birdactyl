package plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	pb "birdactyl-panel-backend/internal/plugins/proto"
	"birdactyl-panel-backend/internal/services"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type PanelServer struct {
	pb.UnimplementedPanelServiceServer
	kv   map[string]string
	kvMu sync.RWMutex
}

func NewPanelServer() *PanelServer {
	return &PanelServer{kv: make(map[string]string)}
}

func (s *PanelServer) GetServer(ctx context.Context, req *pb.IDRequest) (*pb.Server, error) {
	var server models.Server
	if err := database.DB.First(&server, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	return serverToProto(&server), nil
}

func (s *PanelServer) ListServers(ctx context.Context, req *pb.ListServersRequest) (*pb.ListServersResponse, error) {
	var servers []models.Server
	var total int64
	q := database.DB.Model(&models.Server{})
	if req.UserId != "" {
		q = q.Where("user_id = ?", req.UserId)
	}
	if req.NodeId != "" {
		q = q.Where("node_id = ?", req.NodeId)
	}
	q.Count(&total)
	if req.Limit > 0 {
		q = q.Limit(int(req.Limit))
	}
	if req.Offset > 0 {
		q = q.Offset(int(req.Offset))
	}
	q.Find(&servers)
	result := make([]*pb.Server, len(servers))
	for i, srv := range servers {
		result[i] = serverToProto(&srv)
	}
	return &pb.ListServersResponse{Servers: result, Total: int32(total)}, nil
}

func (s *PanelServer) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.Server, error) {
	userID, _ := uuid.Parse(req.UserId)
	nodeID, _ := uuid.Parse(req.NodeId)
	packageID, _ := uuid.Parse(req.PackageId)
	server, err := services.CreateServer(userID, services.CreateServerRequest{
		Name: req.Name, NodeID: nodeID, PackageID: packageID,
		Memory: int(req.Memory), CPU: int(req.Cpu), Disk: int(req.Disk),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	go services.SendCreateServer(server)
	return serverToProto(server), nil
}

func (s *PanelServer) DeleteServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.Id)
	services.SendDeleteServer(serverID)
	database.DB.Delete(&models.Server{}, "id = ?", serverID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) UpdateServer(ctx context.Context, req *pb.UpdateServerRequest) (*pb.Server, error) {
	var server models.Server
	if err := database.DB.First(&server, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Memory > 0 {
		server.Memory = int(req.Memory)
	}
	if req.Cpu > 0 {
		server.CPU = int(req.Cpu)
	}
	if req.Disk > 0 {
		server.Disk = int(req.Disk)
	}
	if req.UserId != "" {
		uid, _ := uuid.Parse(req.UserId)
		server.UserID = uid
	}
	database.DB.Save(&server)
	return serverToProto(&server), nil
}

func (s *PanelServer) SuspendServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.Id)
	database.DB.Model(&models.Server{}).Where("id = ?", serverID).Update("is_suspended", true)
	services.SendStopServer(serverID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) UnsuspendServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Model(&models.Server{}).Where("id = ?", req.Id).Update("is_suspended", false)
	return &pb.Empty{}, nil
}

func (s *PanelServer) StartServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.Id)
	services.SendStartServer(serverID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) StopServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.Id)
	services.SendStopServer(serverID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) RestartServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.Id)
	services.SendRestartServer(serverID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) KillServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.Id)
	services.SendKillServer(serverID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ReinstallServer(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").Preload("Package").First(&server, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	services.SendReinstallServer(&server)
	return &pb.Empty{}, nil
}

func (s *PanelServer) TransferServer(ctx context.Context, req *pb.TransferServerRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	nodeID, _ := uuid.Parse(req.TargetNodeId)
	services.StartTransfer(serverID, nodeID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) GetConsoleLog(ctx context.Context, req *pb.ConsoleLogRequest) (*pb.ConsoleLogResponse, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	lines := services.GetConsoleLog(serverID, int(req.Lines))
	return &pb.ConsoleLogResponse{Lines: lines}, nil
}

func (s *PanelServer) SendCommand(ctx context.Context, req *pb.SendCommandRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	services.SendCommand(serverID, req.Command)
	return &pb.Empty{}, nil
}

func (s *PanelServer) StreamConsole(req *pb.StreamConsoleRequest, stream pb.PanelService_StreamConsoleServer) error {
	serverID, _ := uuid.Parse(req.ServerId)
	if req.IncludeHistory {
		lines := int(req.HistoryLines)
		if lines <= 0 {
			lines = 100
		}
		history := services.GetConsoleLog(serverID, lines)
		for _, line := range history {
			if err := stream.Send(&pb.ConsoleLine{Line: line, Timestamp: 0}); err != nil {
				return err
			}
		}
	}
	ch := services.SubscribeConsole(serverID)
	defer services.UnsubscribeConsole(serverID, ch)
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case line, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(&pb.ConsoleLine{Line: line, Timestamp: time.Now().UnixMilli()}); err != nil {
				return err
			}
		}
	}
}

func (s *PanelServer) GetFullLog(ctx context.Context, req *pb.IDRequest) (*pb.FullLogResponse, error) {
	serverID, _ := uuid.Parse(req.Id)
	content, size := services.GetFullLog(serverID)
	return &pb.FullLogResponse{Content: content, Size: size}, nil
}

func (s *PanelServer) SearchLogs(ctx context.Context, req *pb.SearchLogsRequest) (*pb.SearchLogsResponse, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	matches := services.SearchLogs(serverID, req.Pattern, req.Regex, int(req.Limit), req.Since)
	var pbMatches []*pb.LogMatch
	for _, m := range matches {
		pbMatches = append(pbMatches, &pb.LogMatch{Line: m.Line, LineNumber: int32(m.LineNumber), Timestamp: m.Timestamp})
	}
	return &pb.SearchLogsResponse{Matches: pbMatches}, nil
}

func (s *PanelServer) ListLogFiles(ctx context.Context, req *pb.IDRequest) (*pb.LogFilesResponse, error) {
	serverID, _ := uuid.Parse(req.Id)
	files := services.ListLogFiles(serverID)
	var pbFiles []*pb.LogFileInfo
	for _, f := range files {
		pbFiles = append(pbFiles, &pb.LogFileInfo{Name: f.Name, Size: f.Size, Modified: f.Modified})
	}
	return &pb.LogFilesResponse{Files: pbFiles}, nil
}

func (s *PanelServer) ReadLogFile(ctx context.Context, req *pb.ReadLogFileRequest) (*pb.FullLogResponse, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	content, size := services.ReadLogFile(serverID, req.Filename)
	return &pb.FullLogResponse{Content: content, Size: size}, nil
}

func (s *PanelServer) GetServerStats(ctx context.Context, req *pb.IDRequest) (*pb.ServerStats, error) {
	serverID, _ := uuid.Parse(req.Id)
	stats := services.GetServerStats(serverID)
	if stats == nil {
		return &pb.ServerStats{}, nil
	}
	return &pb.ServerStats{
		MemoryBytes: stats.MemoryBytes,
		MemoryLimit: stats.MemoryLimit,
		CpuPercent:  stats.CPUPercent,
		DiskBytes:   stats.DiskBytes,
		NetworkRx:   stats.NetworkRx,
		NetworkTx:   stats.NetworkTx,
		State:       stats.State,
	}, nil
}

func (s *PanelServer) AddAllocation(ctx context.Context, req *pb.AllocationRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *PanelServer) DeleteAllocation(ctx context.Context, req *pb.AllocationRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *PanelServer) SetPrimaryAllocation(ctx context.Context, req *pb.AllocationRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *PanelServer) UpdateServerVariables(ctx context.Context, req *pb.UpdateVariablesRequest) (*pb.Empty, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	var server models.Server
	database.DB.First(&server, "id = ?", serverID)
	vars := make(map[string]string)
	json.Unmarshal(server.Variables, &vars)
	for k, v := range req.Variables {
		vars[k] = v
	}
	b, _ := json.Marshal(vars)
	database.DB.Model(&server).Update("variables", b)
	return &pb.Empty{}, nil
}

func (s *PanelServer) GetUser(ctx context.Context, req *pb.IDRequest) (*pb.User, error) {
	var user models.User
	if err := database.DB.First(&user, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return userToProto(&user), nil
}

func (s *PanelServer) GetUserByEmail(ctx context.Context, req *pb.EmailRequest) (*pb.User, error) {
	var user models.User
	if err := database.DB.First(&user, "email = ?", req.Email).Error; err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return userToProto(&user), nil
}

func (s *PanelServer) GetUserByUsername(ctx context.Context, req *pb.UsernameRequest) (*pb.User, error) {
	var user models.User
	if err := database.DB.First(&user, "username = ?", req.Username).Error; err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return userToProto(&user), nil
}

func (s *PanelServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	var users []models.User
	var total int64
	q := database.DB.Model(&models.User{})
	if req.Search != "" {
		q = q.Where("username ILIKE ? OR email ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Filter == "admin" {
		q = q.Where("is_admin = ?", true)
	} else if req.Filter == "banned" {
		q = q.Where("is_banned = ?", true)
	}
	q.Count(&total)
	if req.Limit > 0 {
		q = q.Limit(int(req.Limit))
	} else {
		q = q.Limit(100)
	}
	if req.Offset > 0 {
		q = q.Offset(int(req.Offset))
	}
	q.Find(&users)
	result := make([]*pb.User, len(users))
	for i, u := range users {
		result[i] = userToProto(&u)
	}
	return &pb.ListUsersResponse{Users: result, Total: int32(total)}, nil
}

func (s *PanelServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	user, err := services.AdminCreateUser(req.Email, req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return userToProto(user), nil
}

func (s *PanelServer) DeleteUser(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Delete(&models.User{}, "id = ?", req.Id)
	return &pb.Empty{}, nil
}

func (s *PanelServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	var user models.User
	if err := database.DB.First(&user, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Password != "" {
		hash, _ := services.HashPassword(req.Password)
		user.PasswordHash = hash
	}
	database.DB.Save(&user)
	return userToProto(&user), nil
}

func (s *PanelServer) BanUser(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Model(&models.User{}).Where("id = ?", req.Id).Update("is_banned", true)
	database.DB.Where("user_id = ?", req.Id).Delete(&models.Session{})
	return &pb.Empty{}, nil
}

func (s *PanelServer) UnbanUser(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Model(&models.User{}).Where("id = ?", req.Id).Update("is_banned", false)
	return &pb.Empty{}, nil
}

func (s *PanelServer) SetAdmin(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Model(&models.User{}).Where("id = ?", req.Id).Update("is_admin", true)
	return &pb.Empty{}, nil
}

func (s *PanelServer) RevokeAdmin(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Model(&models.User{}).Where("id = ?", req.Id).Update("is_admin", false)
	return &pb.Empty{}, nil
}

func (s *PanelServer) SetUserResources(ctx context.Context, req *pb.SetUserResourcesRequest) (*pb.Empty, error) {
	updates := map[string]interface{}{}
	if req.RamLimit > 0 {
		updates["ram_limit"] = req.RamLimit
	}
	if req.CpuLimit > 0 {
		updates["cpu_limit"] = req.CpuLimit
	}
	if req.DiskLimit > 0 {
		updates["disk_limit"] = req.DiskLimit
	}
	if req.ServerLimit > 0 {
		updates["server_limit"] = req.ServerLimit
	}
	database.DB.Model(&models.User{}).Where("id = ?", req.UserId).Updates(updates)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ForcePasswordReset(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Model(&models.User{}).Where("id = ?", req.Id).Update("force_password_reset", true)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ListSubusers(ctx context.Context, req *pb.IDRequest) (*pb.ListSubusersResponse, error) {
	serverID, _ := uuid.Parse(req.Id)
	subusers, _ := services.GetSubusers(serverID)
	result := make([]*pb.Subuser, len(subusers))
	for i, su := range subusers {
		var perms []string
		json.Unmarshal(su.Permissions, &perms)
		result[i] = &pb.Subuser{Id: su.ID.String(), UserId: su.UserID.String(), Username: su.User.Username, Email: su.User.Email, Permissions: perms}
	}
	return &pb.ListSubusersResponse{Subusers: result}, nil
}

func (s *PanelServer) AddSubuser(ctx context.Context, req *pb.AddSubuserRequest) (*pb.Subuser, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	var user models.User
	if err := database.DB.First(&user, "email = ?", req.Email).Error; err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	su, err := services.AddSubuser(serverID, user.ID, req.Permissions)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Subuser{Id: su.ID.String(), UserId: user.ID.String(), Username: user.Username, Email: user.Email, Permissions: req.Permissions}, nil
}

func (s *PanelServer) UpdateSubuser(ctx context.Context, req *pb.UpdateSubuserRequest) (*pb.Empty, error) {
	perms, _ := json.Marshal(req.Permissions)
	database.DB.Model(&models.Subuser{}).Where("id = ?", req.SubuserId).Update("permissions", perms)
	return &pb.Empty{}, nil
}

func (s *PanelServer) RemoveSubuser(ctx context.Context, req *pb.RemoveSubuserRequest) (*pb.Empty, error) {
	database.DB.Delete(&models.Subuser{}, "id = ?", req.SubuserId)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ListDatabases(ctx context.Context, req *pb.IDRequest) (*pb.ListDatabasesResponse, error) {
	serverID, _ := uuid.Parse(req.Id)
	dbs, _ := services.GetServerDatabases(serverID)
	result := make([]*pb.Database, len(dbs))
	for i, db := range dbs {
		result[i] = &pb.Database{Id: db.ID.String(), Name: db.DatabaseName, Username: db.Username, Password: db.Password, Host: db.Host.Host, Port: int32(db.Host.Port)}
	}
	return &pb.ListDatabasesResponse{Databases: result}, nil
}

func (s *PanelServer) CreateDatabase(ctx context.Context, req *pb.CreateDatabaseRequest) (*pb.Database, error) {
	serverID, _ := uuid.Parse(req.ServerId)
	hostID, _ := uuid.Parse(req.HostId)
	db, err := services.CreateServerDatabase(serverID, hostID, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var host models.DatabaseHost
	database.DB.First(&host, "id = ?", db.HostID)
	return &pb.Database{Id: db.ID.String(), Name: db.DatabaseName, Username: db.Username, Password: db.Password, Host: host.Host, Port: int32(host.Port)}, nil
}

func (s *PanelServer) DeleteDatabase(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	dbID, _ := uuid.Parse(req.Id)
	services.DeleteServerDatabase(dbID)
	return &pb.Empty{}, nil
}

func (s *PanelServer) RotateDatabasePassword(ctx context.Context, req *pb.IDRequest) (*pb.Database, error) {
	dbID, _ := uuid.Parse(req.Id)
	db, err := services.RotateDatabasePassword(dbID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var host models.DatabaseHost
	database.DB.First(&host, "id = ?", db.HostID)
	return &pb.Database{Id: db.ID.String(), Name: db.DatabaseName, Username: db.Username, Password: db.Password, Host: host.Host, Port: int32(host.Port)}, nil
}

func (s *PanelServer) ListDatabaseHosts(ctx context.Context, req *pb.Empty) (*pb.ListDatabaseHostsResponse, error) {
	var hosts []models.DatabaseHost
	database.DB.Find(&hosts)
	result := make([]*pb.DatabaseHost, len(hosts))
	for i, h := range hosts {
		var count int64
		database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", h.ID).Count(&count)
		result[i] = &pb.DatabaseHost{Id: h.ID.String(), Name: h.Name, Host: h.Host, Port: int32(h.Port), Username: h.Username, MaxDatabases: int32(h.MaxDatabases), DatabasesCount: int32(count)}
	}
	return &pb.ListDatabaseHostsResponse{Hosts: result}, nil
}

func (s *PanelServer) CreateDatabaseHost(ctx context.Context, req *pb.CreateDatabaseHostRequest) (*pb.DatabaseHost, error) {
	host := &models.DatabaseHost{Name: req.Name, Host: req.Host, Port: int(req.Port), Username: req.Username, Password: req.Password, MaxDatabases: int(req.MaxDatabases)}
	database.DB.Create(host)
	return &pb.DatabaseHost{Id: host.ID.String(), Name: host.Name, Host: host.Host, Port: int32(host.Port), Username: host.Username, MaxDatabases: int32(host.MaxDatabases)}, nil
}

func (s *PanelServer) UpdateDatabaseHost(ctx context.Context, req *pb.UpdateDatabaseHostRequest) (*pb.Empty, error) {
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Host != "" {
		updates["host"] = req.Host
	}
	if req.Port > 0 {
		updates["port"] = req.Port
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.MaxDatabases > 0 {
		updates["max_databases"] = req.MaxDatabases
	}
	database.DB.Model(&models.DatabaseHost{}).Where("id = ?", req.Id).Updates(updates)
	return &pb.Empty{}, nil
}

func (s *PanelServer) DeleteDatabaseHost(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Delete(&models.DatabaseHost{}, "id = ?", req.Id)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ListFiles(ctx context.Context, req *pb.FilePathRequest) (*pb.ListFilesResponse, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	data, err := services.ProxyGetToNode(&server, "/files?path="+req.Path)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var resp struct {
		Data []struct {
			Name     string `json:"name"`
			IsDir    bool   `json:"is_dir"`
			Size     int64  `json:"size"`
			Modified string `json:"modified"`
			Mime     string `json:"mime"`
		} `json:"data"`
	}
	b, _ := json.Marshal(data)
	json.Unmarshal(b, &resp)
	result := make([]*pb.FileInfo, len(resp.Data))
	for i, f := range resp.Data {
		result[i] = &pb.FileInfo{Name: f.Name, IsDir: f.IsDir, Size: f.Size, Modified: f.Modified, Mime: f.Mime}
	}
	return &pb.ListFilesResponse{Files: result}, nil
}

func (s *PanelServer) ReadFile(ctx context.Context, req *pb.FilePathRequest) (*pb.FileContent, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	data, err := services.ProxyGetToNode(&server, "/files/read?path="+req.Path)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var resp struct {
		Data struct {
			Content string `json:"content"`
		} `json:"data"`
	}
	b, _ := json.Marshal(data)
	json.Unmarshal(b, &resp)
	return &pb.FileContent{Content: []byte(resp.Data.Content)}, nil
}

func (s *PanelServer) WriteFile(ctx context.Context, req *pb.WriteFileRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	body, _ := json.Marshal(map[string]string{"path": req.Path, "content": string(req.Content)})
	services.ProxyPostToNode(&server, "/files/write", body)
	return &pb.Empty{}, nil
}

func (s *PanelServer) DeleteFile(ctx context.Context, req *pb.FilePathRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	services.ProxyDeleteToNode(&server, "/files?path="+req.Path)
	return &pb.Empty{}, nil
}

func (s *PanelServer) CreateFolder(ctx context.Context, req *pb.FilePathRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	body, _ := json.Marshal(map[string]string{"path": req.Path})
	services.ProxyPostToNode(&server, "/files/folder", body)
	return &pb.Empty{}, nil
}

func (s *PanelServer) MoveFile(ctx context.Context, req *pb.MoveFileRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	body, _ := json.Marshal(map[string]string{"from": req.From, "to": req.To})
	services.ProxyPostToNode(&server, "/files/move", body)
	return &pb.Empty{}, nil
}

func (s *PanelServer) CopyFile(ctx context.Context, req *pb.MoveFileRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	body, _ := json.Marshal(map[string]string{"from": req.From, "to": req.To})
	services.ProxyPostToNode(&server, "/files/copy", body)
	return &pb.Empty{}, nil
}

func (s *PanelServer) CompressFiles(ctx context.Context, req *pb.CompressRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	body, _ := json.Marshal(map[string]interface{}{"paths": req.Paths, "destination": req.Destination})
	services.ProxyPostToNode(&server, "/files/compress", body)
	return &pb.Empty{}, nil
}

func (s *PanelServer) DecompressFile(ctx context.Context, req *pb.FilePathRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	body, _ := json.Marshal(map[string]string{"path": req.Path})
	services.ProxyPostToNode(&server, "/files/decompress", body)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ListBackups(ctx context.Context, req *pb.IDRequest) (*pb.ListBackupsResponse, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	data, _ := services.ProxyGetToNode(&server, "/backups")
	var resp struct {
		Data []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Size      int64  `json:"size"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}
	b, _ := json.Marshal(data)
	json.Unmarshal(b, &resp)
	result := make([]*pb.Backup, len(resp.Data))
	for i, bk := range resp.Data {
		result[i] = &pb.Backup{Id: bk.ID, Name: bk.Name, Size: bk.Size, CreatedAt: bk.CreatedAt}
	}
	return &pb.ListBackupsResponse{Backups: result}, nil
}

func (s *PanelServer) CreateBackup(ctx context.Context, req *pb.CreateBackupRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	body, _ := json.Marshal(map[string]string{"name": req.Name})
	services.ProxyPostToNode(&server, "/backups", body)
	return &pb.Empty{}, nil
}

func (s *PanelServer) DeleteBackup(ctx context.Context, req *pb.DeleteBackupRequest) (*pb.Empty, error) {
	var server models.Server
	if err := database.DB.Preload("Node").First(&server, "id = ?", req.ServerId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "server not found")
	}
	services.ProxyDeleteToNode(&server, "/backups/"+req.BackupId)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ListNodes(ctx context.Context, req *pb.Empty) (*pb.ListNodesResponse, error) {
	var nodes []models.Node
	database.DB.Find(&nodes)
	result := make([]*pb.Node, len(nodes))
	for i, n := range nodes {
		lh := ""
		if n.LastHeartbeat != nil {
			lh = n.LastHeartbeat.String()
		}
		result[i] = &pb.Node{Id: n.ID.String(), Name: n.Name, Fqdn: n.FQDN, Port: int32(n.Port), IsOnline: n.IsOnline, LastHeartbeat: lh}
	}
	return &pb.ListNodesResponse{Nodes: result}, nil
}

func (s *PanelServer) GetNode(ctx context.Context, req *pb.IDRequest) (*pb.Node, error) {
	var node models.Node
	if err := database.DB.First(&node, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "node not found")
	}
	lh := ""
	if node.LastHeartbeat != nil {
		lh = node.LastHeartbeat.String()
	}
	return &pb.Node{Id: node.ID.String(), Name: node.Name, Fqdn: node.FQDN, Port: int32(node.Port), IsOnline: node.IsOnline, LastHeartbeat: lh}, nil
}

func (s *PanelServer) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.NodeWithToken, error) {
	node, token, err := services.CreateNode(req.Name, req.Fqdn, int(req.Port))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.NodeWithToken{
		Node:  &pb.Node{Id: node.ID.String(), Name: node.Name, Fqdn: node.FQDN, Port: int32(node.Port)},
		Token: token.TokenID + "." + token.Token,
	}, nil
}

func (s *PanelServer) DeleteNode(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Delete(&models.Node{}, "id = ?", req.Id)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ResetNodeToken(ctx context.Context, req *pb.IDRequest) (*pb.NodeToken, error) {
	nodeID, _ := uuid.Parse(req.Id)
	token, err := services.ResetNodeToken(nodeID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.NodeToken{TokenId: token.TokenID, Token: token.Token}, nil
}

func (s *PanelServer) ListPackages(ctx context.Context, req *pb.Empty) (*pb.ListPackagesResponse, error) {
	var packages []models.Package
	database.DB.Find(&packages)
	result := make([]*pb.Package, len(packages))
	for i, p := range packages {
		result[i] = packageToProto(&p)
	}
	return &pb.ListPackagesResponse{Packages: result}, nil
}

func (s *PanelServer) GetPackage(ctx context.Context, req *pb.IDRequest) (*pb.Package, error) {
	var pkg models.Package
	if err := database.DB.First(&pkg, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "package not found")
	}
	return packageToProto(&pkg), nil
}

func (s *PanelServer) CreatePackage(ctx context.Context, req *pb.CreatePackageRequest) (*pb.Package, error) {
	pkg := &models.Package{
		Name: req.Name, Description: req.Description, DockerImage: req.DockerImage,
		Startup: req.StartupCommand, StopSignal: req.StopCommand,
	}
	database.DB.Create(pkg)
	return packageToProto(pkg), nil
}

func (s *PanelServer) UpdatePackage(ctx context.Context, req *pb.UpdatePackageRequest) (*pb.Package, error) {
	var pkg models.Package
	if err := database.DB.First(&pkg, "id = ?", req.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "package not found")
	}
	if req.Name != "" {
		pkg.Name = req.Name
	}
	if req.Description != "" {
		pkg.Description = req.Description
	}
	if req.DockerImage != "" {
		pkg.DockerImage = req.DockerImage
	}
	if req.StartupCommand != "" {
		pkg.Startup = req.StartupCommand
	}
	if req.StopCommand != "" {
		pkg.StopSignal = req.StopCommand
	}
	database.DB.Save(&pkg)
	return packageToProto(&pkg), nil
}

func (s *PanelServer) DeletePackage(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Delete(&models.Package{}, "id = ?", req.Id)
	return &pb.Empty{}, nil
}

func (s *PanelServer) ListIPBans(ctx context.Context, req *pb.Empty) (*pb.ListIPBansResponse, error) {
	var bans []models.IPBan
	database.DB.Find(&bans)
	result := make([]*pb.IPBan, len(bans))
	for i, b := range bans {
		result[i] = &pb.IPBan{Id: strconv.Itoa(int(b.ID)), Ip: b.IP, Reason: b.Reason, CreatedAt: b.CreatedAt.String()}
	}
	return &pb.ListIPBansResponse{Bans: result}, nil
}

func (s *PanelServer) CreateIPBan(ctx context.Context, req *pb.CreateIPBanRequest) (*pb.IPBan, error) {
	ban := &models.IPBan{IP: req.Ip, Reason: req.Reason}
	database.DB.Create(ban)
	return &pb.IPBan{Id: strconv.Itoa(int(ban.ID)), Ip: ban.IP, Reason: ban.Reason, CreatedAt: ban.CreatedAt.String()}, nil
}

func (s *PanelServer) DeleteIPBan(ctx context.Context, req *pb.IDRequest) (*pb.Empty, error) {
	database.DB.Delete(&models.IPBan{}, "id = ?", req.Id)
	return &pb.Empty{}, nil
}

func (s *PanelServer) GetSettings(ctx context.Context, req *pb.Empty) (*pb.Settings, error) {
	return &pb.Settings{
		RegistrationEnabled:   services.IsRegistrationEnabled(),
		ServerCreationEnabled: services.IsServerCreationEnabled(),
	}, nil
}

func (s *PanelServer) SetRegistrationEnabled(ctx context.Context, req *pb.BoolRequest) (*pb.Empty, error) {
	services.SetSetting("registration_enabled", strconv.FormatBool(req.Value))
	return &pb.Empty{}, nil
}

func (s *PanelServer) SetServerCreationEnabled(ctx context.Context, req *pb.BoolRequest) (*pb.Empty, error) {
	services.SetSetting("server_creation_enabled", strconv.FormatBool(req.Value))
	return &pb.Empty{}, nil
}

func (s *PanelServer) GetActivityLogs(ctx context.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	var logs []models.ActivityLog
	var total int64
	q := database.DB.Model(&models.ActivityLog{})
	if req.Search != "" {
		q = q.Where("username ILIKE ? OR action ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Filter == "admin" {
		q = q.Where("is_admin = ?", true)
	} else if req.Filter == "user" {
		q = q.Where("is_admin = ?", false)
	}
	q.Count(&total)
	q = q.Order("created_at DESC")
	if req.Limit > 0 {
		q = q.Limit(int(req.Limit))
	} else {
		q = q.Limit(100)
	}
	if req.Offset > 0 {
		q = q.Offset(int(req.Offset))
	}
	q.Find(&logs)
	result := make([]*pb.ActivityLog, len(logs))
	for i, l := range logs {
		result[i] = &pb.ActivityLog{
			Id: l.ID.String(), UserId: l.UserID.String(), Username: l.Username,
			Action: l.Action, Description: l.Description, Ip: l.IP,
			IsAdmin: l.IsAdmin, CreatedAt: l.CreatedAt.String(),
		}
	}
	return &pb.GetLogsResponse{Logs: result, Total: int32(total)}, nil
}

func (s *PanelServer) Log(ctx context.Context, req *pb.LogRequest) (*pb.Empty, error) {
	pluginID := getPluginIDFromContext(ctx)
	prefix := "[plugin:" + pluginID + "]"
	switch req.Level {
	case "error":
		log.Printf("%s ERROR: %s", prefix, req.Message)
	case "warn":
		log.Printf("%s WARN: %s", prefix, req.Message)
	case "debug":
		log.Printf("%s DEBUG: %s", prefix, req.Message)
	default:
		log.Printf("%s %s", prefix, req.Message)
	}
	return &pb.Empty{}, nil
}

func (s *PanelServer) GetKV(ctx context.Context, req *pb.KVRequest) (*pb.KVResponse, error) {
	s.kvMu.RLock()
	defer s.kvMu.RUnlock()
	if val, ok := s.kv[req.Key]; ok {
		return &pb.KVResponse{Value: val, Found: true}, nil
	}
	return &pb.KVResponse{Found: false}, nil
}

func (s *PanelServer) SetKV(ctx context.Context, req *pb.KVSetRequest) (*pb.Empty, error) {
	s.kvMu.Lock()
	defer s.kvMu.Unlock()
	s.kv[req.Key] = req.Value
	return &pb.Empty{}, nil
}

func (s *PanelServer) DeleteKV(ctx context.Context, req *pb.KVRequest) (*pb.Empty, error) {
	s.kvMu.Lock()
	defer s.kvMu.Unlock()
	delete(s.kv, req.Key)
	return &pb.Empty{}, nil
}

func (s *PanelServer) QueryDB(ctx context.Context, req *pb.QueryDBRequest) (*pb.QueryDBResponse, error) {
	if !isReadOnlyQuery(req.Query) {
		return nil, status.Error(codes.PermissionDenied, "only SELECT queries allowed")
	}
	args := make([]interface{}, len(req.Args))
	for i, a := range req.Args {
		args[i] = a
	}
	rows, err := database.DB.Raw(req.Query, args...).Rows()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	var result [][]byte
	cols, _ := rows.Columns()
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		rows.Scan(ptrs...)
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		b, _ := json.Marshal(row)
		result = append(result, b)
	}
	return &pb.QueryDBResponse{Rows: result}, nil
}

func (s *PanelServer) BroadcastEvent(ctx context.Context, req *pb.BroadcastEventRequest) (*pb.Empty, error) {
	Emit(EventType(req.EventType), req.Data)
	return &pb.Empty{}, nil
}

func serverToProto(s *models.Server) *pb.Server {
	return &pb.Server{
		Id: s.ID.String(), Name: s.Name, UserId: s.UserID.String(), NodeId: s.NodeID.String(),
		Status: string(s.Status), Memory: int32(s.Memory), Cpu: int32(s.CPU), Disk: int32(s.Disk),
		Suspended: s.IsSuspended, PackageId: s.PackageID.String(),
	}
}

func userToProto(u *models.User) *pb.User {
	user := &pb.User{
		Id: u.ID.String(), Username: u.Username, Email: u.Email,
		IsAdmin: u.IsAdmin, IsBanned: u.IsBanned, ForcePasswordReset: u.ForcePasswordReset,
		CreatedAt: u.CreatedAt.String(),
	}
	if u.RAMLimit != nil {
		user.RamLimit = int32(*u.RAMLimit)
	}
	if u.CPULimit != nil {
		user.CpuLimit = int32(*u.CPULimit)
	}
	if u.DiskLimit != nil {
		user.DiskLimit = int32(*u.DiskLimit)
	}
	if u.ServerLimit != nil {
		user.ServerLimit = int32(*u.ServerLimit)
	}
	return user
}

func packageToProto(p *models.Package) *pb.Package {
	return &pb.Package{
		Id: p.ID.String(), Name: p.Name, Description: p.Description, DockerImage: p.DockerImage,
		StartupCommand: p.Startup, StopCommand: p.StopSignal,
	}
}

func isReadOnlyQuery(q string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(q)), "SELECT")
}

func getPluginIDFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-plugin-id"); len(vals) > 0 {
			return vals[0]
		}
	}
	return "unknown"
}

func (s *PanelServer) HTTPRequest(ctx context.Context, req *pb.PluginHTTPRequest) (*pb.PluginHTTPResponse, error) {
	timeout := 30
	if req.TimeoutSeconds > 0 {
		timeout = int(req.TimeoutSeconds)
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.Url, body)
	if err != nil {
		return &pb.PluginHTTPResponse{Error: err.Error()}, nil
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return &pb.PluginHTTPResponse{Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	headers := make(map[string]string)
	for k := range resp.Header {
		headers[k] = resp.Header.Get(k)
	}

	return &pb.PluginHTTPResponse{
		Status:  int32(resp.StatusCode),
		Headers: headers,
		Body:    respBody,
	}, nil
}

func (s *PanelServer) CallPlugin(ctx context.Context, req *pb.CallPluginRequest) (*pb.CallPluginResponse, error) {
	plugin := GetRegistry().Get(req.PluginId)
	if plugin == nil || !plugin.Online {
		return &pb.CallPluginResponse{Error: "plugin not found or offline"}, nil
	}

	httpReq := &pb.HTTPRequest{
		Method: "POST",
		Path:   "/_internal/" + req.Method,
		Body:   req.Data,
	}

	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := plugin.Client.OnHTTP(callCtx, httpReq)
	if err != nil {
		return &pb.CallPluginResponse{Error: err.Error()}, nil
	}

	if resp.Status != 200 {
		return &pb.CallPluginResponse{Error: "plugin returned status " + strconv.Itoa(int(resp.Status))}, nil
	}

	return &pb.CallPluginResponse{Data: resp.Body}, nil
}