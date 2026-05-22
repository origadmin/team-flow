package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	skillfs "github.com/origadmin/team-flow"
	"github.com/origadmin/team-flow/internal/config"
)

type apiServer struct {
	rootDir string
}

func newAPIServer(rootDir string) *apiServer {
	return &apiServer{rootDir: rootDir}
}

func (s *apiServer) handleTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := fs.ReadDir(skillfs.FS, "teams")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	type teamItem struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		NameZh       string `json:"name_zh"`
		Description  string `json:"description"`
		FlowCount    int    `json:"flow_count"`
		DefaultFlow  string `json:"default_flow"`
	}

	var teams []teamItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, readErr := skillfs.FS.ReadFile("teams/" + entry.Name() + "/team.json")
		if readErr != nil {
			continue
		}
		var meta struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			NameZh      string `json:"name_zh"`
			Description string `json:"description"`
			DefaultFlow string `json:"default_flow"`
			Flows       []struct{} `json:"flows"`
		}
		if jsonErr := json.Unmarshal(data, &meta); jsonErr != nil {
			continue
		}
		teams = append(teams, teamItem{
			ID:          meta.ID,
			Name:        meta.Name,
			NameZh:      meta.NameZh,
			Description: meta.Description,
			FlowCount:   len(meta.Flows),
			DefaultFlow: meta.DefaultFlow,
		})
	}

	writeJSON(w, http.StatusOK, teams)
}

func (s *apiServer) handleTeamFlows(w http.ResponseWriter, r *http.Request, teamID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := skillfs.FS.ReadFile("teams/" + teamID + "/team.json")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "team not found"})
		return
	}

	var meta struct {
		Flows []struct {
			ID          string `json:"id"`
			File        string `json:"file"`
			Description string `json:"description"`
			Default     bool   `json:"default"`
		} `json:"flows"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, meta.Flows)
}

func (s *apiServer) handleTeamFlowJSON(w http.ResponseWriter, r *http.Request, teamID, flowID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := skillfs.FS.ReadFile("teams/" + teamID + "/flows/" + flowID + ".json")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *apiServer) handleProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := config.LoadProjectConfig(s.rootDir)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project config not found"})
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}

func (s *apiServer) handleProjectFlows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flowsDir := filepath.Join(s.rootDir, ".team", "flows")
	entries, err := os.ReadDir(flowsDir)
	if err != nil {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	cfg, _ := config.LoadProjectConfig(s.rootDir)
	defaultFlow := ""
	if cfg != nil {
		defaultFlow = cfg.DefaultFlow
	}

	type flowItem struct {
		ID      string `json:"id"`
		Default bool   `json:"default"`
	}

	var flows []flowItem
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		flows = append(flows, flowItem{
			ID:      id,
			Default: id == defaultFlow,
		})
	}

	writeJSON(w, http.StatusOK, flows)
}

func (s *apiServer) handleProjectFlowJSON(w http.ResponseWriter, r *http.Request, flowName string) {
	switch r.Method {
	case http.MethodGet:
		flowPath := filepath.Join(s.rootDir, ".team", "flows", flowName+".json")
		data, err := os.ReadFile(flowPath)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)

	case http.MethodPut:
		var body json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}

		flowsDir := filepath.Join(s.rootDir, ".team", "flows")
		os.MkdirAll(flowsDir, 0755)
		flowPath := filepath.Join(flowsDir, flowName+".json")

		var indented bytes.Buffer
		if err := json.Indent(&indented, body, "", "  "); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}

		if err := os.WriteFile(flowPath, []byte(indented.String()), 0644); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":   true,
			"path": flowPath,
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) handleValidateFlow(w http.ResponseWriter, r *http.Request, flowName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	result := validateFlowJSON(body)
	writeJSON(w, http.StatusOK, result)
}

func (s *apiServer) handleValidateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result := s.validateTeamFlows()
	writeJSON(w, http.StatusOK, result)
}

type validationError struct {
	NodeID string `json:"node_id,omitempty"`
	Field  string `json:"field"`
	Msg    string `json:"msg"`
}

type validationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []validationError `json:"errors"`
	Warnings []validationError `json:"warnings"`
}

func validateFlowJSON(data json.RawMessage) validationResult {
	result := validationResult{Valid: true}

	var flow map[string]interface{}
	if err := json.Unmarshal(data, &flow); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, validationError{
			Field: "json",
			Msg:   "invalid JSON: " + err.Error(),
		})
		return result
	}

	if _, ok := flow["metadata"]; !ok {
		result.Warnings = append(result.Warnings, validationError{
			Field: "metadata",
			Msg:   "missing metadata section",
		})
	}

	components, ok := flow["components"].(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, validationError{
			Field: "components",
			Msg:   "missing or invalid components section",
		})
		return result
	}

	if _, ok := components["roles"]; !ok {
		result.Warnings = append(result.Warnings, validationError{
			Field: "components.roles",
			Msg:   "no roles defined",
		})
	}

	if _, ok := flow["nodes"]; !ok {
		result.Valid = false
		result.Errors = append(result.Errors, validationError{
			Field: "nodes",
			Msg:   "missing nodes section",
		})
	}

	if _, ok := flow["edges"]; !ok {
		result.Warnings = append(result.Warnings, validationError{
			Field: "edges",
			Msg:   "missing edges section",
		})
	}

	nodes, ok := flow["nodes"].([]interface{})
	if ok {
		nodeIDs := make(map[string]bool)
		for _, n := range nodes {
			node, ok := n.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := node["id"].(string)
			if id == "" {
				result.Valid = false
				result.Errors = append(result.Errors, validationError{
					Field: "nodes[].id",
					Msg:   "node missing id",
				})
			}
			if nodeIDs[id] {
				result.Valid = false
				result.Errors = append(result.Errors, validationError{
					NodeID: id,
					Field:  "nodes[].id",
					Msg:    "duplicate node id",
				})
			}
			nodeIDs[id] = true
		}

		edges, ok := flow["edges"].([]interface{})
		if ok {
			for _, e := range edges {
				edge, ok := e.(map[string]interface{})
				if !ok {
					continue
				}
				from, _ := edge["from"].(string)
				to, _ := edge["to"].(string)
				if from != "" && !nodeIDs[from] {
					result.Valid = false
					result.Errors = append(result.Errors, validationError{
						Field: "edges[].from",
						Msg:   fmt.Sprintf("edge references non-existent node: %s", from),
					})
				}
				if to != "" && !nodeIDs[to] {
					result.Valid = false
					result.Errors = append(result.Errors, validationError{
						Field: "edges[].to",
						Msg:   fmt.Sprintf("edge references non-existent node: %s", to),
					})
				}
			}
		}
	}

	return result
}

type teamValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []validationError `json:"errors"`
	Warnings []validationError `json:"warnings"`
}

func (s *apiServer) validateTeamFlows() teamValidationResult {
	result := teamValidationResult{Valid: true}

	flowsDir := filepath.Join(s.rootDir, ".team", "flows")
	entries, err := os.ReadDir(flowsDir)
	if err != nil || len(entries) == 0 {
		result.Warnings = append(result.Warnings, validationError{
			Field: "flows",
			Msg:   "no flows found in .team/flows/",
		})
		return result
	}

	roleMap := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		flowPath := filepath.Join(flowsDir, entry.Name())
		data, readErr := os.ReadFile(flowPath)
		if readErr != nil {
			continue
		}

		var flow map[string]interface{}
		if jsonErr := json.Unmarshal(data, &flow); jsonErr != nil {
			result.Valid = false
			result.Errors = append(result.Errors, validationError{
				Field: entry.Name(),
				Msg:   "invalid JSON",
			})
			continue
		}

		components, _ := flow["components"].(map[string]interface{})
		roles, _ := components["roles"].([]interface{})
		flowID := strings.TrimSuffix(entry.Name(), ".json")
		for _, r := range roles {
			role, _ := r.(map[string]interface{})
			roleID, _ := role["id"].(string)
			if roleID != "" {
				roleMap[roleID] = append(roleMap[roleID], flowID)
			}
		}
	}

	for roleID, flowIDs := range roleMap {
		if len(flowIDs) == 1 {
			result.Warnings = append(result.Warnings, validationError{
				Field: fmt.Sprintf("role.%s", roleID),
				Msg:   fmt.Sprintf("role only used in %s, consider if it should be in more flows", flowIDs[0]),
			})
		}
	}

	return result
}

func (s *apiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api")

	switch {
	case path == "/teams":
		s.handleTeams(w, r)
	case strings.HasPrefix(path, "/teams/") && strings.HasSuffix(path, "/flows"):
		teamID := strings.TrimPrefix(path, "/teams/")
		teamID = strings.TrimSuffix(teamID, "/flows")
		s.handleTeamFlows(w, r, teamID)
	case strings.HasPrefix(path, "/teams/") && strings.Contains(path, "/flows/"):
		parts := strings.Split(strings.TrimPrefix(path, "/teams/"), "/")
		if len(parts) >= 3 && parts[1] == "flows" {
			s.handleTeamFlowJSON(w, r, parts[0], parts[2])
		} else {
			http.NotFound(w, r)
		}
	case path == "/project":
		s.handleProject(w, r)
	case path == "/project/flows":
		s.handleProjectFlows(w, r)
	case path == "/project/validate-team":
		s.handleValidateTeam(w, r)
	case strings.HasPrefix(path, "/project/flows/"):
		flowName := strings.TrimPrefix(path, "/project/flows/")
		if strings.HasSuffix(flowName, "/validate") {
			flowName = strings.TrimSuffix(flowName, "/validate")
			s.handleValidateFlow(w, r, flowName)
		} else {
			s.handleProjectFlowJSON(w, r, flowName)
		}
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
