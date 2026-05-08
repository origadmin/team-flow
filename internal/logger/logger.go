package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	mu       sync.Mutex
	level    Level
	logFile  *os.File
	module   string
	console  bool
}

var (
	defaultLogger *Logger
	once          sync.Once
	logDir        = ".team"
	logSubDir     = "logs"
)

func Init(module string, level Level) error {
	var err error
	once.Do(func() {
		defaultLogger, err = newLogger(module, level)
	})
	if err != nil {
		return err
	}
	return nil
}

func newLogger(module string, level Level) (*Logger, error) {
	l := &Logger{
		level:  level,
		module: module,
		console: true,
	}

	projectPath, err := os.Getwd()
	if err != nil {
		return l, nil
	}

	logDirPath := filepath.Join(projectPath, logDir, logSubDir)
	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		return l, fmt.Errorf("create log dir: %w", err)
	}

	dateStr := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("team-%s.log", dateStr)
	logFilePath := filepath.Join(logDirPath, logFileName)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return l, fmt.Errorf("open log file: %w", err)
	}
	l.logFile = file

	return l, nil
}

func GetLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger, _ = newLogger("flow", INFO)
	}
	return defaultLogger
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) SetConsole(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.console = enabled
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	_, file, line, _ := runtime.Caller(2)
	fileName := filepath.Base(file)

	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	logLine := fmt.Sprintf("[%s] %s %s/%d - %s\n",
		level.String(), timestamp, fileName, line, message)

	if l.logFile != nil {
		l.logFile.WriteString(logLine)
	}

	if l.console {
		prefix := ""
		switch level {
		case DEBUG:
			prefix = "🔍"
		case INFO:
			prefix = "ℹ️"
		case WARN:
			prefix = "⚠️"
		case ERROR:
			prefix = "❌"
		}
		consoleLine := fmt.Sprintf("%s %s\n", prefix, message)
		fmt.Print(consoleLine)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) WithField(key string, value interface{}) *LogEntry {
	return &LogEntry{
		logger: l,
		fields: map[string]interface{}{key: value},
	}
}

func (l *Logger) WithFields(fields map[string]interface{}) *LogEntry {
	return &LogEntry{
		logger: l,
		fields: fields,
	}
}

func Close() error {
	if defaultLogger != nil && defaultLogger.logFile != nil {
		return defaultLogger.logFile.Close()
	}
	return nil
}

type LogEntry struct {
	logger *Logger
	fields map[string]interface{}
	level  Level
}

func (e *LogEntry) Debug(format string) {
	e.entry(DEBUG, format)
}

func (e *LogEntry) Info(format string) {
	e.entry(INFO, format)
}

func (e *LogEntry) Warn(format string) {
	e.entry(WARN, format)
}

func (e *LogEntry) Error(format string) {
	e.entry(ERROR, format)
}

func (e *LogEntry) entry(level Level, format string) {
	if level < e.logger.level {
		return
	}

	fieldsJSON, _ := json.Marshal(e.fields)
	message := fmt.Sprintf("%s | %s", format, string(fieldsJSON))
	e.logger.log(level, message)
}

type CommandLog struct {
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	ExitCode  int      `json:"exit_code"`
	Output    string   `json:"output,omitempty"`
	ErrorMsg  string   `json:"error,omitempty"`
	Timestamp string   `json:"timestamp"`
}

func (l *Logger) LogCommand(cmd string, args []string, exitCode int, output string, err error) {
	entry := CommandLog{
		Command:   cmd,
		Args:      args,
		ExitCode:  exitCode,
		Output:    output,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	if err != nil {
		entry.ErrorMsg = err.Error()
	}

	logJSON, _ := json.Marshal(entry)
	l.log(DEBUG, "CMD: %s", string(logJSON))
}
