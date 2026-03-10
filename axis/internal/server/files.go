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

	"github.com/spf13/afero"
)

var osFs = afero.NewOsFs()

func GetVFS(serverID string) afero.Fs {
	base := serverDataDir(serverID)
	osFs.MkdirAll(base, 0755)
	
	bfs := afero.NewBasePathFs(osFs, base)

	serverConfigsMu.RLock()
	cfg := serverConfigs[serverID]
	serverConfigsMu.RUnlock()

	if cfg == nil || len(cfg.Mounts) == 0 {
		return bfs
	}

	mountsFs := &mountFs{
		base:   bfs,
		mounts: make(map[string]afero.Fs),
	}

	for _, m := range cfg.Mounts {
		if !m.Navigable {
			continue
		}

		target := filepath.Clean(m.Target)
		if !strings.HasPrefix(target, "/home/container") {
			continue
		}
		
		virtualTarget := strings.TrimPrefix(target, "/home/container")
		if virtualTarget == "" {
			virtualTarget = "/"
		}

		bfs.MkdirAll(virtualTarget, 0755)
		
		realFs := afero.NewBasePathFs(osFs, m.Source)
		if m.ReadOnly {
			realFs = afero.NewReadOnlyFs(realFs)
		}
		mountsFs.mounts[virtualTarget] = realFs
	}
	
	return mountsFs
}

func GetRealPath(serverID, subPath string) (string, error) {
	base := serverDataDir(serverID)
	osFs.MkdirAll(base, 0755)

	p := filepath.Clean("/" + subPath)
	bfs := afero.NewBasePathFs(osFs, base).(*afero.BasePathFs)

	realPath, err := bfs.RealPath(p)
	if err != nil {
		return "", err
	}

	checkPath := realPath
	for {
		eval, err := filepath.EvalSymlinks(checkPath)
		if err == nil {
			if !strings.HasPrefix(eval, base) && eval != base {
				return "", fmt.Errorf("security violation: path escapes the server sandbox")
			}
			break // we are good
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath || parent == base || parent == "/" || parent == "." {
			break
		}
		checkPath = parent
	}

	return realPath, nil
}

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
	fs := GetVFS(serverID)
	target := filepath.Clean("/" + subPath)

	entries, err := afero.ReadDir(fs, target)
	if err != nil {
		if strings.HasPrefix(target, "/.trash") && os.IsNotExist(err) {
			fs.MkdirAll("/.trash", 0755)
			return []FileEntry{}, nil
		}
		return nil, err
	}

	files := make([]FileEntry, 0, len(entries))
	for _, info := range entries {
		files = append(files, FileEntry{
			Name:    info.Name(),
			Size:    info.Size(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
		})
	}
	return files, nil
}

func ListFilesWithHashes(serverID, subPath string) ([]FileEntry, error) {
	fs := GetVFS(serverID)
	target := filepath.Clean("/" + subPath)

	entries, err := afero.ReadDir(fs, target)
	if err != nil {
		return nil, err
	}

	files := make([]FileEntry, 0, len(entries))
	for _, info := range entries {
		entry := FileEntry{
			Name:    info.Name(),
			Size:    info.Size(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
		}

		if !info.IsDir() && info.Size() < 500*1024*1024 {
			filePath := filepath.Join(target, info.Name())
			if hash, err := computeSHA512(fs, filePath); err == nil {
				entry.SHA512 = hash
			}
		}

		files = append(files, entry)
	}
	return files, nil
}

func computeSHA512(fs afero.Fs, filePath string) (string, error) {
	f, err := fs.Open(filePath)
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
	fs := GetVFS(serverID)
	target := filepath.Clean("/" + subPath)

	info, err := fs.Stat(target)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("cannot read directory")
	}
	if info.Size() > 5*1024*1024 {
		return nil, fmt.Errorf("file too large")
	}

	return afero.ReadFile(fs, target)
}

func CreateFolder(serverID, subPath string) error {
	fs := GetVFS(serverID)
	target := filepath.Clean("/" + subPath)
	return fs.MkdirAll(target, 0755)
}

func WriteFile(serverID, subPath string, content []byte) error {
	fs := GetVFS(serverID)
	target := filepath.Clean("/" + subPath)
	return afero.WriteFile(fs, target, content, 0644)
}

func WriteFileStream(serverID, subPath string, r io.Reader) error {
	fs := GetVFS(serverID)
	target := filepath.Clean("/" + subPath)

	f, err := fs.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func SearchFiles(serverID, query string) ([]SearchResult, error) {
	fs := GetVFS(serverID)
	query = strings.ToLower(query)
	var results []SearchResult

	afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.Contains(strings.ToLower(info.Name()), query) {
			relPath := path
			if relPath == "" || relPath == "." {
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
	fs := GetVFS(serverID)
	target := filepath.Clean("/" + subPath)
	if target == "/" || target == "." {
		return fmt.Errorf("cannot delete root workspace")
	}

	if strings.HasPrefix(target, "/.trash") {
		return fs.RemoveAll(target)
	}

	trashDir := "/.trash"
	fs.MkdirAll(trashDir, 0755)

	baseName := filepath.Base(target)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	dirHex := hex.EncodeToString([]byte(filepath.Dir(target)))
	trashTarget := filepath.Join(trashDir, fmt.Sprintf("%s__%s__%s", baseName, timestamp, dirHex))

	err := fs.Rename(target, trashTarget)
	if err != nil {
		return fs.RemoveAll(target)
	}
	return nil
}

func MovePath(serverID, srcPath, destPath string) error {
	fs := GetVFS(serverID)
	src := filepath.Clean("/" + srcPath)
	dest := filepath.Clean("/" + destPath)

	if src == "/" || src == "." {
		return fmt.Errorf("cannot move root workspace")
	}

	fs.MkdirAll(filepath.Dir(dest), 0755)
	return fs.Rename(src, dest)
}

func CopyPath(serverID, srcPath, destPath string) error {
	fs := GetVFS(serverID)
	src := filepath.Clean("/" + srcPath)
	dest := filepath.Clean("/" + destPath)

	if src == "/" || src == "." {
		return fmt.Errorf("cannot copy root workspace")
	}

	info, err := fs.Stat(src)
	if err != nil {
		return err
	}

	fs.MkdirAll(filepath.Dir(dest), 0755)

	if info.IsDir() {
		return copyDirVFS(fs, src, dest)
	}
	return copyFileVFS(fs, src, dest)
}

func GetFilePath(serverID, subPath string) (string, error) {
	return GetRealPath(serverID, subPath)
}

func BulkDelete(serverID string, paths []string) (int, error) {
	fs := GetVFS(serverID)
	deleted := 0

	trashDir := "/.trash"
	fs.MkdirAll(trashDir, 0755)

	for _, p := range paths {
		target := filepath.Clean("/" + p)
		if target == "/" || target == "." {
			continue
		}

		if strings.HasPrefix(target, "/.trash") {
			if err := fs.RemoveAll(target); err == nil {
				deleted++
			}
			continue
		}

		baseName := filepath.Base(target)
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		dirHex := hex.EncodeToString([]byte(filepath.Dir(target)))
		trashTarget := filepath.Join(trashDir, fmt.Sprintf("%s__%s__%s", baseName, timestamp, dirHex))

		if err := fs.Rename(target, trashTarget); err == nil {
			deleted++
		} else {
			if err := fs.RemoveAll(target); err == nil {
				deleted++
			}
		}
	}
	return deleted, nil
}

func BulkCopy(serverID string, paths []string, destDir string) (int, error) {
	fs := GetVFS(serverID)
	dest := filepath.Clean("/" + destDir)
	fs.MkdirAll(dest, 0755)

	copied := 0
	for _, p := range paths {
		src := filepath.Clean("/" + p)
		if src == "/" || src == "." {
			continue
		}
		name := filepath.Base(src)
		target := filepath.Join(dest, name)

		info, err := fs.Stat(src)
		if err != nil {
			continue
		}
		if info.IsDir() {
			if copyDirVFS(fs, src, target) == nil {
				copied++
			}
		} else {
			if copyFileVFS(fs, src, target) == nil {
				copied++
			}
		}
	}
	return copied, nil
}

func copyFileVFS(fs afero.Fs, src, dest string) error {
	in, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := fs.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDirVFS(fs afero.Fs, src, dest string) error {
	fs.MkdirAll(dest, 0755)

	entries, err := afero.ReadDir(fs, src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		destPath := filepath.Join(dest, e.Name())
		if e.IsDir() {
			if err := copyDirVFS(fs, srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFileVFS(fs, srcPath, destPath); err != nil {
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

	fs := GetVFS(serverID)
	target := filepath.Clean("/" + destPath)
	fs.MkdirAll(filepath.Dir(target), 0755)

	resp, err := downloadClient.Get(rawURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	f, err := fs.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
