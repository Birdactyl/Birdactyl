package server

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileEntry struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime int64  `json:"mod_time"`
	Mode    string `json:"mode"`
	SHA512  string `json:"sha512,omitempty"`
}

type SearchResult struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime int64  `json:"mod_time"`
}

func ListFiles(serverID, subPath string) ([]FileEntry, error) {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))

	if !strings.HasPrefix(target, base) {
		return nil, fmt.Errorf("invalid path")
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, err
	}

	files := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, FileEntry{
			Name:    e.Name(),
			Size:    info.Size(),
			IsDir:   e.IsDir(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
		})
	}
	return files, nil
}

func ListFilesWithHashes(serverID, subPath string) ([]FileEntry, error) {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))

	if !strings.HasPrefix(target, base) {
		return nil, fmt.Errorf("invalid path")
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, err
	}

	files := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}

		entry := FileEntry{
			Name:    e.Name(),
			Size:    info.Size(),
			IsDir:   e.IsDir(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
		}

		if !e.IsDir() && info.Size() < 500*1024*1024 {
			filePath := filepath.Join(target, e.Name())
			if hash, err := computeSHA512(filePath); err == nil {
				entry.SHA512 = hash
			}
		}

		files = append(files, entry)
	}
	return files, nil
}

func computeSHA512(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func ReadFile(serverID, subPath string) ([]byte, error) {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))

	if !strings.HasPrefix(target, base) {
		return nil, fmt.Errorf("invalid path")
	}

	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("cannot read directory")
	}
	if info.Size() > 5*1024*1024 {
		return nil, fmt.Errorf("file too large")
	}

	return os.ReadFile(target)
}

func CreateFolder(serverID, subPath string) error {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))

	if !strings.HasPrefix(target, base) {
		return fmt.Errorf("invalid path")
	}

	return os.MkdirAll(target, 0755)
}

func WriteFile(serverID, subPath string, content []byte) error {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))

	if !strings.HasPrefix(target, base) {
		return fmt.Errorf("invalid path")
	}

	return os.WriteFile(target, content, 0644)
}

func WriteFileStream(serverID, subPath string, r io.Reader) error {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))

	if !strings.HasPrefix(target, base) {
		return fmt.Errorf("invalid path")
	}

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func SearchFiles(serverID, query string) ([]SearchResult, error) {
	base := serverDataDir(serverID)
	query = strings.ToLower(query)
	var results []SearchResult

	filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.Contains(strings.ToLower(info.Name()), query) {
			relPath := strings.TrimPrefix(path, base)
			if relPath == "" {
				relPath = "/"
			}
			results = append(results, SearchResult{
				Name:    info.Name(),
				Path:    relPath,
				Size:    info.Size(),
				IsDir:   info.IsDir(),
				ModTime: info.ModTime().Unix(),
			})
		}
		return nil
	})
	return results, nil
}

func DeletePath(serverID, subPath string) error {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))

	if !strings.HasPrefix(target, base) || target == base {
		return fmt.Errorf("invalid path")
	}

	return os.RemoveAll(target)
}

func MovePath(serverID, srcPath, destPath string) error {
	base := serverDataDir(serverID)
	src := filepath.Join(base, filepath.Clean("/"+srcPath))
	dest := filepath.Join(base, filepath.Clean("/"+destPath))

	if !strings.HasPrefix(src, base) || !strings.HasPrefix(dest, base) || src == base {
		return fmt.Errorf("invalid path")
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return os.Rename(src, dest)
}

func CopyPath(serverID, srcPath, destPath string) error {
	base := serverDataDir(serverID)
	src := filepath.Join(base, filepath.Clean("/"+srcPath))
	dest := filepath.Join(base, filepath.Clean("/"+destPath))

	if !strings.HasPrefix(src, base) || !strings.HasPrefix(dest, base) || src == base {
		return fmt.Errorf("invalid path")
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dest)
	}
	return copyFile(src, dest)
}

func GetFilePath(serverID, subPath string) (string, error) {
	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+subPath))
	if !strings.HasPrefix(target, base) {
		return "", fmt.Errorf("invalid path")
	}
	return target, nil
}

func BulkDelete(serverID string, paths []string) (int, error) {
	base := serverDataDir(serverID)
	deleted := 0
	for _, p := range paths {
		target := filepath.Join(base, filepath.Clean("/"+p))
		if !strings.HasPrefix(target, base) || target == base {
			continue
		}
		if err := os.RemoveAll(target); err == nil {
			deleted++
		}
	}
	return deleted, nil
}

func BulkCopy(serverID string, paths []string, destDir string) (int, error) {
	base := serverDataDir(serverID)
	dest := filepath.Join(base, filepath.Clean("/"+destDir))
	if !strings.HasPrefix(dest, base) {
		return 0, fmt.Errorf("invalid destination")
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return 0, err
	}
	copied := 0
	for _, p := range paths {
		src := filepath.Join(base, filepath.Clean("/"+p))
		if !strings.HasPrefix(src, base) || src == base {
			continue
		}
		name := filepath.Base(src)
		target := filepath.Join(dest, name)
		info, err := os.Stat(src)
		if err != nil {
			continue
		}
		if info.IsDir() {
			if copyDir(src, target) == nil {
				copied++
			}
		} else {
			if copyFile(src, target) == nil {
				copied++
			}
		}
	}
	return copied, nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		destPath := filepath.Join(dest, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

var downloadClient = &http.Client{Timeout: 5 * time.Minute}

var blockedHosts = []string{
	"localhost",
	"127.0.0.1",
	"0.0.0.0",
	"::1",
	"169.254.169.254",
	"metadata.google.internal",
	"metadata.internal",
}

var blockedCIDRs = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"fc00::/7",
	"fe80::/10",
}

func isBlockedURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are allowed")
	}

	host := parsed.Hostname()

	for _, blocked := range blockedHosts {
		if strings.EqualFold(host, blocked) {
			return fmt.Errorf("access to internal hosts is not allowed")
		}
	}

	ip := net.ParseIP(host)
	if ip != nil {
		for _, cidr := range blockedCIDRs {
			_, network, _ := net.ParseCIDR(cidr)
			if network.Contains(ip) {
				return fmt.Errorf("access to internal networks is not allowed")
			}
		}
	}

	return nil
}

func DownloadURL(serverID, rawURL, destPath string) error {
	if err := isBlockedURL(rawURL); err != nil {
		return err
	}

	base := serverDataDir(serverID)
	target := filepath.Join(base, filepath.Clean("/"+destPath))

	if !strings.HasPrefix(target, base) {
		return fmt.Errorf("invalid path")
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	resp, err := downloadClient.Get(rawURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
