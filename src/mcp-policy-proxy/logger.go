package main

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     string                 `json:"timestamp"`
	Level         LogLevel               `json:"level"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Component     string                 `json:"component"`
	Message       string                 `json:"message"`
	RequestID     string                 `json:"request_id,omitempty"`
	Method        string                 `json:"method,omitempty"`
	Path          string                 `json:"path,omitempty"`
	StatusCode    int                    `json:"status_code,omitempty"`
	LatencyMs     float64                `json:"latency_ms,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// Logger provides structured JSON logging with correlation ID support
type Logger struct {
	mu        sync.Mutex
	writer    io.Writer
	minLevel  LogLevel
	component string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// NewLogger creates a new Logger instance
func NewLogger(component string) *Logger {
	once.Do(func() {
		defaultLogger = &Logger{
			writer:    os.Stdout,
			minLevel:  LevelInfo,
			component: component,
		}
	})
	return &Logger{
		writer:    os.Stdout,
		minLevel:  LevelInfo,
		component: component,
	}
}

// SetWriter sets the output writer for the logger
func (l *Logger) SetWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer = w
}

// SetMinLevel sets the minimum log level
func (l *Logger) SetMinLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// shouldLog checks if the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}
	return levels[level] >= levels[l.minLevel]
}

// log writes a structured log entry
func (l *Logger) log(level LogLevel, entry LogEntry) {
	if !l.shouldLog(level) {
		return
	}

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	entry.Level = level
	if entry.Component == "" {
		entry.Component = l.component
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer.Write(append(data, '\n'))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...func(*LogEntry)) {
	entry := LogEntry{Message: msg}
	for _, f := range fields {
		f(&entry)
	}
	l.log(LevelDebug, entry)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...func(*LogEntry)) {
	entry := LogEntry{Message: msg}
	for _, f := range fields {
		f(&entry)
	}
	l.log(LevelInfo, entry)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...func(*LogEntry)) {
	entry := LogEntry{Message: msg}
	for _, f := range fields {
		f(&entry)
	}
	l.log(LevelWarn, entry)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...func(*LogEntry)) {
	entry := LogEntry{Message: msg}
	for _, f := range fields {
		f(&entry)
	}
	l.log(LevelError, entry)
}

// Field helpers for log entry configuration

// WithCorrelationID adds a correlation ID to the log entry
func WithCorrelationID(id string) func(*LogEntry) {
	return func(e *LogEntry) {
		e.CorrelationID = id
	}
}

// WithRequestID adds a request ID to the log entry
func WithRequestID(id string) func(*LogEntry) {
	return func(e *LogEntry) {
		e.RequestID = id
	}
}

// WithMethod adds HTTP method to the log entry
func WithMethod(method string) func(*LogEntry) {
	return func(e *LogEntry) {
		e.Method = method
	}
}

// WithPath adds request path to the log entry
func WithPath(path string) func(*LogEntry) {
	return func(e *LogEntry) {
		e.Path = path
	}
}

// WithStatusCode adds status code to the log entry
func WithStatusCode(code int) func(*LogEntry) {
	return func(e *LogEntry) {
		e.StatusCode = code
	}
}

// WithLatency adds latency in milliseconds to the log entry
func WithLatency(d time.Duration) func(*LogEntry) {
	return func(e *LogEntry) {
		e.LatencyMs = float64(d.Nanoseconds()) / 1e6
	}
}

// WithError adds an error to the log entry
func WithError(err error) func(*LogEntry) {
	return func(e *LogEntry) {
		if err != nil {
			e.Error = err.Error()
		}
	}
}

// WithExtra adds extra fields to the log entry
func WithExtra(key string, value interface{}) func(*LogEntry) {
	return func(e *LogEntry) {
		if e.Extra == nil {
			e.Extra = make(map[string]interface{})
		}
		e.Extra[key] = value
	}
}

// GenerateCorrelationID generates a new UUID v4 correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// GetDefaultLogger returns the default logger instance
func GetDefaultLogger() *Logger {
	once.Do(func() {
		defaultLogger = NewLogger("proxy")
	})
	return defaultLogger
}

// Package-level convenience functions using default logger

// Info logs using the default logger
func Info(msg string, fields ...func(*LogEntry)) {
	GetDefaultLogger().Info(msg, fields...)
}

// Error logs using the default logger
func Error(msg string, fields ...func(*LogEntry)) {
	GetDefaultLogger().Error(msg, fields...)
}

// Warn logs using the default logger
func Warn(msg string, fields ...func(*LogEntry)) {
	GetDefaultLogger().Warn(msg, fields...)
}

// Debug logs using the default logger
func Debug(msg string, fields ...func(*LogEntry)) {
	GetDefaultLogger().Debug(msg, fields...)
}
