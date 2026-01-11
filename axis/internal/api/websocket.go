package api

import (
	"bufio"
	"io"
	"sync"
	"time"

	"cauthon-axis/internal/server"
	"encoding/json"

	"github.com/gofiber/contrib/websocket"
)

const (
	logRateLimit   = 200
	logRateWindow  = time.Second
	statusInterval = 500 * time.Millisecond
	pingInterval   = 30 * time.Second
	pongTimeout    = 60 * time.Second
	writeTimeout   = 10 * time.Second
)

func handleServerLogs(c *websocket.Conn) {
	serverID := c.Params("id")
	done := make(chan struct{})
	var closeOnce sync.Once
	var writeMu sync.Mutex
	var statusMu sync.Mutex
	closeDone := func() { closeOnce.Do(func() { close(done) }) }

	var logCount int
	var logWindowStart time.Time
	var throttleNotified bool
	var throttleMu sync.Mutex

	c.SetReadDeadline(time.Now().Add(pongTimeout))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	sendLog := func(line string) {
		throttleMu.Lock()
		now := time.Now()
		if now.Sub(logWindowStart) >= logRateWindow {
			logWindowStart = now
			logCount = 0
			throttleNotified = false
		}
		logCount++
		if logCount > logRateLimit {
			if !throttleNotified {
				throttleNotified = true
				throttleMu.Unlock()
				msg, _ := json.Marshal(map[string]interface{}{"type": "log", "data": "\x1b[36m[Birdactyl] Log output throttled - too many messages\x1b[0m"})
				writeMu.Lock()
				c.SetWriteDeadline(time.Now().Add(writeTimeout))
				c.WriteMessage(websocket.TextMessage, msg)
				writeMu.Unlock()
				return
			}
			throttleMu.Unlock()
			return
		}
		throttleMu.Unlock()

		msg, _ := json.Marshal(map[string]interface{}{"type": "log", "data": line})
		writeMu.Lock()
		c.SetWriteDeadline(time.Now().Add(writeTimeout))
		c.WriteMessage(websocket.TextMessage, msg)
		writeMu.Unlock()
	}

	broadcastCh := server.SubscribeLogs(serverID)
	defer server.UnsubscribeLogs(serverID, broadcastCh)

	if logs, err := server.GetLogs(serverID, "100", false); err == nil {
		scanner := bufio.NewScanner(logs)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) > 8 {
				line = line[8:]
			}
			msg, _ := json.Marshal(map[string]interface{}{"type": "log", "data": string(line)})
			c.WriteMessage(websocket.TextMessage, msg)
		}
		logs.Close()
	}

	initialStatus, _ := server.GetStatus(serverID)
	lastStatus := initialStatus

	statusMsg, _ := json.Marshal(map[string]interface{}{"type": "status", "status": initialStatus})
	c.WriteMessage(websocket.TextMessage, statusMsg)

	if initialStatus == "running" {
		if stats, err := server.GetStats(serverID); err == nil {
			statsMsg, _ := json.Marshal(map[string]interface{}{"type": "stats", "stats": stats})
			c.WriteMessage(websocket.TextMessage, statsMsg)
		}
	}

	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				writeMu.Lock()
				c.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					writeMu.Unlock()
					closeDone()
					return
				}
				writeMu.Unlock()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case line, ok := <-broadcastCh:
				if !ok {
					return
				}
				sendLog(line)
			}
		}
	}()

	go func() {
		var currentLogs io.ReadCloser
		var logDone chan struct{}
		wasStreaming := false

		defer func() {
			if logDone != nil {
				close(logDone)
			}
			if currentLogs != nil {
				currentLogs.Close()
			}
		}()

		for {
			select {
			case <-done:
				return
			default:
			}

			status, _ := server.GetStatus(serverID)
			shouldStream := status == "running"

			if shouldStream && !wasStreaming {
				if currentLogs != nil {
					currentLogs.Close()
					currentLogs = nil
				}
				if logDone != nil {
					close(logDone)
				}
				logDone = make(chan struct{})

				logs, err := server.GetLogs(serverID, "50", true)
				if err == nil {
					currentLogs = logs
					go func(r io.ReadCloser, stopCh chan struct{}) {
						defer r.Close()
						scanner := bufio.NewScanner(r)
						for scanner.Scan() {
							select {
							case <-done:
								return
							case <-stopCh:
								return
							default:
							}
							line := scanner.Bytes()
							if len(line) > 8 {
								line = line[8:]
							}
							sendLog(string(line))
						}
						wasStreaming = false
					}(logs, logDone)
				}
			} else if !shouldStream && wasStreaming {
				if currentLogs != nil {
					currentLogs.Close()
					currentLogs = nil
				}
				if logDone != nil {
					close(logDone)
					logDone = nil
				}
			}
			wasStreaming = shouldStream
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		ticker := time.NewTicker(statusInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				status, _ := server.GetStatus(serverID)
				statusMu.Lock()
				changed := status != lastStatus
				if changed {
					lastStatus = status
				}
				statusMu.Unlock()

				if changed {
					msg, _ := json.Marshal(map[string]interface{}{"type": "status", "status": status})
					writeMu.Lock()
					c.SetWriteDeadline(time.Now().Add(writeTimeout))
					c.WriteMessage(websocket.TextMessage, msg)
					writeMu.Unlock()
				}

				if status == "running" {
					stats, err := server.GetStats(serverID)
					if err == nil {
						msg, _ := json.Marshal(map[string]interface{}{"type": "stats", "stats": stats})
						writeMu.Lock()
						c.SetWriteDeadline(time.Now().Add(writeTimeout))
						c.WriteMessage(websocket.TextMessage, msg)
						writeMu.Unlock()
					}
				}
			}
		}
	}()

	for {
		c.SetReadDeadline(time.Now().Add(pongTimeout))
		_, msg, err := c.ReadMessage()
		if err != nil {
			closeDone()
			return
		}
		var cmd struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		}
		if json.Unmarshal(msg, &cmd) == nil && cmd.Type == "command" {
			server.SendCommand(serverID, cmd.Command)
		}
	}
}
