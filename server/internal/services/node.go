package services

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
)

var (
	ErrNodeNotFound     = errors.New("node not found")
	ErrInvalidNodeToken = errors.New("invalid node token")
	ErrNodeNameTaken    = errors.New("node name already exists")
	ErrPairingFailed    = errors.New("pairing failed")
	ErrPairingRejected  = errors.New("pairing rejected by node")
	ErrPairingTimeout   = errors.New("pairing request timed out")
	ErrNodeNotReady     = errors.New("node not ready for pairing")
)

const heartbeatTimeout = 45 * time.Second

type NodeToken struct {
	TokenID     string `json:"token_id"`
	Token       string `json:"token"`
	DaemonToken string `json:"daemon_token"`
}

func CreateNode(name, fqdn string, port int) (*models.Node, *NodeToken, error) {
	var existing models.Node
	if err := database.DB.Where("name = ?", name).First(&existing).Error; err == nil {
		return nil, nil, ErrNodeNameTaken
	}

	tokenID, token, hash := generateNodeToken()
	daemonToken := tokenID + "." + token

	node := &models.Node{
		Name:        name,
		FQDN:        fqdn,
		Port:        port,
		TokenID:     tokenID,
		TokenHash:   hash,
		DaemonToken: daemonToken,
	}

	if err := database.DB.Create(node).Error; err != nil {
		return nil, nil, err
	}

	return node, &NodeToken{TokenID: tokenID, Token: token, DaemonToken: daemonToken}, nil
}

func GetNodes() ([]models.Node, error) {
	var nodes []models.Node
	err := database.DB.Order("created_at DESC").Find(&nodes).Error

	now := time.Now()
	for i := range nodes {
		if nodes[i].LastHeartbeat != nil {
			nodes[i].IsOnline = now.Sub(*nodes[i].LastHeartbeat) < heartbeatTimeout
		}
	}
	
	return nodes, err
}

func GetOnlineNodes() ([]models.Node, error) {
	var nodes []models.Node
	err := database.DB.Where("is_online = ?", true).Order("name ASC").Find(&nodes).Error

	now := time.Now()
	result := make([]models.Node, 0)
	for _, n := range nodes {
		if n.LastHeartbeat != nil && now.Sub(*n.LastHeartbeat) < heartbeatTimeout {
			n.TokenID = ""
			n.TokenHash = ""
			n.DaemonToken = ""
			result = append(result, n)
		}
	}
	
	return result, err
}

func GetNodeByID(id uuid.UUID) (*models.Node, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", id).First(&node).Error; err != nil {
		return nil, ErrNodeNotFound
	}
	
	if node.LastHeartbeat != nil {
		node.IsOnline = time.Since(*node.LastHeartbeat) < heartbeatTimeout
	}
	
	return &node, nil
}

func DeleteNode(id uuid.UUID) error {
	result := database.DB.Where("id = ?", id).Delete(&models.Node{})
	if result.RowsAffected == 0 {
		return ErrNodeNotFound
	}
	return result.Error
}

func UpdateNode(id uuid.UUID, name, icon string) (*models.Node, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", id).First(&node).Error; err != nil {
		return nil, ErrNodeNotFound
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	updates["icon"] = icon

	if err := database.DB.Model(&node).Updates(updates).Error; err != nil {
		return nil, err
	}

	if node.LastHeartbeat != nil {
		node.IsOnline = time.Since(*node.LastHeartbeat) < heartbeatTimeout
	}

	return &node, nil
}

func ResetNodeToken(id uuid.UUID) (*NodeToken, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", id).First(&node).Error; err != nil {
		return nil, ErrNodeNotFound
	}

	tokenID, token, hash := generateNodeToken()
	daemonToken := tokenID + "." + token

	node.TokenID = tokenID
	node.TokenHash = hash
	node.DaemonToken = daemonToken

	if err := database.DB.Save(&node).Error; err != nil {
		return nil, err
	}

	return &NodeToken{TokenID: tokenID, Token: token, DaemonToken: daemonToken}, nil
}

func ValidateNodeToken(tokenID, token string) (*models.Node, error) {
	var node models.Node
	if err := database.DB.Where("token_id = ?", tokenID).First(&node).Error; err != nil {
		return nil, ErrInvalidNodeToken
	}

	hash := sha256.Sum256([]byte(token))
	hashHex := hex.EncodeToString(hash[:])
	if subtle.ConstantTimeCompare([]byte(hashHex), []byte(node.TokenHash)) != 1 {
		return nil, ErrInvalidNodeToken
	}

	return &node, nil
}

func NodeHeartbeat(nodeID uuid.UUID, systemInfo models.SystemInfo, displayIP string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"is_online":      true,
		"last_heartbeat": now,
		"system_info":    systemInfo,
	}
	if displayIP != "" {
		updates["display_ip"] = displayIP
	}
	return database.DB.Model(&models.Node{}).Where("id = ?", nodeID).Updates(updates).Error
}

func generateNodeToken() (tokenID, token, hash string) {
	idBytes := make([]byte, 8)
	rand.Read(idBytes)
	tokenID = hex.EncodeToString(idBytes)

	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token = hex.EncodeToString(tokenBytes)

	hashBytes := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(hashBytes[:])

	return
}

func RefreshNodes() ([]models.Node, error) {
	var nodes []models.Node
	if err := database.DB.Order("created_at DESC").Find(&nodes).Error; err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}

	var wg sync.WaitGroup
	for i := range nodes {
		wg.Add(1)
		go func(n *models.Node) {
			defer wg.Done()
			pingNode(client, n)
		}(&nodes[i])
	}
	wg.Wait()

	return nodes, nil
}

func pingNode(client *http.Client, node *models.Node) {
	url := fmt.Sprintf("http://%s:%d/api/system", node.FQDN, node.Port)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+node.DaemonToken)
	resp, err := client.Do(req)
	if err != nil {
		node.IsOnline = false
		database.DB.Model(node).Updates(map[string]interface{}{"is_online": false, "auth_error": false})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		node.IsOnline = false
		node.AuthError = true
		database.DB.Model(node).Updates(map[string]interface{}{"is_online": false, "auth_error": true})
		return
	}

	if resp.StatusCode == http.StatusOK {
		var wrapper struct {
			Success bool             `json:"success"`
			Data    models.SystemInfo `json:"data"`
		}
		if json.NewDecoder(resp.Body).Decode(&wrapper) == nil && wrapper.Success {
			now := time.Now()
			node.IsOnline = true
			node.AuthError = false
			node.LastHeartbeat = &now
			node.SystemInfo = wrapper.Data
			database.DB.Model(node).Updates(map[string]interface{}{
				"is_online":      true,
				"auth_error":     false,
				"last_heartbeat": now,
				"system_info":    wrapper.Data,
			})
			return
		}
	}
	node.IsOnline = false
	database.DB.Model(node).Updates(map[string]interface{}{"is_online": false, "auth_error": false})
}

func GeneratePairingCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	code := fmt.Sprintf("%06d", int(b[0])<<16|int(b[1])<<8|int(b[2]))
	if len(code) > 6 {
		code = code[:6]
	}
	return code
}

type PairingResult struct {
	Success bool   `json:"success"`
	TokenID string `json:"token_id,omitempty"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

func PairWithNode(name, fqdn string, port int, panelURL, code string) (*models.Node, *NodeToken, error) {
	var existing models.Node
	if err := database.DB.Where("name = ?", name).First(&existing).Error; err == nil {
		return nil, nil, ErrNodeNameTaken
	}

	client := &http.Client{Timeout: 90 * time.Second}

	reqBody, _ := json.Marshal(map[string]string{
		"panel_url": panelURL,
		"code":      code,
	})

	url := fmt.Sprintf("http://%s:%d/api/pair", fqdn, port)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, ErrNodeNotReady
	}
	defer resp.Body.Close()

	var result PairingResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, ErrPairingFailed
	}

	if !result.Success {
		if result.Error == "Pairing rejected by user" {
			return nil, nil, ErrPairingRejected
		}
		if result.Error == "Pairing request timed out" {
			return nil, nil, ErrPairingTimeout
		}
		if result.Error == "Pairing mode not active. Run 'axis pair' on the node first." {
			return nil, nil, ErrNodeNotReady
		}
		return nil, nil, fmt.Errorf("%s", result.Error)
	}

	hash := sha256.Sum256([]byte(result.Token))
	hashHex := hex.EncodeToString(hash[:])
	daemonToken := result.TokenID + "." + result.Token

	node := &models.Node{
		Name:        name,
		FQDN:        fqdn,
		Port:        port,
		TokenID:     result.TokenID,
		TokenHash:   hashHex,
		DaemonToken: daemonToken,
	}

	if err := database.DB.Create(node).Error; err != nil {
		return nil, nil, err
	}

	return node, &NodeToken{TokenID: result.TokenID, Token: result.Token, DaemonToken: daemonToken}, nil
}
