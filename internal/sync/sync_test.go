package sync

import (
	"testing"
)

func TestClassifyIssue_BugLabel(t *testing.T) {
	issue := &Issue{
		Title:  "App crashes on startup",
		Labels: []Label{{Name: "bug"}},
	}
	issueType, priority := ClassifyIssue(issue)
	if issueType != IssueTypeBug {
		t.Errorf("expected bug, got %s", issueType)
	}
	if priority != 2 {
		t.Errorf("expected P2, got P%d", priority)
	}
}

func TestClassifyIssue_FeatureWithPriority(t *testing.T) {
	issue := &Issue{
		Title:  "Add dark mode",
		Labels: []Label{{Name: "enhancement"}, {Name: "p1"}},
	}
	issueType, priority := ClassifyIssue(issue)
	if issueType != IssueTypeFeature {
		t.Errorf("expected feature, got %s", issueType)
	}
	if priority != 1 {
		t.Errorf("expected P1, got P%d", priority)
	}
}

func TestClassifyIssue_InferFromTitle(t *testing.T) {
	tests := []struct {
		title    string
		expected IssueType
	}{
		{"Fix login crash", IssueTypeBug},
		{"Add export feature", IssueTypeFeature},
		{"Update API documentation", IssueTypeDocs},
		{"Refactor database layer", IssueTypeChange},
		{"Random task", IssueTypeUnknown},
	}
	for _, tt := range tests {
		issue := &Issue{Title: tt.title}
		issueType, _ := ClassifyIssue(issue)
		if issueType != tt.expected {
			t.Errorf("ClassifyIssue(%q) = %s, want %s", tt.title, issueType, tt.expected)
		}
	}
}

func TestExternalRef(t *testing.T) {
	ref := ExternalRef("owner/repo", 42)
	if ref != "github:owner/repo#42" {
		t.Errorf("expected github:owner/repo#42, got %s", ref)
	}
}

func TestParseExternalRef(t *testing.T) {
	repo, number, ok := ParseExternalRef("github:owner/repo#42")
	if !ok {
		t.Fatal("expected ok")
	}
	if repo != "owner/repo" {
		t.Errorf("expected owner/repo, got %s", repo)
	}
	if number != 42 {
		t.Errorf("expected 42, got %d", number)
	}
}

func TestParseExternalRef_Invalid(t *testing.T) {
	_, _, ok := ParseExternalRef("jira:PROJECT-123")
	if ok {
		t.Error("expected not ok for non-github ref")
	}

	_, _, ok = ParseExternalRef("github:invalid")
	if ok {
		t.Error("expected not ok for missing number")
	}
}

func TestParseGithubRepo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git@github.com:owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo", "owner/repo"},
		{"git@gitlab.com:owner/repo.git", ""},
		{"", ""},
	}
	for _, tt := range tests {
		result := parseGithubRepo(tt.input)
		if result != tt.expected {
			t.Errorf("parseGithubRepo(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSyncResult_Summary(t *testing.T) {
	result := &SyncResult{
		Created: []string{"#1", "#2"},
		Skipped: []string{"#3"},
	}
	summary := result.Summary()
	if summary != "2 created, 1 skipped (already synced)" {
		t.Errorf("unexpected summary: %s", summary)
	}
}

func TestSyncResult_EmptySummary(t *testing.T) {
	result := &SyncResult{}
	summary := result.Summary()
	if summary != "No issues to sync" {
		t.Errorf("unexpected summary: %s", summary)
	}
}

func TestFormatIssueSummary(t *testing.T) {
	issue := &Issue{
		Number: 42,
		Title:  "Test issue",
		State:  "open",
		Labels: []Label{{Name: "bug"}, {Name: "p1"}},
		Assignees: []Assignee{{Login: "dev1"}},
		Milestone: &Milestone{Title: "v1.0"},
	}
	summary := FormatIssueSummary(issue)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	if !contains(summary, "#42") || !contains(summary, "Test issue") {
		t.Errorf("summary missing key info: %s", summary)
	}
}

func TestTypePrefixV1(t *testing.T) {
	tests := []struct {
		t        IssueType
		expected string
	}{
		{IssueTypeFeature, "F"},
		{IssueTypeBug, "B"},
		{IssueTypeChange, "C"},
		{IssueTypeAnalysis, "A"},
		{IssueTypeDocs, "D"},
		{IssueTypeHotfix, "H"},
		{IssueTypeUnknown, "T"},
	}
	for _, tt := range tests {
		result := typePrefixV1(tt.t)
		if result != tt.expected {
			t.Errorf("typePrefixV1(%s) = %s, want %s", tt.t, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
