package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type NodeServerConfig struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	DockerImage   string            `json:"docker_image"`
	InstallImage  string            `json:"install_image"`
	InstallScript string            `json:"install_script"`
	Startup       string            `json:"startup"`
	Memory        int               `json:"memory"`
	CPU           int               `json:"cpu"`
	Disk          int               `json:"disk"`
	Ports         []NodePortConfig  `json:"ports"`
	Variables     map[string]string `json:"variables"`
	StopSignal    string            `json:"stop_signal"`
	StopCommand   string            `json:"stop_command"`
	StopTimeout   int               `json:"stop_timeout"`
}

type NodePortConfig struct {
	Host      int    `json:"host"`
	Container int    `json:"container"`
	Protocol  string `json:"protocol"`
}

var httpClient = &http.Client{Timeout: 2 * time.Minute}
var transferClient = &http.Client{Timeout: 0}

func getNodeURL(node *models.Node) string {
	return fmt.Sprintf("http://%s:%d", node.FQDN, node.Port)
}

func SendCreateServer(server *models.Server) error {
	var node models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&node).Error; err != nil {
		return fmt.Errorf("node not found")
	}

	var pkg models.Package
	if err := database.DB.Where("id = ?", server.PackageID).First(&pkg).Error; err != nil {
		return fmt.Errorf("package not found")
	}

	var pkgPorts []models.PackagePort
	json.Unmarshal(pkg.Ports, &pkgPorts)

	var serverPorts []models.ServerPort
	json.Unmarshal(server.Ports, &serverPorts)

	ports := make([]NodePortConfig, 0)
	for i, pp := range pkgPorts {
		hostPort := pp.Default
		if i < len(serverPorts) {
			hostPort = serverPorts[i].Port
		}
		ports = append(ports, NodePortConfig{
			Host:      hostPort,
			Container: pp.Default,
			Protocol:  pp.Protocol,
		})
	}

	var pkgVars []models.PackageVariable
	json.Unmarshal(pkg.Variables, &pkgVars)

	var serverVars map[string]string
	json.Unmarshal(server.Variables, &serverVars)
	if serverVars == nil {
		serverVars = make(map[string]string)
	}

	finalVars := make(map[string]string)
	for _, pv := range pkgVars {
		finalVars[pv.Name] = pv.Default
	}
	for k, v := range serverVars {
		finalVars[k] = v
	}

	if len(ports) > 0 {
		finalVars["SERVER_PORT"] = fmt.Sprintf("%d", ports[0].Container)
	}
	finalVars["SERVER_MEMORY"] = fmt.Sprintf("%d", server.Memory)

	cfg := NodeServerConfig{
		ID:            server.ID.String(),
		Name:          server.Name,
		DockerImage:   orDefault(server.DockerImage, pkg.DockerImage),
		InstallImage:  pkg.InstallImage,
		InstallScript: pkg.InstallScript,
		Startup:       orDefault(server.Startup, pkg.Startup),
		Memory:        server.Memory,
		CPU:           server.CPU,
		Disk:          server.Disk,
		Ports:         ports,
		Variables:     finalVars,
		StopSignal:    pkg.StopSignal,
		StopCommand:   pkg.StopCommand,
		StopTimeout:   pkg.StopTimeout,
	}

	return sendToNode(&node, "POST", "/api/servers", cfg)
}

func SendStartServer(serverID uuid.UUID) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}


	
	var pkg models.Package
	if err := database.DB.Where("id = ?", server.PackageID).First(&pkg).Error; err != nil {
		return fmt.Errorf("package not found")
	}

	var pkgPorts []models.PackagePort
	json.Unmarshal(pkg.Ports, &pkgPorts)

	var serverPorts []models.ServerPort
	json.Unmarshal(server.Ports, &serverPorts)

	ports := make([]NodePortConfig, 0)
	for i, pp := range pkgPorts {
		hostPort := pp.Default
		if i < len(serverPorts) {
			hostPort = serverPorts[i].Port
		}
		ports = append(ports, NodePortConfig{
			Host:      hostPort,
			Container: pp.Default,
			Protocol:  pp.Protocol,
		})
	}

	var pkgVars []models.PackageVariable
	json.Unmarshal(pkg.Variables, &pkgVars)

	var serverVars map[string]string
	json.Unmarshal(server.Variables, &serverVars)
	if serverVars == nil {
		serverVars = make(map[string]string)
	}

	finalVars := make(map[string]string)
	for _, pv := range pkgVars {
		finalVars[pv.Name] = pv.Default
	}
	for k, v := range serverVars {
		finalVars[k] = v
	}

	if len(ports) > 0 {
		finalVars["SERVER_PORT"] = fmt.Sprintf("%d", ports[0].Container)
	}
	finalVars["SERVER_MEMORY"] = fmt.Sprintf("%d", server.Memory)

	cfg := NodeServerConfig{
		ID:            server.ID.String(),
		Name:          server.Name,
		DockerImage:   orDefault(server.DockerImage, pkg.DockerImage),
		InstallImage:  pkg.InstallImage,
		InstallScript: pkg.InstallScript,
		Startup:       orDefault(server.Startup, pkg.Startup),
		Memory:        server.Memory,
		CPU:           server.CPU,
		Disk:          server.Disk,
		Ports:         ports,
		Variables:     finalVars,
		StopSignal:    pkg.StopSignal,
		StopCommand:   pkg.StopCommand,
		StopTimeout:   pkg.StopTimeout,
	}

	return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/start", server.ID), cfg)
}

func SendStopServer(serverID uuid.UUID) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}

	var pkg models.Package
	if err := database.DB.Where("id = ?", server.PackageID).First(&pkg).Error; err != nil {
		return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/stop", server.ID), nil)
	}

	stopCfg := map[string]interface{}{
		"stop_command": pkg.StopCommand,
		"stop_signal":  pkg.StopSignal,
		"stop_timeout": pkg.StopTimeout,
	}

	return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/stop", server.ID), stopCfg)
}

func SendKillServer(serverID uuid.UUID) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}
	return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/kill", server.ID), nil)
}

func SendRestartServer(serverID uuid.UUID) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}

	var pkg models.Package
	if err := database.DB.Where("id = ?", server.PackageID).First(&pkg).Error; err != nil {
		return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/restart", server.ID), nil)
	}

	stopCfg := map[string]interface{}{
		"stop_command": pkg.StopCommand,
		"stop_signal":  pkg.StopSignal,
		"stop_timeout": pkg.StopTimeout,
	}

	return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/restart", server.ID), stopCfg)
}

func SendReinstallServer(server *models.Server) error {
	var node models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&node).Error; err != nil {
		return fmt.Errorf("node not found")
	}

	var pkg models.Package
	if err := database.DB.Where("id = ?", server.PackageID).First(&pkg).Error; err != nil {
		return fmt.Errorf("package not found")
	}

	var pkgPorts []models.PackagePort
	json.Unmarshal(pkg.Ports, &pkgPorts)

	var serverPorts []models.ServerPort
	json.Unmarshal(server.Ports, &serverPorts)

	ports := make([]NodePortConfig, 0)
	for i, pp := range pkgPorts {
		hostPort := pp.Default
		if i < len(serverPorts) {
			hostPort = serverPorts[i].Port
		}
		ports = append(ports, NodePortConfig{
			Host:      hostPort,
			Container: pp.Default,
			Protocol:  pp.Protocol,
		})
	}

	var pkgVars []models.PackageVariable
	json.Unmarshal(pkg.Variables, &pkgVars)

	var serverVars map[string]string
	json.Unmarshal(server.Variables, &serverVars)
	if serverVars == nil {
		serverVars = make(map[string]string)
	}

	finalVars := make(map[string]string)
	for _, pv := range pkgVars {
		finalVars[pv.Name] = pv.Default
	}
	for k, v := range serverVars {
		finalVars[k] = v
	}

	if len(ports) > 0 {
		finalVars["SERVER_PORT"] = fmt.Sprintf("%d", ports[0].Container)
	}
	finalVars["SERVER_MEMORY"] = fmt.Sprintf("%d", server.Memory)

	cfg := NodeServerConfig{
		ID:            server.ID.String(),
		Name:          server.Name,
		DockerImage:   orDefault(server.DockerImage, pkg.DockerImage),
		InstallImage:  pkg.InstallImage,
		InstallScript: pkg.InstallScript,
		Startup:       orDefault(server.Startup, pkg.Startup),
		Memory:        server.Memory,
		CPU:           server.CPU,
		Disk:          server.Disk,
		Ports:         ports,
		Variables:     finalVars,
		StopSignal:    pkg.StopSignal,
		StopCommand:   pkg.StopCommand,
		StopTimeout:   pkg.StopTimeout,
	}

	return sendToNode(&node, "POST", fmt.Sprintf("/api/servers/%s/reinstall", server.ID), cfg)
}

func SendDeleteServer(serverID uuid.UUID) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}
	return sendToNode(node, "DELETE", fmt.Sprintf("/api/servers/%s", server.ID), nil)
}

func getServerAndNode(serverID uuid.UUID) (*models.Server, *models.Node, error) {
	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return nil, nil, fmt.Errorf("server not found")
	}

	var node models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&node).Error; err != nil {
		return nil, nil, fmt.Errorf("node not found")
	}

	return &server, &node, nil
}

func sendToNode(node *models.Node, method, path string, body interface{}) error {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	url := getNodeURL(node) + path
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if errMsg, ok := result["error"].(string); ok {
			return fmt.Errorf("node error: %s", errMsg)
		}
		return fmt.Errorf("node returned status %d", resp.StatusCode)
	}

	return nil
}

type NodeResponse struct {
	StatusCode int
	Body       []byte
}

func ProxyToNode(server *models.Server, method, path string, body interface{}) (*NodeResponse, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&node).Error; err != nil {
		return nil, fmt.Errorf("node not found")
	}

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	url := getNodeURL(&node) + path
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return &NodeResponse{StatusCode: resp.StatusCode, Body: respBody}, nil
}

func ProxyUploadToNode(server *models.Server, path string, body io.Reader, contentType string) (*NodeResponse, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&node).Error; err != nil {
		return nil, fmt.Errorf("node not found")
	}

	url := getNodeURL(&node) + path
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return &NodeResponse{StatusCode: resp.StatusCode, Body: respBody}, nil
}

func StreamDownloadFromNode(server *models.Server, path string) (*http.Response, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&node).Error; err != nil {
		return nil, fmt.Errorf("node not found")
	}

	url := getNodeURL(&node) + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)

	return httpClient.Do(req)
}

func ProxyGetToNode(server *models.Server, path string) (map[string]interface{}, error) {
	resp, err := ProxyToNode(server, "GET", fmt.Sprintf("/api/servers/%s%s", server.ID, path), nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	json.Unmarshal(resp.Body, &result)
	return result, nil
}

func ProxyPostToNode(server *models.Server, path string, body []byte) (map[string]interface{}, error) {
	var bodyData interface{}
	if len(body) > 0 {
		json.Unmarshal(body, &bodyData)
	}
	resp, err := ProxyToNode(server, "POST", fmt.Sprintf("/api/servers/%s%s", server.ID, path), bodyData)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	json.Unmarshal(resp.Body, &result)
	return result, nil
}

func ProxyDeleteToNode(server *models.Server, path string) (map[string]interface{}, error) {
	resp, err := ProxyToNode(server, "DELETE", fmt.Sprintf("/api/servers/%s%s", server.ID, path), nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	json.Unmarshal(resp.Body, &result)
	return result, nil
}

func GetNodeProxyURL(server *models.Server, path string) (string, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", server.NodeID).First(&node).Error; err != nil {
		return "", fmt.Errorf("node not found")
	}
	return fmt.Sprintf("%s/api/servers/%s%s?token=%s", getNodeURL(&node), server.ID, path, node.DaemonToken), nil
}

type ServerStatsResult struct {
	MemoryBytes int64
	MemoryLimit int64
	CPUPercent  float64
	DiskBytes   int64
	NetworkRx   int64
	NetworkTx   int64
	State       string
}

func GetServerStats(serverID uuid.UUID) *ServerStatsResult {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		log.Printf("[nodeclient] GetServerStats: %v", err)
		return nil
	}
	url := fmt.Sprintf("%s/api/servers/%s/status", getNodeURL(node), server.ID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[nodeclient] GetServerStats request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var result struct {
		Data struct {
			Status string `json:"status"`
			Stats  struct {
				Memory    int64   `json:"memory"`
				MemLimit  int64   `json:"memory_limit"`
				CPU       float64 `json:"cpu"`
				Disk      int64   `json:"disk"`
				NetworkRx int64   `json:"network_rx"`
				NetworkTx int64   `json:"network_tx"`
			} `json:"stats"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return &ServerStatsResult{
		MemoryBytes: result.Data.Stats.Memory,
		MemoryLimit: result.Data.Stats.MemLimit,
		CPUPercent:  result.Data.Stats.CPU,
		DiskBytes:   result.Data.Stats.Disk,
		NetworkRx:   result.Data.Stats.NetworkRx,
		NetworkTx:   result.Data.Stats.NetworkTx,
		State:       result.Data.Status,
	}
}

func GetConsoleLog(serverID uuid.UUID, lines int) []string {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		log.Printf("[nodeclient] GetConsoleLog: %v", err)
		return nil
	}
	url := fmt.Sprintf("%s/api/servers/%s/logs?lines=%d", getNodeURL(node), server.ID, lines)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[nodeclient] GetConsoleLog request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var result struct {
		Lines []string `json:"lines"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Lines
}

func SendCommand(serverID uuid.UUID, command string) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}
	return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/command", server.ID), map[string]string{"command": command})
}

func CreateServerArchive(serverID uuid.UUID) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}
	return sendToNode(node, "POST", fmt.Sprintf("/api/servers/%s/archive", server.ID), nil)
}

func DownloadServerArchive(serverID uuid.UUID) (*http.Response, error) {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return nil, err
	}

	url := getNodeURL(node) + fmt.Sprintf("/api/servers/%s/archive/download", server.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)

	return transferClient.Do(req)
}

func DeleteServerArchive(serverID uuid.UUID) error {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return err
	}
	return sendToNode(node, "DELETE", fmt.Sprintf("/api/servers/%s/archive", server.ID), nil)
}

func ImportServerArchive(targetNode *models.Node, serverID string, sourceNode *models.Node) error {
	url := getNodeURL(targetNode) + fmt.Sprintf("/api/servers/%s/import", serverID)
	sourceURL := getNodeURL(sourceNode) + fmt.Sprintf("/api/servers/%s/archive/download", serverID)

	body := map[string]string{
		"url":   sourceURL,
		"token": sourceNode.DaemonToken,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+targetNode.DaemonToken)

	resp, err := transferClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to target node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if errMsg, ok := result["error"].(string); ok {
			return fmt.Errorf("target node error: %s", errMsg)
		}
		return fmt.Errorf("target node returned status %d", resp.StatusCode)
	}

	return nil
}

func orDefault(val, def string) string {
	if val != "" {
		return val
	}
	return def
}

type LogMatch struct {
	Line       string
	LineNumber int
	Timestamp  int64
}

type LogFileInfo struct {
	Name     string
	Size     int64
	Modified string
}

var (
	consoleSubscribers   = make(map[uuid.UUID][]chan string)
	consoleSubscribersMu sync.RWMutex
)

func SubscribeConsole(serverID uuid.UUID) chan string {
	ch := make(chan string, 100)
	consoleSubscribersMu.Lock()
	consoleSubscribers[serverID] = append(consoleSubscribers[serverID], ch)
	consoleSubscribersMu.Unlock()
	go streamFromNode(serverID, ch)
	return ch
}

func UnsubscribeConsole(serverID uuid.UUID, ch chan string) {
	consoleSubscribersMu.Lock()
	defer consoleSubscribersMu.Unlock()
	subs := consoleSubscribers[serverID]
	for i, sub := range subs {
		if sub == ch {
			consoleSubscribers[serverID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

func streamFromNode(serverID uuid.UUID, ch chan string) {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		return
	}
	url := fmt.Sprintf("ws://%s:%d/api/servers/%s/ws?token=%s", node.FQDN, node.Port, server.ID, node.DaemonToken)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var data struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}
		if json.Unmarshal(msg, &data) == nil && data.Type == "log" {
			select {
			case ch <- data.Data:
			default:
			}
		}
	}
}

func GetFullLog(serverID uuid.UUID) ([]byte, int64) {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		log.Printf("[nodeclient] GetFullLog: %v", err)
		return nil, 0
	}
	url := fmt.Sprintf("%s/api/servers/%s/logs/full", getNodeURL(node), server.ID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[nodeclient] GetFullLog request failed: %v", err)
		return nil, 0
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	return content, int64(len(content))
}

func SearchLogs(serverID uuid.UUID, pattern string, regex bool, limit int, since int64) []LogMatch {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		log.Printf("[nodeclient] SearchLogs: %v", err)
		return nil
	}
	url := fmt.Sprintf("%s/api/servers/%s/logs/search?pattern=%s&regex=%t&limit=%d&since=%d",
		getNodeURL(node), server.ID, pattern, regex, limit, since)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[nodeclient] SearchLogs request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var result struct {
		Matches []LogMatch `json:"matches"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Matches
}

func ListLogFiles(serverID uuid.UUID) []LogFileInfo {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		log.Printf("[nodeclient] ListLogFiles: %v", err)
		return nil
	}
	url := fmt.Sprintf("%s/api/servers/%s/logs/files", getNodeURL(node), server.ID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[nodeclient] ListLogFiles request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var result struct {
		Files []LogFileInfo `json:"files"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Files
}

func ReadLogFile(serverID uuid.UUID, filename string) ([]byte, int64) {
	server, node, err := getServerAndNode(serverID)
	if err != nil {
		log.Printf("[nodeclient] ReadLogFile: %v", err)
		return nil, 0
	}
	url := fmt.Sprintf("%s/api/servers/%s/logs/file/%s", getNodeURL(node), server.ID, filename)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[nodeclient] ReadLogFile request failed: %v", err)
		return nil, 0
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	return content, int64(len(content))
}
