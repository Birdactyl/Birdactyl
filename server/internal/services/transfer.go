package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
)

type TransferStage string

const (
	TransferStagePending     TransferStage = "pending"
	TransferStageStopping    TransferStage = "stopping"
	TransferStageArchiving   TransferStage = "archiving"
	TransferStageDownloading TransferStage = "downloading"
	TransferStageUploading   TransferStage = "uploading"
	TransferStageImporting   TransferStage = "importing"
	TransferStageCleanup     TransferStage = "cleanup"
	TransferStageComplete    TransferStage = "complete"
	TransferStageFailed      TransferStage = "failed"
)

type TransferStatus struct {
	ID           string        `json:"id"`
	ServerID     string        `json:"server_id"`
	ServerName   string        `json:"server_name"`
	FromNodeID   string        `json:"from_node_id"`
	FromNodeName string        `json:"from_node_name"`
	ToNodeID     string        `json:"to_node_id"`
	ToNodeName   string        `json:"to_node_name"`
	Stage        TransferStage `json:"stage"`
	Progress     int           `json:"progress"`
	Error        string        `json:"error,omitempty"`
	StartedAt    time.Time     `json:"started_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
}

var (
	transfers   = make(map[string]*TransferStatus)
	transfersMu sync.RWMutex
)

func GetTransferStatus(transferID string) *TransferStatus {
	transfersMu.RLock()
	defer transfersMu.RUnlock()
	return transfers[transferID]
}

func GetAllTransfers() []*TransferStatus {
	transfersMu.RLock()
	defer transfersMu.RUnlock()
	result := make([]*TransferStatus, 0, len(transfers))
	for _, t := range transfers {
		result = append(result, t)
	}
	return result
}

func StartTransfer(serverID, targetNodeID uuid.UUID) (string, error) {
	var server models.Server
	if err := database.DB.Preload("Node").Where("id = ?", serverID).First(&server).Error; err != nil {
		return "", fmt.Errorf("server not found")
	}

	if server.NodeID == targetNodeID {
		return "", fmt.Errorf("server is already on this node")
	}

	var targetNode models.Node
	if err := database.DB.Where("id = ?", targetNodeID).First(&targetNode).Error; err != nil {
		return "", fmt.Errorf("target node not found")
	}

	if !targetNode.IsOnline {
		return "", fmt.Errorf("target node is offline")
	}

	if server.Node != nil && !server.Node.IsOnline {
		return "", fmt.Errorf("source node is offline")
	}

	transferID := uuid.New().String()[:8]

	status := &TransferStatus{
		ID:           transferID,
		ServerID:     serverID.String(),
		ServerName:   server.Name,
		FromNodeID:   server.NodeID.String(),
		FromNodeName: server.Node.Name,
		ToNodeID:     targetNodeID.String(),
		ToNodeName:   targetNode.Name,
		Stage:        TransferStagePending,
		Progress:     0,
		StartedAt:    time.Now(),
	}

	transfersMu.Lock()
	transfers[transferID] = status
	transfersMu.Unlock()

	go runTransfer(status, &server, &targetNode)

	return transferID, nil
}

func updateTransfer(status *TransferStatus, stage TransferStage, progress int) {
	transfersMu.Lock()
	status.Stage = stage
	status.Progress = progress
	transfersMu.Unlock()
}

func failTransfer(status *TransferStatus, err error) {
	transfersMu.Lock()
	status.Stage = TransferStageFailed
	status.Error = err.Error()
	now := time.Now()
	status.CompletedAt = &now
	transfersMu.Unlock()
}

func runTransfer(status *TransferStatus, server *models.Server, targetNode *models.Node) {
	serverID := server.ID
	log.Printf("[Transfer] Starting transfer for server %s to node %s", serverID, targetNode.Name)

	var sourceNode models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&sourceNode).Error; err != nil {
		failTransfer(status, fmt.Errorf("source node not found"))
		return
	}

	updateTransfer(status, TransferStageStopping, 5)
	if server.Status == models.ServerStatusRunning {
		log.Printf("[Transfer] Stopping server %s", serverID)
		SendStopServer(serverID)
		UpdateServerStatus(serverID, models.ServerStatusStopped, "")
		time.Sleep(2 * time.Second)
	}

	updateTransfer(status, TransferStageArchiving, 15)
	log.Printf("[Transfer] Creating archive for %s", serverID)
	start := time.Now()
	if err := CreateServerArchive(serverID); err != nil {
		failTransfer(status, fmt.Errorf("failed to create archive: %w", err))
		return
	}
	log.Printf("[Transfer] Archive created in %v", time.Since(start))

	updateTransfer(status, TransferStageUploading, 40)
	log.Printf("[Transfer] Telling target node to fetch from source node")
	start = time.Now()
	err := ImportServerArchive(targetNode, serverID.String(), &sourceNode)
	if err != nil {
		DeleteServerArchive(serverID)
		failTransfer(status, fmt.Errorf("failed to import: %w", err))
		return
	}
	log.Printf("[Transfer] Transfer completed in %v", time.Since(start))

	updateTransfer(status, TransferStageCleanup, 85)
	DeleteServerArchive(serverID)
	SendDeleteServer(serverID)

	updateTransfer(status, TransferStageImporting, 95)
	newPorts := AllocatePortsForNode(uuid.MustParse(status.ToNodeID), server.Ports)
	database.DB.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"node_id": status.ToNodeID,
		"ports":   newPorts,
	})

	updateTransfer(status, TransferStageComplete, 100)
	now := time.Now()
	transfersMu.Lock()
	status.CompletedAt = &now
	transfersMu.Unlock()

	go func() {
		time.Sleep(5 * time.Minute)
		transfersMu.Lock()
		delete(transfers, status.ID)
		transfersMu.Unlock()
	}()
}
