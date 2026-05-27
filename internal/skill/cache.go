package skill

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type CacheManager struct {
	projectRoot string
	cachePath   string
	ttl         time.Duration
}

func NewCacheManager(projectRoot string) *CacheManager {
	return &CacheManager{
		projectRoot: projectRoot,
		cachePath:   filepath.Join(projectRoot, ".team", SkillsCacheFile),
		ttl:         DefaultCacheTTL,
	}
}

func NewCacheManagerWithTTL(projectRoot string, ttl time.Duration) *CacheManager {
	return &CacheManager{
		projectRoot: projectRoot,
		cachePath:   filepath.Join(projectRoot, ".team", SkillsCacheFile),
		ttl:         ttl,
	}
}

func (cm *CacheManager) Get() (*SkillsFile, error) {
	if _, err := os.Stat(cm.cachePath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(cm.cachePath)
	if err != nil {
		return nil, err
	}

	var skillsFile = &SkillsFile{}
	if err := yaml.Unmarshal(data, skillsFile); err != nil {
		return nil, err
	}

	return skillsFile, nil
}

func (cm *CacheManager) Set(skillsFile *SkillsFile) error {
	teamDir := filepath.Dir(cm.cachePath)
	if _, err := os.Stat(teamDir); os.IsNotExist(err) {
		if err := os.MkdirAll(teamDir, 0755); err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(skillsFile)
	if err != nil {
		return err
	}

	return os.WriteFile(cm.cachePath, data, 0644)
}

func (cm *CacheManager) IsFresh() (bool, error) {
	info, err := os.Stat(cm.cachePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if time.Since(info.ModTime()) > cm.ttl {
		return false, nil
	}

	return true, nil
}

func (cm *CacheManager) Invalidate() error {
	if _, err := os.Stat(cm.cachePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(cm.cachePath)
}

func (cm *CacheManager) GetCacheMetadata() (*CacheMetadata, error) {
	skillsFile, err := cm.Get()
	if err != nil {
		return nil, err
	}
	return &skillsFile.CacheMetadata, nil
}

func (cm *CacheManager) GetCachePath() string {
	return cm.cachePath
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
