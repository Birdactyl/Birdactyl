package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"cauthon-axis/internal/docker"
	"cauthon-axis/internal/logger"

	"github.com/docker/docker/api/types/container"
)

var (
	logSubscribers   = make(map[string][]chan string)
	logSubscribersMu sync.RWMutex
	recentMessages   = make(map[string]map[string]time.Time)
	recentMessagesMu sync.Mutex
)

func init() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			recentMessagesMu.Lock()
			now := time.Now()
			for serverID, msgs := range recentMessages {
				for msg, t := range msgs {
					if now.Sub(t) > 5*time.Second {
						delete(msgs, msg)
					}
				}
				if len(msgs) == 0 {
					delete(recentMessages, serverID)
				}
			}
			recentMessagesMu.Unlock()
		}
	}()
}

func BroadcastLog(serverID, message string) {
	message = stripANSI(message)
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}

	recentMessagesMu.Lock()
	if recentMessages[serverID] == nil {
		recentMessages[serverID] = make(map[string]time.Time)
	}
	if lastSent, exists := recentMessages[serverID][message]; exists && time.Since(lastSent) < time.Second {
		recentMessagesMu.Unlock()
		return
	}
	recentMessages[serverID][message] = time.Now()
	for msg, t := range recentMessages[serverID] {
		if time.Since(t) > 5*time.Second {
			delete(recentMessages[serverID], msg)
		}
	}
	recentMessagesMu.Unlock()

	line := fmt.Sprintf("\033[36m[Birdactyl Axis]\033[0m %s", message)
	logger.Server(serverID, message)

	logSubscribersMu.RLock()
	subs := logSubscribers[serverID]
	logSubscribersMu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- line:
		default:
		}
	}
}

func SubscribeLogs(serverID string) chan string {
	ch := make(chan string, 100)
	logSubscribersMu.Lock()
	logSubscribers[serverID] = append(logSubscribers[serverID], ch)
	logSubscribersMu.Unlock()
	return ch
}

func UnsubscribeLogs(serverID string, ch chan string) {
	logSubscribersMu.Lock()
	defer logSubscribersMu.Unlock()
	subs := logSubscribers[serverID]
	for i, sub := range subs {
		if sub == ch {
			logSubscribers[serverID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

func GetLogs(serverID string, tail string, follow bool) (io.ReadCloser, error) {
	ctx := context.Background()
	id, err := docker.GetContainerID(ctx, containerName(serverID))
	if err != nil {
		return nil, err
	}
	return docker.GetContainerLogs(ctx, id, tail, follow)
}

func GetLogLines(serverID string, lines int) ([]string, error) {
	logs, err := GetLogs(serverID, fmt.Sprintf("%d", lines), false)
	if err != nil {
		return nil, err
	}
	defer logs.Close()
	var result []string
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 8 {
			line = line[8:]
		}
		result = append(result, line)
	}
	return result, nil
}

type LogMatch struct {
	Line       string `json:"line"`
	LineNumber int    `json:"line_number"`
	Timestamp  int64  `json:"timestamp"`
}

type LogFileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

func GetFullLog(serverID string) ([]byte, error) {
	logs, err := GetLogs(serverID, "all", false)
	if err != nil {
		return nil, err
	}
	defer logs.Close()
	var buf bytes.Buffer
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 8 {
			line = line[8:]
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

func SearchLogs(serverID, pattern string, regex bool, limit int, since int64) ([]LogMatch, error) {
	logs, err := GetLogs(serverID, "all", false)
	if err != nil {
		return nil, err
	}
	defer logs.Close()
	var matches []LogMatch
	var re *regexp.Regexp
	if regex {
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}
	scanner := bufio.NewScanner(logs)
	lineNum := 0
	for scanner.Scan() && len(matches) < limit {
		lineNum++
		line := scanner.Text()
		if len(line) > 8 {
			line = line[8:]
		}
		matched := false
		if regex && re != nil {
			matched = re.MatchString(line)
		} else {
			matched = strings.Contains(strings.ToLower(line), strings.ToLower(pattern))
		}
		if matched {
			matches = append(matches, LogMatch{Line: line, LineNumber: lineNum, Timestamp: time.Now().UnixMilli()})
		}
	}
	return matches, nil
}

func ListLogFiles(serverID string) ([]LogFileInfo, error) {
	dataDir := serverDataDir(serverID)
	logsDir := filepath.Join(dataDir, "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogFileInfo{}, nil
		}
		return nil, err
	}
	var files []LogFileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, LogFileInfo{
			Name:     e.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format(time.RFC3339),
		})
	}
	return files, nil
}

func ReadLogFile(serverID, filename string) ([]byte, error) {
	dataDir := serverDataDir(serverID)
	safeName := filepath.Base(filename)
	logPath := filepath.Join(dataDir, "logs", safeName)
	return os.ReadFile(logPath)
}

func SendCommand(serverID string, command string) error {
	ctx := context.Background()
	id, err := docker.GetContainerID(ctx, containerName(serverID))
	if err != nil {
		return err
	}
	return docker.SendCommand(ctx, id, command)
}

func streamInstallLogs(ctx context.Context, serverID, containerID string) {
	logs, err := docker.Client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return
	}
	defer logs.Close()

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 8 {
			line = line[8:]
		}
		text := string(line)
		if strings.TrimSpace(text) != "" {
			BroadcastLog(serverID, text)
		}
	}
}
