package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FailedRequest represents a request that failed and was sent to the DLQ
type FailedRequest struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	Error       string            `json:"error"`
	RetryCount  int               `json:"retry_count"`
	Source      string            `json:"source"` // Where the request originated (e.g., "mcp-backend", "tool-call")
	LastRetryAt time.Time         `json:"last_retry_at"`
}

// DLQMessage is the on-disk representation of a failed request
type DLQMessage struct {
	ID          string        `json:"id"`
	Timestamp   time.Time     `json:"timestamp"`
	Filename    string        `json:"filename"`
	RetryCount  int           `json:"retry_count"`
	LastRetryAt time.Time     `json:"last_retry_at"`
	FailedReq   FailedRequest `json:"failed_request"`
}

// DLQ implements a file-based Dead Letter Queue for failed requests
type DLQ struct {
	path     string
	ttlHours int
	mu       sync.RWMutex
	logger   func(format string, args ...interface{})
}

// DLQConfig holds DLQ configuration
type DLQConfig struct {
	Path     string
	TTLHours int
}

// NewDLQ creates a new DLQ instance
func NewDLQ(config *DLQConfig) (*DLQ, error) {
	if config == nil {
		config = &DLQConfig{
			Path:     "data/dlq",
			TTLHours: 24,
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create DLQ directory: %w", err)
	}

	dlq := &DLQ{
		path:     config.Path,
		ttlHours: config.TTLHours,
		logger:   func(format string, args ...interface{}) {},
	}

	return dlq, nil
}

// SetLogger sets the logging function
func (d *DLQ) SetLogger(logFunc func(format string, args ...interface{})) {
	if logFunc != nil {
		d.logger = logFunc
	}
}

// Enqueue adds a failed request to the DLQ
func (d *DLQ) Enqueue(ctx context.Context, req *FailedRequest) error {
	if req == nil {
		return fmt.Errorf("cannot enqueue nil request")
	}

	// Generate ID if not set
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// Set timestamp if not set
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Create message for storage
	msg := DLQMessage{
		ID:          req.ID,
		Timestamp:   req.Timestamp,
		Filename:    fmt.Sprintf("%d_%s.json", req.Timestamp.UnixNano(), req.ID),
		RetryCount:  req.RetryCount,
		LastRetryAt: req.LastRetryAt,
		FailedReq:   *req,
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	// Write to file
	filename := filepath.Join(d.path, msg.Filename)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write DLQ file: %w", err)
	}

	d.logger("[DLQ] Enqueued request %s to %s", req.ID, filename)

	return nil
}

// Replay processes all messages in the DLQ using the provided handler
// Returns the count of successfully processed messages
func (d *DLQ) Replay(ctx context.Context, handler func(*FailedRequest) error) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Get all pending messages
	messages, err := d.getPendingMessages()
	if err != nil {
		return fmt.Errorf("failed to get pending messages: %w", err)
	}

	if len(messages) == 0 {
		d.logger("[DLQ] No messages to replay")
		return nil
	}

	d.logger("[DLQ] Replaying %d messages", len(messages))

	var processed int
	for _, msg := range messages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute handler
		if err := handler(&msg.FailedReq); err != nil {
			// Handler failed for this message - log and continue
			d.logger("[DLQ] Replay handler failed for %s: %v", msg.ID, err)
			continue
		}

		// Success - remove from DLQ
		filename := filepath.Join(d.path, msg.Filename)
		if err := os.Remove(filename); err != nil {
			d.logger("[DLQ] Warning: failed to remove %s after replay: %v", filename, err)
		} else {
			d.logger("[DLQ] Successfully replayed and removed %s", msg.ID)
		}
		processed++
	}

	d.logger("[DLQ] Replay complete: %d/%d messages processed", processed, len(messages))
	return nil
}

// Size returns the number of pending messages in the DLQ
func (d *DLQ) Size(ctx context.Context) (int, error) {
	messages, err := d.getPendingMessages()
	if err != nil {
		return 0, err
	}
	return len(messages), nil
}

// getPendingMessages returns all messages in the DLQ, sorted by timestamp (oldest first)
func (d *DLQ) getPendingMessages() ([]DLQMessage, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.readPendingMessagesLocked()
}

// readPendingMessagesLocked reads messages from disk (caller must hold lock)
func (d *DLQ) readPendingMessagesLocked() ([]DLQMessage, error) {
	entries, err := os.ReadDir(d.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read DLQ directory: %w", err)
	}

	var messages []DLQMessage
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .json files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filename := filepath.Join(d.path, entry.Name())
		data, err := os.ReadFile(filename)
		if err != nil {
			d.logger("[DLQ] Warning: failed to read %s: %v", filename, err)
			continue
		}

		var msg DLQMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			d.logger("[DLQ] Warning: failed to parse %s: %v", filename, err)
			continue
		}

		messages = append(messages, msg)
	}

	// Sort by timestamp (oldest first - FIFO)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	return messages, nil
}

// Cleanup removes messages older than TTL
func (d *DLQ) Cleanup() (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	messages, err := d.readPendingMessagesLocked()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-time.Duration(d.ttlHours) * time.Hour)
	var removed int

	for _, msg := range messages {
		if msg.Timestamp.Before(cutoff) {
			filename := filepath.Join(d.path, msg.Filename)
			if err := os.Remove(filename); err != nil {
				d.logger("[DLQ] Warning: failed to cleanup %s: %v", filename, err)
				continue
			}
			d.logger("[DLQ] Cleaned up expired message %s (age: %v)", msg.ID, time.Since(msg.Timestamp))
			removed++
		}
	}

	return removed, nil
}

// GetMessages returns all messages for inspection (without processing)
func (d *DLQ) GetMessages(ctx context.Context) ([]DLQMessage, error) {
	return d.getPendingMessages()
}

// Peek returns a single message without removing it
func (d *DLQ) Peek(ctx context.Context) (*DLQMessage, error) {
	messages, err := d.getPendingMessages()
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, io.EOF
	}

	return &messages[0], nil
}

// Remove deletes a specific message by ID
func (d *DLQ) Remove(ctx context.Context, id string) error {
	// First, get the message to find its filename (read operation)
	messages, err := d.getPendingMessages()
	if err != nil {
		return err
	}

	var filename string
	for _, msg := range messages {
		if msg.ID == id {
			filename = filepath.Join(d.path, msg.Filename)
			break
		}
	}

	if filename == "" {
		return fmt.Errorf("message not found: %s", id)
	}

	// Now remove the file (write operation)
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to remove message: %w", err)
	}
	d.logger("[DLQ] Removed message %s", id)
	return nil
}
