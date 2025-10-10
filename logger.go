package tmigo

import (
	"fmt"
	"log"
	"os"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// Logger interface for logging
type Logger interface {
	SetLevel(level string) error
	Trace(msg string)
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

// DefaultLogger implements the Logger interface
type DefaultLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new default logger
func NewLogger() *DefaultLogger {
	return &DefaultLogger{
		level:  LogLevelError,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// SetLevel sets the logging level
func (l *DefaultLogger) SetLevel(level string) error {
	switch level {
	case "trace":
		l.level = LogLevelTrace
	case "debug":
		l.level = LogLevelDebug
	case "info":
		l.level = LogLevelInfo
	case "warn":
		l.level = LogLevelWarn
	case "error":
		l.level = LogLevelError
	case "fatal":
		l.level = LogLevelFatal
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

func (l *DefaultLogger) log(level LogLevel, prefix string, msg string) {
	if level >= l.level {
		l.logger.Printf("[%s] %s", prefix, msg)
	}
}

// Trace logs a trace message
func (l *DefaultLogger) Trace(msg string) {
	l.log(LogLevelTrace, "TRACE", msg)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string) {
	l.log(LogLevelDebug, "DEBUG", msg)
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string) {
	l.log(LogLevelInfo, "INFO", msg)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string) {
	l.log(LogLevelWarn, "WARN", msg)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string) {
	l.log(LogLevelError, "ERROR", msg)
}

// Fatal logs a fatal message and exits
func (l *DefaultLogger) Fatal(msg string) {
	l.log(LogLevelFatal, "FATAL", msg)
	os.Exit(1)
}
