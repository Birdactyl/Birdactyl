package panel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cauthon-axis/internal/config"
	"cauthon-axis/internal/system"
)

type Client struct {
	httpClient *http.Client
	panelURL   string
	token      string
}

func NewClient() *Client {
	cfg := config.Get()

	transport := &http.Transport{}

	return &Client{
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
		panelURL: cfg.Panel.URL,
		token:    cfg.Panel.Token,
	}
}

func (c *Client) SendHeartbeat() error {
	cfg := config.Get()
	payload := map[string]interface{}{
		"system": system.GetInfo(),
	}
	if cfg.Node.DisplayIP != "" {
		payload["display_ip"] = cfg.Node.DisplayIP
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", c.panelURL+"/api/v1/internal/nodes/heartbeat", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to panel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("panel returned status %d", resp.StatusCode)
	}

	return nil
}
