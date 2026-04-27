package outgoing

import "time"

// RequestEntry holds structured data for a single intercepted outgoing HTTP request.
type RequestEntry struct {
	ID           string    `json:"id"`
	Method       string    `json:"method"`
	URL          string    `json:"url"`
	StatusCode   int       `json:"status_code"`
	DurationMs   int64     `json:"duration_ms"`
	Body         string    `json:"body"`
	ResponseBody string    `json:"response_body"`
	Timestamp    time.Time `json:"timestamp"`
}

// CacheEntry holds structured data for a single intercepted cache command.
type CacheEntry struct {
	ID         string    `json:"id"`
	Command    string    `json:"command"`
	Key        string    `json:"key"`
	DurationMs int64     `json:"duration_ms"`
	Error      string    `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
}
