package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

type ModpackIndex struct {
	FormatVersion int               `json:"formatVersion"`
	Game          string            `json:"game"`
	VersionID     string            `json:"versionId"`
	Name          string            `json:"name"`
	Summary       string            `json:"summary"`
	Files         []ModpackFile     `json:"files"`
	Dependencies  map[string]string `json:"dependencies"`
}

type ModpackFile struct {
	Path      string            `json:"path"`
	Hashes    map[string]string `json:"hashes"`
	Env       *ModpackEnv       `json:"env"`
	Downloads []string          `json:"downloads"`
	FileSize  int64             `json:"fileSize"`
}

type ModpackEnv struct {
	Client string `json:"client"`
	Server string `json:"server"`
}

type CurseForgeManifest struct {
	Minecraft       CFMinecraft `json:"minecraft"`
	ManifestType    string      `json:"manifestType"`
	ManifestVersion int         `json:"manifestVersion"`
	Name            string      `json:"name"`
	Version         string      `json:"version"`
	Author          string      `json:"author"`
	Files           []CFFile    `json:"files"`
	Overrides       string      `json:"overrides"`
}

type CFMinecraft struct {
	Version    string     `json:"version"`
	ModLoaders []CFLoader `json:"modLoaders"`
}

type CFLoader struct {
	ID      string `json:"id"`
	Primary bool   `json:"primary"`
}

type CFFile struct {
	ProjectID int  `json:"projectID"`
	FileID    int  `json:"fileID"`
	Required  bool `json:"required"`
}

type ModpackInstallResult struct {
	Name           string   `json:"name"`
	FilesInstalled int      `json:"files_installed"`
	FilesFailed    int      `json:"files_failed"`
	FailedFiles    []string `json:"failed_files,omitempty"`
}

type ModpackInstallRequest struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	APIKey string `json:"api_key"`
}

func InstallModpack(serverID string, req ModpackInstallRequest) (*ModpackInstallResult, error) {
	if err := isBlockedURL(req.URL); err != nil {
		return nil, err
	}

	fs := GetVFS(serverID)
	tempDir := "/.modpack_temp"
	packPath := filepath.Join(tempDir, "modpack.zip")

	fs.RemoveAll(tempDir)
	if err := fs.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer fs.RemoveAll(tempDir)

	if err := DownloadURL(serverID, req.URL, packPath); err != nil {
		return nil, fmt.Errorf("failed to download modpack: %w", err)
	}

	if req.Type == "curseforge" {
		return installCurseForgeModpack(serverID, packPath, tempDir, req.APIKey)
	}

	return installModrinthModpack(serverID, packPath, tempDir)
}

func installModrinthModpack(serverID, packPath, tempDir string) (*ModpackInstallResult, error) {
	index, err := extractAndParseModrinth(serverID, packPath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modpack: %w", err)
	}

	result := &ModpackInstallResult{Name: index.Name}
	var serverFiles []ModpackFile
	for _, f := range index.Files {
		if f.Env != nil && f.Env.Server == "unsupported" {
			continue
		}
		serverFiles = append(serverFiles, f)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 5)

	vfs := GetVFS(serverID)

	for _, f := range serverFiles {
		wg.Add(1)
		go func(file ModpackFile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			destPath := filepath.Clean("/" + file.Path)

			if err := vfs.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				mu.Lock()
				result.FilesFailed++
				result.FailedFiles = append(result.FailedFiles, file.Path)
				mu.Unlock()
				return
			}

			var downloaded bool
			for _, dUrl := range file.Downloads {
				if isBlockedURL(dUrl) == nil {
					if err := DownloadURL(serverID, dUrl, destPath); err == nil {
						downloaded = true
						break
					}
				}
			}

			mu.Lock()
			if downloaded {
				result.FilesInstalled++
			} else {
				result.FilesFailed++
				result.FailedFiles = append(result.FailedFiles, file.Path)
			}
			mu.Unlock()
		}(f)
	}

	wg.Wait()

	overridesDir := filepath.Join(tempDir, "overrides")
	if info, err := vfs.Stat(overridesDir); err == nil && info.IsDir() {
		copyDirVFS(vfs, overridesDir, "/")
	}

	serverOverridesDir := filepath.Join(tempDir, "server-overrides")
	if info, err := vfs.Stat(serverOverridesDir); err == nil && info.IsDir() {
		copyDirVFS(vfs, serverOverridesDir, "/")
	}

	return result, nil
}

func installCurseForgeModpack(serverID, packPath, tempDir, apiKey string) (*ModpackInstallResult, error) {
	manifest, err := extractAndParseCurseForge(serverID, packPath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modpack: %w", err)
	}

	result := &ModpackInstallResult{Name: manifest.Name}
	vfs := GetVFS(serverID)
	modsDir := "/mods"

	if err := vfs.MkdirAll(modsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create mods dir: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	semaphore := make(chan struct{}, 5)

	for _, f := range manifest.Files {
		wg.Add(1)
		go func(file CFFile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			downloadURL, fileName, err := getCurseForgeDownloadURL(file.ProjectID, file.FileID, apiKey)
			if err != nil {
				mu.Lock()
				result.FilesFailed++
				result.FailedFiles = append(result.FailedFiles, fmt.Sprintf("project:%d/file:%d", file.ProjectID, file.FileID))
				mu.Unlock()
				return
			}

			destPath := filepath.Join(modsDir, fileName)
			if err := DownloadURL(serverID, downloadURL, destPath); err != nil {
				mu.Lock()
				result.FilesFailed++
				result.FailedFiles = append(result.FailedFiles, fileName)
				mu.Unlock()
				return
			}

			mu.Lock()
			result.FilesInstalled++
			mu.Unlock()
		}(f)
	}

	wg.Wait()

	overridesName := manifest.Overrides
	if overridesName == "" {
		overridesName = "overrides"
	}
	overridesDir := filepath.Join(tempDir, overridesName)
	if info, err := vfs.Stat(overridesDir); err == nil && info.IsDir() {
		copyDirVFS(vfs, overridesDir, "/")
	}

	return result, nil
}

func getCurseForgeDownloadURL(projectID, fileID int, apiKey string) (string, string, error) {
	url := fmt.Sprintf("https://api.curseforge.com/v1/mods/%d/files/%d", projectID, fileID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := downloadClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			DownloadURL string `json:"downloadUrl"`
			FileName    string `json:"fileName"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", err
	}

	if result.Data.DownloadURL == "" {
		return "", "", fmt.Errorf("download not available (distribution disabled)")
	}

	return result.Data.DownloadURL, result.Data.FileName, nil
}

func extractAndParseModrinth(serverID, packPath, tempDir string) (*ModpackIndex, error) {
	vfs := GetVFS(serverID)
	f, err := vfs.Open(packPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, _ := f.Stat()
	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, err
	}

	var index *ModpackIndex

	for _, zf := range r.File {
		cleanName := filepath.Clean("/" + zf.Name)
		destPath := filepath.Join(tempDir, cleanName)

		if zf.FileInfo().IsDir() {
			vfs.MkdirAll(destPath, 0755)
			continue
		}

		if err := vfs.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			continue
		}

		if cleanName == "/modrinth.index.json" {
			rc, err := zf.Open()
			if err != nil {
				return nil, err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}

			index = &ModpackIndex{}
			if err := json.Unmarshal(data, index); err != nil {
				return nil, err
			}
			continue
		}

		if strings.HasPrefix(cleanName, "/overrides/") || strings.HasPrefix(cleanName, "/server-overrides/") {
			rc, err := zf.Open()
			if err != nil {
				continue
			}
			outFile, err := vfs.Create(destPath)
			if err != nil {
				rc.Close()
				continue
			}
			io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
		}
	}

	if index == nil {
		return nil, fmt.Errorf("modrinth.index.json not found")
	}

	return index, nil
}

func extractAndParseCurseForge(serverID, packPath, tempDir string) (*CurseForgeManifest, error) {
	vfs := GetVFS(serverID)
	f, err := vfs.Open(packPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, _ := f.Stat()
	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, err
	}

	var manifest *CurseForgeManifest

	for _, zf := range r.File {
		cleanName := filepath.Clean("/" + zf.Name)
		destPath := filepath.Join(tempDir, cleanName)

		if zf.FileInfo().IsDir() {
			vfs.MkdirAll(destPath, 0755)
			continue
		}

		if err := vfs.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			continue
		}

		if cleanName == "/manifest.json" {
			rc, err := zf.Open()
			if err != nil {
				return nil, err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}

			manifest = &CurseForgeManifest{}
			if err := json.Unmarshal(data, manifest); err != nil {
				return nil, err
			}
			continue
		}

		if strings.HasPrefix(cleanName, "/overrides/") {
			rc, err := zf.Open()
			if err != nil {
				continue
			}
			outFile, err := vfs.Create(destPath)
			if err != nil {
				rc.Close()
				continue
			}
			io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
		}
	}

	if manifest == nil {
		return nil, fmt.Errorf("manifest.json not found")
	}

	return manifest, nil
}
