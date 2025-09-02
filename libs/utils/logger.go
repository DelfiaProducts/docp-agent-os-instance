package utils

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DEBUG = iota
	INFO
	WARN
)

// DocpLogger is struct for logger the docp
type DocpLogger struct {
	logger *slog.Logger
}

func getLogLevel() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewDocpLoggerJSON return instance of docp logger
// with formatter json
func NewDocpLoggerJSON(writter io.Writer) *DocpLogger {
	return &DocpLogger{
		logger: slog.New(slog.NewJSONHandler(writter, &slog.HandlerOptions{
			Level: getLogLevel(),
		})),
	}
}

// NewDocpLogger return instance of docp logger
// with formatter text
func NewDocpLoggerText(writter io.Writer) *DocpLogger {
	return &DocpLogger{
		logger: slog.New(slog.NewTextHandler(writter, &slog.HandlerOptions{
			Level: getLogLevel(),
		})),
	}
}

// Debug execute logging the debug
func (d *DocpLogger) Debug(msg string, args ...any) {
	d.logger.Debug(msg, args...)
}

// Info execute logging the info
func (d *DocpLogger) Info(msg string, args ...any) {
	d.logger.Info(msg, args...)
}

// Warn execute logging the warning
func (d *DocpLogger) Warn(msg string, args ...any) {
	d.logger.Warn(msg, args...)
}

// Error execut logging the error
func (d *DocpLogger) Error(msg string, args ...any) {
	d.logger.Error(msg, args...)
}

// Close close logger
func (d *DocpLogger) Close() error {
	return nil
}

// DocpLoggerWindows is struct for logger the docp on windows
type DocpLoggerWindows struct {
	logger *log.Logger
	level  int
	mutex  sync.Mutex
}

func getLogLevelWindows() int {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {

	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	default:
		return INFO
	}
}

// NewDocpLoggerFileText return instance of docp logger redirect for file
// with formatter text
func NewDocpLoggerWindowsFileText(logPath string) *DocpLoggerWindows {
	logFile := &lumberjack.Logger{
		Filename:   logPath, // Caminho do arquivo de log.
		MaxSize:    5,       // Tamanho máximo em megabytes antes da rotação.
		MaxBackups: 2,       // Número máximo de arquivos de backup antigos para manter.
		MaxAge:     1,       // Número máximo de dias para manter os arquivos de log antigos.
		Compress:   false,   // Compactar arquivos de log antigos.
	}

	multiWriter := io.MultiWriter(logFile, os.Stdout)
	logger := log.New(multiWriter, "", log.LstdFlags)
	return &DocpLoggerWindows{
		logger: logger,
		level:  getLogLevelWindows(),
	}
}

func (d *DocpLoggerWindows) formatArgs(msg string, args ...any) string {
	format := msg
	for range args {
		format += ", %v"
	}
	format += "\n"
	return fmt.Sprintf(format, args...)
}

// Debug execute logging the debug
func (d *DocpLoggerWindows) Debug(msg string, args ...any) {
	if d.level == DEBUG {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.logger.Printf(d.formatArgs(msg, args...))
	}
}

// Info execute logging the info
func (d *DocpLoggerWindows) Info(msg string, args ...any) {
	if d.level == INFO {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.logger.Printf(d.formatArgs(msg, args...))
	}
}

// Warn execute logging the warning
func (d *DocpLoggerWindows) Warn(msg string, args ...any) {
	if d.level == WARN {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.logger.Printf(d.formatArgs(msg, args...))
	}
}

// Error execut logging the error
func (d *DocpLoggerWindows) Error(msg string, args ...any) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.logger.Printf(d.formatArgs(msg, args...))
}

// Close close logger
func (d *DocpLoggerWindows) Close() error {
	return nil
}
