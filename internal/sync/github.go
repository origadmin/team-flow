package sync

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Issue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	URL       string `json:"url"`
	Labels    []Label `json:"labels"`
	Assignees []Assignee `json:"assignees"`
	Milestone *Milestone `json:"milestone"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Assignee struct {
	Login string `json:"login"`
}

type Milestone struct {
	Title string `json:"title"`
}

var ghPath string

func FindGhPath() string {
	if ghPath != "" {
		return ghPath
	}
	path, err := exec.LookPath("gh")
	if err == nil {
		ghPath = path
	}
	return ghPath
}

func IsGhAvailable() bool {
	return FindGhPath() != ""
}

func FetchIssues(repo string, state string, label string) ([]Issue, error) {
	if !IsGhAvailable() {
		return nil, fmt.Errorf("gh CLI not found. Install: https://cli.github.com")
	}

	args := []string{"issue", "list", "--repo", repo, "--json",
		"number,title,body,state,url,labels,assignees,milestone,createdAt,updatedAt",
		"--limit", "100"}

	if state != "" {
		args = append(args, "--state", state)
	} else {
		args = append(args, "--state", "open")
	}

	if label != "" {
		args = append(args, "--label", label)
	}

	cmd := exec.Command(FindGhPath(), args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh issue list failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh issue list failed: %w", err)
	}

	var issues []Issue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("parse gh output: %w", err)
	}

	return issues, nil
}

func FetchIssue(repo string, number int) (*Issue, error) {
	if !IsGhAvailable() {
		return nil, fmt.Errorf("gh CLI not found")
	}

	args := []string{"issue", "view", fmt.Sprintf("%d", number), "--repo", repo,
		"--json", "number,title,body,state,url,labels,assignees,milestone,createdAt,updatedAt"}

	cmd := exec.Command(FindGhPath(), args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh issue view failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh issue view failed: %w", err)
	}

	var issue Issue
	if err := json.Unmarshal(output, &issue); err != nil {
		return nil, fmt.Errorf("parse gh output: %w", err)
	}

	return &issue, nil
}

type IssueType string

const (
	IssueTypeFeature  IssueType = "feature"
	IssueTypeBug      IssueType = "bug"
	IssueTypeChange   IssueType = "change"
	IssueTypeAnalysis IssueType = "analysis"
	IssueTypeDocs     IssueType = "docs"
	IssueTypeHotfix   IssueType = "hotfix"
	IssueTypeUnknown  IssueType = "task"
)

var labelTypeMap = map[string]IssueType{
	"bug":            IssueTypeBug,
	"enhancement":    IssueTypeFeature,
	"feature":        IssueTypeFeature,
	"feature-request": IssueTypeFeature,
	"change":         IssueTypeChange,
	"breaking-change": IssueTypeChange,
	"analysis":       IssueTypeAnalysis,
	"docs":           IssueTypeDocs,
	"documentation":  IssueTypeDocs,
	"hotfix":         IssueTypeHotfix,
	"critical":       IssueTypeHotfix,
	"p0":             IssueTypeHotfix,
}

var priorityMap = map[string]int{
	"critical": 0,
	"p0":       0,
	"urgent":   0,
	"p1":       1,
	"high":     1,
	"p2":       2,
	"medium":   2,
	"p3":       3,
	"low":      3,
	"p4":       4,
}

func ClassifyIssue(issue *Issue) (IssueType, int) {
	issueType := IssueTypeUnknown
	priority := 2

	for _, label := range issue.Labels {
		name := strings.ToLower(label.Name)
		if t, ok := labelTypeMap[name]; ok {
			issueType = t
		}
		if p, ok := priorityMap[name]; ok {
			priority = p
		}
	}

	if issueType == IssueTypeUnknown {
		issueType = inferTypeFromTitle(issue.Title)
	}

	return issueType, priority
}

var bugPatterns = regexp.MustCompile(`(?i)\b(bug|fix|crash|error|broken|fail|issue|wrong|regression)\b`)
var featurePatterns = regexp.MustCompile(`(?i)\b(add|new|feature|support|implement|create|enable)\b`)
var changePatterns = regexp.MustCompile(`(?i)\b(change|update|refactor|improve|migrate|upgrade|rename)\b`)
var docsPatterns = regexp.MustCompile(`(?i)\b(doc|docs|documentation|readme|guide|tutorial|example)\b`)

func inferTypeFromTitle(title string) IssueType {
	if bugPatterns.MatchString(title) {
		return IssueTypeBug
	}
	if docsPatterns.MatchString(title) {
		return IssueTypeDocs
	}
	if featurePatterns.MatchString(title) {
		return IssueTypeFeature
	}
	if changePatterns.MatchString(title) {
		return IssueTypeChange
	}
	return IssueTypeUnknown
}

func ExternalRef(repo string, number int) string {
	return fmt.Sprintf("github:%s#%d", repo, number)
}

func ParseExternalRef(ref string) (repo string, number int, ok bool) {
	if !strings.HasPrefix(ref, "github:") {
		return "", 0, false
	}
	rest := strings.TrimPrefix(ref, "github:")
	parts := strings.SplitN(rest, "#", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	repo = parts[0]
	_, err := fmt.Sscanf(parts[1], "%d", &number)
	if err != nil {
		return "", 0, false
	}
	return repo, number, true
}

func FormatIssueSummary(issue *Issue) string {
	labels := make([]string, 0, len(issue.Labels))
	for _, l := range issue.Labels {
		labels = append(labels, l.Name)
	}
	labelStr := strings.Join(labels, ", ")
	if labelStr == "" {
		labelStr = "(none)"
	}

	assignees := make([]string, 0, len(issue.Assignees))
	for _, a := range issue.Assignees {
		assignees = append(assignees, a.Login)
	}
	assigneeStr := strings.Join(assignees, ", ")
	if assigneeStr == "" {
		assigneeStr = "(unassigned)"
	}

	milestone := "(none)"
	if issue.Milestone != nil {
		milestone = issue.Milestone.Title
	}

	return fmt.Sprintf("#%-4d [%s] %s\n       Labels: %s | Milestone: %s | Assignees: %s",
		issue.Number, issue.State, issue.Title, labelStr, milestone, assigneeStr)
}
