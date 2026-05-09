package toolrunner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type ToolConfig struct {
	Name        string            `mapstructure:"name"`
	Description string            `mapstructure:"description"`
	Binary      BinaryConfig      `mapstructure:"binary"`
	Install     InstallConfig     `mapstructure:"install"`
	Commands    map[string]Command `mapstructure:"commands"`
	Aliases     map[string]string  `mapstructure:"aliases"`
}

type BinaryConfig struct {
	Path         string            `mapstructure:"path"`
	AutoDownload InstallConfig     `mapstructure:"auto_download"`
	EnvPrefix    string            `mapstructure:"env_prefix"`
	Args         []string          `mapstructure:"args"`
	Env          map[string]string `mapstructure:"env"`
}

type InstallConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	URL      string `mapstructure:"url"`
	Method   string `mapstructure:"method"`
	Checksum string `mapstructure:"checksum"`
}

type Command struct {
	Name        string            `mapstructure:"name"`
	Description string            `mapstructure:"description"`
	Args        []string          `mapstructure:"args"`
	Env        map[string]string `mapstructure:"env"`
	Output     OutputConfig      `mapstructure:"output"`
	Examples    []string          `mapstructure:"examples"`
}

type OutputConfig struct {
	Format string `mapstructure:"format"`
}

type ToolsManager struct {
	tools map[string]*ToolConfig
	viper *viper.Viper
}

func NewToolsManager(configPath string) (*ToolsManager, error) {
	tm := &ToolsManager{
		tools:  make(map[string]*ToolConfig),
		viper: viper.New(),
	}

	tm.viper.SetConfigType("yaml")

	if configPath != "" {
		tm.viper.SetConfigFile(configPath)
	} else {
		cwd, _ := os.Getwd()
		projectToolsPath := filepath.Join(cwd, ".team", "tools.yaml")
		if _, err := os.Stat(projectToolsPath); err == nil {
			tm.viper.SetConfigFile(projectToolsPath)
		} else {
			home, _ := os.UserHomeDir()
			defaultPath := filepath.Join(home, ".flow", "tools.yaml")
			tm.viper.SetConfigFile(defaultPath)
		}
	}

	if err := tm.viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	} else {
		if err := tm.viper.Unmarshal(&tm.tools); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	return tm, nil
}

func (tm *ToolsManager) Register(tool *ToolConfig) {
	tm.tools[tool.Name] = tool
}

func (tm *ToolsManager) Get(name string) (*ToolConfig, bool) {
	tool, ok := tm.tools[name]
	if !ok {
		for toolName, t := range tm.tools {
			if strings.EqualFold(toolName, name) {
				return t, true
			}
		}
	}
	return tool, ok
}

func (tm *ToolsManager) List() []*ToolConfig {
	tools := make([]*ToolConfig, 0, len(tm.tools))
	for _, t := range tm.tools {
		tools = append(tools, t)
	}
	return tools
}

func (tm *ToolsManager) Run(toolName, command string, args []string) (string, error) {
	tool, ok := tm.Get(toolName)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	cmd, ok := tool.Commands[command]
	if !ok {
		return "", fmt.Errorf("command not found: %s %s", toolName, command)
	}

	binaryPath := tool.Binary.Path
	if binaryPath == "" {
		binaryPath = toolName
	}

	fullArgs := append(cmd.Args, args...)

	return execCommand(tool.Binary.Env, binaryPath, fullArgs)
}

func execCommand(envVars map[string]string, name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)

	for k, v := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

func (tm *ToolsManager) IsAvailable(toolName string) bool {
	tool, ok := tm.Get(toolName)
	if !ok {
		return false
	}

	binaryPath := tool.Binary.Path
	if binaryPath == "" {
		binaryPath = toolName
	}

	_, err := exec.LookPath(binaryPath)
	return err == nil
}

func (tm *ToolsManager) Install(toolName string) error {
	tool, ok := tm.Get(toolName)
	if !ok {
		return fmt.Errorf("tool not found: %s", toolName)
	}

	if !tool.Install.Enabled {
		return fmt.Errorf("auto-install not enabled for %s", toolName)
	}

	if tool.Install.URL == "" {
		return fmt.Errorf("no install URL specified for %s", toolName)
	}

	return fmt.Errorf("install not implemented: %s", tool.Install.URL)
}
