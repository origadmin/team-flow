package proc

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

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
	current := result.Current

	if current.IsTerminal {
		return formatTerminalText(w, result)
	}

	width := 60

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%s\n", boxTop(width))
	fmt.Fprintf(w, "║  %-58s║\n", current.Name)
	fmt.Fprintf(w, "%s\n", boxMid(width))

	fmt.Fprintf(w, "║  NODE: %-52s║\n", current.NodeID)
	fmt.Fprintf(w, "║  TYPE: %-52s║\n", current.NodeType)
	if current.Role != "" {
		fmt.Fprintf(w, "║  ROLE: %-52s║\n", current.Role)
	}

	if len(current.Rules) > 0 {
		fmt.Fprintf(w, "║  RULES:%-53s║\n", "")
		for _, r := range current.Rules {
			label := r.Ref
			if r.Name != "" {
				label = r.Name
			}
			if r.Source != "" && r.Source != "builtin" {
				label += " (" + r.Source + ")"
			}
			if r.Enforcement != "" {
				label += " [" + strings.ToUpper(r.Enforcement) + "]"
			}
			fmt.Fprintf(w, "║    - %-54s║\n", label)
			if r.Description != "" {
				desc := r.Description
				for len(desc) > 52 {
					fmt.Fprintf(w, "║      %-54s║\n", desc[:52])
					desc = desc[52:]
				}
				if desc != "" {
					fmt.Fprintf(w, "║      %-54s║\n", desc)
				}
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
			fmt.Fprintf(w, "║      → %-52s║\n", d.Path)
			if d.Description != "" {
				fmt.Fprintf(w, "║        %-52s║\n", d.Description)
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
		fmt.Fprintf(w, "║%-60s║\n", "")
		fmt.Fprintf(w, "║  AI BEHAVIOR RULES:%-40s║\n", "")
		fmt.Fprintf(w, "║    1. Evaluate ALL conditions%-33s║\n", "")
		fmt.Fprintf(w, "║    2. Prefer is_default branch when uncertain%-23s║\n", "")
		fmt.Fprintf(w, "║    3. Explain reasoning for non-default selection%-20s║\n", "")
		fmt.Fprintf(w, "║    4. Pause and ask user when completely uncertain%-19s║\n", "")
	}

	fmt.Fprintf(w, "%s\n", boxMid(width))

	if len(result.NextOptions) > 0 {
		fmt.Fprintf(w, "║  NEXT OPTIONS:%-45s║\n", "")
		for i, opt := range result.NextOptions {
			label := fmt.Sprintf("[%d] → %s", i+1, opt.NodeID)
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
		for _, opt := range result.NextOptions {
			if opt.IsDefault {
				defaultNodeID = opt.NodeID
				break
			}
		}
		if defaultNodeID == "" && len(result.NextOptions) > 0 {
			defaultNodeID = result.NextOptions[0].NodeID
		}
		if defaultNodeID != "" {
			fmt.Fprintf(w, "║  NEXT STEP: flow proc run %-31s║\n", defaultNodeID)
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
