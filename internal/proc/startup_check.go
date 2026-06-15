package proc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SessionStartupStatus tracks the state of Session Startup Protocol
type SessionStartupStatus struct {
	Step1Detected       bool   `json:"step1_detected"`
	Step1ProjectRoot    string `json:"step1_project_root"`
	Step1ProjectName    string `json:"step1_project_name"`
	Step2SessionCreated bool   `json:"step2_session_created"`
	Step2SessionID      string `json:"step2_session_id"`
	Step3RoleAdopted    bool   `json:"step3_role_adopted"`
	Step3RoleAlias      string `json:"step3_role_alias"`
	Completed           bool   `json:"completed"`
}

// CheckSessionStartup verifies if Session Startup Protocol has been completed
func CheckSessionStartup(projectRoot string) (SessionStartupStatus, error) {
	status := SessionStartupStatus{}

	// Step 1: Check if flow project detect has been run
	// Look for .team/project.md or .beads directory
	projectMD := filepath.Join(projectRoot, ".team", "project.md")
	beadsDir := filepath.Join(projectRoot, ".beads")
	
	if _, err := os.Stat(projectMD); err == nil {
		status.Step1Detected = true
		status.Step1ProjectRoot = projectRoot
		content, _ := os.ReadFile(projectMD)
		if idx := strings.Index(string(content), "name:"); idx != -1 {
			rest := strings.TrimSpace(string(content)[idx+5:])
			if end := strings.Index(rest, "\n"); end != -1 {
				status.Step1ProjectName = rest[:end]
			} else {
				status.Step1ProjectName = rest
			}
		}
	} else if _, err := os.Stat(beadsDir); err == nil {
		status.Step1Detected = true
		status.Step1ProjectRoot = projectRoot
		status.Step1ProjectName = filepath.Base(projectRoot)
	}

	// Step 2: Check if session has been created
	// Look for session.json in .team or temp directory
	sessionFile := filepath.Join(projectRoot, ".team", "session.json")
	if _, err := os.Stat(sessionFile); err == nil {
		var sessionData map[string]interface{}
		if data, err := os.ReadFile(sessionFile); err == nil {
			if json.Unmarshal(data, &sessionData) == nil {
				if id, ok := sessionData["id"].(string); ok {
					status.Step2SessionCreated = true
					status.Step2SessionID = id
				}
			}
		}
	}

	// Step 3: Check if role has been adopted
	// Look for role information in session or project config
	if status.Step2SessionCreated {
		if data, err := os.ReadFile(filepath.Join(projectRoot, ".team", "session.json")); err == nil {
			var sessionData map[string]interface{}
			if json.Unmarshal(data, &sessionData) == nil {
				if role, ok := sessionData["role"].(map[string]interface{}); ok {
					if alias, ok := role["alias"].(string); ok {
						status.Step3RoleAdopted = true
						status.Step3RoleAlias = alias
					}
				}
			}
		}
	}

	status.Completed = status.Step1Detected && status.Step2SessionCreated && status.Step3RoleAdopted
	return status, nil
}

// GetStartupViolationMessage returns a message if startup protocol is violated
func GetStartupViolationMessage(status SessionStartupStatus) string {
	if status.Completed {
		return ""
	}

	var violations []string
	
	if !status.Step1Detected {
		violations = append(violations, "Step 1: flow project detect 未执行")
	}
	if !status.Step2SessionCreated {
		violations = append(violations, "Step 2: flow proc run --new 未执行")
	}
	if !status.Step3RoleAdopted {
		violations = append(violations, "Step 3: 角色未采用")
	}

	return fmt.Sprintf("⛔ SESSION STARTUP PROTOCOL VIOLATION\n\n以下步骤未完成:\n\n%s\n\n请按顺序执行:\n1. flow project detect\n2. flow proc run --new\n3. 读取输出中的角色信息并采用", 
		strings.Join(violations, "\n"))
}

// EnsureStartupCompleted enforces the Session Startup Protocol
func EnsureStartupCompleted(projectRoot string) error {
	status, err := CheckSessionStartup(projectRoot)
	if err != nil {
		return fmt.Errorf("check startup: %w", err)
	}
	
	if !status.Completed {
		return fmt.Errorf("%s", GetStartupViolationMessage(status))
	}
	
	return nil
}