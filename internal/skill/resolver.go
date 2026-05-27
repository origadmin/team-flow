package skill

type SkillResolver struct {
	defaultTagMap map[string][]ResolvedSkill
}

func NewSkillResolver() *SkillResolver {
	return &SkillResolver{
		defaultTagMap: DefaultSkillTagMap,
	}
}

func (sr *SkillResolver) ResolveFromTags(tags []string) []ResolvedSkill {
	var skills []ResolvedSkill
	seen := make(map[string]bool)

	for _, tag := range tags {
		if tagSkills, ok := sr.defaultTagMap[tag]; ok {
			for _, skill := range tagSkills {
				if !seen[skill.ID] {
					skill.Enabled = true
					skills = append(skills, skill)
					seen[skill.ID] = true
				}
			}
		}
	}

	return skills
}

func (sr *SkillResolver) MergeSkills(projectSkills, teamSkills, globalSkills []ResolvedSkill) []ResolvedSkill {
	merged := make(map[string]ResolvedSkill)

	for _, skill := range globalSkills {
		merged[skill.ID] = skill
	}

	for _, skill := range teamSkills {
		merged[skill.ID] = skill
	}

	for _, skill := range projectSkills {
		merged[skill.ID] = skill
	}

	result := make([]ResolvedSkill, 0, len(merged))
	for _, skill := range merged {
		result = append(result, skill)
	}

	return result
}

func (sr *SkillResolver) Deduplicate(skills []ResolvedSkill) []ResolvedSkill {
	seen := make(map[string]bool)
	var result []ResolvedSkill

	for _, skill := range skills {
		if !seen[skill.ID] {
			result = append(result, skill)
			seen[skill.ID] = true
		}
	}

	return result
}

func (sr *SkillResolver) ApplyDisabled(skills []ResolvedSkill, disabled []string) []ResolvedSkill {
	disabledMap := make(map[string]bool)
	for _, id := range disabled {
		disabledMap[id] = true
	}

	var result []ResolvedSkill
	for _, skill := range skills {
		if !disabledMap[skill.ID] {
			result = append(result, skill)
		}
	}

	return result
}

func (sr *SkillResolver) ApplyOverrides(skills []ResolvedSkill, overrides map[string]string) []ResolvedSkill {
	var result []ResolvedSkill

	for _, skill := range skills {
		if newPath, ok := overrides[skill.ID]; ok {
			skill.Path = newPath
		}
		result = append(result, skill)
	}

	return result
}

// ResolveSkills is a backward-compatible function to resolve skills from tags
func ResolveSkills(tags []string) []ResolvedSkill {
	sr := NewSkillResolver()
	return sr.ResolveFromTags(tags)
}

// DeduplicateSkills is a backward-compatible function to deduplicate skills
func DeduplicateSkills(skills []ResolvedSkill) []ResolvedSkill {
	sr := NewSkillResolver()
	return sr.Deduplicate(skills)
}
