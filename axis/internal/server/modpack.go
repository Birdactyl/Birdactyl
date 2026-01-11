package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ModpackIndex struct {
	FormatVersion int                    `json:"formatVersion"`
	Game          string                 `json:"game"`
	VersionID     string                 `json:"versionId"`
	Name          string                 `json:"name"`
	Summary       string                 `json:"summary"`
	Files         []ModpackFile          `json:"files"`
	Dependencies  map[string]string      `json:"dependencies"`
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
	Name          string   `json:"name"`
	FilesInstalled int     `json:"files_installed"`
	FilesFailed    int     `json:"files_failed"`
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

	base := serverDataDir(serverID)
	tempDir := filepath.Join(base, ".modpack_temp")
	packPath := filepath.Join(tempDir, "modpack.zip")

	os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := downloadFile(req.URL, packPath); err != nil {
		return nil, fmt.Errorf("failed to download modpack: %w", err)
	}

	if req.Type == "curseforge" {
		return installCurseForgeModpack(packPath, tempDir, base, req.APIKey)
	}

	return installModrinthModpack(packPath, tempDir, base)
}

func installModrinthModpack(packPath, tempDir, base string) (*ModpackInstallResult, error) {
	index, err := extractAndParseModrinth(packPath, tempDir)
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

	for _, f := range serverFiles {
		wg.Add(1)
		go func(file ModpackFile) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			destPath := filepath.Join(base, filepath.Clean("/"+file.Path))
			if !strings.HasPrefix(destPath, base) {
				mu.Lock()
				result.FilesFailed++
				result.FailedFiles = append(result.FailedFiles, file.Path)
				mu.Unlock()
				return
			}

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				mu.Lock()
				result.FilesFailed++
				result.FailedFiles = append(result.FailedFiles, file.Path)
				mu.Unlock()
				return
			}

			var downloaded bool
			for _, url := range file.Downloads {
				if isBlockedURL(url) == nil {
					if err := downloadFile(url, destPath); err == nil {
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
	if _, err := os.Stat(overridesDir); err == nil {
		copyDir(overridesDir, base)
	}

	serverOverridesDir := filepath.Join(tempDir, "server-overrides")
	if _, err := os.Stat(serverOverridesDir); err == nil {
		copyDir(serverOverridesDir, base)
	}

	return result, nil
}

func installCurseForgeModpack(packPath, tempDir, base, apiKey string) (*ModpackInstallResult, error) {
	manifest, err := extractAndParseCurseForge(packPath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modpack: %w", err)
	}

	result := &ModpackInstallResult{Name: manifest.Name}

	modsDir := filepath.Join(base, "mods")
	if err := os.MkdirAll(modsDir, 0755); err != nil {
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
			if err := downloadFile(downloadURL, destPath); err != nil {
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
	if _, err := os.Stat(overridesDir); err == nil {
		copyDir(overridesDir, base)
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

func downloadFile(url, destPath string) error {
	resp, err := downloadClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractAndParseModrinth(packPath, tempDir string) (*ModpackIndex, error) {
	r, err := zip.OpenReader(packPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var index *ModpackIndex

	for _, f := range r.File {
		destPath := filepath.Join(tempDir, f.Name)

		if !strings.HasPrefix(destPath, tempDir) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			continue
		}

		if f.Name == "modrinth.index.json" {
			rc, err := f.Open()
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

		if strings.HasPrefix(f.Name, "overrides/") || strings.HasPrefix(f.Name, "server-overrides/") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			outFile, err := os.Create(destPath)
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

func extractAndParseCurseForge(packPath, tempDir string) (*CurseForgeManifest, error) {
	r, err := zip.OpenReader(packPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var manifest *CurseForgeManifest

	for _, f := range r.File {
		destPath := filepath.Join(tempDir, f.Name)

		if !strings.HasPrefix(destPath, tempDir) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			continue
		}

		if f.Name == "manifest.json" {
			rc, err := f.Open()
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

		if strings.HasPrefix(f.Name, "overrides/") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			outFile, err := os.Create(destPath)
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
