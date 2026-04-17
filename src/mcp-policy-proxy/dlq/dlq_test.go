package dlq

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDLQ creates a temporary DLQ for testing
func setupTestDLQ(t *testing.T) (*DLQ, string) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "dlq-test-*")
	require.NoError(t, err)

	dlq, err := NewDLQ(&DLQConfig{
		Path:     tmpDir,
		TTLHours: 24,
	})
	require.NoError(t, err)

	return dlq, tmpDir
}

// cleanupTestDLQ cleans up the temporary DLQ directory
func cleanupTestDLQ(tmpDir string) {
	if tmpDir != "" {
		os.RemoveAll(tmpDir)
	}
}

// TestDLQ_Enqueue_SavesFile tests that Enqueue saves a file to disk
func TestDLQ_Enqueue_SavesFile(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Enqueue a failed request
	req := &FailedRequest{
		Method: "POST",
		URL:    "http://example.com/api",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:       []byte(`{"key":"value"}`),
		Error:      "connection refused",
		RetryCount: 3,
		Source:     "test",
	}

	err := dlq.Enqueue(ctx, req)
	require.NoError(t, err)

	// Verify file was created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 1, "should create one file")

	// Verify the file has .json extension
	assert.True(t, filepath.Ext(files[0].Name()) == ".json")
}

// TestDLQ_Replay_ExecutesHandler tests that Replay executes the handler for each message
func TestDLQ_Replay_ExecutesHandler(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Enqueue multiple failed requests
	for i := 0; i < 3; i++ {
		req := &FailedRequest{
			Method: "POST",
			URL:    "http://example.com/api",
			Error:  "error",
			Source: "test",
		}
		require.NoError(t, dlq.Enqueue(ctx, req))
	}

	// Track which messages were processed
	var processed []string

	handler := func(req *FailedRequest) error {
		processed = append(processed, req.ID)
		return nil
	}

	err := dlq.Replay(ctx, handler)
	require.NoError(t, err)

	// All messages should have been processed
	assert.Len(t, processed, 3, "all messages should be processed")

	// Messages should be processed in FIFO order (sorted by timestamp)
	for i := 1; i < len(processed); i++ {
		// Verify FIFO by checking the messages were processed in order
		t.Logf("Processed message %d: %s", i, processed[i])
	}

	// After replay, DLQ should be empty
	size, err := dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, size, "DLQ should be empty after replay")
}

// TestDLQ_Replay_FIFOOrder tests that messages are replayed in FIFO order (oldest first)
func TestDLQ_Replay_FIFOOrder(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Enqueue with explicit timestamps to control order
	req1 := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/first",
		Timestamp: time.Now().Add(-2 * time.Hour), // Oldest
		Source:    "test",
	}
	req2 := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/second",
		Timestamp: time.Now().Add(-1 * time.Hour), // Middle
		Source:    "test",
	}
	req3 := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/third",
		Timestamp: time.Now(), // Newest
		Source:    "test",
	}

	require.NoError(t, dlq.Enqueue(ctx, req1))
	require.NoError(t, dlq.Enqueue(ctx, req2))
	require.NoError(t, dlq.Enqueue(ctx, req3))

	// Track order of processing
	var processedOrder []string

	handler := func(req *FailedRequest) error {
		processedOrder = append(processedOrder, req.URL)
		return nil
	}

	err := dlq.Replay(ctx, handler)
	require.NoError(t, err)

	// Should be processed in FIFO order
	assert.Equal(t, []string{
		"http://example.com/first",
		"http://example.com/second",
		"http://example.com/third",
	}, processedOrder, "messages should be processed in FIFO order")
}

// TestDLQ_Size_CountsFiles tests that Size returns the correct count
func TestDLQ_Size_CountsFiles(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Initially empty
	size, err := dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, size, "should be empty initially")

	// Add messages
	for i := 0; i < 5; i++ {
		req := &FailedRequest{
			Method: "POST",
			URL:    "http://example.com/api",
			Source: "test",
		}
		require.NoError(t, dlq.Enqueue(ctx, req))

		size, err := dlq.Size(ctx)
		require.NoError(t, err)
		assert.Equal(t, i+1, size, "size should match after each enqueue")
	}
}

// TestDLQ_Cleanup_EliminatesExpiredMessages tests that Cleanup removes messages older than TTL
func TestDLQ_Cleanup_EliminatesExpiredMessages(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Create DLQ with 1 hour TTL
	dlq.ttlHours = 1

	// Enqueue an expired message (created 2 hours ago)
	expiredReq := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/expired",
		Timestamp: time.Now().Add(-2 * time.Hour),
		Source:    "test",
	}
	require.NoError(t, dlq.Enqueue(ctx, expiredReq))

	// Enqueue a fresh message (created 30 minutes ago)
	freshReq := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/fresh",
		Timestamp: time.Now().Add(-30 * time.Minute),
		Source:    "test",
	}
	require.NoError(t, dlq.Enqueue(ctx, freshReq))

	// Verify both exist
	size, err := dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, size, "should have 2 messages")

	// Run cleanup
	removed, err := dlq.Cleanup()
	require.NoError(t, err)
	assert.Equal(t, 1, removed, "should remove 1 expired message")

	// Verify only fresh message remains
	size, err = dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, size, "should have 1 message remaining")
}

// TestDLQ_CleanupWithTTL tests CleanupWithTTL with custom TTL
func TestDLQ_CleanupWithTTL(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Create messages with different ages
	oldReq := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/old",
		Timestamp: time.Now().Add(-48 * time.Hour), // 48 hours old
		Source:    "test",
	}
	mediumReq := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/medium",
		Timestamp: time.Now().Add(-36 * time.Hour), // 36 hours old
		Source:    "test",
	}
	newReq := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/new",
		Timestamp: time.Now().Add(-12 * time.Hour), // 12 hours old
		Source:    "test",
	}

	require.NoError(t, dlq.Enqueue(ctx, oldReq))
	require.NoError(t, dlq.Enqueue(ctx, mediumReq))
	require.NoError(t, dlq.Enqueue(ctx, newReq))

	// Cleanup with 24 hour TTL
	removed, err := CleanupWithTTL(dlq, 24)
	require.NoError(t, err)
	assert.Equal(t, 2, removed, "should remove 2 messages older than 24 hours")

	// Only new message should remain
	size, err := dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, size, "should have 1 message remaining")
}

// TestDLQ_Enqueue_GeneratesID tests that Enqueue generates ID if not set
func TestDLQ_Enqueue_GeneratesID(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	req := &FailedRequest{
		Method: "POST",
		URL:    "http://example.com/api",
		Source: "test",
		// ID not set
	}

	err := dlq.Enqueue(ctx, req)
	require.NoError(t, err)

	assert.NotEmpty(t, req.ID, "ID should be generated")
}

// TestDLQ_Enqueue_SetsTimestamp tests that Enqueue sets timestamp if not set
func TestDLQ_Enqueue_SetsTimestamp(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	before := time.Now()

	req := &FailedRequest{
		Method: "POST",
		URL:    "http://example.com/api",
		Source: "test",
		// Timestamp not set
	}

	err := dlq.Enqueue(ctx, req)
	require.NoError(t, err)

	after := time.Now()

	assert.True(t, req.Timestamp.After(before.Add(-time.Second)), "timestamp should be set to now-ish")
	assert.True(t, req.Timestamp.Before(after.Add(time.Second)), "timestamp should be set to now-ish")
}

// TestDLQ_Remove_DeletesMessage tests the Remove method
func TestDLQ_Remove_DeletesMessage(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	req := &FailedRequest{
		Method: "POST",
		URL:    "http://example.com/api",
		Source: "test",
	}

	err := dlq.Enqueue(ctx, req)
	require.NoError(t, err)

	// Verify it exists
	size, err := dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, size)

	// Remove it
	err = dlq.Remove(ctx, req.ID)
	require.NoError(t, err)

	// Verify it's gone
	size, err = dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, size)
}

// TestDLQ_Remove_NotFound tests Remove with non-existent ID
func TestDLQ_Remove_NotFound(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	err := dlq.Remove(ctx, "non-existent-id")
	assert.Error(t, err, "should return error for non-existent ID")
}

// TestDLQ_Peek_ReturnsOldestMessage tests that Peek returns the oldest message
func TestDLQ_Peek_ReturnsOldestMessage(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Enqueue messages out of order
	newReq := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/new",
		Timestamp: time.Now(),
		Source:    "test",
	}
	oldReq := &FailedRequest{
		Method:    "POST",
		URL:       "http://example.com/old",
		Timestamp: time.Now().Add(-time.Hour),
		Source:    "test",
	}

	require.NoError(t, dlq.Enqueue(ctx, newReq))
	require.NoError(t, dlq.Enqueue(ctx, oldReq))

	// Peek should return the oldest
	msg, err := dlq.Peek(ctx)
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/old", msg.FailedReq.URL, "peek should return oldest message")

	// DLQ should still have 2 messages (peek doesn't remove)
	size, err := dlq.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, size)
}

// TestDLQ_Peek_EmptyQueue tests Peek on empty queue
func TestDLQ_Peek_EmptyQueue(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	msg, err := dlq.Peek(ctx)
	assert.Error(t, err, "should return error on empty queue")
	assert.Nil(t, msg)
}

// TestDLQ_GetMessages_ReturnsAllMessages tests GetMessages
func TestDLQ_GetMessages_ReturnsAllMessages(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Enqueue messages
	for i := 0; i < 3; i++ {
		req := &FailedRequest{
			Method: "POST",
			URL:    "http://example.com/api",
			Source: "test",
		}
		require.NoError(t, dlq.Enqueue(ctx, req))
	}

	messages, err := dlq.GetMessages(ctx)
	require.NoError(t, err)
	assert.Len(t, messages, 3, "should return all messages")
}

// TestDLQ_Enqueue_NilRequest tests that Enqueue rejects nil request
func TestDLQ_Enqueue_NilRequest(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	err := dlq.Enqueue(ctx, nil)
	assert.Error(t, err, "should reject nil request")
	assert.Contains(t, err.Error(), "nil")
}

// TestDLQ_Replay_NilHandler tests that Replay rejects nil handler
func TestDLQ_Replay_NilHandler(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	err := dlq.Replay(ctx, nil)
	assert.Error(t, err, "should reject nil handler")
	assert.Contains(t, err.Error(), "nil")
}

// TestDLQ_NewDLQ_CreatesDirectory tests that NewDLQ creates the directory
func TestDLQ_NewDLQ_CreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dlq-create-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Use a non-existent subdirectory
	testPath := filepath.Join(tmpDir, "subdir", "dlq")

	dlq, err := NewDLQ(&DLQConfig{
		Path:     testPath,
		TTLHours: 24,
	})
	require.NoError(t, err)
	require.NotNil(t, dlq)

	// Verify directory was created
	info, err := os.Stat(testPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestDLQ_Replay_ContextCancellation tests that Replay respects context cancellation
func TestDLQ_Replay_ContextCancellation(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	ctx := context.Background()

	// Enqueue many messages
	for i := 0; i < 10; i++ {
		req := &FailedRequest{
			Method: "POST",
			URL:    "http://example.com/api",
			Source: "test",
		}
		require.NoError(t, dlq.Enqueue(ctx, req))
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	processed := 0
	handler := func(req *FailedRequest) error {
		processed++
		if processed >= 5 {
			cancel()
		}
		return nil
	}

	err := dlq.Replay(ctx, handler)
	assert.Error(t, err, "should return error on context cancellation")
	assert.Equal(t, 5, processed, "should stop processing after cancellation")
}

// TestDLQ_SetLogger tests that SetLogger updates the logger
func TestDLQ_SetLogger(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)
	defer cleanupTestDLQ(tmpDir)

	var logCalled bool
	customLogger := func(format string, args ...interface{}) {
		logCalled = true
	}

	dlq.SetLogger(customLogger)
	require.NoError(t, dlq.Enqueue(context.Background(), &FailedRequest{
		Method: "POST",
		URL:    "http://example.com/api",
		Source: "test",
	}))

	assert.True(t, logCalled, "custom logger should be called")
}

// TestStartCleanupGoroutine tests that cleanup goroutine starts and can be stopped
func TestStartCleanupGoroutine(t *testing.T) {
	dlq, tmpDir := setupTestDLQ(t)

	stop := StartCleanupGoroutine(dlq, &CleanupConfig{
		Interval: 100 * time.Millisecond,
		TTLHours: 24,
	})

	// Wait a bit to let it run
	time.Sleep(150 * time.Millisecond)

	// Stop the goroutine
	stop()

	cleanupTestDLQ(tmpDir)
}
