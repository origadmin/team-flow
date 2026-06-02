package proc

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
	"github.com/spf13/cobra"
)

func validFlowJSON() []byte {
	f := &flow.Flow{
		Version:  "v3",
		Metadata: flow.FlowMetadata{Name: "test-proc", Description: "A test process"},
		Nodes: []flow.FlowNode{
			{ID: "triage", Type: flow.NodeTypePhase, Name: "Triage"},
			{ID: "done", Type: flow.NodeTypeTerminal, Name: "Done"},
		},
		Edges: []flow.FlowEdge{
			{From: "triage", To: "done"},
		},
	}
	data, _ := json.MarshalIndent(f, "", "  ")
	return data
}

func invalidFlowJSON() []byte {
	return []byte(`{
		"version": "",
		"metadata": {},
		"nodes": [],
		"edges": []
	}`)
}

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	presetDir := filepath.Join(dir, "v3", "flows")
	if err := os.MkdirAll(presetDir, 0755); err != nil {
		t.Fatalf("create preset dir: %v", err)
	}

	projectDir := filepath.Join(dir, ".team", "flows")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(presetDir, "standard.json"), validFlowJSON(), 0644); err != nil {
		t.Fatalf("write preset process: %v", err)
	}

	if err := os.WriteFile(filepath.Join(projectDir, "custom.json"), validFlowJSON(), 0644); err != nil {
		t.Fatalf("write project process: %v", err)
	}

	return dir
}

func TestListProcs_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	entries := listProcs(dir)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries in empty dir, got %d", len(entries))
	}
}

func TestListProcs_PresetAndProject(t *testing.T) {
	dir := setupTestDir(t)
	entries := listProcs(dir)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	presetFound := false
	projectFound := false
	for _, e := range entries {
		if e.Name == "standard" && e.Source == "preset" {
			presetFound = true
		}
		if e.Name == "custom" && e.Source == "project" {
			projectFound = true
		}
	}

	if !presetFound {
		t.Error("expected preset process 'standard' not found")
	}
	if !projectFound {
		t.Error("expected project process 'custom' not found")
	}
}

func TestListProcs_PresetOnly(t *testing.T) {
	dir := t.TempDir()
	presetDir := filepath.Join(dir, "v3", "flows")
	os.MkdirAll(presetDir, 0755)
	os.WriteFile(filepath.Join(presetDir, "bugfix.json"), validFlowJSON(), 0644)

	entries := listProcs(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "bugfix" {
		t.Errorf("expected name bugfix, got %s", entries[0].Name)
	}
	if entries[0].Source != "preset" {
		t.Errorf("expected source preset, got %s", entries[0].Source)
	}
}

func TestPrintProcs_Empty(t *testing.T) {
	var buf bytes.Buffer
	printProcs(&buf, nil)
	output := buf.String()
	if !strings.Contains(output, "No processes found") {
		t.Errorf("expected 'No processes found' in output, got: %s", output)
	}
}

func TestPrintProcs_WithEntries(t *testing.T) {
	var buf bytes.Buffer
	entries := []ProcEntry{
		{Name: "test-proc", Source: "preset", Path: "/v3/flows/test-proc.json"},
		{Name: "my-proc", Source: "project", Path: "/.team/flows/my-proc.json"},
	}
	printProcs(&buf, entries)
	output := buf.String()

	if !strings.Contains(output, "test-proc") {
		t.Error("output should contain process name 'test-proc'")
	}
	if !strings.Contains(output, "preset") {
		t.Error("output should contain source 'preset'")
	}
	if !strings.Contains(output, "my-proc") {
		t.Error("output should contain process name 'my-proc'")
	}
	if !strings.Contains(output, "Total: 2 flow(s)") {
		t.Error("output should contain total count")
	}
}

func TestShowProc(t *testing.T) {
	f := &flow.Flow{
		Version:  "v3",
		Metadata: flow.FlowMetadata{Name: "test-proc", Description: "A test process", Author: "dev"},
		Nodes: []flow.FlowNode{
			{ID: "triage", Type: flow.NodeTypePhase, Name: "Triage"},
			{ID: "done", Type: flow.NodeTypeTerminal, Name: "Done"},
		},
		Edges: []flow.FlowEdge{
			{From: "triage", To: "done"},
		},
	}

	var buf bytes.Buffer
	showProc(&buf, f)
	output := buf.String()

	if !strings.Contains(output, "Name:        test-proc") {
		t.Errorf("expected name in output, got: %s", output)
	}
	if !strings.Contains(output, "Version:     v3") {
		t.Errorf("expected version in output, got: %s", output)
	}
	if !strings.Contains(output, "Nodes:       2") {
		t.Errorf("expected node count in output, got: %s", output)
	}
}

func TestShowProc_WithTagsAndConfig(t *testing.T) {
	autoDispatch := true
	f := &flow.Flow{
		Version:  "v3",
		Metadata: flow.FlowMetadata{Name: "tagged-proc", Tags: []string{"bug", "fast-track"}},
		Config:   &flow.FlowConfig{TaskType: flow.TaskTypeBug, AutoDispatch: &autoDispatch},
		Nodes: []flow.FlowNode{
			{ID: "start", Type: flow.NodeTypePhase, Name: "Start"},
			{ID: "end", Type: flow.NodeTypeTerminal, Name: "End"},
		},
		Edges: []flow.FlowEdge{
			{From: "start", To: "end", Type: flow.EdgeTypeConditional},
		},
	}

	var buf bytes.Buffer
	showProc(&buf, f)
	output := buf.String()

	if !strings.Contains(output, "Tags:        bug, fast-track") {
		t.Errorf("expected tags in output, got: %s", output)
	}
	if !strings.Contains(output, "Task Type:   bug") {
		t.Errorf("expected task type in output, got: %s", output)
	}
}

func TestPrintValidation_Valid(t *testing.T) {
	f := &flow.Flow{
		Version:  "v3",
		Metadata: flow.FlowMetadata{Name: "valid-proc"},
		Nodes: []flow.FlowNode{
			{ID: "start", Type: flow.NodeTypePhase, Name: "Start"},
			{ID: "done", Type: flow.NodeTypeTerminal, Name: "Done"},
		},
		Edges: []flow.FlowEdge{{From: "start", To: "done"}},
	}

	result := flow.ValidateFlow(f)

	var buf bytes.Buffer
	printValidation(&buf, f, result)
	output := buf.String()

	if !strings.Contains(output, "✓ Process is valid") {
		t.Errorf("expected valid indicator, got: %s", output)
	}
}

func TestPrintValidation_Invalid(t *testing.T) {
	f := &flow.Flow{
		Version:  "",
		Metadata: flow.FlowMetadata{},
		Nodes:    []flow.FlowNode{},
		Edges:    []flow.FlowEdge{},
	}

	result := flow.ValidateFlow(f)

	var buf bytes.Buffer
	printValidation(&buf, f, result)
	output := buf.String()

	if !strings.Contains(output, "✗ Process is invalid") {
		t.Errorf("expected invalid indicator, got: %s", output)
	}
}

func TestPrintValidation_WithWarnings(t *testing.T) {
	f := &flow.Flow{
		Version:  "v3",
		Metadata: flow.FlowMetadata{Name: "orphan-proc"},
		Nodes: []flow.FlowNode{
			{ID: "start", Type: flow.NodeTypePhase, Name: "Start"},
			{ID: "orphan", Type: flow.NodeTypePhase, Name: "Orphan"},
			{ID: "done", Type: flow.NodeTypeTerminal, Name: "Done"},
		},
		Edges: []flow.FlowEdge{{From: "start", To: "done"}},
	}

	result := flow.ValidateFlow(f)

	var buf bytes.Buffer
	printValidation(&buf, f, result)
	output := buf.String()

	if !strings.Contains(output, "Warnings") {
		t.Errorf("expected warnings section, got: %s", output)
	}
}

func TestResolveProcPath_ByProjectID(t *testing.T) {
	dir := setupTestDir(t)

	result := resolveProcPath(dir, "custom")
	expected := filepath.Join(dir, ".team", "flows", "custom.json")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestResolveProcPath_ByPresetID(t *testing.T) {
	dir := setupTestDir(t)

	result := resolveProcPath(dir, "standard")
	expected := filepath.Join(dir, "v3", "flows", "standard.json")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestResolveProcPath_AbsolutePath(t *testing.T) {
	dir := setupTestDir(t)
	absPath := filepath.Join(dir, ".team", "flows", "custom.json")

	result := resolveProcPath(dir, absPath)
	if result != absPath {
		t.Errorf("expected %s, got %s", absPath, result)
	}
}

func TestResolveProcPath_NotFound(t *testing.T) {
	dir := setupTestDir(t)

	result := resolveProcPath(dir, "nonexistent")
	if result != "" {
		t.Errorf("expected empty string for nonexistent process, got %s", result)
	}
}

func TestCreateProc_Blank(t *testing.T) {
	dir := t.TempDir()

	targetPath, err := createProc(dir, "my-proc", "")
	if err != nil {
		t.Fatalf("createProc returned error: %v", err)
	}

	expectedPath := filepath.Join(dir, ".team", "flows", "my-proc.json")
	if targetPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, targetPath)
	}

	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("created file should exist: %v", err)
	}

	f, err := flow.ParseFlowFile(targetPath)
	if err != nil {
		t.Fatalf("parse created process: %v", err)
	}

	if f.Metadata.Name != "my-proc" {
		t.Errorf("expected name my-proc, got %s", f.Metadata.Name)
	}
	if f.Version != "v3" {
		t.Errorf("expected version v3, got %s", f.Version)
	}

	result := flow.ValidateFlow(f)
	if !result.Valid {
		t.Errorf("blank process should be valid, got errors: %v", result.Errors)
	}
}

func TestCreateProc_FromTemplate(t *testing.T) {
	dir := setupTestDir(t)

	targetPath, err := createProc(dir, "from-template", "standard")
	if err != nil {
		t.Fatalf("createProc from template returned error: %v", err)
	}

	f, err := flow.ParseFlowFile(targetPath)
	if err != nil {
		t.Fatalf("parse created process: %v", err)
	}

	if f.Metadata.Name != "test-proc" {
		t.Errorf("expected name from template 'test-proc', got %s", f.Metadata.Name)
	}
}

func TestCreateProc_AlreadyExists(t *testing.T) {
	dir := setupTestDir(t)

	_, err := createProc(dir, "custom", "")
	if err == nil {
		t.Fatal("expected error for already existing process, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestCreateProc_TemplateNotFound(t *testing.T) {
	dir := setupTestDir(t)

	_, err := createProc(dir, "new-proc", "nonexistent-template")
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
	if !strings.Contains(err.Error(), "template not found") {
		t.Errorf("expected 'template not found' error, got: %v", err)
	}
}

func TestRunListCommand(t *testing.T) {
	dir := setupTestDir(t)

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	procRootDir = dir
	defer func() { procRootDir = "" }()

	err := runList(cmd, nil)
	if err != nil {
		t.Fatalf("runList returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "standard") {
		t.Errorf("output should contain preset process 'standard', got: %s", output)
	}
	if !strings.Contains(output, "Total: 2 flow(s)") {
		t.Errorf("output should contain total count, got: %s", output)
	}
}

func TestRunShowCommand(t *testing.T) {
	dir := setupTestDir(t)

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	procRootDir = dir
	defer func() { procRootDir = "" }()

	err := runShow(cmd, []string{"standard"})
	if err != nil {
		t.Fatalf("runShow returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-proc") {
		t.Errorf("output should contain process name, got: %s", output)
	}
}

func TestRunShowCommand_NotFound(t *testing.T) {
	dir := setupTestDir(t)

	procRootDir = dir
	defer func() { procRootDir = "" }()

	err := runShow(&cobra.Command{}, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent process, got nil")
	}
	if !strings.Contains(err.Error(), "process not found") {
		t.Errorf("expected 'process not found' error, got: %v", err)
	}
}

func TestRunValidateCommand_ValidFile(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "valid.json")
	os.WriteFile(validPath, validFlowJSON(), 0644)

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	err := runValidate(cmd, []string{validPath})
	if err != nil {
		t.Fatalf("runValidate returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓ Process is valid") {
		t.Errorf("output should indicate valid process, got: %s", output)
	}
}

func TestRunValidateCommand_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "invalid.json")
	os.WriteFile(invalidPath, invalidFlowJSON(), 0644)

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	err := runValidate(cmd, []string{invalidPath})
	if err != ErrValidationFailed {
		t.Fatalf("expected ErrValidationFailed, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✗ Process is invalid") {
		t.Errorf("output should indicate invalid process, got: %s", output)
	}
}

func TestRunValidateCommand_FileNotFound(t *testing.T) {
	err := runValidate(&cobra.Command{}, []string{"/nonexistent/flow.json"})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestRunCreateCommand_Blank(t *testing.T) {
	dir := t.TempDir()

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	procRootDir = dir
	procTemplate = ""
	defer func() { procRootDir = ""; procTemplate = "" }()

	err := runCreate(cmd, []string{"new-proc"})
	if err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created process '") {
		t.Errorf("output should indicate process created, got: %s", output)
	}
}

func TestRunCreateCommand_FromTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping template test in short mode (requires v3/flows/standard.json)")
	}
	dir := setupTestDir(t)

	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	procRootDir = dir
	procTemplate = "standard"
	defer func() { procRootDir = ""; procTemplate = "" }()

	err := runCreate(cmd, []string{"templated-proc"})
	if err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created process '") {
		t.Errorf("output should indicate process created, got: %s", output)
	}
}

func TestCmdStructure(t *testing.T) {
	if Cmd.Use != "proc" {
		t.Errorf("expected Cmd.Use 'proc', got %s", Cmd.Use)
	}
	if Cmd.Short != "Process lifecycle management" {
		t.Errorf("unexpected Cmd.Short: %s", Cmd.Short)
	}

	subcommands := make(map[string]bool)
	for _, sub := range Cmd.Commands() {
		subcommands[sub.Use] = true
	}

	for _, name := range []string{"run", "list", "show", "validate", "create", "next"} {
		found := false
		for k := range subcommands {
			if strings.HasPrefix(k, name) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand '%s' not found", name)
		}
	}

	for _, name := range []string{"start", "resume", "back", "status", "reset"} {
		for k := range subcommands {
			if strings.HasPrefix(k, name) {
				t.Errorf("removed subcommand '%s' should not exist", name)
			}
		}
	}
}

func TestCmdHasRootFlag(t *testing.T) {
	flag := Cmd.PersistentFlags().Lookup("root")
	if flag == nil {
		t.Error("expected persistent flag 'root' not found")
	}
}

func TestCreateCmdHasTemplateFlag(t *testing.T) {
	flag := createCmd.Flags().Lookup("template")
	if flag == nil {
		t.Error("expected flag 'template' on create command not found")
	}
}
