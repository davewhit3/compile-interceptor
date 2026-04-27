package outgoing

import "time"

// RequestEntry holds structured data for a single outgoing HTTP request.
type RequestEntry struct {
	ID         string    `json:"id"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	DurationMs int64     `json:"duration_ms"`
	Body       string    `json:"body"`
	Timestamp  time.Time `json:"timestamp"`
}
