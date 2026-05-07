package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	graphDepth      int
	graphDetail     string
	graphOutput     string
	graphPattern    string
	graphTarget     string
	graphQuery      string
	graphFiles      []string
	graphBase       string
	graphMaxResults int
)

var Cmd = &cobra.Command{
	Use:   "graph",
	Short: "Code graph analysis (code-review-graph)",
	Long: `Code graph analysis powered by code-review-graph.

Provides semantic search, impact analysis, change detection, and flow tracking.
Requires: pip install code-review-graph`,
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build or update the code graph",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPythonModule("build")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show graph statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPythonModule("status")
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Semantic search for code nodes",
	Long:  "Search the code graph using natural language or symbol names.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		pyCode := fmt.Sprintf(
			"from code_review_graph.tools import semantic_search_nodes; "+
				"import json; r=semantic_search_nodes(query=%q,limit=%d,detail_level=%q);"+
				"print(json.dumps(r,ensure_ascii=False,indent=2))",
			query, graphMaxResults, graphDetail,
		)
		return runPythonCode(pyCode)
	},
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Structured graph query (callers_of, tests_for, etc.)",
	Long: `Structured graph query patterns:
  callers_of    - Find callers of a function
  callees_of    - Find functions called by a function
  tests_for     - Find tests for a function
  importers_of  - Find files importing a module`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if graphPattern == "" || graphTarget == "" {
			return fmt.Errorf("both --pattern and --target are required")
		}
		pyCode := fmt.Sprintf(
			"from code_review_graph.tools import query_graph; "+
				"import json; r=query_graph(pattern=%q,target=%q,detail_level=%q);"+
				"print(json.dumps(r,ensure_ascii=False,indent=2))",
			graphPattern, graphTarget, graphDetail,
		)
		return runPythonCode(pyCode)
	},
}

var impactCmd = &cobra.Command{
	Use:   "impact [files...]",
	Short: "Analyze impact radius of changes",
	Long:  "Calculate blast radius for changed files. Shows directly changed nodes and impacted nodes within N hops.",
	RunE: func(cmd *cobra.Command, args []string) error {
		files := args
		if len(files) == 0 {
			detected, err := detectChangedFiles()
			if err != nil {
				return fmt.Errorf("no files specified and auto-detect failed: %w", err)
			}
			files = detected
		}
		filesJSON, _ := json.Marshal(files)
		pyCode := fmt.Sprintf(
			"from code_review_graph.tools import get_impact_radius; "+
				"import json; r=get_impact_radius(changed_files=json.loads(%s),max_depth=%d,max_results=%d,detail_level=%q);"+
				"print(json.dumps(r,ensure_ascii=False,indent=2))",
			string(filesJSON), graphDepth, graphMaxResults, graphDetail,
		)
		return runPythonCode(pyCode)
	},
}

var flowsCmd = &cobra.Command{
	Use:   "flows [files...]",
	Short: "Trace affected data flows",
	Long:  "Trace which execution flows are affected by changes in the specified files.",
	RunE: func(cmd *cobra.Command, args []string) error {
		files := args
		if len(files) == 0 {
			detected, err := detectChangedFiles()
			if err != nil {
				return fmt.Errorf("no files specified and auto-detect failed: %w", err)
			}
			files = detected
		}
		filesJSON, _ := json.Marshal(files)
		pyCode := fmt.Sprintf(
			"from code_review_graph.tools import get_affected_flows_func; "+
				"import json; r=get_affected_flows_func(changed_files=json.loads(%s),base=%q);"+
				"print(json.dumps(r,ensure_ascii=False,indent=2))",
			string(filesJSON), graphBase,
		)
		return runPythonCode(pyCode)
	},
}

var changesCmd = &cobra.Command{
	Use:   "changes",
	Short: "Detect and analyze code changes",
	Long:  "Detect staged/unstaged changes and analyze their impact with risk scoring.",
	RunE: func(cmd *cobra.Command, args []string) error {
		pyCode := fmt.Sprintf(
			"from code_review_graph.tools import detect_changes_func; "+
				"import json; r=detect_changes_func();"+
				"print(json.dumps(r,ensure_ascii=False,indent=2))",
		)
		return runPythonCode(pyCode)
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review [files...]",
	Short: "Get review context for changes",
	Long:  "Build a full review context with source snippets for the specified files.",
	RunE: func(cmd *cobra.Command, args []string) error {
		files := args
		if len(files) == 0 {
			detected, err := detectChangedFiles()
			if err != nil {
				return fmt.Errorf("no files specified and auto-detect failed: %w", err)
			}
			files = detected
		}
		filesJSON, _ := json.Marshal(files)
		pyCode := fmt.Sprintf(
			"from code_review_graph.tools import get_review_context; "+
				"import json; r=get_review_context(changed_files=json.loads(%s));"+
				"print(json.dumps(r,ensure_ascii=False,indent=2))",
			string(filesJSON),
		)
		return runPythonCode(pyCode)
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server (stdio transport)",
	Long:  "Start the code-review-graph MCP server for AI IDE integration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPythonModule("serve")
	},
}

func init() {
	Cmd.AddCommand(buildCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(searchCmd)
	Cmd.AddCommand(queryCmd)
	Cmd.AddCommand(impactCmd)
	Cmd.AddCommand(flowsCmd)
	Cmd.AddCommand(changesCmd)
	Cmd.AddCommand(reviewCmd)
	Cmd.AddCommand(serveCmd)

	searchCmd.Flags().IntVar(&graphMaxResults, "limit", 20, "Max results")
	searchCmd.Flags().StringVar(&graphDetail, "detail", "standard", "Detail level: minimal/standard/full")

	queryCmd.Flags().StringVar(&graphPattern, "pattern", "", "Query pattern: callers_of, callees_of, tests_for, importers_of")
	queryCmd.Flags().StringVar(&graphTarget, "target", "", "Target symbol name")
	queryCmd.Flags().StringVar(&graphDetail, "detail", "standard", "Detail level")

	impactCmd.Flags().IntVar(&graphDepth, "depth", 2, "Impact depth (hops)")
	impactCmd.Flags().IntVar(&graphMaxResults, "limit", 500, "Max results")
	impactCmd.Flags().StringVar(&graphDetail, "detail", "standard", "Detail level")

	flowsCmd.Flags().StringVar(&graphBase, "base", "HEAD~1", "Git base for comparison")

	changesCmd.Flags().StringVar(&graphDetail, "detail", "standard", "Detail level")

	reviewCmd.Flags().StringVar(&graphDetail, "detail", "standard", "Detail level")
}

func runPythonModule(subCmd string) error {
	cmd := exec.Command("python", "-m", "code_review_graph", subCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runPythonCode(code string) error {
	cmd := exec.Command("python", "-c", code)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func detectChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff: %s: %w", string(output), err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no changed files detected")
	}
	return files, nil
}
