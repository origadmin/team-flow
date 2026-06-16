package skill

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ── Plugin Skill 机制 ──────────────────────────────────────────────────────
//
// 一个 skill plugin 就是一个 GitHub repo（owner/repo）。
// 结构是若干子目录，每个子目录含一个 SKILL.md 文件，每个子目录就是一个 skill。
//
// 目录约定：
//   .team/
//     skill-plugins/
//       <owner>-<repo>/           ← 规范化的 plugin 目录名
//         pm-execution/
//           skills/
//             create-prd/SKILL.md
//             brainstorm-okrs/SKILL.md
//             ...
//         pm-product-discovery/
//           skills/
//             opportunity-solution-tree/SKILL.md
//             ...
//         pm-toolkit/
//           skills/
//             draft-nda/SKILL.md
//             ...
//         ...
//         plugin.json             ← 自动写入的元信息
//
// plugin.json:
//   {
//     "name": "phuryn-pm-skills",
//     "source": "github.com/phuryn/pm-skills",
//     "version": "main",
//     "installed_at": "2024-01-01T00:00:00Z"
//   }
//
// 安装流程：
//   1. flow skill install <owner>/<repo>
//   2. 通过 HTTPS 下载 repo 的 tarball（main 分支）到临时目录
//   3. 解压到 .team/skill-plugins/<owner>-<repo>/
//   4. 递归扫描所有含 SKILL.md 的子目录作为可用 skill
//   5. 写入 plugin.json 元信息
//
// 扫描流程：
//   - 在 .team/skill-plugins/<plugin-name>/ 下递归寻找所有含 SKILL.md 的目录
//   - 每个子目录的目录名 = skill ID
//   - 提供两种引用方式：
//       * 完整引用: "<plugin-name>:<skill-dir>"  （如 "phuryn-pm-skills:create-prd"）
//       * 短引用:   "<skill-dir>"                  （如 "create-prd"）

const (
	SkillSourcePlugin = "plugin"
	PluginDirName     = "skill-plugins"
	PluginMetaFile    = "plugin.json"
	PluginSkillPrefix = ":" // skill 引用分隔符："pm-skills:create-prd"
)

// SkillPlugin 表示一个已安装的 skill plugin
type SkillPlugin struct {
	Name        string          `json:"name"`
	Source      string          `json:"source"`           // github.com/phuryn/pm-skills
	Version     string          `json:"version"`          // main / commit hash
	InstalledAt string          `json:"installed_at"`     // ISO 时间
	Skills      []ResolvedSkill `json:"skills,omitempty"` // 扫描到的 skills
	InstallDir  string          `json:"-"`                // 不存盘
}

// PluginManager 管理 skill plugin 的安装/扫描/卸载
type PluginManager struct {
	projectRoot string
	pluginsDir  string
	scanner     *SkillScanner
}

func NewPluginManager(projectRoot string) *PluginManager {
	return &PluginManager{
		projectRoot: projectRoot,
		pluginsDir:  filepath.Join(projectRoot, ".team", PluginDirName),
		scanner:     NewSkillScanner(),
	}
}

// Install 从 GitHub 下载并安装一个 skill plugin
// repo 格式: "owner/repo" 或 "owner/repo@branch"
func (pm *PluginManager) Install(repo string) (*SkillPlugin, error) {
	normalized, branch, err := parseRepoRef(repo)
	if err != nil {
		return nil, err
	}

	pluginDir := filepath.Join(pm.pluginsDir, normalized)

	// 已存在则先删除（允许覆盖更新）
	if _, err := os.Stat(pluginDir); err == nil {
		if err := os.RemoveAll(pluginDir); err != nil {
			return nil, fmt.Errorf("remove existing plugin: %w", err)
		}
	}

	// 准备临时目录
	tmpDir, err := os.MkdirTemp("", "skill-plugin-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 原始 owner/repo（从 normalized 还原）
	originalPath := strings.Replace(normalized, "-", "/", 1)

	// 下载 tarball（先试指定的 branch，失败再试 master）
	var usedBranch string
	var tarballPath string

	tarballPath = filepath.Join(tmpDir, "repo.tar.gz")
	err = downloadTarball(originalPath, branch, tarballPath)
	if err != nil {
		// 备用分支
		altBranch := "master"
		if err2 := downloadTarball(originalPath, altBranch, tarballPath); err2 != nil {
			return nil, fmt.Errorf("download plugin %s@%s: %w (also tried %s: %v)", originalPath, branch, err, altBranch, err2)
		}
		usedBranch = altBranch
	} else {
		usedBranch = branch
	}

	// 解压
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir extracted: %w", err)
	}
	if err := extractTarGz(tarballPath, extractDir); err != nil {
		return nil, fmt.Errorf("extract tarball: %w", err)
	}

	// 找到解压后的 repo 根目录
	repoRoot := findRepoRoot(extractDir)
	if repoRoot == "" {
		return nil, fmt.Errorf("no repo root found in extracted archive")
	}

	// 移动到最终位置
	if err := os.MkdirAll(filepath.Dir(pluginDir), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir plugins dir: %w", err)
	}
	if err := moveDir(repoRoot, pluginDir); err != nil {
		return nil, fmt.Errorf("move plugin dir: %w", err)
	}

	// 扫描 skills
	plugin := &SkillPlugin{
		Name:        normalized,
		Source:      "github.com/" + originalPath,
		Version:     usedBranch,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
		InstallDir:  pluginDir,
	}

	skills, err := pm.scanPluginSkills(plugin)
	if err != nil {
		return nil, fmt.Errorf("scan plugin skills: %w", err)
	}
	plugin.Skills = skills

	// 写入 plugin.json 元信息
	meta := map[string]interface{}{
		"name":         plugin.Name,
		"source":       plugin.Source,
		"version":      plugin.Version,
		"installed_at": plugin.InstalledAt,
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(pluginDir, PluginMetaFile), metaBytes, 0o644); err != nil {
		return nil, fmt.Errorf("write plugin.json: %w", err)
	}

	return plugin, nil
}

// List 列出所有已安装的 plugin
func (pm *PluginManager) List() ([]*SkillPlugin, error) {
	if _, err := os.Stat(pm.pluginsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(pm.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("read plugins dir: %w", err)
	}

	var plugins []*SkillPlugin
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// 跳过隐藏目录
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		plugin, err := pm.loadPlugin(entry.Name())
		if err != nil {
			continue
		}
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}

// Uninstall 卸载一个 plugin
func (pm *PluginManager) Uninstall(name string) error {
	pluginDir := filepath.Join(pm.pluginsDir, name)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", name)
	}
	return os.RemoveAll(pluginDir)
}

// ScanAll 扫描所有 plugin 中的 skills，返回完整的 skill 列表
func (pm *PluginManager) ScanAll() ([]ResolvedSkill, error) {
	plugins, err := pm.List()
	if err != nil {
		return nil, err
	}

	var allSkills []ResolvedSkill
	for _, plugin := range plugins {
		skills, err := pm.scanPluginSkills(plugin)
		if err != nil {
			continue
		}
		allSkills = append(allSkills, skills...)
	}
	return allSkills, nil
}

// ── 内部方法 ───────────────────────────────────────────────────────────────

func (pm *PluginManager) loadPlugin(dirName string) (*SkillPlugin, error) {
	pluginDir := filepath.Join(pm.pluginsDir, dirName)

	plugin := &SkillPlugin{
		Name:       dirName,
		InstallDir: pluginDir,
	}

	// 尝试加载 plugin.json 元信息
	metaPath := filepath.Join(pluginDir, PluginMetaFile)
	if data, err := os.ReadFile(metaPath); err == nil {
		var meta map[string]interface{}
		if json.Unmarshal(data, &meta) == nil {
			if v, ok := meta["source"].(string); ok {
				plugin.Source = v
			}
			if v, ok := meta["version"].(string); ok {
				plugin.Version = v
			}
			if v, ok := meta["installed_at"].(string); ok {
				plugin.InstalledAt = v
			}
		}
	}

	skills, err := pm.scanPluginSkills(plugin)
	if err == nil {
		plugin.Skills = skills
	}
	return plugin, nil
}

// scanPluginSkills 递归扫描单个 plugin 目录下的所有含 SKILL.md 的子目录
func (pm *PluginManager) scanPluginSkills(plugin *SkillPlugin) ([]ResolvedSkill, error) {
	var skills []ResolvedSkill

	err := filepath.Walk(plugin.InstallDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}

		// 跳过隐藏目录和明显非 skill 的目录
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "__pycache__" || name == "node_modules" {
			return filepath.SkipDir
		}

		// 如果当前目录有 SKILL.md
		skillFile := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			skillDir := path
			skillID := filepath.Base(skillDir)

			// 排除顶层 plugin 目录自身（不太可能有 SKILL.md，保险起见）
			if filepath.Clean(skillDir) == filepath.Clean(plugin.InstallDir) {
				return nil
			}

			content, err := os.ReadFile(skillFile)
			if err != nil {
				return nil
			}
			trigger := pm.scanner.parseTrigger(content)

			// 完整引用: plugin-name:skill-id
			qualified := ResolvedSkill{
				ID:      plugin.Name + PluginSkillPrefix + skillID,
				Source:  SkillSourcePlugin,
				Trigger: trigger,
				Path:    skillDir,
				Enabled: true,
			}
			skills = append(skills, qualified)

			// 短引用: skill-id（方便在只装一个 plugin 时直接用短名）
			shortQualified := ResolvedSkill{
				ID:      skillID,
				Source:  SkillSourcePlugin,
				Trigger: trigger,
				Path:    skillDir,
				Enabled: true,
			}
			skills = append(skills, shortQualified)
		}
		return nil
	})

	if err != nil {
		return skills, nil
	}
	return skills, nil
}

// parseRepoRef 解析 "owner/repo" 或 "owner/repo@branch"
// 返回: 规范化目录名（owner-repo）, 分支名, 错误
func parseRepoRef(ref string) (string, string, error) {
	ref = strings.TrimSpace(ref)

	// 去掉可能的协议/域名前缀
	ref = strings.TrimPrefix(ref, "https://")
	ref = strings.TrimPrefix(ref, "http://")
	ref = strings.TrimPrefix(ref, "github.com/")

	// 解析 @branch
	var branch string
	if idx := strings.Index(ref, "@"); idx > 0 {
		branch = ref[idx+1:]
		ref = ref[:idx]
	}

	parts := strings.Split(ref, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid plugin reference: %s (expected owner/repo)", ref)
	}
	owner := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid plugin reference: %s", ref)
	}

	if branch == "" {
		branch = "main"
	}

	normalized := strings.ReplaceAll(owner+"-"+repo, "/", "-")
	return normalized, branch, nil
}

// ── 文件工具函数 ───────────────────────────────────────────────────────────

// downloadTarball 从 GitHub 下载 tarball
func downloadTarball(ownerRepo, branch, destPath string) error {
	url := fmt.Sprintf("https://github.com/%s/archive/refs/heads/%s.tar.gz", ownerRepo, branch)

	// 先用 curl（如果系统有），失败再用 net/http
	if _, err := exec.LookPath("curl"); err == nil {
		cmd := exec.Command("curl", "-fsSL", "--connect-timeout", "10", "--max-time", "120", "-o", destPath, url)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("curl failed: %w: %s", err, string(output))
		}
	} else {
		// 回退到 net/http
		client := &http.Client{Timeout: 120 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("http get: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("http %d", resp.StatusCode)
		}
		f, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create tarball file: %w", err)
		}
		defer f.Close()
		if _, err := io.Copy(f, resp.Body); err != nil {
			return fmt.Errorf("write tarball: %w", err)
		}
	}

	info, err := os.Stat(destPath)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("downloaded file is empty or missing: %s", url)
	}
	return nil
}

// extractTarGz 解压 tar.gz 文件到目标目录
func extractTarGz(srcPath, destDir string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		// 清理路径，防止 zip-slip 攻击
		cleanName := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			continue
		}

		targetPath := filepath.Join(destDir, cleanName)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", targetPath, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return fmt.Errorf("mkdir parent %s: %w", targetPath, err)
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("write file %s: %w", targetPath, err)
			}
			outFile.Close()
			// 保留权限
			if header.Mode != 0 {
				_ = os.Chmod(targetPath, os.FileMode(header.Mode))
			}
		case tar.TypeSymlink:
			// 只允许相对路径的软连接
			linkTarget := filepath.Clean(header.Linkname)
			if !filepath.IsAbs(linkTarget) && !strings.HasPrefix(linkTarget, "..") {
				_ = os.Symlink(header.Linkname, targetPath)
			}
		}
	}

	return nil
}

// findRepoRoot 找到解压后的 repo 根目录
// tar.gz 通常顶层只有一个目录（如 phuryn-pm-skills-main/）
func findRepoRoot(extractDir string) string {
	// 先尝试找第一层子目录（正常情况）
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			return filepath.Join(extractDir, entry.Name())
		}
	}

	// 如果解压后直接就是文件（极少见），返回 extractDir 本身
	return extractDir
}

// moveDir 将 src 目录移动到 dst（跨设备时 fallback 到 copy + delete）
func moveDir(src, dst string) error {
	// 先尝试直接 rename（同设备最快）
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// fallback: 递归复制
	if err := copyDir(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", src)
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}
