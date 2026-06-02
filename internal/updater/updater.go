package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubAPIReleasesAPI = "https://api.github.com/repos/origadmin/team-flow/releases/latest"
	githubRepo             = "origadmin/team-flow"
	defaultCheckInterval  = 24 * time.Hour
	cacheFileName          = ".team/.flow_update_cache.json"
	httpTimeout            = 2 * time.Second // 更短的超时时间
)

type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

type UpdateCheck struct {
	CurrentVersion    string
	LatestVersion    string
	HasUpdate        bool
	DownloadURL      string
	UpdateMessage    string
}

// UpdateCache 本地缓存结构
type UpdateCache struct {
	LastCheck    time.Time      `json:"last_check"`
	LastResult  *UpdateCheck   `json:"last_result"`
	CheckInterval time.Duration `json:"check_interval"`
	Disabled     bool           `json:"disabled"`
}

// CheckForUpdateWithCache 带缓存的更新检查（高性能版本）
func CheckForUpdateWithCache(currentVersion, projectRoot string, forceCheck bool) (*UpdateCheck, error) {
	cachePath := filepath.Join(projectRoot, cacheFileName)

	// 1. 尝试读取缓存
	cache, _ := loadCache(cachePath)

	// 2. 从 project.md 读取配置
	readProjectConfig(projectRoot, cache)

	// 3. 检查是否禁用
	if cache.Disabled {
		return nil, fmt.Errorf("update check disabled")
	}

	// 3. 检查缓存是否有效
	if !forceCheck && cache.LastResult != nil {
		interval := cache.CheckInterval
		if interval == 0 {
			interval = defaultCheckInterval
		}
		if time.Since(cache.LastCheck) < interval {
			return cache.LastResult, nil // 返回缓存结果
		}
	}

	// 4. 执行实际检查（带超时）
	check, err := CheckForUpdate(currentVersion)

	// 5. 即使检查失败也更新缓存时间（避免频繁重试）
	newCache := &UpdateCache{
		LastCheck:    time.Now(),
		LastResult:  check, // 即使失败也保存最新状态
		CheckInterval: cache.CheckInterval,
		Disabled:     cache.Disabled,
	}
	saveCache(cachePath, newCache)

	return check, err
}

// loadCache 从文件加载缓存
func loadCache(path string) (*UpdateCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &UpdateCache{}, err
	}
	var cache UpdateCache
	err = json.Unmarshal(data, &cache)
	return &cache, err
}

// saveCache 保存缓存到文件
func saveCache(path string, cache *UpdateCache) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// readProjectConfig 从 .team/project.md 读取更新配置
func readProjectConfig(projectRoot string, cache *UpdateCache) {
	projectConfigPath := filepath.Join(projectRoot, ".team", "project.md")
	data, err := os.ReadFile(projectConfigPath)
	if err != nil {
		return // 配置文件不存在，使用默认值
	}
	content := string(data)

	// 解析 update_check_disabled
	if strings.Contains(content, "update_check_disabled: true") {
		cache.Disabled = true
	}

	// 解析 update_check_interval
	if idx := strings.Index(content, "update_check_interval:"); idx != -1 {
		rest := content[idx+len("update_check_interval:"):]
		if endIdx := strings.IndexAny(rest, "\n#"); endIdx != -1 {
			intervalStr := strings.TrimSpace(rest[:endIdx])
			if d, err := time.ParseDuration(intervalStr); err == nil {
				cache.CheckInterval = d
			}
		}
	}
}

func CheckForUpdate(currentVersion string) (*UpdateCheck, error) {
	client := &http.Client{
		Timeout: httpTimeout,
	}

	resp, err := client.Get(githubAPIReleasesAPI)
	if err != nil {
		return nil, fmt.Errorf("check update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	check := &UpdateCheck{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      isNewerVersion(release.TagName, currentVersion),
	}

	if check.HasUpdate {
		check.DownloadURL = findMatchingAsset(release.Assets)
		check.UpdateMessage = fmt.Sprintf("New version available: %s → %s",
			currentVersion, release.TagName)
	}

	return check, nil
}

func isNewerVersion(latest, current string) bool {
	if current == "dev" || current == "" {
		return false
	}

	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	for i := 0; i < max(len(latestParts), len(currentParts)); i++ {
		var lp, cp int
		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &lp)
		}
		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &cp)
		}
		if lp > cp {
			return true
		}
		if lp < cp {
			return false
		}
	}
	return false
}

func findMatchingAsset(assets []Asset) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var suffix string
	switch goos {
	case "windows":
		suffix = fmt.Sprintf("_%s_%s.exe", goos, goarch)
	case "darwin", "linux":
		suffix = fmt.Sprintf("_%s_%s", goos, goarch)
	}

	for _, asset := range assets {
		if strings.Contains(asset.Name, suffix) {
			return asset.DownloadURL
		}
	}
	if len(assets) > 0 {
		return assets[0].DownloadURL
	}
	return ""
}

func ShowUpdateCheck(check *UpdateCheck) {
	if !check.HasUpdate {
		fmt.Println("  ✓ flow CLI is up to date")
		return
	}

	fmt.Printf("  ⬆ Update available: %s → %s\n", check.CurrentVersion, check.LatestVersion)
	if check.DownloadURL != "" {
		fmt.Printf("    Download: %s\n", check.DownloadURL)
	}
	fmt.Println("    Run: flow update")
}

func GetCurrentExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exePath)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// is404Error checks if the error is a 404 Not Found from GitHub API.
func is404Error(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "404")
}
