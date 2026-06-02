package proc

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

func formatTeamIntro(w io.Writer, intro *TeamIntroData) {
	teamName := intro.TeamName
	if intro.TeamNameZh != "" {
		teamName = intro.TeamNameZh + "(" + intro.TeamName + ")"
	}

	fmt.Fprintln(w, "╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Fprintf(w, "║  🏠 Welcome to %s\n", padLine(teamName, 62))
	fmt.Fprintln(w, "╠══════════════════════════════════════════════════════════════════════╣")

	if intro.TeamDescription != "" {
		fmt.Fprintf(w, "║  %s\n", padLine(intro.TeamDescription, 70))
		fmt.Fprintln(w, "║                                                                    ║")
	}

	fmt.Fprintf(w, "║  FLOW: %s ⭐ (default)%s\n", intro.FlowName, padLine("", 62-len(intro.FlowName)-14))
	fmt.Fprintln(w, "║                                                                    ║")
	fmt.Fprintln(w, "║  YOUR TEAM:                                                        ║")

	sort.Slice(intro.Roles, func(i, j int) bool {
		if intro.Roles[i].Principal != intro.Roles[j].Principal {
			return intro.Roles[i].Principal
		}
		return intro.Roles[i].Alias < intro.Roles[j].Alias
	})

	for _, r := range intro.Roles {
		icon := "🔧"
		if r.Principal {
			icon = "⭐"
		}
		label := r.Alias
		if r.AliasEn != "" {
			label = r.Alias + "(" + r.AliasEn + ")"
		}
		desc := fmt.Sprintf("%s · %s", label, r.RoleName)
		fmt.Fprintf(w, "║    %s %s\n", icon, padLine(desc, 67))
	}

	fmt.Fprintln(w, "║                                                                    ║")
	fmt.Fprintln(w, "║  HOW IT WORKS:                                                     ║")
	for i, step := range intro.HowItWorks {
		fmt.Fprintf(w, "║    %d. %s\n", i+1, padLine(step, 66))
	}

	fmt.Fprintln(w, "║                                                                    ║")
	fmt.Fprintln(w, "║  AVAILABLE FLOWS:                                                  ║")
	flowStrs := make([]string, 0, len(intro.Flows))
	for _, f := range intro.Flows {
		s := f.ID
		if f.IsDefault {
			s += " ⭐"
		}
		flowStrs = append(flowStrs, s)
	}
	flowLine := strings.Join(flowStrs, "  ")
	fmt.Fprintf(w, "║    %s\n", padLine(flowLine, 68))

	fmt.Fprintln(w, "║                                                                    ║")
	fmt.Fprintln(w, "║  SAY: \"help\" for commands, \"status\" for current progress           ║")
	fmt.Fprintln(w, "╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(w)
}

func padLine(s string, width int) string {
	if len(s) > width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func formatPathValidation(w io.Writer, pv *PathValidation) {
	if pv == nil || pv.IsMatch {
		return
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "╔══════════════════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(w, "║  ⛔  PATH MISMATCH DETECTED                                                ║")
	fmt.Fprintln(w, "╠══════════════════════════════════════════════════════════════════════════╣")
	fmt.Fprintf(w, "║  PROJECT_ROOT: %-60s║\n", padLine(pv.ProjectRoot, 60))
	fmt.Fprintf(w, "║  CURRENT_CWD:   %-60s║\n", padLine(pv.CurrentCwd, 60))
	fmt.Fprintln(w, "╠══════════════════════════════════════════════════════════════════════════╣")
	fmt.Fprintln(w, "║  ⛔  AI MUST operate in PROJECT_ROOT, not in CURRENT_CWD!                 ║")
	fmt.Fprintln(w, "║                                                                        ║")
	if pv.Warning != "" {
		fmt.Fprintf(w, "║  ⚠️  %s\n", padLine(pv.Warning, 67))
	}
	if pv.RequiredAction != "" {
		fmt.Fprintf(w, "║  REQUIRED ACTION:%-57s║\n", "")
		fmt.Fprintf(w, "║    → %-65s║\n", padLine(pv.RequiredAction, 65))
	} else {
		fmt.Fprintf(w, "║  REQUIRED ACTION:%-57s║\n", "")
		fmt.Fprintf(w, "║    → cd %-65s║\n", padLine(pv.ProjectRoot, 65))
	}
	fmt.Fprintln(w, "║                                                                        ║")
	fmt.Fprintln(w, "║  [AI: Execute 'cd {PROJECT_ROOT}' BEFORE any other commands]              ║")
	fmt.Fprintln(w, "╚══════════════════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(w, "")
}

func FormatJSON(w io.Writer, result *ProcRunResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	_, err = w.Write(data)
	if err == nil {
		w.Write([]byte("\n"))
	}
	return err
}

func FormatText(w io.Writer, result *ProcRunResult) error {
	if result.TeamIntro != nil {
		formatTeamIntro(w, result.TeamIntro)
	}

	current := result.Current

	if current.IsTerminal {
		return formatTerminalText(w, result)
	}

	width := 60

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%s\n", boxTop(width))
	fmt.Fprintf(w, "║  %-58s║\n", current.Name)
	fmt.Fprintf(w, "%s\n", boxMid(width))

	fmt.Fprintf(w, "║  NODE: %-52s║\n", current.Name)
	fmt.Fprintf(w, "║  TYPE: %-52s║\n", current.NodeType)
	if result.ResumedFrom != "" {
		switch result.ResumedFrom {
		case "gate-auto-advance":
			fmt.Fprintf(w, "║  📌 AUTO-ADVANCED: Passed gate, moved to next node%-14s║\n", "")
		case "auto-advance":
			fmt.Fprintf(w, "║  📌 AUTO-ADVANCED: Moved to next node%-22s║\n", "")
		default:
			fmt.Fprintf(w, "║  📌 RESUMED: Auto-resumed from last session%-18s║\n", "")
		}
	}
	if result.Workspace != "" {
		fmt.Fprintf(w, "║  WORKSPACE: %-46s║\n", result.Workspace)
	}
	if result.ProjectRoot != "" {
		fmt.Fprintf(w, "║  PROJECT_ROOT: %-43s║\n", result.ProjectRoot)
	}
	if result.TeamRoot != "" && result.TeamRoot != result.ProjectRoot {
		fmt.Fprintf(w, "║  TEAM_ROOT: %-46s║\n", result.TeamRoot)
		fmt.Fprintf(w, "║  (config from TEAM_ROOT, work in PROJECT_ROOT)%-20s║\n", "")
	}

	if result.PathValidation != nil {
		formatPathValidation(w, result.PathValidation)
	}
	if current.Role != "" {
		roleLabel := current.RoleName
		if roleLabel == "" {
			roleLabel = current.Role
		}
		if current.Alias != "" {
			roleLabel = current.Alias
			if current.AliasEn != "" {
				roleLabel += "(" + current.AliasEn + ")"
			}
			roleLabel += " · " + current.RoleName
		}
		if current.Principal {
			roleLabel += " ⭐"
		}
		fmt.Fprintf(w, "║  ROLE: %-52s║\n", roleLabel)
	}
	if current.Persona != "" {
		fmt.Fprintf(w, "║  PERSONA: %-49s║\n", current.Persona)
	}
	if len(current.Traits) > 0 {
		fmt.Fprintf(w, "║  TRAITS: %-50s║\n", strings.Join(current.Traits, ", "))
	}
	if current.Guidance != "" {
		fmt.Fprintf(w, "║  GUIDANCE: %-48s║\n", current.Guidance)
	}

	if current.Principal {
		fmt.Fprintf(w, "║%-60s║\n", "")
		fmt.Fprintf(w, "║  ⛔ DISPATCH ONLY — DO NOT execute tasks directly%-12s║\n", "")
		fmt.Fprintf(w, "║  You MUST classify input and dispatch to sub-roles%-13s║\n", "")
		fmt.Fprintf(w, "║  NEVER write code or modify files yourself%-19s║\n", "")
		fmt.Fprintf(w, "║  Use Task tool to dispatch, then verify deliverables%-10s║\n", "")
		fmt.Fprintf(w, "║%-60s║\n", "")
		fmt.Fprintf(w, "║  ═══ SUB-AGENT CONTEXT (include in EVERY task description) ═══%-9s║\n", "")
		fmt.Fprintf(w, "║    flow:     %-45s║\n", result.Flow.Name)
		fmt.Fprintf(w, "║    node:     %-45s║\n", current.Name)
		if len(current.Rules) > 0 {
			fmt.Fprintf(w, "║    rules for sub-agents to enforce:%-22s║\n", "")
			for _, r := range current.Rules {
				label := r.Name
				if label == "" {
					label = r.Ref
				}
				fmt.Fprintf(w, "║      - %-48s║\n", label)
			}
		}
		fmt.Fprintf(w, "║  Sub-agents MUST obey flow gates and rules above%-16s║\n", "")
		fmt.Fprintf(w, "║  GATE BREACH = task rejected, return to Triage%-21s║\n", "")
	}

	if len(current.Rules) > 0 {
		fmt.Fprintf(w, "║  RULES:%-53s║\n", "")
		for _, r := range current.Rules {
			label := r.Name
			if label == "" {
				label = r.Ref
			}
			if r.Source != "" && r.Source != "builtin" {
				label += " (" + r.Source + ")"
			}
			if r.Enforcement != "" {
				label += " [" + strings.ToUpper(r.Enforcement) + "]"
			}
			fmt.Fprintf(w, "║    - %-54s║\n", label)
			if r.Instruction != "" {
				fmt.Fprintf(w, "║      %-54s║\n", r.Instruction)
			}
			if r.RuleRef != "" {
				fmt.Fprintf(w, "║      → %-52s║\n", r.RuleRef)
			}
		}
	}

	if len(current.Tools) > 0 {
		fmt.Fprintf(w, "║  TOOLS:%-53s║\n", "")
		for _, t := range current.Tools {
			label := t.Ref
			if len(t.Commands) > 0 {
				label += " [" + strings.Join(t.Commands, ", ") + "]"
			}
			if t.Source != "" && t.Source != "builtin" {
				label += " (" + t.Source + ")"
			}
			fmt.Fprintf(w, "║    - %-54s║\n", label)
		}
	}

	if len(current.Skills) > 0 {
		fmt.Fprintf(w, "║  SKILLS:%-52s║\n", "")
		for _, s := range current.Skills {
			label := s.Ref
			if s.Source != "" && s.Source != "builtin" {
				label += " (" + s.Source + ")"
			}
			fmt.Fprintf(w, "║    - %-54s║\n", label)
		}
	}

	if len(current.Docs) > 0 {
		fmt.Fprintf(w, "║  DOCS:%-54s║\n", "")
		for _, d := range current.Docs {
			req := ""
			if d.Required {
				req = " [REQUIRED]"
			}
			fmt.Fprintf(w, "║    - %-54s║\n", d.Name+req)
			displayPath := d.Path
			unresolvedVars := extractUnresolvedVars(d.Path)
			if len(unresolvedVars) > 0 {
				hint := " (use --task <id> to resolve)"
				if len(unresolvedVars) == 1 && unresolvedVars[0] == "task_id" {
					hint = " (use: flow task ready → flow proc run --task <id> " + current.NodeID + ")"
				}
				displayPath = d.Path + hint
			}
			fmt.Fprintf(w, "║      → %-52s║\n", displayPath)
			if d.Description != "" {
				fmt.Fprintf(w, "║        %-52s║\n", d.Description)
			}
			if d.Template != "" {
				fmt.Fprintf(w, "║        template: %-38s║\n", d.Template)
			}
			for i, rule := range d.ContentRules {
				if i == 0 {
					fmt.Fprintf(w, "║        rules:%-43s║\n", "")
				}
				fmt.Fprintf(w, "║          %d. %-42s║\n", i+1, rule)
			}
		}
	}

	if len(current.GateConditions) > 0 {
		fmt.Fprintf(w, "║  GATE CONDITIONS:%-42s║\n", "")
		for _, c := range current.GateConditions {
			reqLabel := "[ADVISORY]"
			if c.Required {
				reqLabel = "[REQUIRED]"
			}
			line := c.Type
			if c.Threshold != "" {
				line += " — threshold: " + c.Threshold
			}
			fmt.Fprintf(w, "║    %s %-47s║\n", reqLabel, line)
		}

		if len(result.GateCheckResults) > 0 {
			fmt.Fprintf(w, "║%-60s║\n", "")
			fmt.Fprintf(w, "║  GATE CHECK RESULTS (automated):%-30s║\n", "")
			for _, r := range result.GateCheckResults {
				statusIcon := "✓"
				if !r.Passed {
					statusIcon = "✗"
				}
				if r.Skipped {
					statusIcon = "⊘"
				}
				modeLabel := "[AUTO]"
				if !r.Auto {
					modeLabel = "[AI]"
				}
				reqMark := ""
				if r.Required {
					reqMark = " *"
				}
				fmt.Fprintf(w, "║    %s %s %s%s\n", statusIcon, modeLabel, padLine(r.Type+reqMark, 24), padLine("", 24))
				msgLines := wrapMessage(r.Message, 58)
				for _, line := range msgLines {
					fmt.Fprintf(w, "║      %s\n", padLine(line, 56))
				}
			}
			passed, summary := GateOverallResult(result.GateCheckResults)
			overallIcon := "✓"
			if !passed {
				overallIcon = "✗"
			}
			fmt.Fprintf(w, "║%-60s║\n", "")
			fmt.Fprintf(w, "║  %s OVERALL: %s\n", overallIcon, padLine(summary, 50))
		}

		fmt.Fprintf(w, "║%-60s║\n", "")
		fmt.Fprintf(w, "║  AI BEHAVIOR RULES:%-40s║\n", "")
		fmt.Fprintf(w, "║    1. Evaluate ALL conditions%-33s║\n", "")
		fmt.Fprintf(w, "║    2. Prefer is_default branch when uncertain%-23s║\n", "")
		fmt.Fprintf(w, "║    3. Explain reasoning for non-default selection%-20s║\n", "")
		fmt.Fprintf(w, "║    4. Pause and ask user when completely uncertain%-19s║\n", "")
	}

	if len(current.ParallelBranches) > 0 {
		fmt.Fprintf(w, "║  PARALLEL BRANCHES:%-40s║\n", "")
		strategy := current.ParallelStrategy
		if strategy == "" {
			strategy = "all_success"
		}
		merge := current.MergeStrategy
		if merge == "" {
			merge = "wait_all"
		}
		fmt.Fprintf(w, "║    strategy: %-46s║\n", strategy)
		fmt.Fprintf(w, "║    merge: %-49s║\n", merge)
		for i, b := range current.ParallelBranches {
			label := fmt.Sprintf("[%d] %s", i+1, b.Name)
		if b.Name == "" {
			label = fmt.Sprintf("[%d] %s", i+1, b.NodeID)
		}
			if b.Alias != "" {
				label += " (" + b.Alias
				if b.RoleName != "" {
					label += " · " + b.RoleName
				}
				label += ")"
			} else if b.RoleName != "" {
				label += " (" + b.RoleName + ")"
			}
			if b.Name != "" {
				label += " — " + b.Name
			}
			fmt.Fprintf(w, "║    %-56s║\n", label)
		}
		fmt.Fprintf(w, "║%-60s║\n", "")
		fmt.Fprintf(w, "║  AI BEHAVIOR RULES:%-40s║\n", "")
		fmt.Fprintf(w, "║    1. Dispatch ALL branches concurrently via sub-agents%-11s║\n", "")
		fmt.Fprintf(w, "║    2. Wait for ALL branches to complete (wait_all)%-14s║\n", "")
		fmt.Fprintf(w, "║    3. If any branch fails → report to principal%-16s║\n", "")
		fmt.Fprintf(w, "║    4. Only proceed to next node when all branches pass%-10s║\n", "")
	}

	if current.SubflowRef != "" {
		fmt.Fprintf(w, "║  SUBFLOW:%-51s║\n", "")
		cleanRef := strings.TrimPrefix(current.SubflowRef, "builtin:")
		displayName := cleanRef
		if current.SubflowName != "" {
			displayName = fmt.Sprintf("%s (%s)", current.SubflowName, cleanRef)
		}
		fmt.Fprintf(w, "║    → %-54s║\n", displayName)
		fmt.Fprintf(w, "║%-60s║\n", "")
		fmt.Fprintf(w, "║  AI BEHAVIOR RULES:%-40s║\n", "")
		fmt.Fprintf(w, "║    1. Run 'flow proc run --flow %s' to enter subflow%-8s║\n", cleanRef, "")
		fmt.Fprintf(w, "║    2. Complete subflow execution, then return to parent%-8s║\n", "")
		fmt.Fprintf(w, "║    3. Use 'flow proc run' (no --flow) to resume parent%-7s║\n", "")
	}

	fmt.Fprintf(w, "%s\n", boxMid(width))

	if len(result.NextOptions) > 0 {
		fmt.Fprintf(w, "║  NEXT OPTIONS:%-45s║\n", "")
		for i, opt := range result.NextOptions {
			displayName := opt.Name
			if displayName == "" {
				displayName = opt.NodeID
			}
			label := fmt.Sprintf("[%d] → %s", i+1, displayName)
			if opt.Role != "" {
				label += " (" + opt.Role + ")"
			}
			if opt.Condition != nil {
				label += " [" + *opt.Condition + "]"
			}
			if opt.IsDefault {
				label += " [DEFAULT]"
			}
			fmt.Fprintf(w, "║    %-56s║\n", label)
		}
	}

	fmt.Fprintf(w, "%s\n", boxMid(width))

	if len(result.NextOptions) > 0 {
		defaultNodeID := ""
		defaultNodeName := ""
		for _, opt := range result.NextOptions {
			if opt.IsDefault {
				defaultNodeID = opt.NodeID
				defaultNodeName = opt.Name
				break
			}
		}
		if defaultNodeID == "" && len(result.NextOptions) > 0 {
			defaultNodeID = result.NextOptions[0].NodeID
			defaultNodeName = result.NextOptions[0].Name
		}
		if defaultNodeID != "" {
			nextLabel := defaultNodeID
			if defaultNodeName != "" {
				nextLabel = defaultNodeName + " (" + defaultNodeID + ")"
			}
			fmt.Fprintf(w, "║  NEXT STEP: flow proc run %-31s║\n", nextLabel)
		}
	}

	fmt.Fprintf(w, "%s\n", boxBottom(width))
	fmt.Fprintf(w, "\n")

	return nil
}

func formatTerminalText(w io.Writer, result *ProcRunResult) error {
	width := 60
	current := result.Current

	statusIcon := "✓"
	if current.TerminalStatus != "success" {
		statusIcon = "✗"
	}

	msg := current.Name
	if current.TerminalMessage != "" {
		msg = current.TerminalMessage
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%s\n", boxTop(width))
	fmt.Fprintf(w, "║  %s %s\n", statusIcon, padRight(msg, 57))
	fmt.Fprintf(w, "║%-60s║\n", "")
	fmt.Fprintf(w, "%s\n", boxMid(width))
	fmt.Fprintf(w, "║  SESSION RESET RULES:%-39s║\n", "")
	fmt.Fprintf(w, "║    1. Check if there are any in_progress tasks%-16s║\n", "")
	fmt.Fprintf(w, "║    2. If no in_progress tasks, reset to Triage role%-12s║\n", "")
	fmt.Fprintf(w, "║    3. Classify the new input and dispatch to the flow%-11s║\n", "")
	fmt.Fprintf(w, "║    4. Never skip Triage — all new input must be classified%-5s║\n", "")
	fmt.Fprintf(w, "%s\n", boxBottom(width))
	fmt.Fprintf(w, "\n")

	return nil
}

func boxTop(w int) string {
	return "╔" + strings.Repeat("═", w) + "╗"
}

func boxMid(w int) string {
	return "╠" + strings.Repeat("═", w) + "╣"
}

func boxBottom(w int) string {
	return "╚" + strings.Repeat("═", w) + "╝"
}

func padRight(s string, width int) string {
	if width <= 0 {
		return s
	}
	for len(s) < width {
		s += " "
	}
	if len(s) > width {
		return s[:width]
	}
	return s
}

var unresolvedVarRe = regexp.MustCompile(`\{(\w+)\}`)

func wrapMessage(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	var lines []string
	for len(s) > width {
		idx := width
		for i := width; i > 0; i-- {
			if s[i] == ' ' || s[i] == ',' || s[i] == '\n' {
				idx = i + 1
				break
			}
		}
		if idx == width {
			idx = width
		}
		lines = append(lines, s[:idx])
		s = s[idx:]
	}
	if s != "" {
		lines = append(lines, s)
	}
	return lines
}

func extractUnresolvedVars(path string) []string {
	matches := unresolvedVarRe.FindAllStringSubmatch(path, -1)
	vars := make([]string, 0, len(matches))
	for _, m := range matches {
		vars = append(vars, m[1])
	}
	return vars
}
