package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/bd"
)

type SyncResult struct {
	Created  []string
	Skipped  []string
	Failed   []string
	Updated  []string
}

func (r *SyncResult) Summary() string {
	parts := []string{}
	if len(r.Created) > 0 {
		parts = append(parts, fmt.Sprintf("%d created", len(r.Created)))
	}
	if len(r.Updated) > 0 {
		parts = append(parts, fmt.Sprintf("%d updated", len(r.Updated)))
	}
	if len(r.Skipped) > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped (already synced)", len(r.Skipped)))
	}
	if len(r.Failed) > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", len(r.Failed)))
	}
	if len(parts) == 0 {
		return "No issues to sync"
	}
	return strings.Join(parts, ", ")
}

type SyncOptions struct {
	Repo      string
	State     string
	Label     string
	Version   string
	Overwrite bool
	DryRun    bool
}

func SyncFromGithub(opts SyncOptions) (*SyncResult, error) {
	issues, err := FetchIssues(opts.Repo, opts.State, opts.Label)
	if err != nil {
		return nil, fmt.Errorf("fetch issues: %w", err)
	}

	if len(issues) == 0 {
		return &SyncResult{}, nil
	}

	result := &SyncResult{}

	switch opts.Version {
	case "v1":
		syncV1(issues, opts, result)
	case "v2", "v3":
		syncBeads(issues, opts, result)
	default:
		return nil, fmt.Errorf("unsupported version: %s", opts.Version)
	}

	return result, nil
}

func syncV1(issues []Issue, opts SyncOptions, result *SyncResult) {
	taskPoolPath := filepath.Join(".team", "task-pool.md")

	if _, err := os.Stat(taskPoolPath); os.IsNotExist(err) {
		result.Failed = append(result.Failed, "task-pool.md not found (run: flow init --v1)")
		return
	}

	existingRefs := loadV1ExternalRefs(taskPoolPath)

	for _, issue := range issues {
		ref := ExternalRef(opts.Repo, issue.Number)

		if _, exists := existingRefs[ref]; exists && !opts.Overwrite {
			result.Skipped = append(result.Skipped, fmt.Sprintf("#%d", issue.Number))
			continue
		}

		issueType, priority := ClassifyIssue(&issue)
		prefix := typePrefixV1(issueType)

		entry := fmt.Sprintf("\n## %s%03d — %s\n\n", prefix, issue.Number, issue.Title)
		entry += fmt.Sprintf("- **Source**: %s\n", issue.URL)
		entry += fmt.Sprintf("- **External Ref**: %s\n", ref)
		entry += fmt.Sprintf("- **Type**: %s\n", string(issueType))
		entry += fmt.Sprintf("- **Priority**: P%d\n", priority)
		entry += fmt.Sprintf("- **Status**: %s\n", issue.State)

		if len(issue.Labels) > 0 {
			labels := make([]string, 0, len(issue.Labels))
			for _, l := range issue.Labels {
				labels = append(labels, l.Name)
			}
			entry += fmt.Sprintf("- **Labels**: %s\n", strings.Join(labels, ", "))
		}

		if issue.Body != "" {
			body := issue.Body
			if len(body) > 500 {
				body = body[:500] + "..."
			}
			entry += fmt.Sprintf("\n%s\n", body)
		}

		if opts.DryRun {
			fmt.Printf("  [DRY] Would add %s%03d: %s\n", prefix, issue.Number, issue.Title)
			result.Created = append(result.Created, fmt.Sprintf("#%d", issue.Number))
			continue
		}

		f, err := os.OpenFile(taskPoolPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("#%d: %v", issue.Number, err))
			continue
		}
		if _, err := f.WriteString(entry); err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("#%d: %v", issue.Number, err))
			f.Close()
			continue
		}
		f.Close()

		result.Created = append(result.Created, fmt.Sprintf("#%d", issue.Number))
	}
}

func loadV1ExternalRefs(taskPoolPath string) map[string]bool {
	refs := make(map[string]bool)
	data, err := os.ReadFile(taskPoolPath)
	if err != nil {
		return refs
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, "External Ref**:") {
			parts := strings.SplitN(line, "**External Ref**:", 2)
			if len(parts) == 2 {
				ref := strings.TrimSpace(parts[1])
				refs[ref] = true
			}
		}
	}
	return refs
}

func typePrefixV1(t IssueType) string {
	switch t {
	case IssueTypeFeature:
		return "F"
	case IssueTypeBug:
		return "B"
	case IssueTypeChange:
		return "C"
	case IssueTypeAnalysis:
		return "A"
	case IssueTypeDocs:
		return "D"
	case IssueTypeHotfix:
		return "H"
	default:
		return "T"
	}
}

func syncBeads(issues []Issue, opts SyncOptions, result *SyncResult) {
	if !bd.IsAvailable() {
		result.Failed = append(result.Failed, "bd CLI not found (run: flow init)")
		return
	}

	existingRefs := loadBeadsExternalRefs()

	for _, issue := range issues {
		ref := ExternalRef(opts.Repo, issue.Number)

		if taskID, exists := existingRefs[ref]; exists {
			if opts.Overwrite {
				updateBeadsTask(taskID, &issue, ref, result)
			} else {
				result.Skipped = append(result.Skipped, fmt.Sprintf("#%d → %s", issue.Number, taskID))
			}
			continue
		}

		issueType, priority := ClassifyIssue(&issue)
		typeFlag := string(issueType)

		title := fmt.Sprintf("[GH#%d] %s", issue.Number, issue.Title)

		if opts.DryRun {
			fmt.Printf("  [DRY] Would create task: %s (type=%s, P%d, ref=%s)\n",
				title, typeFlag, priority, ref)
			result.Created = append(result.Created, fmt.Sprintf("#%d", issue.Number))
			continue
		}

		args := []string{
			"create", title,
			"-t", typeFlag,
			"-p", fmt.Sprintf("%d", priority),
			"--external-ref", ref,
		}

		if len(issue.Labels) > 0 {
			for _, l := range issue.Labels {
				args = append(args, "--add-label", l.Name)
			}
		}
		args = append(args, "--add-label", "source:github")

		output, err := bd.Run(args...)
		if err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("#%d: %v", issue.Number, err))
			continue
		}

		taskID := extractTaskID(output)
		if taskID != "" {
			if issue.Body != "" {
				noteArgs := []string{"note", taskID, "--note", truncateBody(issue.Body, 1000)}
				bd.Run(noteArgs...)
			}
			result.Created = append(result.Created, fmt.Sprintf("#%d → %s", issue.Number, taskID))
		} else {
			result.Created = append(result.Created, fmt.Sprintf("#%d", issue.Number))
		}
	}
}

func loadBeadsExternalRefs() map[string]string {
	refs := make(map[string]string)

	if !bd.IsAvailable() {
		return refs
	}

	output, err := bd.Run("list", "--json")
	if err != nil {
		return refs
	}

	var tasks []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &tasks); err != nil {
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				taskID := fields[0]
				rest := strings.Join(fields[1:], " ")
				if strings.Contains(rest, "github:") {
					parts := strings.SplitN(rest, "github:", 2)
					if len(parts) == 2 {
						ref := "github:" + strings.Fields(parts[1])[0]
						refs[ref] = taskID
					}
				}
			}
		}
		return refs
	}

	for _, task := range tasks {
		taskID, _ := task["id"].(string)
		extRef, _ := task["external_ref"].(string)
		if taskID != "" && extRef != "" && strings.HasPrefix(extRef, "github:") {
			refs[extRef] = taskID
		}
	}

	return refs
}

func updateBeadsTask(taskID string, issue *Issue, ref string, result *SyncResult) {
	args := []string{"update", taskID}

	if issue.State == "closed" {
		args = append(args, "--status", "closed")
	}

	_, err := bd.Run(args...)
	if err != nil {
		result.Failed = append(result.Failed, fmt.Sprintf("#%d (update %s): %v", issue.Number, taskID, err))
		return
	}

	result.Updated = append(result.Updated, fmt.Sprintf("#%d → %s", issue.Number, taskID))
}

func extractTaskID(output string) string {
	output = strings.TrimSpace(output)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Created ") || strings.Contains(line, "created") {
			fields := strings.Fields(line)
			for _, f := range fields {
				if len(f) >= 3 && strings.Contains(f, "-") {
					return f
				}
			}
		}
	}
	lastLine := lines[len(lines)-1]
	fields := strings.Fields(lastLine)
	if len(fields) > 0 {
		last := fields[len(fields)-1]
		if len(last) >= 3 {
			return last
		}
	}
	return ""
}

func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "..."
}
