package skill

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadSkillsFile(path string) (*SkillsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read skills file: %w", err)
	}
	var sf SkillsFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parse skills file: %w", err)
	}
	return &sf, nil
}

func SaveSkillsFile(path string, sf *SkillsFile) error {
	data, err := yaml.Marshal(sf)
	if err != nil {
		return fmt.Errorf("marshal skills file: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func NewSkillsFile(teamID string, tags []string, skills []ResolvedSkill) *SkillsFile {
	return &SkillsFile{
		Version:    1,
		ResolvedAt: time.Now().Format(time.RFC3339),
		Team:       teamID,
		SkillTags:  tags,
		Skills:     skills,
	}
}
