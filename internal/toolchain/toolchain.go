package toolchain

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	pythonPath     string
	pythonChecked  bool
	pipPath        string
	pipChecked     bool
	nodePath       string
	nodeChecked    bool
	goPath         string
	goChecked      bool
	gitPath        string
	gitChecked     bool
)

func FindPythonPath() string {
	if pythonChecked {
		return pythonPath
	}
	pythonChecked = true

	if runtime.GOOS == "windows" {
		paths := []string{
			`C:\Python312\python.exe`,
			`C:\Python311\python.exe`,
			`C:\Python310\python.exe`,
			`C:\Python39\python.exe`,
			`C:\Program Files\Python312\python.exe`,
			`C:\Program Files\Python311\python.exe`,
			`C:\Program Files\Python310\python.exe`,
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				pythonPath = p
				return p
			}
		}
	}

	cmd := exec.Command("where.exe", "python")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				pythonPath = line
				return line
			}
		}
	}

	cmd = exec.Command("which", "python")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}

	return ""
}

func FindPipPath() string {
	if pipChecked {
		return pipPath
	}
	pipChecked = true

	pyPath := FindPythonPath()
	if pyPath == "" {
		return ""
	}

	if runtime.GOOS == "windows" {
		paths := []string{
			strings.Replace(pyPath, "python.exe", "Scripts\\pip.exe", 1),
			strings.Replace(pyPath, "python.exe", "Scripts\\pip3.exe", 1),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				pipPath = p
				return p
			}
		}
	}

	cmd := exec.Command("where.exe", "pip")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				pipPath = line
				return line
			}
		}
	}

	cmd = exec.Command(pyPath, "-m", "pip", "--version")
	if err := cmd.Run(); err == nil {
		pipPath = pyPath + " -m pip"
		return pipPath
	}

	return ""
}

func FindNodePath() string {
	if nodeChecked {
		return nodePath
	}
	nodeChecked = true

	if runtime.GOOS == "windows" {
		cmd := exec.Command("where.exe", "node")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					nodePath = line
					return line
				}
			}
		}
	}

	cmd := exec.Command("which", "node")
	if output, err := cmd.Output(); err == nil {
		nodePath = strings.TrimSpace(string(output))
	}
	return nodePath
}

func FindGoPath() string {
	if goChecked {
		return goPath
	}
	goChecked = true

	if runtime.GOOS == "windows" {
		cmd := exec.Command("where.exe", "go")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					goPath = line
					return line
				}
			}
		}
	}

	cmd := exec.Command("which", "go")
	if output, err := cmd.Output(); err == nil {
		goPath = strings.TrimSpace(string(output))
	}
	return goPath
}

func FindGitPath() string {
	if gitChecked {
		return gitPath
	}
	gitChecked = true

	if runtime.GOOS == "windows" {
		cmd := exec.Command("where.exe", "git")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					gitPath = line
					return line
				}
			}
		}
	}

	cmd := exec.Command("which", "git")
	if output, err := cmd.Output(); err == nil {
		gitPath = strings.TrimSpace(string(output))
	}
	return gitPath
}

func CheckCommand(name string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("where.exe", name)
	} else {
		cmd = exec.Command("which", name)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command not found: %s", name)
	}
	return nil
}
