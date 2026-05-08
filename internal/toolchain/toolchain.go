package toolchain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func FindBdPath() string {
	if path, err := exec.LookPath("bd"); err == nil {
		return path
	}

	home, _ := os.UserHomeDir()
	localAppData := os.Getenv("LOCALAPPDATA")

	candidates := []string{}
	switch runtime.GOOS {
	case "windows":
		if localAppData != "" {
			candidates = append(candidates,
				filepath.Join(localAppData, "Programs", "bd", "bd.exe"),
				filepath.Join(localAppData, "bd", "bd.exe"),
			)
		}
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, "AppData", "Local", "Programs", "bd", "bd.exe"),
				filepath.Join(home, ".local", "bin", "bd.exe"),
			)
		}
	case "darwin":
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "bd"),
				filepath.Join("/usr/local/bin", "bd"),
				filepath.Join("/opt/homebrew/bin", "bd"),
			)
		}
	default:
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "bd"),
				filepath.Join("/usr/local/bin", "bd"),
			)
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

func FindPythonPath() string {
	for _, name := range []string{"python", "python3"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	home, _ := os.UserHomeDir()
	localAppData := os.Getenv("LOCALAPPDATA")
	programFiles := os.Getenv("ProgramFiles")

	candidates := []string{}
	switch runtime.GOOS {
	case "windows":
		if localAppData != "" {
			candidates = append(candidates,
				filepath.Join(localAppData, "Programs", "Python", "Python311", "python.exe"),
				filepath.Join(localAppData, "Programs", "Python", "Python312", "python.exe"),
				filepath.Join(localAppData, "Programs", "Python", "Python313", "python.exe"),
				filepath.Join(localAppData, "Programs", "Python", "Python310", "python.exe"),
			)
		}
		if programFiles != "" {
			candidates = append(candidates,
				filepath.Join(programFiles, "Python311", "python.exe"),
				filepath.Join(programFiles, "Python312", "python.exe"),
				filepath.Join(programFiles, "Python313", "python.exe"),
			)
		}
		candidates = append(candidates,
			`C:\Python311\python.exe`,
			`C:\Python312\python.exe`,
			`C:\Python310\python.exe`,
		)
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, "AppData", "Local", "Programs", "Python", "Python311", "python.exe"),
				filepath.Join(home, "AppData", "Local", "Programs", "Python", "Python312", "python.exe"),
				filepath.Join(home, "AppData", "Local", "Programs", "Python", "Python313", "python.exe"),
			)
		}
	case "darwin":
		candidates = append(candidates,
			"/usr/bin/python3",
			"/usr/local/bin/python3",
			"/opt/homebrew/bin/python3",
		)
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "python3"),
			)
		}
	default:
		candidates = append(candidates,
			"/usr/bin/python3",
			"/usr/local/bin/python3",
		)
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "python3"),
			)
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

func FindPipPath() string {
	for _, name := range []string{"pip", "pip3"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

func AddBdToPath(bdPath string) bool {
	bdDir := filepath.Dir(bdPath)

	currentPath := os.Getenv("PATH")
	if currentPath == "" {
		return false
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	os.Setenv("PATH", currentPath+separator+bdDir)

	if runtime.GOOS == "windows" {
		return addToUserPathWindows(bdDir)
	}

	return true
}

func addToUserPathWindows(dir string) bool {
	psScript := fmt.Sprintf(
		`$current = [Environment]::GetEnvironmentVariable('Path', 'User'); if ($current -notlike '*%s*') { [Environment]::SetEnvironmentVariable('Path', $current + ';%s', 'User') }`,
		dir, dir,
	)
	cmd := exec.Command("powershell", "-Command", psScript)
	return cmd.Run() == nil
}

func IsBdOnPath() bool {
	_, err := exec.LookPath("bd")
	return err == nil
}
