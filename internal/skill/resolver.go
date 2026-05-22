package skill

func ResolveSkills(tags []string) []ResolvedSkill {
	var skills []ResolvedSkill
	for _, tag := range tags {
		if mapped, ok := DefaultSkillTagMap[tag]; ok {
			skills = append(skills, mapped...)
		}
	}
	return DeduplicateSkills(skills)
}

func DeduplicateSkills(skills []ResolvedSkill) []ResolvedSkill {
	seen := make(map[string]bool)
	var result []ResolvedSkill
	for _, s := range skills {
		if !seen[s.ID] {
			seen[s.ID] = true
			result = append(result, s)
		}
	}
	return result
}
