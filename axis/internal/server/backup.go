package server

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cauthon-axis/internal/config"
)

type Backup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt int64  `json:"created_at"`
	Completed bool   `json:"completed"`
}

var (
	inProgressBackups   = make(map[string]map[string]*Backup)
	inProgressBackupsMu sync.RWMutex
)

func backupDir(serverID string) string {
	cfg := config.Get()
	return filepath.Join(cfg.Node.BackupDir, serverID)
}

func ListBackups(serverID string) ([]Backup, error) {
	dir := backupDir(serverID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	inProgressBackupsMu.RLock()
	serverInProgress := inProgressBackups[serverID]
	inProgressBackupsMu.RUnlock()

	var backups []Backup

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".tar.gz") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".tar.gz")

		if serverInProgress != nil {
			if inProg, ok := serverInProgress[id]; ok {
				backups = append(backups, Backup{
					ID:        id,
					Name:      inProg.Name,
					Size:      info.Size(),
					CreatedAt: inProg.CreatedAt,
					Completed: false,
				})
				continue
			}
		}

		backups = append(backups, Backup{
			ID:        id,
			Name:      id,
			Size:      info.Size(),
			CreatedAt: info.ModTime().Unix(),
			Completed: true,
		})
	}

	if serverInProgress != nil {
		for id, b := range serverInProgress {
			found := false
			for _, existing := range backups {
				if existing.ID == id {
					found = true
					break
				}
			}
			if !found {
				backups = append(backups, *b)
			}
		}
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

func CreateBackup(serverID, name string) (*Backup, error) {
	if name == "" {
		name = fmt.Sprintf("Backup at %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	dir := backupDir(serverID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	safeID := strings.ReplaceAll(name, " ", "-")
	safeID = strings.ReplaceAll(safeID, ":", "-")
	filename := fmt.Sprintf("%s.tar.gz", safeID)
	destPath := filepath.Join(dir, filename)

	srcDir := serverDataDir(serverID)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		if err := os.MkdirAll(srcDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create server directory: %v", err)
		}
	}

	backup := &Backup{
		ID:        safeID,
		Name:      name,
		Size:      0,
		CreatedAt: time.Now().Unix(),
		Completed: false,
	}

	inProgressBackupsMu.Lock()
	if inProgressBackups[serverID] == nil {
		inProgressBackups[serverID] = make(map[string]*Backup)
	}
	inProgressBackups[serverID][safeID] = backup
	inProgressBackupsMu.Unlock()

	go func() {
		BroadcastLog(serverID, fmt.Sprintf("Creating backup: %s", name))
		
		var success bool
		containerName := "birdactyl-" + serverID
		
		status, _ := GetStatus(serverID)
		if status == "running" {
			cmd := exec.Command("docker", "exec", containerName, "tar", "-czf", "-", "-C", "/home/container", ".")
			outFile, err := os.Create(destPath)
			if err != nil {
				BroadcastLog(serverID, fmt.Sprintf("Backup failed: %v", err))
				inProgressBackupsMu.Lock()
				delete(inProgressBackups[serverID], safeID)
				inProgressBackupsMu.Unlock()
				return
			}
			cmd.Stdout = outFile
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err = cmd.Run()
			outFile.Close()
			if err != nil {
				BroadcastLog(serverID, fmt.Sprintf("Backup failed: %v - %s", err, stderr.String()))
				os.Remove(destPath)
			} else {
				success = true
			}
		} else {
			cmd := exec.Command("tar", "-czf", destPath, "-C", srcDir, ".")
			output, err := cmd.CombinedOutput()
			if err != nil {
				BroadcastLog(serverID, fmt.Sprintf("Backup failed: %v - %s", err, string(output)))
				os.Remove(destPath)
			} else {
				success = true
			}
		}

		if success {
			BroadcastLog(serverID, "Backup completed")
		}

		inProgressBackupsMu.Lock()
		delete(inProgressBackups[serverID], safeID)
		inProgressBackupsMu.Unlock()
	}()

	return backup, nil
}

func DeleteBackup(serverID, backupID string) error {
	path := filepath.Join(backupDir(serverID), backupID+".tar.gz")
	return os.Remove(path)
}

func GetBackupPath(serverID, backupID string) (string, error) {
	path := filepath.Join(backupDir(serverID), backupID+".tar.gz")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("backup not found")
	}
	return path, nil
}

func ArchiveServer(serverID string) (string, error) {
	cfg := config.Get()
	archiveDir := filepath.Join(cfg.Node.BackupDir, "transfers")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return "", err
	}

	archivePath := filepath.Join(archiveDir, fmt.Sprintf("%s-transfer.tar.gz", serverID))
	srcDir := serverDataDir(serverID)

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return "", fmt.Errorf("server data directory not found")
	}

	cmd := exec.Command("tar", "-czf", archivePath, "-C", srcDir, ".")
	if err := cmd.Run(); err != nil {
		os.Remove(archivePath)
		return "", fmt.Errorf("failed to create archive: %v", err)
	}

	return archivePath, nil
}

func RestoreBackup(serverID, backupID string) error {
	backupPath, err := GetBackupPath(serverID, backupID)
	if err != nil {
		return err
	}

	destDir := serverDataDir(serverID)

	entries, err := os.ReadDir(destDir)
	if err == nil {
		for _, entry := range entries {
			os.RemoveAll(filepath.Join(destDir, entry.Name()))
		}
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %v", err)
	}

	BroadcastLog(serverID, fmt.Sprintf("Restoring backup: %s", backupID))

	cmd := exec.Command("tar", "-xzf", backupPath, "-C", destDir)
	if err := cmd.Run(); err != nil {
		BroadcastLog(serverID, fmt.Sprintf("Restore failed: %v", err))
		return fmt.Errorf("failed to extract backup: %v", err)
	}

	uid, _ := strconv.Atoi(GetServerUID())
	filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			os.Chown(path, uid, uid)
		}
		return nil
	})

	BroadcastLog(serverID, "Backup restored successfully")
	return nil
}

func GetArchivePath(serverID string) (string, error) {
	cfg := config.Get()
	path := filepath.Join(cfg.Node.BackupDir, "transfers", fmt.Sprintf("%s-transfer.tar.gz", serverID))
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("archive not found")
	}
	return path, nil
}

func DeleteArchive(serverID string) error {
	cfg := config.Get()
	path := filepath.Join(cfg.Node.BackupDir, "transfers", fmt.Sprintf("%s-transfer.tar.gz", serverID))
	return os.Remove(path)
}

func ImportServer(serverID string, archivePath string) error {
	destDir := serverDataDir(serverID)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %v", err)
	}

	cmd := exec.Command("tar", "-xzf", archivePath, "-C", destDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract archive: %v", err)
	}

	uid, _ := strconv.Atoi(GetServerUID())
	filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			os.Chown(path, uid, uid)
		}
		return nil
	})

	return nil
}
