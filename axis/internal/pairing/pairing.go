package pairing

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"cauthon-axis/internal/config"
	"cauthon-axis/internal/logger"
)

type PairingState struct {
	Active      bool
	ExpiresAt   time.Time
	PendingCode string
	PendingURL  string
	ResultChan  chan PairingResult
	mu          sync.Mutex
}

type PairingResult struct {
	Accepted bool
	TokenID  string
	Token    string
}

type PairingRequest struct {
	PanelURL string `json:"panel_url"`
	Code     string `json:"code"`
}

type PairingResponse struct {
	Success bool   `json:"success"`
	TokenID string `json:"token_id,omitempty"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

var state = &PairingState{}

func StartPairingMode(duration time.Duration) {
	state.mu.Lock()
	defer state.mu.Unlock()

	state.Active = true
	state.ExpiresAt = time.Now().Add(duration)
	state.PendingCode = ""
	state.PendingURL = ""
	state.ResultChan = nil

	logger.Info("Pairing mode active for %v", duration)
	logger.Info("Waiting for pairing request from panel...")

	go func() {
		time.Sleep(duration)
		state.mu.Lock()
		if state.Active && time.Now().After(state.ExpiresAt) {
			state.Active = false
			logger.Info("Pairing mode expired")
		}
		state.mu.Unlock()
	}()
}

func IsActive() bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	return state.Active && time.Now().Before(state.ExpiresAt)
}

func HandlePairingRequest(panelURL, code string) PairingResponse {
	state.mu.Lock()

	if !state.Active || time.Now().After(state.ExpiresAt) {
		state.mu.Unlock()
		return PairingResponse{Success: false, Error: "Pairing mode not active"}
	}

	state.PendingCode = code
	state.PendingURL = panelURL
	state.ResultChan = make(chan PairingResult, 1)
	resultChan := state.ResultChan

	state.mu.Unlock()

	logger.Info("")
	logger.Info("========================================")
	logger.Info("  PAIRING REQUEST RECEIVED")
	logger.Info("========================================")
	logger.Info("  Panel URL: %s", panelURL)
	logger.Info("  Code: %s", code)
	logger.Info("========================================")
	logger.Info("")
	logger.Info("Does this match what you see in the panel?")
	fmt.Print("Accept pairing? [y/N]: ")

	go func() {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		state.mu.Lock()
		defer state.mu.Unlock()

		if state.ResultChan == nil {
			return
		}

		if input == "y" || input == "yes" {
			tokenID, token := generateToken()
			state.ResultChan <- PairingResult{Accepted: true, TokenID: tokenID, Token: token}
		} else {
			state.ResultChan <- PairingResult{Accepted: false}
		}
	}()

	select {
	case result := <-resultChan:
		state.mu.Lock()
		state.Active = false
		state.ResultChan = nil
		state.mu.Unlock()

		if result.Accepted {
			saveToken(result.TokenID, result.Token, panelURL)
			logger.Success("Pairing accepted! Node is now connected to panel.")
			return PairingResponse{Success: true, TokenID: result.TokenID, Token: result.Token}
		}
		logger.Warn("Pairing rejected by user")
		return PairingResponse{Success: false, Error: "Pairing rejected by user"}

	case <-time.After(60 * time.Second):
		state.mu.Lock()
		state.ResultChan = nil
		state.mu.Unlock()
		logger.Warn("Pairing request timed out")
		return PairingResponse{Success: false, Error: "Pairing request timed out"}
	}
}

func generateToken() (tokenID, token string) {
	idBytes := make([]byte, 8)
	rand.Read(idBytes)
	tokenID = hex.EncodeToString(idBytes)

	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token = hex.EncodeToString(tokenBytes)

	return
}

func saveToken(tokenID, token, panelURL string) {
	cfg := config.Get()
	cfg.Panel.URL = panelURL
	cfg.Panel.Token = tokenID + "." + token

	hash := sha256.Sum256([]byte(token))
	_ = hex.EncodeToString(hash[:])

	config.Save()
}
