package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/config"
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
	scanCmd.Flags().StringVarP(&sourceFlag, "source", "s", "", "Filter by source (trae/local/team/global/plugin")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(scanCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(uninstallCmd)
	Cmd.AddCommand(listPluginsCmd)
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
	dir, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(dir)
	if projectRoot == "" {
		return fmt.Errorf("no project root found")
	}

	manager := NewSkillManager(projectRoot)

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
	dir, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(dir)
	if projectRoot == "" {
		return fmt.Errorf("no project root found")
	}

	fmt.Println("Updating skill cache...")

	manager := NewSkillManager(projectRoot)

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
	dir, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(dir)
	if projectRoot == "" {
		return fmt.Errorf("no project root found")
	}

	fmt.Println("Scanning available skills...")
	fmt.Println()

	manager := NewSkillManager(projectRoot)

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

	dir, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(dir)
	if projectRoot == "" {
		return fmt.Errorf("no project root found")
	}

	manager := NewSkillManager(projectRoot)

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

// ── Plugin 命令 ─────────────────────────────────────────────────────────

var installCmd = &cobra.Command{
	Use:   "install <owner/repo> [@branch]",
	Short: "Install a skill plugin from GitHub",
	Long:  "Download and install a skill plugin from a GitHub repository.\n\nExample:\n  flow skill install phuryn/pm-skills\n  flow skill install phuryn/pm-skills@main",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <plugin-name>",
	Short: "Uninstall a skill plugin",
	Long:  "Remove an installed skill plugin from the project.",
	Args:  cobra.ExactArgs(1),
	RunE:  runUninstall,
}

var listPluginsCmd = &cobra.Command{
	Use:   "list-plugins",
	Short: "List installed skill plugins",
	Long:  "List all skill plugins installed in the project, with their skills.",
	RunE:  runListPlugins,
}

func runInstall(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(dir)
	if projectRoot == "" {
		return fmt.Errorf("no project root found")
	}

	manager := NewSkillManager(projectRoot)
	pm := manager.PluginManager()

	fmt.Printf("Installing plugin: %s\n", args[0])

	plugin, err := pm.Install(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("✓ Plugin installed: %s\n", plugin.Name)
	fmt.Printf("  Source: %s\n", plugin.Source)
	fmt.Printf("  Branch: %s\n", plugin.Version)
	fmt.Printf("  Installed at: %s\n", plugin.InstalledAt)
	fmt.Printf("  Directory: %s\n", plugin.InstallDir)

	if len(plugin.Skills) > 0 {
		fmt.Printf("\nSkills (%d skills available):\n", len(plugin.Skills))
		seen := make(map[string]bool)
		for _, skill := range plugin.Skills {
			if seen[skill.ID] {
				continue
			}
			seen[skill.ID] = true
			fmt.Printf("  - %s\n", skill.ID)
		}
	} else {
		fmt.Println("\n(No SKILL.md files found in plugin)")
	}

	// 刷新 skill 缓存
	if err := manager.InvalidateCache(); err == nil {
		fmt.Println("\n✓ Skill cache refreshed")
	}

	return nil
}

func runUninstall(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(dir)
	if projectRoot == "" {
		return fmt.Errorf("no project root found")
	}

	manager := NewSkillManager(projectRoot)
	pm := manager.PluginManager()

	pluginName := args[0]
	fmt.Printf("Uninstalling plugin: %s\n", pluginName)

	if err := pm.Uninstall(pluginName); err != nil {
		return err
	}

	fmt.Printf("✓ Plugin uninstalled: %s\n", pluginName)

	// 刷新 skill 缓存
	if err := manager.InvalidateCache(); err == nil {
		fmt.Println("✓ Skill cache refreshed")
	}

	return nil
}

func runListPlugins(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(dir)
	if projectRoot == "" {
		return fmt.Errorf("no project root found")
	}

	manager := NewSkillManager(projectRoot)
	pm := manager.PluginManager()

	plugins, err := pm.List()
	if err != nil {
		return err
	}

	if len(plugins) == 0 {
		fmt.Println("No skill plugins installed")
		fmt.Println("\nTo install a plugin, run:")
		fmt.Println("  flow skill install <owner>/<repo>")
		return nil
	}

	fmt.Printf("Installed skill plugins (%d):\n\n", len(plugins))

	for _, plugin := range plugins {
		fmt.Printf("Name: %s\n", plugin.Name)
		fmt.Printf("  Source: %s\n", plugin.Source)
		fmt.Printf("  Branch: %s\n", plugin.Version)
		fmt.Printf("  Installed: %s\n", plugin.InstalledAt)
		fmt.Printf("  Directory: %s\n", plugin.InstallDir)

		if len(plugin.Skills) > 0 {
			fmt.Printf("  Skills (%d):\n", len(plugin.Skills))
			seen := make(map[string]bool)
			for _, skill := range plugin.Skills {
				if seen[skill.ID] {
					continue
				}
				seen[skill.ID] = true
				fmt.Printf("    - %s\n", skill.ID)
			}
		} else {
			fmt.Println("  (no SKILL.md files found)")
		}
		fmt.Println()
	}

	return nil
}
