package skill

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SkillScanner struct{}

func NewSkillScanner() *SkillScanner {
	return &SkillScanner{}
}

func (ss *SkillScanner) ScanDirectory(dir string, source string) ([]ResolvedSkill, error) {
	var skills []ResolvedSkill

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return skills, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(dir, entry.Name())
		skill, err := ss.ScanSkillDir(skillDir, source)
		if err != nil {
			continue
		}
		if skill != nil {
			skills = append(skills, *skill)
		}
	}

	return skills, nil
}

func (ss *SkillScanner) ScanSkillDir(skillDir string, source string) (*ResolvedSkill, error) {
	skillFile := filepath.Join(skillDir, SkillFileName)
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		return nil, nil
	}

	skillID := filepath.Base(skillDir)

	content, err := os.ReadFile(skillFile)
	if err != nil {
		return nil, err
	}

	trigger := ss.parseTrigger(content)

	return &ResolvedSkill{
		ID:      skillID,
		Source:  source,
		Trigger: trigger,
		Path:    skillDir,
		Enabled: true,
	}, nil
}

func (ss *SkillScanner) parseTrigger(content []byte) string {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "trigger:") || strings.HasPrefix(line, "Trigger:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func (ss *SkillScanner) GetLatestMtime(dir string) (time.Time, error) {
	var latestMtime time.Time

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return latestMtime, nil
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.ModTime().After(latestMtime) {
			latestMtime = info.ModTime()
		}
		return nil
	})

	return latestMtime, err
}

func (ss *SkillScanner) ParseSkillFrontmatter(path string) (map[string]interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	lines := strings.Split(string(content), "\n")
	inFrontmatter := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"'")
				result[key] = value
			}
		}
	}

	return result, nil
}
