package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_CreatesStructuredJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("test")
	logger.SetWriter(&buf)
	logger.SetMinLevel(LevelInfo)

	logger.Info("test message",
		WithCorrelationID("test-correlation-123"),
		WithMethod("GET"),
		WithPath("/api/test"),
		WithStatusCode(200),
	)

	output := buf.String()
	assert.NotEmpty(t, output)

	// Parse JSON
	var entry LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "INFO", string(entry.Level))
	assert.Equal(t, "test message", entry.Message)
	assert.Equal(t, "test-correlation-123", entry.CorrelationID)
	assert.Equal(t, "GET", entry.Method)
	assert.Equal(t, "/api/test", entry.Path)
	assert.Equal(t, 200, entry.StatusCode)
	assert.NotEmpty(t, entry.Timestamp)

	// Verify timestamp format (RFC3339Nano)
	_, err = time.Parse(time.RFC3339Nano, entry.Timestamp)
	assert.NoError(t, err, "Timestamp should be valid RFC3339Nano")
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		minLevel  LogLevel
		logLevel  LogLevel
		shouldLog bool
	}{
		{"debug when min is debug", LevelDebug, LevelDebug, true},
		{"debug when min is info", LevelInfo, LevelDebug, false},
		{"info when min is info", LevelInfo, LevelInfo, true},
		{"info when min is warn", LevelWarn, LevelInfo, false},
		{"warn when min is warn", LevelWarn, LevelWarn, true},
		{"error when min is error", LevelError, LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger("test")
			logger.SetWriter(&buf)
			logger.SetMinLevel(tt.minLevel)

			switch tt.logLevel {
			case LevelDebug:
				logger.Debug("debug message")
			case LevelInfo:
				logger.Info("info message")
			case LevelWarn:
				logger.Warn("warn message")
			case LevelError:
				logger.Error("error message")
			}

			if tt.shouldLog {
				assert.NotEmpty(t, buf.String(), "Should log message")
			} else {
				assert.Empty(t, buf.String(), "Should not log message")
			}
		})
	}
}

func TestLogger_WithError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("test")
	logger.SetWriter(&buf)

	logger.Error("something failed",
		WithCorrelationID("corr-123"),
		WithError(assert.AnError),
	)

	var entry LogEntry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	// Error should contain "assert.AnError"
	assert.Contains(t, entry.Error, "assert.AnError")
}

func TestLogger_WithExtra(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("test")
	logger.SetWriter(&buf)

	logger.Info("request processed",
		WithExtra("user_id", "user-42"),
		WithExtra("request_size", 1024),
	)

	var entry LogEntry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.NotNil(t, entry.Extra)
	assert.Equal(t, "user-42", entry.Extra["user_id"])
	assert.Equal(t, float64(1024), entry.Extra["request_size"])
}

func TestLogger_WithLatency(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("test")
	logger.SetWriter(&buf)

	logger.Info("request completed",
		WithLatency(150*time.Millisecond),
	)

	var entry LogEntry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	// Latency should be in milliseconds (150ms = 150.0)
	assert.InDelta(t, 150.0, entry.LatencyMs, 0.1)
}

func TestGenerateCorrelationID(t *testing.T) {
	id1 := GenerateCorrelationID()
	id2 := GenerateCorrelationID()

	// Should be non-empty
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)

	// Should be unique
	assert.NotEqual(t, id1, id2)

	// Should be valid UUID format (8-4-4-4-12)
	parts := strings.Split(id1, "-")
	assert.Len(t, parts, 5)
	assert.Len(t, parts[0], 8)
	assert.Len(t, parts[1], 4)
	assert.Len(t, parts[2], 4)
	assert.Len(t, parts[3], 4)
	assert.Len(t, parts[4], 12)
}

func TestLogger_ComponentDefault(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("custom-component")
	logger.SetWriter(&buf)

	logger.Info("test")

	var entry LogEntry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "custom-component", entry.Component)
}

func TestLogger_AllLevels(t *testing.T) {
	levels := []LogLevel{LevelDebug, LevelInfo, LevelWarn, LevelError}
	expected := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for i, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger("test")
			logger.SetWriter(&buf)
			logger.SetMinLevel(LevelDebug)

			switch level {
			case LevelDebug:
				logger.Debug("debug msg")
			case LevelInfo:
				logger.Info("info msg")
			case LevelWarn:
				logger.Warn("warn msg")
			case LevelError:
				logger.Error("error msg")
			}

			var entry LogEntry
			err := json.Unmarshal(buf.Bytes(), &entry)
			require.NoError(t, err)
			assert.Equal(t, expected[i], string(entry.Level))
		})
	}
}

func TestLogEntry_AllFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("full-test")
	logger.SetWriter(&buf)
	logger.SetMinLevel(LevelInfo)

	logger.Error("comprehensive test",
		WithCorrelationID("corr-full"),
		WithRequestID("req-full"),
		WithMethod("POST"),
		WithPath("/api/comprehensive"),
		WithStatusCode(500),
		WithLatency(250*time.Millisecond),
		WithError(assert.AnError),
		WithExtra("extra_field", "extra_value"),
	)

	var entry LogEntry
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "ERROR", string(entry.Level))
	assert.Equal(t, "comprehensive test", entry.Message)
	assert.Equal(t, "corr-full", entry.CorrelationID)
	assert.Equal(t, "req-full", entry.RequestID)
	assert.Equal(t, "POST", entry.Method)
	assert.Equal(t, "/api/comprehensive", entry.Path)
	assert.Equal(t, 500, entry.StatusCode)
	assert.InDelta(t, 250.0, entry.LatencyMs, 0.1)
	assert.NotEmpty(t, entry.Error)
	assert.Equal(t, "extra_value", entry.Extra["extra_field"])
}

func TestGetCorrelationID(t *testing.T) {
	// Test with empty context
	ctx := context.Background()
	id := getCorrelationID(ctx)
	assert.Empty(t, id)

	// Test with correlation ID in context
	ctx = context.WithValue(ctx, correlationIDKey, "test-id-456")
	id = getCorrelationID(ctx)
	assert.Equal(t, "test-id-456", id)
}
