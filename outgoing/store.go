package outgoing

import (
	"fmt"
	"sync"
	"time"
)

const storeMaxSize = 100

// ringBuffer is an unexported, thread-safe fixed-size ring buffer.
// Each type gets its own instance with its own mutex and sequence counter.
type ringBuffer[T any] struct {
	entries []T
	mu      sync.Mutex
	cursor  int
	seq     uint64
}

func newRingBuffer[T any]() *ringBuffer[T] {
	return &ringBuffer[T]{entries: make([]T, 0, storeMaxSize)}
}

// nextID increments the sequence and returns a hex ID string.
// Caller must hold rb.mu.
func (rb *ringBuffer[T]) nextID() string {
	rb.seq++
	return fmt.Sprintf("%016x", rb.seq)
}

// push appends or overwrites the oldest slot in the ring.
// Caller must hold rb.mu.
func (rb *ringBuffer[T]) push(e T) {
	if len(rb.entries) < storeMaxSize {
		rb.entries = append(rb.entries, e)
	} else {
		rb.entries[rb.cursor] = e
	}
	rb.cursor = (rb.cursor + 1) % storeMaxSize
}

// listAll returns all entries newest-first.
func (rb *ringBuffer[T]) listAll() []T {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	n := len(rb.entries)
	result := make([]T, n)
	for i := 0; i < n; i++ {
		if n < storeMaxSize {
			result[i] = rb.entries[n-1-i]
		} else {
			idx := (rb.cursor - 1 - i + storeMaxSize) % storeMaxSize
			result[i] = rb.entries[idx]
		}
	}
	return result
}

// clear resets the buffer, preserving the sequence counter.
func (rb *ringBuffer[T]) clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.entries = make([]T, 0, storeMaxSize)
	rb.cursor = 0
}

// ── per-type stores ───────────────────────────────────────────────────────────

var requests = newRingBuffer[RequestEntry]()
var commands = newRingBuffer[CacheEntry]()

// ── HTTP requests ─────────────────────────────────────────────────────────────

// AddRequest stores an outgoing HTTP request in the ring buffer.
func AddRequest(method, url string, code int, dur time.Duration, body, responseBody string) {
	requests.mu.Lock()
	defer requests.mu.Unlock()

	e := RequestEntry{
		ID:           requests.nextID(),
		Method:       method,
		URL:          url,
		StatusCode:   code,
		DurationMs:   dur.Milliseconds(),
		Body:         body,
		ResponseBody: responseBody,
		Timestamp:    time.Now().UTC(),
	}
	requests.push(e)
}

// ListRequests returns all stored RequestEntry values, newest first.
func ListRequests() []RequestEntry {
	return requests.listAll()
}

// ResetRequests clears the HTTP request store.
func ResetRequests() {
	requests.clear()
}

// ── cache commands ────────────────────────────────────────────────────────────

// AddCommand stores an intercepted cache command in the ring buffer.
func AddCommand(command, key string, dur time.Duration, errStr string) {
	commands.mu.Lock()
	defer commands.mu.Unlock()

	e := CacheEntry{
		ID:         commands.nextID(),
		Command:    command,
		Key:        key,
		DurationMs: dur.Milliseconds(),
		Error:      errStr,
		Timestamp:  time.Now().UTC(),
	}
	commands.push(e)
}

// ListCommands returns all stored CacheEntry values, newest first.
func ListCommands() []CacheEntry {
	return commands.listAll()
}

// ResetCommands clears the cache command store.
func ResetCommands() {
	commands.clear()
}

// ── global reset ──────────────────────────────────────────────────────────────

// Reset clears all stores.
func Reset() {
	requests.clear()
	commands.clear()
}
