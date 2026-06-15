package skill

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/flow"
)

var (
	ErrSkillNotFound   = errors.New("skill not found")
	ErrCacheCorrupted  = errors.New("cache file corrupted")
	ErrNoProjectConfig = errors.New("no project config found")
	ErrNoTeamConfig    = errors.New("no team config found")
)

type SkillManager struct {
	projectRoot   string
	resolver      *SkillResolver
	scanner       *SkillScanner
	cacheMgr      *CacheManager
	pluginManager *PluginManager
}

func NewSkillManager(projectRoot string) *SkillManager {
	return &SkillManager{
		projectRoot:   projectRoot,
		resolver:      NewSkillResolver(),
		scanner:       NewSkillScanner(),
		cacheMgr:      NewCacheManager(projectRoot),
		pluginManager: NewPluginManager(projectRoot),
	}
}

func (sm *SkillManager) CachePath() string {
	return sm.cacheMgr.GetCachePath()
}

func (sm *SkillManager) Resolve(roleID string) (*SkillsFile, error) {
	return sm.ResolveWithOptions(roleID, ResolveOptions{})
}

func (sm *SkillManager) ResolveWithOptions(roleID string, opts ResolveOptions) (*SkillsFile, error) {
	if !opts.ForceRefresh {
		if cached, err := sm.cacheMgr.Get(); err == nil {
			return cached, nil
		}
	}

	return sm.resolveFresh(roleID, opts)
}

func (sm *SkillManager) resolveFresh(roleID string, opts ResolveOptions) (*SkillsFile, error) {
	var allTags []string
	var teamConfig *flow.TeamDefinition
	var projectConfig *config.ProjectConfig

	projectConfig, _ = config.LoadProjectConfig(sm.projectRoot)

	teamConfig, _ = sm.loadTeamConfig()

	if projectConfig != nil {
		allTags = append(allTags, projectConfig.Skills.Tags...)
	}

	if teamConfig != nil {
		allTags = append(allTags, teamConfig.SkillTags...)
		if roleID != "" {
			if role := sm.getRoleByID(teamConfig, roleID); role != nil {
				allTags = append(allTags, role.SkillTags...)
			}
		}
	}

	allTags = append(allTags, opts.ExtraTags...)

	tagSkills := sm.resolver.ResolveFromTags(allTags)

	projectSkillsDir := filepath.Join(sm.projectRoot, ".team", SkillsDirName)
	projectSkills, _ := sm.scanner.ScanDirectory(projectSkillsDir, SkillSourceLocal)

	// 扫描 plugin skills
	pluginSkills, _ := sm.pluginManager.ScanAll()

	merged := sm.resolver.MergeSkills(projectSkills)
	merged = append(merged, tagSkills...)
	merged = append(merged, pluginSkills...)
	merged = sm.resolver.Deduplicate(merged)

	if projectConfig != nil {
		merged = sm.resolver.ApplyDisabled(merged, append(projectConfig.Skills.Disabled, opts.Disabled...))
		merged = sm.resolver.ApplyOverrides(merged, mergeMaps(projectConfig.Skills.Overrides, opts.Overrides))
	}

	teamID := ""
	projectName := ""
	if teamConfig != nil {
		teamID = teamConfig.ID
	}
	if projectConfig != nil {
		projectName = projectConfig.Name
	}

	skillsFile := &SkillsFile{
		Version:    1,
		ResolvedAt: formatTime(time.Now()),
		Team:       teamID,
		Project:    projectName,
		Role:       roleID,
		SkillTags:  allTags,
		Skills:     merged,
	}

	projectSkillsMtime, _ := sm.scanner.GetLatestMtime(projectSkillsDir)
	teamConfigMtime := sm.getFileMtime(sm.getTeamConfigPath())
	projectConfigMtime := sm.getFileMtime(filepath.Join(sm.projectRoot, ".team", "project.yaml"))

	skillsFile.CacheMetadata = CacheMetadata{
		ProjectSkillsMtime: formatTime(projectSkillsMtime),
		TeamConfigMtime:    formatTime(teamConfigMtime),
		ProjectConfigMtime: formatTime(projectConfigMtime),
	}

	if err := sm.cacheMgr.Set(skillsFile); err != nil {
		return skillsFile, nil
	}

	return skillsFile, nil
}

func (sm *SkillManager) GetSkill(skillID string) (*ResolvedSkill, error) {
	skillsFile, err := sm.cacheMgr.Get()
	if err == nil {
		for _, skill := range skillsFile.Skills {
			if skill.ID == skillID {
				return &skill, nil
			}
		}
	}

	projectSkillsDir := filepath.Join(sm.projectRoot, ".team", SkillsDirName)
	if skill, err := sm.scanner.ScanSkillDir(filepath.Join(projectSkillsDir, skillID), SkillSourceLocal); err == nil && skill != nil {
		return skill, nil
	}

	for _, skills := range DefaultSkillTagMap {
		for _, skill := range skills {
			if skill.ID == skillID {
				return &skill, nil
			}
		}
	}

	return nil, ErrSkillNotFound
}

func (sm *SkillManager) ListSkills() ([]ResolvedSkill, error) {
	skillsFile, err := sm.cacheMgr.Get()
	if err == nil {
		return skillsFile.Skills, nil
	}

	skillsFile, err = sm.Resolve("")
	if err != nil {
		return nil, err
	}

	return skillsFile.Skills, nil
}

func (sm *SkillManager) InvalidateCache() error {
	return sm.cacheMgr.Invalidate()
}

func (sm *SkillManager) PluginManager() *PluginManager {
	return sm.pluginManager
}

func (sm *SkillManager) ScanAvailable(source string) ([]ResolvedSkill, error) {
	var allSkills []ResolvedSkill

	if source == "" || source == "all" || source == SkillSourceLocal {
		projectSkillsDir := filepath.Join(sm.projectRoot, ".team", SkillsDirName)
		if skills, err := sm.scanner.ScanDirectory(projectSkillsDir, SkillSourceLocal); err == nil {
			allSkills = append(allSkills, skills...)
		}
	}

	if source == "" || source == "all" || source == SkillSourceTrae {
		seen := make(map[string]bool)
		for _, skills := range DefaultSkillTagMap {
			for _, skill := range skills {
				if !seen[skill.ID] {
					allSkills = append(allSkills, skill)
					seen[skill.ID] = true
				}
			}
		}
	}

	return sm.resolver.Deduplicate(allSkills), nil
}

func (sm *SkillManager) loadTeamConfig() (*flow.TeamDefinition, error) {
	teamConfigPath := sm.getTeamConfigPath()
	if teamConfigPath == "" {
		return nil, ErrNoTeamConfig
	}

	data, err := os.ReadFile(teamConfigPath)
	if err != nil {
		return nil, err
	}

	var teamConfig flow.TeamDefinition
	if err := json.Unmarshal(data, &teamConfig); err != nil {
		return nil, err
	}

	return &teamConfig, nil
}

func (sm *SkillManager) getTeamConfigPath() string {
	candidates := []string{
		filepath.Join(sm.projectRoot, ".team", "team.json"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func (sm *SkillManager) getRoleByID(teamConfig *flow.TeamDefinition, roleID string) *flow.RoleDefinition {
	for _, role := range teamConfig.Roles {
		if role.ID == roleID {
			return &role
		}
	}
	return nil
}

func (sm *SkillManager) getFileMtime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func mergeMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}
