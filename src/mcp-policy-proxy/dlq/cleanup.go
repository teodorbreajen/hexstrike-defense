package dlq

import (
	"context"
	"os"
	"sync"
	"time"
)

// CleanupConfig holds cleanup goroutine configuration
type CleanupConfig struct {
	Interval time.Duration // How often to run cleanup (default: 1 hour)
	TTLHours int           // Messages older than this are removed (default: 24)
}

// StartCleanupGoroutine starts a background goroutine that periodically cleans up expired messages
// Returns a stop function that should be called when shutting down
func StartCleanupGoroutine(dlq *DLQ, config *CleanupConfig) (stop func()) {
	if config == nil {
		config = &CleanupConfig{
			Interval: 1 * time.Hour,
			TTLHours: 24,
		}
	}

	if config.Interval <= 0 {
		config.Interval = 1 * time.Hour
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		// Run initial cleanup on startup
		runCleanup(dlq)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCleanup(dlq)
			}
		}
	}()

	return func() {
		cancel()
		wg.Wait()
	}
}

// runCleanup performs the actual cleanup operation
func runCleanup(dlq *DLQ) {
	if dlq == nil {
		return
	}

	removed, err := dlq.Cleanup()
	if err != nil {
		dlq.logger("[DLQ Cleanup] Error during cleanup: %v", err)
		return
	}

	if removed > 0 {
		dlq.logger("[DLQ Cleanup] Removed %d expired messages", removed)
	}
}

// CleanupNow triggers an immediate cleanup (useful for testing)
func CleanupNow(dlq *DLQ) (int, error) {
	if dlq == nil {
		return 0, nil
	}
	return dlq.Cleanup()
}

// CleanupWithTTL removes messages older than the specified TTL
// This is useful for manual cleanup with custom TTL
func CleanupWithTTL(dlq *DLQ, ttlHours int) (int, error) {
	if dlq == nil {
		return 0, nil
	}

	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	messages, err := dlq.readPendingMessagesLocked()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-time.Duration(ttlHours) * time.Hour)
	var removed int

	for _, msg := range messages {
		if msg.Timestamp.Before(cutoff) {
			filename := dlq.path + "/" + msg.Filename
			if err := os.Remove(filename); err != nil {
				dlq.logger("[DLQ] Warning: failed to cleanup %s: %v", filename, err)
				continue
			}
			dlq.logger("[DLQ] Cleaned up expired message %s (age: %v)", msg.ID, time.Since(msg.Timestamp))
			removed++
		}
	}

	return removed, nil
}
