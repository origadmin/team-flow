package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills",
	Long:  "Manage skills for team-flow: list, update, scan, and show skills",
}

var (
	roleFlag   string
	refreshFlag bool
	sourceFlag string
)

func init() {
	listCmd.Flags().StringVarP(&roleFlag, "role", "r", "", "Filter by role")
	listCmd.Flags().BoolVar(&refreshFlag, "refresh", false, "Force refresh cache")
	scanCmd.Flags().StringVarP(&sourceFlag, "source", "s", "", "Filter by source (trae/local/team/global)")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(scanCmd)
	Cmd.AddCommand(showCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resolved skills",
	Long:  "List skills resolved for the current project and role",
	RunE:  runList,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update skill cache",
	Long:  "Force update the skill cache",
	RunE:  runUpdate,
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan available skills",
	Long:  "Scan all available skills from all sources",
	RunE:  runScan,
}

var showCmd = &cobra.Command{
	Use:   "show <skill-id>",
	Short: "Show skill details",
	Long:  "Show detailed information about a specific skill",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runList(cmd *cobra.Command, args []string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	manager := NewSkillManager(projectRoot, "", "")

	opts := ResolveOptions{
		ForceRefresh: refreshFlag,
	}

	skillsFile, err := manager.ResolveWithOptions(roleFlag, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Skills for role: %s (team: %s, project: %s)\n", 
		orDefault(skillsFile.Role, "none"),
		orDefault(skillsFile.Team, "none"),
		orDefault(skillsFile.Project, "none"))
	fmt.Printf("Tags: %s\n", strings.Join(skillsFile.SkillTags, ", "))
	fmt.Println()

	if len(skillsFile.Skills) == 0 {
		fmt.Println("No skills found")
		return nil
	}

	fmt.Println("  ID                      Source   Trigger")
	fmt.Println("  ─────────────────────────────────────────────────")
	
	enabledCount := 0
	for _, skill := range skillsFile.Skills {
		if skill.Enabled {
			enabledCount++
		}
		fmt.Printf("  %-25s %-8s %s\n", skill.ID, skill.Source, skill.Trigger)
	}
	
	fmt.Printf("\nTotal: %d skills (%d enabled, %d disabled)\n", 
		len(skillsFile.Skills), enabledCount, len(skillsFile.Skills)-enabledCount)

	return nil
}

func runUpdate(cmd *cobra.Command, args []string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	fmt.Println("Updating skill cache...")

	manager := NewSkillManager(projectRoot, "", "")

	opts := ResolveOptions{
		ForceRefresh: true,
	}

	skillsFile, err := manager.ResolveWithOptions(roleFlag, opts)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Resolved %d skill tags\n", len(skillsFile.SkillTags))
	fmt.Printf("✓ Resolved %d skills\n", len(skillsFile.Skills))
	fmt.Printf("✓ Cache updated at: %s\n", skillsFile.ResolvedAt)
	fmt.Printf("✓ Saved to: %s\n", manager.cacheMgr.GetCachePath())

	return nil
}

func runScan(cmd *cobra.Command, args []string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	fmt.Println("Scanning available skills...")
	fmt.Println()

	manager := NewSkillManager(projectRoot, "", "")

	skills, err := manager.ScanAvailable(sourceFlag)
	if err != nil {
		return err
	}

	bySource := make(map[string][]ResolvedSkill)
	for _, skill := range skills {
		bySource[skill.Source] = append(bySource[skill.Source], skill)
	}

	sources := []string{SkillSourceLocal, SkillSourceTeam, SkillSourceGlobal, SkillSourceTrae}
	if sourceFlag != "" && sourceFlag != "all" {
		sources = []string{sourceFlag}
	}

	for _, source := range sources {
		if sourceSkills, ok := bySource[source]; ok {
			fmt.Printf("Source: %s\n", source)
			for _, skill := range sourceSkills {
				fmt.Printf("  %-25s %s\n", skill.ID, skill.Trigger)
			}
			fmt.Println()
		}
	}

	fmt.Printf("Total available: %d skills\n", len(skills))

	return nil
}

func runShow(cmd *cobra.Command, args []string) error {
	skillID := args[0]

	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	manager := NewSkillManager(projectRoot, "", "")

	skill, err := manager.GetSkill(skillID)
	if err != nil {
		return fmt.Errorf("skill not found: %s", skillID)
	}

	fmt.Printf("Skill: %s\n", skill.ID)
	fmt.Printf("Source: %s\n", skill.Source)
	fmt.Printf("Trigger: %s\n", skill.Trigger)
	if skill.Path != "" {
		fmt.Printf("Path: %s\n", skill.Path)
	}
	fmt.Printf("Enabled: %v\n", skill.Enabled)

	if skill.Path != "" {
		fmt.Println("\nDirectory structure:")
		printDirectory(skill.Path, "  ")
	}

	return nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		teamDir := filepath.Join(dir, ".team")
		if _, err := os.Stat(teamDir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no project root found (looking for .team directory)")
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func printDirectory(path string, indent string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		fmt.Printf("%s%s\n", indent, name)
		if entry.IsDir() {
			printDirectory(filepath.Join(path, name), indent+"  ")
		}
	}

	return nil
}
