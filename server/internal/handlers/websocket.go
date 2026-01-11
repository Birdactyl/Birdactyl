package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
)

const (
	wsPingInterval  = 30 * time.Second
	wsPongTimeout   = 60 * time.Second
	wsWriteTimeout  = 10 * time.Second
)

func ServerLogsWS(c *websocket.Conn) {
	serverID := c.Params("id")
	userID := c.Locals("userID")
	isAdmin := c.Locals("isAdmin")

	parsedID, err := uuid.Parse(serverID)
	if err != nil {
		c.WriteJSON(map[string]string{"error": "Invalid server ID"})
		return
	}

	if userID == nil {
		c.WriteJSON(map[string]string{"error": "Unauthorized"})
		return
	}

	uid := userID.(uuid.UUID)

	var server models.Server
	if err := database.DB.Preload("Node").Where("id = ?", parsedID).First(&server).Error; err != nil {
		c.WriteJSON(map[string]string{"error": "Server not found"})
		return
	}

	canRead := isAdmin == true || server.UserID == uid || services.HasServerPermission(uid, parsedID, false, models.PermConsoleRead)
	canWrite := isAdmin == true || server.UserID == uid || services.HasServerPermission(uid, parsedID, false, models.PermConsoleWrite)

	if !canRead {
		c.WriteJSON(map[string]string{"error": "Permission denied"})
		return
	}

	if server.Node == nil {
		c.WriteJSON(map[string]string{"error": "Node not found"})
		return
	}

	nodeURL := fmt.Sprintf("ws://%s:%d/api/servers/%s/ws?token=%s",
		server.Node.FQDN, server.Node.Port, server.ID.String(),
		url.QueryEscape(server.Node.DaemonToken))

	dialer := gorilla.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	nodeConn, _, err := dialer.Dial(nodeURL, nil)
	if err != nil {
		c.WriteJSON(map[string]string{"error": "Failed to connect to node: " + err.Error()})
		return
	}
	defer nodeConn.Close()

	done := make(chan struct{})
	var closeOnce sync.Once
	closeDone := func() { closeOnce.Do(func() { close(done) }) }
	var writeMu sync.Mutex

	nodeConn.SetReadDeadline(time.Now().Add(wsPongTimeout))
	nodeConn.SetPongHandler(func(string) error {
		nodeConn.SetReadDeadline(time.Now().Add(wsPongTimeout))
		return nil
	})

	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(wsPongTimeout))
		return nil
	})

	go func() {
		ticker := time.NewTicker(wsPingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				writeMu.Lock()
				c.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					writeMu.Unlock()
					closeDone()
					return
				}
				writeMu.Unlock()

				nodeConn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
				if err := nodeConn.WriteMessage(gorilla.PingMessage, nil); err != nil {
					closeDone()
					return
				}
			}
		}
	}()

	go func() {
		defer closeDone()
		for {
			nodeConn.SetReadDeadline(time.Now().Add(wsPongTimeout))
			_, msg, err := nodeConn.ReadMessage()
			if err != nil {
				return
			}
			writeMu.Lock()
			c.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			err = c.WriteMessage(websocket.TextMessage, msg)
			writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		default:
		}

		c.SetReadDeadline(time.Now().Add(wsPongTimeout))
		_, msg, err := c.ReadMessage()
		if err != nil {
			closeDone()
			return
		}

		var wsMsg struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(msg, &wsMsg) == nil && wsMsg.Type == "command" {
			if !canWrite {
				continue
			}
		}

		nodeConn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
		if err := nodeConn.WriteMessage(gorilla.TextMessage, msg); err != nil {
			closeDone()
			return
		}
	}
}
