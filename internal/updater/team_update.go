package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TeamFrameworkConfig 定义 team-flow 框架更新配置
type TeamFrameworkConfig struct {
	GitRepo       string // GitHub repo, e.g., "origadmin/team-flow"
	Branch        string // Branch, default "main"
	BasePath      string // Base path in repo, default ""
	CheckInterval time.Duration // Check interval, default 24h
}

// TeamFrameworkUpdateCheckResult 框架更新检查结果
type TeamFrameworkUpdateCheckResult struct {
	HasUpdate     bool
	LocalVersion  string
	RemoteVersion string
	Changes       []string
	LastCheck     time.Time
}

// TeamFrameworkUpdater 负责 team-flow 框架更新管理
type TeamFrameworkUpdater struct {
	config TeamFrameworkConfig
	cache  struct {
		lastCheck time.Time
		result    *TeamFrameworkUpdateCheckResult
	}
}

// NewTeamFrameworkUpdater 创建新的框架更新管理器
func NewTeamFrameworkUpdater(config TeamFrameworkConfig) *TeamFrameworkUpdater {
	if config.GitRepo == "" {
		config.GitRepo = "origadmin/team-flow"
	}
	if config.Branch == "" {
		config.Branch = "main"
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 24 * time.Hour
	}
	return &TeamFrameworkUpdater{config: config}
}

// CheckForTeamUpdate 检查 team-flow 框架是否有更新
func (u *TeamFrameworkUpdater) CheckForTeamUpdate(projectRoot string) (*TeamFrameworkUpdateCheckResult, error) {
	result := &TeamFrameworkUpdateCheckResult{
		LastCheck: time.Now(),
	}

	// Read local version from .team/version
	localVersionPath := filepath.Join(projectRoot, ".team", "version")
	if data, err := os.ReadFile(localVersionPath); err == nil {
		result.LocalVersion = strings.TrimSpace(string(data))
	} else {
		result.LocalVersion = "unknown"
	}

	// Try to get latest version from GitHub (simplified)
	// This is a placeholder - real implementation would check git tags or commits
	// For now, we'll just return a simulated result
	result.RemoteVersion = result.LocalVersion // Assume same for now
	result.HasUpdate = false

	u.cache.result = result
	u.cache.lastCheck = time.Now()

	return result, nil
}

// ShowTeamUpdateInfo 显示团队框架更新信息
func ShowTeamUpdateInfo(result *TeamFrameworkUpdateCheckResult) {
	if !result.HasUpdate {
		fmt.Println("  ✓ team-flow framework is up to date")
		return
	}

	fmt.Printf("  ⬆ team-flow framework update available: %s → %s\n", result.LocalVersion, result.RemoteVersion)
	if len(result.Changes) > 0 {
		fmt.Println("  Changes:")
		for _, change := range result.Changes {
			fmt.Printf("    - %s\n", change)
		}
	}
	fmt.Println("  To update, check the project documentation")
}

// UpdateTeamFiles 更新项目中的 team-flow 相关文件
func (u *TeamFrameworkUpdater) UpdateTeamFiles(projectRoot string) error {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Download the latest files from GitHub
	// 2. Backup existing files
	// 3. Replace/update files

	fmt.Println("  team-flow framework update not implemented yet")
	fmt.Println("  Please check the project repository for updates")
	return nil
}

// CheckForUpdatesInVerCommand 在 ver 命令中同时检查 CLI 和框架更新
func CheckForUpdatesInVerCommand(currentCLIVersion string, projectRoot string, forceCheck bool) {
	// Check CLI update
	var updateCheck *UpdateCheck
	var err error
	if forceCheck {
		updateCheck, err = CheckForUpdate(currentCLIVersion)
	} else {
		updateCheck, err = CheckForUpdateWithCache(currentCLIVersion, projectRoot, forceCheck)
	}
	if err == nil {
		ShowUpdateCheck(updateCheck)
	} else {
		fmt.Printf("  ⚠ CLI update check failed: %v\n", err)
	}

	fmt.Println()

	// Check team framework update
	updater := NewTeamFrameworkUpdater(TeamFrameworkConfig{})
	teamResult, err := updater.CheckForTeamUpdate(projectRoot)
	if err == nil {
		ShowTeamUpdateInfo(teamResult)
	} else {
		fmt.Printf("  ⚠ team-flow framework update check failed: %v\n", err)
	}
}

// DownloadFile 下载文件（helper 函数）
func DownloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
