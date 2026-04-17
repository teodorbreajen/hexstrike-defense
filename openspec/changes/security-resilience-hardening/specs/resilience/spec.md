# Resilience Spec — New

## Purpose

Define retry logic with exponential backoff and dead letter queue for failed operations to ensure system resilience.

## ADDED Requirements

### Requirement: HTTP Client Retry with Exponential Backoff

The HTTP client wrapper MUST implement retry logic with exponential backoff for transient failures. The system SHALL retry failed operations up to 3 times with delays of 1s, 2s, and 4s.

**Retryable Conditions:**
- Network timeout (context deadline exceeded)
- HTTP 5xx responses from backend
- Connection reset/refused errors

**Non-Retryable Conditions:**
- HTTP 4xx client errors (except 429 Too Many Requests)
- HTTP 200-299 success responses
- Body parsing errors

**Backoff Schedule:**
| Attempt | Delay |
|---------|-------|
| 1 | 1 second |
| 2 | 2 seconds |
| 3 | 4 seconds |

#### Scenario: Successful request on first attempt

- GIVEN HTTP endpoint is healthy
- WHEN `HTTPClient.Do()` is called
- THEN request SHALL complete on first attempt
- AND response SHALL be returned immediately

#### Scenario: Retry succeeds on second attempt

- GIVEN first request times out, second succeeds
- WHEN `HTTPClient.Do()` is called
- THEN first attempt SHALL fail after timeout
- AND second attempt SHALL be made after 1 second delay
- AND response SHALL be returned
- AND log SHALL show "attempt 1/3 failed, retrying in 1s"

#### Scenario: All retries exhausted

- GIVEN all 3 attempts fail
- WHEN `HTTPClient.Do()` is called
- THEN third retry SHALL occur after 4 second delay
- AND error SHALL be returned after final failure
- AND log SHALL show "attempt 3/3 failed: max retries exceeded"
- AND message SHALL be sent to DLQ if configured

#### Scenario: Non-retryable error (4xx) fails immediately

- GIVEN HTTP endpoint returns 400 Bad Request
- WHEN `HTTPClient.Do()` is called
- THEN request SHALL NOT be retried
- AND error SHALL be returned immediately
- AND log SHALL show "non-retryable error: HTTP 400"

### Requirement: Dead Letter Queue (DLQ)

The system MUST implement a dead letter queue for messages that fail after maximum retry attempts. Failed messages SHALL be persisted temporarily for replay.

**DLQ Behavior:**
- Messages are persisted to disk in a designated directory
- Each message is stored as a JSON file with metadata
- TTL for messages: 24 hours (configurable via `DLQ_TTL_HOURS`)
- DLQ directory: `data/dlq/` (configurable via `DLQ_PATH`)

**Message Format:**
```json
{
  "id": "uuid",
  "timestamp": "2026-04-17T10:30:00Z",
  "payload": { /* original request */ },
  "error": "last error message",
  "retry_count": 3,
  "source": "http-client" | "db-operation"
}
```

#### Scenario: Failed message is sent to DLQ

- GIVEN HTTP request fails after 3 retries
- WHEN `HTTPClient.Do()` returns final error
- THEN message SHALL be persisted to DLQ
- AND file SHALL be created at `{DLQ_PATH}/{uuid}.json`
- AND log SHALL show "message queued to DLQ: {uuid}"

#### Scenario: DLQ messages have TTL expiration

- GIVEN a DLQ message older than `DLQ_TTL_HOURS`
- WHEN cleanup job runs
- THEN expired messages SHALL be deleted
- AND log SHALL show "cleaned up {count} expired DLQ messages"

#### Scenario: DLQ can be replayed

- GIVEN DLQ contains messages
- WHEN replay script/endpoint is triggered
- THEN messages SHALL be reprocessed in FIFO order
- AND successful reprocessing SHALL remove message from DLQ
- AND log SHALL show "DLQ replay: {count} messages processed"

### Requirement: Retry Logging

The system SHALL log retry attempts with sufficient detail for debugging and monitoring.

**Required Log Fields:**
- `retry_attempt`: Current attempt number (1-3)
- `max_retries`: Maximum retries configured
- `delay_ms`: Actual delay before retry
- `error`: Error message from failed attempt
- `will_retry`: Boolean indicating if more retries remain

#### Scenario: Retry attempt is logged

- GIVEN a request fails and will be retried
- WHEN retry occurs
- THEN log entry SHALL include:
  - `level: WARN`
  - `message: "retry attempt {n}/{max} after {delay}ms"`
  - `error: "{error_details}"`
