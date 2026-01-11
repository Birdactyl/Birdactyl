package services

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sync"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

var (
	ErrServerNotFound   = errors.New("server not found")
	ErrNodeOffline      = errors.New("node is offline")
	ErrServerSuspended  = errors.New("server is suspended")
	portAllocationMu    sync.Mutex
)

type ResourceUsage struct {
	Servers int
	RAM     int
	CPU     int
	Disk    int
}

func GetUserResourceUsage(userID uuid.UUID) ResourceUsage {
	var servers []models.Server
	database.DB.Where("user_id = ?", userID).Find(&servers)
	
	usage := ResourceUsage{Servers: len(servers)}
	for _, s := range servers {
		usage.RAM += s.Memory
		usage.CPU += s.CPU
		usage.Disk += s.Disk
	}
	return usage
}

type CreateServerRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	NodeID      uuid.UUID         `json:"node_id"`
	PackageID   uuid.UUID         `json:"package_id"`
	Memory      int               `json:"memory"`
	CPU         int               `json:"cpu"`
	Disk        int               `json:"disk"`
	Ports       []models.ServerPort `json:"ports"`
	Variables   map[string]string `json:"variables"`
}

func allocatePort(nodeID uuid.UUID) int {
	portAllocationMu.Lock()
	defer portAllocationMu.Unlock()

	var servers []models.Server
	database.DB.Where("node_id = ?", nodeID).Find(&servers)
	
	usedPorts := make(map[int]bool)
	for _, s := range servers {
		var ports []models.ServerPort
		json.Unmarshal(s.Ports, &ports)
		for _, p := range ports {
			usedPorts[p.Port] = true
		}
	}
	
	for attempts := 0; attempts < 1000; attempts++ {
		port := 25565 + rand.Intn(4436)
		if !usedPorts[port] {
			return port
		}
	}
	for port := 25565; port <= 30000; port++ {
		if !usedPorts[port] {
			return port
		}
	}
	return 25565
}

func CreateServer(userID uuid.UUID, req CreateServerRequest) (*models.Server, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", req.NodeID).First(&node).Error; err != nil {
		return nil, ErrNodeNotFound
	}
	if !node.IsOnline {
		return nil, ErrNodeOffline
	}

	var pkg models.Package
	if err := database.DB.Where("id = ?", req.PackageID).First(&pkg).Error; err != nil {
		return nil, ErrPackageNotFound
	}

	for i := range req.Ports {
		req.Ports[i].Port = allocatePort(req.NodeID)
	}

	portsJSON, _ := json.Marshal(req.Ports)
	varsJSON, _ := json.Marshal(req.Variables)

	server := &models.Server{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
		NodeID:      req.NodeID,
		PackageID:   req.PackageID,
		Status:      models.ServerStatusInstalling,
		Memory:      req.Memory,
		CPU:         req.CPU,
		Disk:        req.Disk,
		Ports:       portsJSON,
		Variables:   varsJSON,
	}

	if err := database.DB.Create(server).Error; err != nil {
		return nil, err
	}

	database.DB.Preload("Node").Preload("Package").First(server, server.ID)

	return server, nil
}

func GetServersByUser(userID uuid.UUID) ([]models.Server, error) {
	var servers []models.Server
	
	var subuserServerIDs []uuid.UUID
	database.DB.Model(&models.Subuser{}).Where("user_id = ?", userID).Pluck("server_id", &subuserServerIDs)
	
	err := database.DB.Where("user_id = ? OR id IN ?", userID, subuserServerIDs).
		Preload("Node").Preload("Package").Preload("User").
		Order("created_at DESC").Find(&servers).Error
	if err != nil {
		return servers, err
	}

	for i := range servers {
		if stats := GetServerStats(servers[i].ID); stats != nil {
			servers[i].Status = models.ServerStatus(stats.State)
		}
	}

	return servers, nil
}

func GetServerByID(serverID, userID uuid.UUID, isAdmin bool) (*models.Server, error) {
	var server models.Server
	query := database.DB.Preload("Node").Preload("Package").Preload("User")
	
	if isAdmin {
		query = query.Where("id = ?", serverID)
	} else {
		var subuserCount int64
		database.DB.Model(&models.Subuser{}).Where("server_id = ? AND user_id = ?", serverID, userID).Count(&subuserCount)
		if subuserCount > 0 {
			query = query.Where("id = ?", serverID)
		} else {
			query = query.Where("id = ? AND user_id = ?", serverID, userID)
		}
	}
	
	if err := query.First(&server).Error; err != nil {
		return nil, ErrServerNotFound
	}
	return &server, nil
}

func GetAllServers() ([]models.Server, error) {
	var servers []models.Server
	err := database.DB.Preload("User").Preload("Node").Preload("Package").
		Order("created_at DESC").Find(&servers).Error
	return servers, err
}

func UpdateServerStatus(serverID uuid.UUID, status models.ServerStatus, containerID string) error {
	updates := map[string]interface{}{"status": status}
	if containerID != "" {
		updates["container_id"] = containerID
	}
	return database.DB.Model(&models.Server{}).Where("id = ?", serverID).Updates(updates).Error
}

func DeleteServer(serverID, userID uuid.UUID, isAdmin bool) error {
	query := database.DB.Where("id = ?", serverID)
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}
	
	var count int64
	query.Model(&models.Server{}).Count(&count)
	if count == 0 {
		return ErrServerNotFound
	}
	
	database.DB.Where("server_id = ?", serverID).Delete(&models.ServerDatabase{})
	database.DB.Where("server_id = ?", serverID).Delete(&models.Subuser{})
	database.DB.Where("server_id = ?", serverID).Delete(&models.Schedule{})
	
	result := database.DB.Where("id = ?", serverID).Delete(&models.Server{})
	return result.Error
}

func AddAllocation(serverID, userID uuid.UUID, isAdmin bool) (*models.Server, error) {
	server, err := GetServerByID(serverID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	var ports []models.ServerPort
	json.Unmarshal(server.Ports, &ports)

	newPort := allocatePort(server.NodeID)
	ports = append(ports, models.ServerPort{Port: newPort, Primary: false})

	portsJSON, _ := json.Marshal(ports)
	server.Ports = portsJSON

	if err := database.DB.Save(server).Error; err != nil {
		return nil, err
	}

	return server, nil
}

func SetPrimaryAllocation(serverID, userID uuid.UUID, port int, isAdmin bool) (*models.Server, error) {
	server, err := GetServerByID(serverID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	var ports []models.ServerPort
	json.Unmarshal(server.Ports, &ports)

	found := false
	for i := range ports {
		if ports[i].Port == port {
			ports[i].Primary = true
			found = true
		} else {
			ports[i].Primary = false
		}
	}

	if !found {
		return nil, errors.New("port not found")
	}

	portsJSON, _ := json.Marshal(ports)
	server.Ports = portsJSON

	if err := database.DB.Save(server).Error; err != nil {
		return nil, err
	}

	return server, nil
}

func DeleteAllocation(serverID, userID uuid.UUID, port int, isAdmin bool) (*models.Server, error) {
	server, err := GetServerByID(serverID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	var ports []models.ServerPort
	json.Unmarshal(server.Ports, &ports)

	idx := -1
	for i, p := range ports {
		if p.Port == port {
			if p.Primary {
				return nil, errors.New("cannot delete primary allocation")
			}
			idx = i
			break
		}
	}

	if idx == -1 {
		return nil, errors.New("port not found")
	}

	ports = append(ports[:idx], ports[idx+1:]...)

	portsJSON, _ := json.Marshal(ports)
	server.Ports = portsJSON

	if err := database.DB.Save(server).Error; err != nil {
		return nil, err
	}

	return server, nil
}

func UpdateServerResources(serverID, userID uuid.UUID, memory, cpu, disk int, isAdmin bool) (*models.Server, error) {
	server, err := GetServerByID(serverID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	if memory > 0 {
		server.Memory = memory
	}
	if cpu > 0 {
		server.CPU = cpu
	}
	if disk > 0 {
		server.Disk = disk
	}

	if err := database.DB.Save(server).Error; err != nil {
		return nil, err
	}

	return server, nil
}

func UpdateServerResourcesAdmin(serverID uuid.UUID, memory, cpu, disk int) (*models.Server, error) {
	var server models.Server
	if err := database.DB.Preload("Node").Preload("Package").Where("id = ?", serverID).First(&server).Error; err != nil {
		return nil, ErrServerNotFound
	}

	if memory > 0 {
		server.Memory = memory
	}
	if cpu > 0 {
		server.CPU = cpu
	}
	if disk > 0 {
		server.Disk = disk
	}

	if err := database.DB.Save(&server).Error; err != nil {
		return nil, err
	}

	return &server, nil
}

func UpdateServerName(serverID, userID uuid.UUID, name string, isAdmin bool) (*models.Server, error) {
	server, err := GetServerByID(serverID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	server.Name = name

	if err := database.DB.Save(server).Error; err != nil {
		return nil, err
	}

	return server, nil
}

func UpdateServerVariables(serverID, userID uuid.UUID, variables map[string]string, startup, dockerImage string, isAdmin bool) (*models.Server, error) {
	server, err := GetServerByID(serverID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	varsJSON, _ := json.Marshal(variables)
	server.Variables = varsJSON
	server.Startup = startup
	server.DockerImage = dockerImage

	if err := database.DB.Save(server).Error; err != nil {
		return nil, err
	}

	return server, nil
}

func SuspendServer(serverID uuid.UUID) error {
	return database.DB.Model(&models.Server{}).Where("id = ?", serverID).Update("is_suspended", true).Error
}

func DeleteServerAdmin(serverID uuid.UUID) error {
	result := database.DB.Where("id = ?", serverID).Delete(&models.Server{})
	if result.RowsAffected == 0 {
		return ErrServerNotFound
	}
	return result.Error
}

func UnsuspendServer(serverID uuid.UUID) error {
	return database.DB.Model(&models.Server{}).Where("id = ?", serverID).Update("is_suspended", false).Error
}

func IsServerSuspended(serverID uuid.UUID) (bool, error) {
	var server models.Server
	if err := database.DB.Select("is_suspended").Where("id = ?", serverID).First(&server).Error; err != nil {
		return false, err
	}
	return server.IsSuspended, nil
}

func AllocatePortsForNode(nodeID uuid.UUID, existingPorts datatypes.JSON) datatypes.JSON {
	var ports []models.ServerPort
	json.Unmarshal(existingPorts, &ports)

	for i := range ports {
		ports[i].Port = allocatePort(nodeID)
	}

	newPorts, _ := json.Marshal(ports)
	return newPorts
}
