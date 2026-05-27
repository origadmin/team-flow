package migrate

import (
	"bufio"
	"encoding/json"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	skillfs "github.com/origadmin/team-flow"
	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/ide"
	"github.com/origadmin/team-flow/internal/logger"
	"github.com/spf13/cobra"
)

var (
	migrateDryRun bool
	migrateForce  bool
	migrateFlow   string
	migrateYes    bool
	log           *logger.Logger
)

type TaskPoolEntry struct {
	ID       string
	Type     string
	Title    string
	Priority string
	Status   string
	Phase    string
	Assignee string
	Docs     string
}

var Cmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate project version (v1→v2, v2→v3)",
	Long: `Migrate project between team-flow versions.

Subcommands:
  migrate v2    Migrate v1 (task-pool) → v2 (beads-native)
  migrate v3    Migrate v2 (beads-native) → v3 (flow-engine)

Without subcommand, auto-detects current version and migrates to next.`,
	RunE: runMigrateAuto,
}

var v2Cmd = &cobra.Command{
	Use:   "v2",
	Short: "Migrate v1 → v2",
	Long: `Migrate project from v1 (task-pool.md) to v2 (beads-native + code-review-graph).

Steps:
  1. Parse task-pool.md entries
  2. Initialize beads database (bd init)
  3. Create beads issues from task-pool entries
  4. Build code-review-graph
  5. Update project configuration`,
	RunE: runMigrateV2,
}

var v3Cmd = &cobra.Command{
	Use:   "v3",
	Short: "Migrate v2 → v3",
	Long: `Migrate project from v2 (beads-native) to v3 (flow-engine).

Steps:
  1. Pre-check: verify current version is v2
  2. Backup: .team/project.md → .team/project.md.v2.bak
  3. Update project.md: docs_path → docs_internal, add docs_external, add default_flow
  4. Update .team/version: v2 → v3
  5. Install v3 skills to IDE skill directory (v2 preserved in v2/ subdirectory)
  6. Update IDE bridge file to point to v3 SKILL
  7. Verify: flow proc run --flow <flow-name>`,
	RunE: runMigrateV3,
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback v3 → v2",
	Long: `Rollback from v3 (flow-engine) to v2 (beads-native).

Prerequisites:
  - .team/project.md.v2.bak must exist (created during migrate v3)
  - v2 fallback skills must exist in IDE skill directory (v2/ subdirectory)

Steps:
  1. Verify .team/version is v3
  2. Restore .team/project.md from .team/project.md.v2.bak
  3. Update .team/version: v3 → v2
  4. Restore v2 SKILL.md from v2/ subdirectory to skill root
  5. Update IDE bridge file to point to v2 SKILL
  6. Verify`,
	RunE: runMigrateRollback,
}

func init() {
	Cmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Dry run, do not execute changes")
	Cmd.Flags().BoolVar(&migrateForce, "force", false, "Force migration")
	Cmd.Flags().BoolVarP(&migrateYes, "yes", "y", false, "Auto-confirm all prompts")

	v3Cmd.Flags().StringVar(&migrateFlow, "flow", "dev-flow", "Default flow to bind (dev-flow, novel-flow, etc.)")
	v3Cmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Dry run, do not execute changes")
	v3Cmd.Flags().BoolVar(&migrateForce, "force", false, "Force migration even if already v3")
	v3Cmd.Flags().BoolVarP(&migrateYes, "yes", "y", false, "Auto-confirm all prompts")

	v2Cmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Dry run, do not execute changes")
	v2Cmd.Flags().BoolVar(&migrateForce, "force", false, "Force migration even if .beads exists")

	Cmd.AddCommand(v2Cmd)
	Cmd.AddCommand(v3Cmd)
	Cmd.AddCommand(rollbackCmd)
}

func runMigrateAuto(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	versionFile := filepath.Join(projectPath, ".team", "version")
	currentVersion := "unknown"
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion = strings.TrimSpace(string(data))
	}

	switch currentVersion {
	case "v1":
		fmt.Println("Detected v1 project. Migrating to v2...")
		return runMigrateV2(cmd, args)
	case "v2":
		fmt.Println("Detected v2 project. Migrating to v3...")
		return runMigrateV3(cmd, args)
	case "v3":
		fmt.Println("Already on v3. No migration needed.")
		return nil
	default:
		fmt.Println("Unknown or missing version. Run: flow init")
		return nil
	}
}

func runMigrateV3(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()
	log.Debug("Starting v2 → v3 migration")

	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        v2 → v3 Migration                 ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  Project: %s\n", projectPath)
	fmt.Printf("  Target flow: %s\n\n", migrateFlow)

	fmt.Println("━━━ Step 1: Pre-check ━━━")
	versionFile := filepath.Join(projectPath, ".team", "version")
	currentVersion := "unknown"
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion = strings.TrimSpace(string(data))
	}

	if currentVersion == "v3" && !migrateForce {
		fmt.Println("  Already on v3. Use --force to re-migrate.")
		return nil
	}
	if currentVersion != "v2" && currentVersion != "unknown" && !migrateForce {
		fmt.Printf("  ⚠ Current version is %q, expected v2. Use --force to override.\n", currentVersion)
		return nil
	}
	fmt.Printf("  Current version: %s\n", currentVersion)

	projectMdPath := filepath.Join(projectPath, ".team", "project.md")
	if _, err := os.Stat(projectMdPath); err != nil {
		return fmt.Errorf(".team/project.md not found. Run: flow init")
	}
	fmt.Println("  ✓ .team/project.md found")

	beadsDir := filepath.Join(projectPath, ".beads")
	if _, err := os.Stat(beadsDir); err == nil {
		fmt.Println("  ✓ .beads/ found (will be preserved)")
	} else {
		fmt.Println("  ⚠ .beads/ not found (beads tasks will not be available)")
	}

	fmt.Println("\n━━━ Step 2: Backup ━━━")
	backupPath := projectMdPath + ".v2.bak"
	if _, err := os.Stat(backupPath); err == nil && !migrateForce {
		fmt.Println("  Backup already exists: " + backupPath)
	} else {
		data, err := os.ReadFile(projectMdPath)
		if err != nil {
			return fmt.Errorf("read project.md: %w", err)
		}
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return fmt.Errorf("create backup: %w", err)
		}
		fmt.Println("  ✓ Backup created: .team/project.md.v2.bak")
	}

	if migrateDryRun {
		fmt.Println("\n=== DRY RUN ===")
		fmt.Println("Would update .team/project.md:")
		fmt.Println("  - docs_path: → docs_internal:")
		fmt.Println("  - Add docs_external: docs/")
		fmt.Printf("  - Add default_flow: %s\n", migrateFlow)
		fmt.Println("  - Registered Projects table: docs_path → docs_internal")
		fmt.Println("Would update .team/version: v2 → v3")
		fmt.Println("Would install v3 skills")
		fmt.Println("Would update IDE bridge file")
		return nil
	}

	fmt.Println("\n━━━ Step 3: Update project.md ━━━")
	data, err := os.ReadFile(projectMdPath)
	if err != nil {
		return fmt.Errorf("read project.md: %w", err)
	}
	content := string(data)

	updated := content

	updated = strings.Replace(updated, "docs_path:", "docs_internal:", -1)

	if !strings.Contains(updated, "docs_external:") {
		docsInternalLine := ""
		for _, line := range strings.Split(updated, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "docs_internal:") {
				docsInternalLine = line
				break
			}
		}
		if docsInternalLine != "" {
			updated = strings.Replace(updated, docsInternalLine, docsInternalLine+"\ndocs_external: docs/", 1)
		}
	}

	if !strings.Contains(updated, "default_flow:") {
		lines := strings.Split(updated, "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "docs_external:") {
				lines[i] = line + "\ndefault_flow: " + migrateFlow
				updated = strings.Join(lines, "\n")
				break
			}
		}
	} else {
		re := regexp.MustCompile(`default_flow:\s*\S+`)
		updated = re.ReplaceAllString(updated, "default_flow: "+migrateFlow)
	}

	updated = strings.Replace(updated, "| docs_path |", "| docs_internal |", 1)
	updated = regexp.MustCompile(`\| (_docs/\S+) \|`).ReplaceAllString(updated, "| $1 |")

	updated = strings.Replace(updated, "- **Team Version**: v2", "- **Team Version**: v3", 1)

	if err := os.WriteFile(projectMdPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write project.md: %w", err)
	}
	fmt.Println("  ✓ docs_path → docs_internal")
	fmt.Println("  ✓ Added docs_external: docs/")
	fmt.Printf("  ✓ Added default_flow: %s\n", migrateFlow)
	fmt.Println("  ✓ Updated Registered Projects table")

	fmt.Println("\n━━━ Step 4: Update version ━━━")
	if err := os.WriteFile(versionFile, []byte("v3"), 0644); err != nil {
		fmt.Printf("  ⚠ Failed to update .team/version: %v\n", err)
	} else {
		fmt.Println("  ✓ .team/version updated to v3")
	}

	fmt.Println("\n━━━ Step 5: Install v3 skills ━━━")
	installV3Skills(projectPath, skillfs.FS)

	fmt.Println("\n━━━ Step 5.5: Install preset flows ━━━")
	installPresetFlows(projectPath, skillfs.FS)

	fmt.Println("\n━━━ Step 6: Update IDE bridge ━━━")
	updateBridgeFile(projectPath)

	fmt.Println("\n━━━ Step 7: Verify ━━━")
	flowFile := filepath.Join(projectPath, "v3", "flows", migrateFlow+".json")
	if _, err := os.Stat(flowFile); err == nil {
		fmt.Printf("  ✓ Flow file found: v3/flows/%s.json\n", migrateFlow)
	} else {
		flowFileAlt := filepath.Join(projectPath, ".team", "flows", migrateFlow+".json")
		if _, err := os.Stat(flowFileAlt); err == nil {
			fmt.Printf("  ✓ Flow file found: .team/flows/%s.json\n", migrateFlow)
		} else {
			fmt.Printf("  ⚠ Flow file not found: %s.json\n", migrateFlow)
			fmt.Println("    Create one with: flow proc create")
		}
	}

	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║        v2 → v3 Migration Complete!       ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println("\nWhat changed:")
	fmt.Println("  • .team/version: v2 → v3")
	fmt.Println("  • .team/project.md: docs_path → docs_internal + docs_external + default_flow")
	fmt.Println("  • v3 skills installed to IDE skill directory")
	fmt.Println("  • IDE bridge file updated to v3 SKILL")
	fmt.Println("\nWhat was preserved:")
	fmt.Println("  • .beads/ database (beads is version-agnostic)")
	fmt.Println("  • _docs/ directory (path unchanged, variable name changed)")
	fmt.Println("  • .team/project.md.v2.bak (backup)")
	fmt.Println("\nRollback:")
	fmt.Println("  1. cp -f .team/project.md.v2.bak .team/project.md")
	fmt.Println("  2. echo v2 > .team/version")
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. flow proc run --flow %s\n", migrateFlow)
	fmt.Println("  2. AI reads v3 SKILL.md → exec skill → flow engine drives execution")

	return nil
}

func installV3Skills(projectPath string, fsys embed.FS) {
	srcDir := skillfs.SkillRoot + "/v3"
	ides := detectIDEs(projectPath)
	installed := false

	for _, ide := range ides {
		if !ide.Detected {
			continue
		}

		skillDir := filepath.Join(ide.SkillDir, "team-flow")

		skillEntry := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillEntry); err == nil || migrateForce {
			entryData, readErr := fsys.ReadFile(skillfs.SkillRoot + "/v3/SKILL.md")
			if readErr != nil {
				fmt.Printf("  ⚠ Read SKILL.md error: %v\n", readErr)
			} else {
				os.MkdirAll(skillDir, 0755)
				if writeErr := os.WriteFile(skillEntry, entryData, 0644); writeErr != nil {
					fmt.Printf("  ⚠ Write SKILL.md error: %v\n", writeErr)
				}
			}
		}

		copied, _, err := copyFromFS(fsys, srcDir, skillDir, true, nil)
		if err != nil {
			fmt.Printf("  ⚠ %s skill copy error: %v\n", ide.Name, err)
			continue
		}

		fmt.Printf("  ✓ %s: Copied %d v3 skill files\n", ide.Name, copied)

		installMigrateFallback(fsys, skillDir, "v2", ide.Name)

		installed = true
	}

	if !installed {
		fmt.Println("  No IDE detected. Installing to .agents/skills/team-flow/")
		skillDir := filepath.Join(projectPath, ".agents", "skills", "team-flow")
		entryData, readErr := fsys.ReadFile(skillfs.SkillRoot + "/v3/SKILL.md")
		if readErr == nil {
			os.MkdirAll(skillDir, 0755)
			os.WriteFile(filepath.Join(skillDir, "SKILL.md"), entryData, 0644)
		}
		copied, _, err := copyFromFS(fsys, srcDir, skillDir, true, nil)
		if err != nil {
			fmt.Printf("  ⚠ Skill copy error: %v\n", err)
		} else {
			fmt.Printf("  ✓ Copied %d v3 skill files\n", copied)
		}

		installMigrateFallback(fsys, skillDir, "v2", "local")
	}
}

func installMigrateFallback(fsys embed.FS, skillDir, fallbackVersion, ideName string) {
	fallbackSrcDir := skillfs.SkillRoot + "/" + fallbackVersion
	fallbackDstDir := filepath.Join(skillDir, fallbackVersion)

	fallbackSkillEntry := filepath.Join(fallbackDstDir, "SKILL.md")
	if _, err := os.Stat(fallbackSkillEntry); err == nil && !migrateForce {
		fmt.Printf("  ✓ %s: %s fallback already exists\n", ideName, fallbackVersion)
		return
	}

	copied, _, err := copyFromFS(fsys, fallbackSrcDir, fallbackDstDir, migrateForce, nil)
	if err != nil {
		fmt.Printf("  ⚠ %s: %s fallback copy error: %v\n", ideName, fallbackVersion, err)
		return
	}

	fmt.Printf("  ✓ %s: %s fallback preserved (%d files)\n", ideName, fallbackVersion, copied)
}

func installPresetFlows(projectPath string, fsys embed.FS) {
	presetSrcDir := skillfs.FlowsRoot
	projectFlowsDir := filepath.Join(projectPath, ".team", "flows")

	if _, err := fs.ReadDir(fsys, presetSrcDir); err != nil {
		srcFlowsDir := filepath.Join(projectPath, "v3", "flows")
		if _, err := os.Stat(srcFlowsDir); err != nil {
			fmt.Println("  ⚠ No preset flows found to install")
			return
		}
		entries, _ := os.ReadDir(srcFlowsDir)
		os.MkdirAll(projectFlowsDir, 0755)
		copied := 0
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			src := filepath.Join(srcFlowsDir, e.Name())
			dst := filepath.Join(projectFlowsDir, e.Name())
			data, err := os.ReadFile(src)
			if err != nil {
				continue
			}
			if err := os.WriteFile(dst, data, 0644); err != nil {
				fmt.Printf("  ⚠ Failed to copy %s: %v\n", e.Name(), err)
				continue
			}
			copied++
		}
		fmt.Printf("  ✓ Copied %d preset flows to .team/flows/\n", copied)
		return
	}

	os.MkdirAll(projectFlowsDir, 0755)
	copied, _, err := copyFromFS(fsys, presetSrcDir, projectFlowsDir, false, nil)
	if err != nil {
		fmt.Printf("  ⚠ Preset flows copy error: %v\n", err)
	} else {
		fmt.Printf("  ✓ Copied %d preset flows to .team/flows/\n", copied)
	}

	registerFlowsInProjectMd(projectPath, projectFlowsDir)
}

func registerFlowsInProjectMd(projectPath, flowsDir string) {
	projectMdPath := filepath.Join(projectPath, ".team", "project.md")
	data, err := os.ReadFile(projectMdPath)
	if err != nil {
		return
	}
	content := string(data)

	if strings.Contains(content, "## Flows") {
		fmt.Println("  ✓ Flows section already exists in project.md")
		return
	}

	entries, err := os.ReadDir(flowsDir)
	if err != nil || len(entries) == 0 {
		return
	}

	var flowRows []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		flowPath := filepath.Join(flowsDir, e.Name())
		flowData, err := os.ReadFile(flowPath)
		if err != nil {
			continue
		}

		desc := extractFlowDescription(flowData, name)
		isDefault := name == migrateFlow
		defaultMark := ""
		if isDefault {
			defaultMark = " ⭐"
		}
		flowRows = append(flowRows, fmt.Sprintf("| %s | preset | %s |%s", name, desc, defaultMark))
	}

	if len(flowRows) == 0 {
		return
	}

	var sb strings.Builder
	sb.WriteString("\n## Flows\n\n")
	sb.WriteString("| Flow | Source | Description | Default |\n")
	sb.WriteString("|------|--------|-------------|--------|\n")
	for _, row := range flowRows {
		sb.WriteString(row + " |\n")
	}

	insertPoint := "## Toolchain"
	if idx := strings.Index(content, insertPoint); idx > 0 {
		content = content[:idx] + sb.String() + "\n" + content[idx:]
	} else {
		content += sb.String()
	}

	if err := os.WriteFile(projectMdPath, []byte(content), 0644); err != nil {
		fmt.Printf("  ⚠ Failed to update project.md: %v\n", err)
	} else {
		fmt.Printf("  ✓ Registered %d flows in project.md\n", len(flowRows))
	}
}

func extractFlowDescription(flowData []byte, name string) string {
	var flow struct {
		Metadata struct {
			Description string `json:"description"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(flowData, &flow); err == nil && flow.Metadata.Description != "" {
		return flow.Metadata.Description
	}
	return name
}

func updateBridgeFile(projectPath string, targetVersion ...string) {
	target := "v3"
	if len(targetVersion) > 0 {
		target = targetVersion[0]
	}

	ides := detectIDEs(projectPath)
	if len(ides) == 0 {
		ides = append(ides, ide.IDEInfo{
			Name:       "Trae",
			ConfigDir:  filepath.Join(projectPath, ".trae"),
			BridgePath: ".trae/rules/team-flow.md",
			BridgeFmt:  "trae",
			Detected:   true,
		})
	}

	for _, ide := range ides {
		if !ide.Detected {
			continue
		}

		bridgeAbsPath := filepath.Join(projectPath, ide.BridgePath)
		if _, err := os.Stat(bridgeAbsPath); err != nil {
			bridgeAbsPath = filepath.Join(filepath.Dir(projectPath), ide.BridgePath)
			if _, err := os.Stat(bridgeAbsPath); err != nil {
				continue
			}
		}

		data, err := os.ReadFile(bridgeAbsPath)
		if err != nil {
			continue
		}
		content := string(data)

		if target == "v3" {
			if strings.Contains(content, "v2") || strings.Contains(content, "assets/skill/v2") {
				content = strings.Replace(content, "assets/skill/v2/SKILL.md", "assets/skill/v3/SKILL.md", -1)
				content = strings.Replace(content, "team-flow/v2", "team-flow/v3", -1)

				if strings.Contains(content, "bd ready") {
					content = strings.Replace(content, "bd ready", "flow proc run", -1)
				}
				if strings.Contains(content, "bd show") {
					content = strings.Replace(content, "bd show", "flow task show", -1)
				}
				if strings.Contains(content, "bd create") {
					content = strings.Replace(content, "bd create", "flow task create", -1)
				}
				if strings.Contains(content, "bd update") {
					content = strings.Replace(content, "bd update", "flow task update", -1)
				}
				if strings.Contains(content, "bd close") {
					content = strings.Replace(content, "bd close", "flow task close", -1)
				}
			}

			statusLine := "## Status Line (MANDATORY)\n\n" +
				"Every response MUST start with:\n\n" +
				"```\n" +
				"[Role: {alias} | Flow: {flow-name} | Node: {node-id} ({node-name}) | Phase: {phase}]\n" +
				"```\n\n" +
				"- Role: alias from flow proc run output\n" +
				"- Flow: flow name (e.g., dev-flow)\n" +
				"- Node: current node ID and name\n" +
				"- Phase: current execution phase\n\n"

			if !strings.Contains(content, "Flow: {flow-name}") {
				oldStatus := "## Status Line (MANDATORY)"
				if idx := strings.Index(content, oldStatus); idx >= 0 {
					endIdx := strings.Index(content[idx:], "## ")
					if endIdx > 0 && endIdx < 500 {
						endIdx += idx
					} else {
						endIdx = len(content)
					}
					content = content[:idx] + statusLine + content[endIdx:]
				}
			}

			skillRef := "Read `.team/version` for active version (v1, v2, or v3).\n" +
				"If v3: Load and follow ALL rules in the v3 SKILL.md as the entry point.\n" +
				"Run `flow proc run` to start the flow engine.\n"

			if strings.Contains(content, "Load and follow ALL rules in `") {
				re := regexp.MustCompile("Load and follow ALL rules in `[^`]+` as the entry point\\.")
				content = re.ReplaceAllString(content, skillRef)
			}
		} else if target == "v2" {
			content = strings.Replace(content, "assets/skill/v3/SKILL.md", "assets/skill/v2/SKILL.md", -1)
			content = strings.Replace(content, "team-flow/v3", "team-flow/v2", -1)
			content = strings.Replace(content, "flow proc run", "bd ready", -1)
			content = strings.Replace(content, "flow task show", "bd show", -1)
			content = strings.Replace(content, "flow task create", "bd create", -1)
			content = strings.Replace(content, "flow task update", "bd update", -1)
			content = strings.Replace(content, "flow task close", "bd close", -1)

			skillRef := "Read `.team/version` for active version.\n" +
				"If v2: Load and follow ALL rules in the v2 SKILL.md as the entry point.\n" +
				"Run `bd ready` to find available work.\n"

			if strings.Contains(content, "Read `.team/version`") {
				re := regexp.MustCompile("Read `\\.team/version`[^\\n]*\\n(If v3:[^\\n]*\\n|Run `flow proc run`[^\\n]*\\n)*")
				content = re.ReplaceAllString(content, skillRef)
			}
		}

		if err := os.WriteFile(bridgeAbsPath, []byte(content), 0644); err != nil {
			fmt.Printf("  ⚠ %s bridge update failed: %v\n", ide.Name, err)
		} else {
			fmt.Printf("  ✓ %s bridge file updated to %s\n", ide.Name, target)
		}
	}
}

func detectIDEs(projectPath string) []ide.IDEInfo {
	return ide.DetectIDEsWithParentSearch(projectPath, skillfs.FS)
}

func copyFromFS(fsys embed.FS, srcDir, dst string, overwrite bool, excludeDirs []string) (copied, skipped int, err error) {
	err = fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		var relPath string
		if srcDir == "." {
			relPath = path
			if relPath == "." {
				return nil
			}
		} else {
			if path == srcDir {
				return nil
			}
			relPath = strings.TrimPrefix(path, srcDir+"/")
		}

		parts := strings.Split(relPath, "/")
		for _, part := range parts {
			for _, exc := range excludeDirs {
				if part == exc {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
			}
		}

		dstPath := filepath.Join(dst, filepath.FromSlash(relPath))

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if !overwrite {
			if _, err := os.Stat(dstPath); err == nil {
				skipped++
				return nil
			}
		}

		data, readErr := fsys.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		if mkdirErr := os.MkdirAll(filepath.Dir(dstPath), 0755); mkdirErr != nil {
			return mkdirErr
		}

		if writeErr := os.WriteFile(dstPath, data, 0644); writeErr != nil {
			return writeErr
		}
		copied++
		return nil
	})
	return
}

func runMigrateV2(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()
	log.Debug("Starting v1 → v2 migration")

	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	log.Debugf("Starting migration: %s", projectPath)

	versionFile := filepath.Join(projectPath, ".team", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion := strings.TrimSpace(string(data))
		if currentVersion == "v2" {
			fmt.Println("Already on v2. No migration needed.")
			return nil
		}
		if currentVersion != "v1" {
			fmt.Printf("Unknown version %q. Expected v1.\n", currentVersion)
			fmt.Println("Run: flow init --v1")
			return nil
		}
	} else {
		fmt.Println("⚠ .team/version not found. Assuming v1 project.")
	}

	fmt.Println("=== v1 → v2 Migration ===")
	fmt.Printf("Project: %s\n\n", projectPath)

	taskPoolPath := findTaskPool(projectPath)
	if taskPoolPath == "" {
		return fmt.Errorf("task-pool.md not found. Searched: .team/, .trae/skills/team-flow/, .cursor/skills/team-flow/, .team-flow/skill/")
	}
	fmt.Printf("Found task-pool: %s\n", taskPoolPath)

	entries, err := parseTaskPool(taskPoolPath)
	if err != nil {
		return fmt.Errorf("parse task-pool: %w", err)
	}
	fmt.Printf("Parsed %d task entries\n\n", len(entries))

	if migrateDryRun {
		fmt.Println("=== DRY RUN ===")
		for _, e := range entries {
			fmt.Printf("  %s [%s] %s (P%s, %s)\n", e.ID, e.Type, e.Title, e.Priority, e.Status)
		}
		fmt.Printf("\nWould create %d beads issues\n", len(entries))
		return nil
	}

	bdCmd := bd.FindPath()
	if bdCmd == "" {
		fmt.Println("\n⚠ beads (bd CLI) not found.")
		fmt.Println("Installing beads...")
		if err := bd.Install(); err != nil {
			return fmt.Errorf("bd installation failed: %w\nManual install: https://github.com/steveyegge/beads", err)
		}
		bdCmd = bd.FindPath()
		if bdCmd != "" {
			bd.EnsureOnPath()
		}
	}

	if bdCmd == "" {
		return fmt.Errorf("bd not available after installation")
	}

	if err := bd.EnsureOnPath(); err != nil {
		fmt.Printf("  ⚠ %v\n", err)
	}

	beadsDir := filepath.Join(projectPath, ".beads")
	if _, err := os.Stat(beadsDir); os.IsNotExist(err) || migrateForce {
		fmt.Println("Initializing beads database...")
		if _, err := bd.Run("init"); err != nil {
			return fmt.Errorf("bd init: %w", err)
		}
		fmt.Println("  ✓ bd init done")
	} else {
		fmt.Println("  .beads already exists, skipping bd init")
	}

	fmt.Printf("\nMigrating %d tasks to beads...\n", len(entries))
	migrated := 0
	for _, e := range entries {
		issueType := mapType(e.Type)
		priority := mapPriority(e.Priority)

		createArgs := []string{
			"create", e.Title,
			"-t", issueType,
			"-p", fmt.Sprintf("%d", priority),
			"--external-ref", e.ID,
			"--json",
		}

		output, _ := bd.Run(createArgs...)

		jsonOutput, jsonErr := extractJSON(output)
		if jsonErr != nil {
			fmt.Printf("  ✗ %s: failed to extract JSON: %v\n", e.ID, jsonErr)
			fmt.Printf("    Raw output: %s\n", truncateString(output, 200))
			continue
		}

		var results []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonOutput), &results); err != nil {
			var singleResult map[string]interface{}
			if err2 := json.Unmarshal([]byte(jsonOutput), &singleResult); err2 != nil {
				fmt.Printf("  ✗ %s: parse error: %v\n", e.ID, err)
				continue
			}
			results = []map[string]interface{}{singleResult}
		}

		if len(results) > 0 {
			result := results[0]
			if beadsID, ok := result["id"].(string); ok {
				var title string
				if t, ok := result["title"].(string); ok {
					title = t
					fmt.Printf("  ✓ %s → %s (%s)\n", e.ID, beadsID, title)
				} else {
					fmt.Printf("  ✓ %s → %s\n", e.ID, beadsID)
				}

				if e.Status == "Doing" || e.Status == "Review" {
					bd.Run("update", beadsID, "--status", "in_progress")
				}
				if e.Assignee != "" && e.Assignee != "-" {
					bd.Run("update", beadsID, "--assignee", e.Assignee)
				}
				if title != "" {
					log.Debugf("Created issue %s: %s", beadsID, title)
				}
				notes := fmt.Sprintf("MIGRATED FROM v1 task-pool. Original ID: %s, Status: %s, Phase: %s", e.ID, e.Status, e.Phase)
				if e.Docs != "" && e.Docs != "-" {
					notes += fmt.Sprintf(", Docs: %s", e.Docs)
				}
				bd.Run("update", beadsID, "--notes", notes)
			}
		}

		if _, ok := results[0]["id"].(string); ok {
			migrated++
		}
	}

	fmt.Printf("\n✅ Migrated %d/%d tasks\n", migrated, len(entries))

	if migrated == 0 {
		fmt.Println("\n⚠ No tasks migrated successfully. Not updating version.")
		fmt.Println("Please check the errors above and try again.")
		return fmt.Errorf("migration failed: no tasks migrated")
	}

	versionFile = filepath.Join(projectPath, ".team", "version")
	if err := os.WriteFile(versionFile, []byte("v2"), 0644); err != nil {
		fmt.Printf("  ⚠ Failed to update .team/version: %v\n", err)
	} else {
		fmt.Println("  ✓ .team/version updated to v2")
	}

	graphDir := filepath.Join(projectPath, ".code-review-graph")
	if _, err := os.Stat(graphDir); os.IsNotExist(err) {
		fmt.Println("\nBuilding code-review-graph...")
		if err := runCmd("python", "-m", "code_review_graph", "build"); err != nil {
			fmt.Printf("  ⚠ code-review-graph build failed: %v\n", err)
			fmt.Println("  You can build it later: python -m code_review_graph build")
		} else {
			fmt.Println("  ✓ Graph built successfully")
		}
	}

	fmt.Println("\n=== v1 → v2 Migration Complete ===")
	fmt.Println("Next steps:")
	fmt.Println("  1. Verify: bd ready --json")
	fmt.Println("  2. Check graph: python -m code_review_graph status")
	fmt.Println("  3. Update AI config to use v2 SKILL.md")
	fmt.Println("  4. Backup and remove old task-pool.md")

	return nil
}

func findTaskPool(projectPath string) string {
	candidates := []string{
		filepath.Join(projectPath, ".team", "task-pool.md"),
		filepath.Join(projectPath, ".trae", "skills", "team-flow", "task-pool.md"),
		filepath.Join(projectPath, ".cursor", "skills", "team-flow", "task-pool.md"),
		filepath.Join(projectPath, ".claude", "skills", "team-flow", "task-pool.md"),
		filepath.Join(projectPath, ".team-flow", "skill", "task-pool.md"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func parseTaskPool(path string) ([]TaskPoolEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []TaskPoolEntry
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	tableRegex := regexp.MustCompile(`^\|\s*(F|B|C|A|D)\d{3}\s*\|`)

	for scanner.Scan() {
		line := scanner.Text()
		if !tableRegex.MatchString(line) {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}

		cleanFields := make([]string, len(fields))
		for i, f := range fields {
			cleanFields[i] = strings.TrimSpace(f)
		}

		entry := TaskPoolEntry{
			ID:       cleanFields[1],
			Title:    cleanFields[2],
			Priority: cleanFields[3],
			Status:   cleanFields[4],
			Phase:    cleanFields[5],
			Docs:     cleanFields[6],
		}

		if len(entry.ID) > 0 {
			switch strings.ToUpper(entry.ID[:1]) {
			case "F":
				entry.Type = "feature"
			case "B":
				entry.Type = "bug"
			case "C":
				entry.Type = "task"
			case "A":
				entry.Type = "task"
			case "D":
				entry.Type = "task"
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func mapType(t string) string {
	switch t {
	case "feature":
		return "feature"
	case "bug":
		return "bug"
	default:
		return "task"
	}
}

func mapPriority(p string) int {
	switch strings.TrimPrefix(p, "P") {
	case "0":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	default:
		return 2
	}
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractJSON(output string) (string, error) {
	start := strings.Index(output, "[")
	if start == -1 {
		start = strings.Index(output, "{")
	}
	if start == -1 {
		return "", fmt.Errorf("no JSON found in output")
	}

	end := strings.LastIndex(output, "]")
	if end == -1 || !strings.Contains(output[:end], "[") {
		end = strings.LastIndex(output, "}")
	}

	return strings.TrimSpace(output[start : end+1]), nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func runMigrateRollback(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        v3 → v2 Rollback                  ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  Project: %s\n", projectPath)

	versionFile := filepath.Join(projectPath, ".team", "version")
	currentVersion := "unknown"
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion = strings.TrimSpace(string(data))
	}

	fmt.Println("\n━━━ Step 1: Pre-check ━━━")
	fmt.Printf("  Current version: %s\n", currentVersion)
	if currentVersion != "v3" {
		return fmt.Errorf("rollback only works from v3, current version is %s", currentVersion)
	}

	bakFile := filepath.Join(projectPath, ".team", "project.md.v2.bak")
	if _, err := os.Stat(bakFile); err != nil {
		return fmt.Errorf("backup file not found: %s (rollback requires backup created during migrate v3)", bakFile)
	}
	fmt.Println("  ✓ Backup file found")

	fmt.Println("\n━━━ Step 2: Restore project.md ━━━")
	projectMd := filepath.Join(projectPath, ".team", "project.md")
	data, err := os.ReadFile(bakFile)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}
	if err := os.WriteFile(projectMd, data, 0644); err != nil {
		return fmt.Errorf("restore project.md: %w", err)
	}
	fmt.Println("  ✓ .team/project.md restored from backup")

	fmt.Println("\n━━━ Step 3: Update version ━━━")
	if err := os.WriteFile(versionFile, []byte("v2\n"), 0644); err != nil {
		return fmt.Errorf("update version: %w", err)
	}
	fmt.Println("  ✓ .team/version updated to v2")

	fmt.Println("\n━━━ Step 4: Restore v2 skills ━━━")
	ides := detectIDEs(projectPath)
	restored := false
	for _, ide := range ides {
		if !ide.Detected {
			continue
		}

		skillDir := filepath.Join(ide.SkillDir, "team-flow")
		v2SkillMd := filepath.Join(skillDir, "v2", "SKILL.md")
		rootSkillMd := filepath.Join(skillDir, "SKILL.md")

		if _, err := os.Stat(v2SkillMd); err != nil {
			fmt.Printf("  ⚠ %s: v2 fallback SKILL.md not found at %s\n", ide.Name, v2SkillMd)
			continue
		}

		v2Data, readErr := os.ReadFile(v2SkillMd)
		if readErr != nil {
			fmt.Printf("  ⚠ %s: read v2 SKILL.md error: %v\n", ide.Name, readErr)
			continue
		}

		if writeErr := os.WriteFile(rootSkillMd, v2Data, 0644); writeErr != nil {
			fmt.Printf("  ⚠ %s: write SKILL.md error: %v\n", ide.Name, writeErr)
			continue
		}

		fmt.Printf("  ✓ %s: v2 SKILL.md restored to skill root\n", ide.Name)
		restored = true
	}

	if !restored {
		fmt.Println("  ⚠ No IDE skill directories found for v2 restore")
		fmt.Println("  → Run 'flow init' to reinstall v2 skills")
	}

	fmt.Println("\n━━━ Step 5: Update IDE bridge ━━━")
	updateBridgeFile(projectPath, "v2")

	fmt.Println("\n━━━ Step 6: Verify ━━━")
	if data, err := os.ReadFile(versionFile); err == nil {
		v := strings.TrimSpace(string(data))
		fmt.Printf("  ✓ .team/version: %s\n", v)
	}
	fmt.Println("  ✓ Rollback complete")

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        v3 → v2 Rollback Complete!        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("What changed:")
	fmt.Println("  • .team/version: v3 → v2")
	fmt.Println("  • .team/project.md: restored from .team/project.md.v2.bak")
	fmt.Println("  • IDE SKILL.md: restored from v2/ fallback subdirectory")
	fmt.Println("  • IDE bridge file: updated to v2 SKILL")
	fmt.Println()
	fmt.Println("To re-upgrade:")
	fmt.Println("  flow migrate v3 --flow <flow-name>")
	fmt.Println()

	return nil
}
