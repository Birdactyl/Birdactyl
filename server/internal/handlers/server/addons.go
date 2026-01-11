package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

var addonHTTPClient = &http.Client{Timeout: 30 * time.Second}

func getSourceHeaders(source *models.AddonSource) map[string]string {
	headers := make(map[string]string)

	for k, v := range source.Headers {
		headers[k] = v
	}

	if source.APIKey == "" {
		return headers
	}

	cfg := config.Get()
	if cfg == nil || cfg.APIKeys == nil {
		return headers
	}

	apiKeyCfg, ok := cfg.APIKeys[source.APIKey]
	if !ok || apiKeyCfg.Key == "" {
		return headers
	}

	for k, v := range apiKeyCfg.Headers {
		headers[k] = strings.ReplaceAll(v, "{{key}}", apiKeyCfg.Key)
	}

	return headers
}

func GetAddonSources(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}

	if server.Package == nil {
		return c.JSON(fiber.Map{"success": true, "data": []interface{}{}})
	}

	var sources []models.AddonSource
	json.Unmarshal(server.Package.AddonSources, &sources)

	safeSources := make([]map[string]interface{}, len(sources))
	for i, s := range sources {
		safeSources[i] = map[string]interface{}{
			"id":           s.ID,
			"name":         s.Name,
			"icon":         s.Icon,
			"type":         s.Type,
			"install_path": s.InstallPath,
		}
	}

	return c.JSON(fiber.Map{"success": true, "data": safeSources})
}

func SearchAddons(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}

	sourceID := c.Query("source")
	query := c.Query("q")

	if server.Package == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No package configured"})
	}

	var sources []models.AddonSource
	json.Unmarshal(server.Package.AddonSources, &sources)

	var source *models.AddonSource
	for i := range sources {
		if sources[i].ID == sourceID {
			source = &sources[i]
			break
		}
	}

	if source == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid addon source"})
	}

	vars := getServerVariables(server)
	vars["query"] = query
	vars["limit"] = c.Query("limit", "20")
	vars["offset"] = c.Query("offset", "0")

	searchURL := interpolateURL(source.SearchURL, vars)

	req, _ := http.NewRequest("GET", searchURL, nil)
	headers := getSourceHeaders(source)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "Birdactyl/1.0")

	if source.APIKey != "" {
		cfg := config.Get()
		if cfg == nil || cfg.APIKeys == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "API key '" + source.APIKey + "' not configured"})
		}
		if _, ok := cfg.APIKeys[source.APIKey]; !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "API key '" + source.APIKey + "' not configured"})
		}
	}

	resp, err := addonHTTPClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "Failed to fetch addons: " + err.Error()})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "Addon API returned " + resp.Status, "body": string(body)})
	}

	results := mapAddonResults(string(body), source.Mapping)

	return c.JSON(fiber.Map{"success": true, "data": results})
}

func GetAddonVersions(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}

	sourceID := c.Query("source")
	addonID := c.Query("addon")

	if server.Package == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No package configured"})
	}

	var sources []models.AddonSource
	json.Unmarshal(server.Package.AddonSources, &sources)

	var source *models.AddonSource
	for i := range sources {
		if sources[i].ID == sourceID {
			source = &sources[i]
			break
		}
	}

	if source == nil || source.VersionsURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid addon source"})
	}

	vars := getServerVariables(server)
	vars["id"] = addonID

	versionsURL := interpolateURL(source.VersionsURL, vars)

	req, _ := http.NewRequest("GET", versionsURL, nil)
	for k, v := range getSourceHeaders(source) {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "Birdactyl/1.0")

	resp, err := addonHTTPClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "Failed to fetch versions"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	versions := mapAddonVersions(string(body), source.Mapping)

	return c.JSON(fiber.Map{"success": true, "data": versions})
}

func InstallAddon(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileWrite)
	if err != nil {
		return nil
	}

	var req struct {
		SourceID    string `json:"source_id"`
		DownloadURL string `json:"download_url"`
		FileName    string `json:"file_name"`
		ModID       string `json:"mod_id"`
		FileID      string `json:"file_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if server.Package == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No package configured"})
	}

	var sources []models.AddonSource
	json.Unmarshal(server.Package.AddonSources, &sources)

	var source *models.AddonSource
	for i := range sources {
		if sources[i].ID == req.SourceID {
			source = &sources[i]
			break
		}
	}

	if source == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid addon source"})
	}

	downloadURL := req.DownloadURL
	
	if downloadURL == "" && source.APIKey == "curseforge" && req.ModID != "" && req.FileID != "" {
		cfURL := fmt.Sprintf("https://api.curseforge.com/v1/mods/%s/files/%s", req.ModID, req.FileID)
		cfReq, _ := http.NewRequest("GET", cfURL, nil)
		headers := getSourceHeaders(source)
		for k, v := range headers {
			cfReq.Header.Set(k, v)
		}
		cfReq.Header.Set("User-Agent", "Birdactyl/1.0")
		
		cfg := config.Get()
		if cfg == nil || cfg.APIKeys == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "API key 'curseforge' not configured", "source_id": source.ID})
		}
		if _, ok := cfg.APIKeys["curseforge"]; !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "API key 'curseforge' not configured", "source_id": source.ID})
		}
		
		cfResp, err := addonHTTPClient.Do(cfReq)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "Failed to fetch file info: " + err.Error()})
		}
		defer cfResp.Body.Close()
		
		body, _ := io.ReadAll(cfResp.Body)
		
		if cfResp.StatusCode != http.StatusOK {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "CurseForge returned " + cfResp.Status, "url": cfURL, "response": string(body)})
		}
		
		downloadURL = gjson.Get(string(body), "data.downloadUrl").String()
	}

	if downloadURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No download URL available"})
	}

	fileName := sanitizeFileName(req.FileName)
	if fileName == "" {
		fileName = "addon_" + uuid.New().String()[:8]
	}

	if source.Type != "" && plugins.GetAddonTypeRegistry().Has(source.Type) {
		serverVars := getServerVariables(server)
		sourceInfo := map[string]string{
			"id":           source.ID,
			"name":         source.Name,
			"install_path": source.InstallPath,
			"api_key":      source.APIKey,
		}

		addonReq := plugins.AddonTypeRequest{
			TypeID:          source.Type,
			ServerID:        server.ID.String(),
			NodeID:          server.NodeID.String(),
			DownloadURL:     downloadURL,
			FileName:        fileName,
			InstallPath:     source.InstallPath,
			SourceInfo:      sourceInfo,
			ServerVariables: serverVars,
		}

		resp, err := plugins.DispatchAddonType(addonReq)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Plugin error: " + err.Error()})
		}
		if resp != nil {
			if !resp.Success {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": resp.Error})
			}
			return executeAddonActions(c, server, resp.Actions, resp.Message)
		}
	}

	installPath := filepath.Join(source.InstallPath, fileName)

	resp, err := services.ProxyToNode(server, "POST", "/api/servers/"+server.ID.String()+"/files/download-url", fiber.Map{
		"url":  downloadURL,
		"path": installPath,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(resp.StatusCode).Send(resp.Body)
}

func ListInstalledAddons(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}

	sourceID := c.Query("source")

	if server.Package == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No package configured"})
	}

	var sources []models.AddonSource
	json.Unmarshal(server.Package.AddonSources, &sources)

	var source *models.AddonSource
	for i := range sources {
		if sources[i].ID == sourceID {
			source = &sources[i]
			break
		}
	}

	if source == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid addon source"})
	}

	resp, err := services.ProxyToNode(server, "GET", "/api/servers/"+server.ID.String()+"/files/hashes?path="+url.QueryEscape(source.InstallPath), nil)
	if err != nil {
		return proxyGetWithQuery(c, server, "/files", "path", source.InstallPath)
	}

	var nodeResp struct {
		Success bool `json:"success"`
		Data    []struct {
			Name    string `json:"name"`
			Size    int64  `json:"size"`
			IsDir   bool   `json:"is_dir"`
			ModTime int64  `json:"mod_time"`
			Mode    string `json:"mode"`
			SHA512  string `json:"sha512"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &nodeResp); err != nil || !nodeResp.Success {
		return proxyGetWithQuery(c, server, "/files", "path", source.InstallPath)
	}

	var hashes []string
	hashToFile := make(map[string]int)
	for i, f := range nodeResp.Data {
		if !f.IsDir && f.SHA512 != "" {
			hashes = append(hashes, f.SHA512)
			hashToFile[f.SHA512] = i
		}
	}

	iconMap := make(map[string]string)
	projectMap := make(map[string]string)
	if len(hashes) > 0 {
		modrinthData := lookupModrinthByHashes(hashes)
		for hash, info := range modrinthData {
			iconMap[hash] = info.Icon
			projectMap[hash] = info.ProjectID
		}
	}

	type EnrichedFile struct {
		Name      string `json:"name"`
		Size      int64  `json:"size"`
		IsDir     bool   `json:"is_dir"`
		ModTime   int64  `json:"mod_time"`
		Mode      string `json:"mode"`
		Icon      string `json:"icon,omitempty"`
		ProjectID string `json:"project_id,omitempty"`
	}

	result := make([]EnrichedFile, len(nodeResp.Data))
	for i, f := range nodeResp.Data {
		result[i] = EnrichedFile{
			Name:      f.Name,
			Size:      f.Size,
			IsDir:     f.IsDir,
			ModTime:   f.ModTime,
			Mode:      f.Mode,
			Icon:      iconMap[f.SHA512],
			ProjectID: projectMap[f.SHA512],
		}
	}

	return c.JSON(fiber.Map{"success": true, "data": result})
}

type modrinthVersionInfo struct {
	Icon      string
	ProjectID string
}

func lookupModrinthByHashes(hashes []string) map[string]modrinthVersionInfo {
	result := make(map[string]modrinthVersionInfo)

	if len(hashes) == 0 {
		return result
	}

	reqBody := struct {
		Hashes    []string `json:"hashes"`
		Algorithm string   `json:"algorithm"`
	}{
		Hashes:    hashes,
		Algorithm: "sha512",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.modrinth.com/v2/version_files", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Birdactyl/1.0")

	resp, err := addonHTTPClient.Do(req)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var versionFiles map[string]struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(body, &versionFiles); err != nil {
		return result
	}

	var projectIDs []string
	projectIDSet := make(map[string]bool)
	hashToProject := make(map[string]string)

	for hash, vf := range versionFiles {
		hashToProject[hash] = vf.ProjectID
		if !projectIDSet[vf.ProjectID] {
			projectIDSet[vf.ProjectID] = true
			projectIDs = append(projectIDs, vf.ProjectID)
		}
	}

	if len(projectIDs) == 0 {
		return result
	}

	idsJSON, _ := json.Marshal(projectIDs)
	projectsURL := "https://api.modrinth.com/v2/projects?ids=" + url.QueryEscape(string(idsJSON))
	req2, _ := http.NewRequest("GET", projectsURL, nil)
	req2.Header.Set("User-Agent", "Birdactyl/1.0")

	resp2, err := addonHTTPClient.Do(req2)
	if err != nil {
		return result
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)

	var projects []struct {
		ID      string `json:"id"`
		IconURL string `json:"icon_url"`
	}
	if err := json.Unmarshal(body2, &projects); err != nil {
		return result
	}

	projectIcons := make(map[string]string)
	for _, p := range projects {
		projectIcons[p.ID] = p.IconURL
	}

	for hash, projectID := range hashToProject {
		result[hash] = modrinthVersionInfo{
			Icon:      projectIcons[projectID],
			ProjectID: projectID,
		}
	}

	return result
}

func DeleteAddon(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileDelete)
	if err != nil {
		return nil
	}

	var req struct {
		SourceID string `json:"source_id"`
		FileName string `json:"file_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if server.Package == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No package configured"})
	}

	var sources []models.AddonSource
	json.Unmarshal(server.Package.AddonSources, &sources)

	var source *models.AddonSource
	for i := range sources {
		if sources[i].ID == req.SourceID {
			source = &sources[i]
			break
		}
	}

	if source == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid addon source"})
	}

	filePath := filepath.Join(source.InstallPath, sanitizeFileName(req.FileName))

	resp, err := services.ProxyToNode(server, "DELETE", "/api/servers/"+server.ID.String()+"/files?path="+url.QueryEscape(filePath), nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(resp.StatusCode).Send(resp.Body)
}

func interpolateURL(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", url.QueryEscape(v))
	}
	return result
}

func getServerVariables(server *models.Server) map[string]string {
	vars := make(map[string]string)

	if server.Package != nil && server.Package.Variables != nil {
		var pkgVars []struct {
			Name    string `json:"name"`
			Default string `json:"default"`
		}
		if json.Unmarshal(server.Package.Variables, &pkgVars) == nil {
			for _, v := range pkgVars {
				vars[v.Name] = v.Default
			}
		}
	}

	if server.Variables != nil {
		var serverVars map[string]string
		if json.Unmarshal(server.Variables, &serverVars) == nil && serverVars != nil {
			for k, v := range serverVars {
				vars[k] = v
			}
		}
	}

	return vars
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(name, "_")
	return name
}

func mapAddonResults(jsonBody string, mapping models.AddonSourceMapping) []map[string]interface{} {
	var results []map[string]interface{}

	resultsPath := mapping.Results
	if resultsPath == "" {
		resultsPath = "@this"
	}

	items := gjson.Get(jsonBody, resultsPath)
	if !items.IsArray() {
		return results
	}

	items.ForEach(func(_, item gjson.Result) bool {
		addon := map[string]interface{}{
			"id":          getJSONValue(item, mapping.ID, "id"),
			"name":        getJSONValue(item, mapping.Name, "name"),
			"description": getJSONValue(item, mapping.Description, "description"),
			"icon":        getJSONValue(item, mapping.Icon, "icon"),
			"author":      getJSONValue(item, mapping.Author, "author"),
			"downloads":   getJSONValue(item, mapping.Downloads, "downloads"),
		}
		results = append(results, addon)
		return true
	})

	return results
}

func mapAddonVersions(jsonBody string, mapping models.AddonSourceMapping) []map[string]interface{} {
	var versions []map[string]interface{}

	parsed := gjson.Parse(jsonBody)
	if !parsed.IsArray() {
		if gjson.Get(jsonBody, "data").IsArray() {
			parsed = gjson.Get(jsonBody, "data")
		} else if gjson.Get(jsonBody, "versions").IsArray() {
			parsed = gjson.Get(jsonBody, "versions")
		}
	}

	if !parsed.IsArray() {
		return versions
	}

	parsed.ForEach(func(_, item gjson.Result) bool {
		version := map[string]interface{}{
			"id":           getJSONValue(item, mapping.VersionID, "id"),
			"name":         getJSONValue(item, mapping.VersionName, "version_number"),
			"download_url": getJSONValue(item, mapping.DownloadURL, "files.0.url"),
			"file_name":    getJSONValue(item, mapping.FileName, "files.0.filename"),
			"mod_id":       item.Get("modId").Value(),
		}
		versions = append(versions, version)
		return true
	})

	return versions
}

func getJSONValue(item gjson.Result, path, fallback string) interface{} {
	if path == "" {
		path = fallback
	}
	val := item.Get(path)
	if !val.Exists() {
		return nil
	}
	return val.Value()
}

func SearchModpacks(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}

	query := c.Query("q")
	loader := c.Query("loader", "fabric")

	vars := getServerVariables(server)
	gameVersion := vars["MC_VERSION"]
	if gameVersion == "" {
		gameVersion = "1.21.4"
	}

	searchURL := "https://api.modrinth.com/v2/search?query=" + url.QueryEscape(query) +
		"&facets=" + url.QueryEscape(`[["project_type:modpack"],["categories:`+loader+`"],["versions:`+gameVersion+`"]]`) +
		"&limit=" + c.Query("limit", "20") +
		"&offset=" + c.Query("offset", "0")

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Birdactyl/1.0")

	resp, err := addonHTTPClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "Failed to fetch modpacks"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	mapping := models.AddonSourceMapping{
		Results:     "hits",
		ID:          "project_id",
		Name:        "title",
		Description: "description",
		Icon:        "icon_url",
		Author:      "author",
		Downloads:   "downloads",
	}

	results := mapAddonResults(string(body), mapping)
	return c.JSON(fiber.Map{"success": true, "data": results})
}

func GetModpackVersions(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}

	modpackID := c.Query("id")
	loader := c.Query("loader", "fabric")

	vars := getServerVariables(server)
	gameVersion := vars["MC_VERSION"]
	if gameVersion == "" {
		gameVersion = "1.21.4"
	}

	versionsURL := "https://api.modrinth.com/v2/project/" + url.PathEscape(modpackID) +
		"/version?loaders=" + url.QueryEscape(`["`+loader+`"]`) +
		"&game_versions=" + url.QueryEscape(`["`+gameVersion+`"]`)

	req, _ := http.NewRequest("GET", versionsURL, nil)
	req.Header.Set("User-Agent", "Birdactyl/1.0")

	resp, err := addonHTTPClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "Failed to fetch versions"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var versions []map[string]interface{}
	parsed := gjson.ParseBytes(body)

	parsed.ForEach(func(_, item gjson.Result) bool {
		var mrpackURL string
		files := item.Get("files")
		files.ForEach(func(_, file gjson.Result) bool {
			if file.Get("filename").String() != "" && strings.HasSuffix(file.Get("filename").String(), ".mrpack") {
				mrpackURL = file.Get("url").String()
				return false
			}
			return true
		})

		versions = append(versions, map[string]interface{}{
			"id":           item.Get("id").String(),
			"name":         item.Get("version_number").String(),
			"download_url": mrpackURL,
		})
		return true
	})

	return c.JSON(fiber.Map{"success": true, "data": versions})
}

func InstallModpack(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileWrite)
	if err != nil {
		return nil
	}

	var req struct {
		DownloadURL string `json:"download_url"`
		SourceID    string `json:"source_id"`
		ModID       string `json:"mod_id"`
		FileID      string `json:"file_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if req.DownloadURL == "" && req.SourceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "download_url or source_id required"})
	}

	payload := fiber.Map{}

	if strings.HasPrefix(req.SourceID, "curseforge") {
		cfg := config.Get()
		var apiKey string
		if cfg != nil && cfg.APIKeys != nil {
			if keyCfg, ok := cfg.APIKeys["curseforge"]; ok && keyCfg.Key != "" {
				apiKey = keyCfg.Key
			}
		}
		if apiKey == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "API key 'curseforge' not configured"})
		}
		payload["type"] = "curseforge"
		payload["api_key"] = apiKey
		if req.DownloadURL != "" {
			payload["url"] = req.DownloadURL
		} else if req.ModID != "" && req.FileID != "" {
			fileURL := fmt.Sprintf("https://api.curseforge.com/v1/mods/%s/files/%s/download-url", req.ModID, req.FileID)
			httpReq, _ := http.NewRequest("GET", fileURL, nil)
			httpReq.Header.Set("x-api-key", apiKey)
			httpReq.Header.Set("Accept", "application/json")
			resp, err := addonHTTPClient.Do(httpReq)
			if err != nil {
				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "Failed to get download URL"})
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			downloadURL := gjson.Get(string(body), "data").String()
			if downloadURL == "" {
				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "No download URL available"})
			}
			payload["url"] = downloadURL
		}
	} else {
		payload["type"] = "modrinth"
		payload["url"] = req.DownloadURL
	}

	resp, err := services.ProxyToNode(server, "POST", "/api/servers/"+server.ID.String()+"/modpack/install", payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(resp.StatusCode).Send(resp.Body)
}

func executeAddonActions(c *fiber.Ctx, server *models.Server, actions []plugins.AddonInstallAction, message string) error {
	for _, action := range actions {
		var err error
		switch action.Type {
		case 0:
			_, err = services.ProxyToNode(server, "POST", "/api/servers/"+server.ID.String()+"/files/download-url", fiber.Map{
				"url":     action.URL,
				"path":    action.Path,
				"headers": action.Headers,
			})
		case 1:
			_, err = services.ProxyToNode(server, "POST", "/api/servers/"+server.ID.String()+"/files/decompress", fiber.Map{
				"path": action.Path,
			})
		case 2:
			_, err = services.ProxyToNode(server, "DELETE", "/api/servers/"+server.ID.String()+"/files?path="+url.QueryEscape(action.Path), nil)
		case 3:
			_, err = services.ProxyToNode(server, "POST", "/api/servers/"+server.ID.String()+"/files/folder", fiber.Map{
				"path": action.Path,
			})
		case 4:
			_, err = services.ProxyToNode(server, "POST", "/api/servers/"+server.ID.String()+"/files/write", fiber.Map{
				"path":    action.Path,
				"content": string(action.Content),
			})
		case 6:
			if action.NodeEndpoint != "" {
				var payload interface{}
				if len(action.NodePayload) > 0 {
					json.Unmarshal(action.NodePayload, &payload)
				}
				_, err = services.ProxyToNode(server, "POST", action.NodeEndpoint, payload)
			}
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Action failed: " + err.Error()})
		}
	}

	return c.JSON(fiber.Map{"success": true, "message": message})
}
