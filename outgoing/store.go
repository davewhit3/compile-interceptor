package outgoing

import (
	"fmt"
	"sync"
	"time"
)

const outgoingStoreMaxSize = 100

var (
	outgoingStore       []string
	outgoingStoreMutex  sync.Mutex
	outgoingStoreCursor int

	entryStore       []RequestEntry
	entryStoreMutex  sync.Mutex
	entryStoreCursor int
	entryStoreSeq    uint64
)

func init() {
	createStore()
}

func createStore() {
	outgoingStore = make([]string, 0, outgoingStoreMaxSize)
	outgoingStoreCursor = 0

	entryStore = make([]RequestEntry, 0, outgoingStoreMaxSize)
	entryStoreCursor = 0
}

func Add(url string) {
	outgoingStoreMutex.Lock()
	defer outgoingStoreMutex.Unlock()

	if len(outgoingStore) < outgoingStoreMaxSize {
		outgoingStore = append(outgoingStore, url)
	} else {
		outgoingStore[outgoingStoreCursor] = url
	}
	outgoingStoreCursor = (outgoingStoreCursor + 1) % outgoingStoreMaxSize
}

func List() []string {
	outgoingStoreMutex.Lock()
	defer outgoingStoreMutex.Unlock()

	n := len(outgoingStore)
	result := make([]string, n)

	for i := 0; i < n; i++ {
		if n < outgoingStoreMaxSize {
			result[i] = outgoingStore[n-1-i]
		} else {
			idx := (outgoingStoreCursor - 1 - i + outgoingStoreMaxSize) % outgoingStoreMaxSize
			result[i] = outgoingStore[idx]
		}
	}

	return result
}

// AddRequest stores a structured RequestEntry in the ring buffer.
func AddRequest(method, url string, code int, dur time.Duration, body string) {
	entryStoreMutex.Lock()
	defer entryStoreMutex.Unlock()

	entryStoreSeq++
	e := RequestEntry{
		ID:         fmt.Sprintf("%016x", entryStoreSeq),
		Method:     method,
		URL:        url,
		StatusCode: code,
		DurationMs: dur.Milliseconds(),
		Body:       body,
		Timestamp:  time.Now().UTC(),
	}

	if len(entryStore) < outgoingStoreMaxSize {
		entryStore = append(entryStore, e)
	} else {
		entryStore[entryStoreCursor] = e
	}
	entryStoreCursor = (entryStoreCursor + 1) % outgoingStoreMaxSize
}

// ListEntries returns all stored RequestEntry values, newest first.
func ListEntries() []RequestEntry {
	entryStoreMutex.Lock()
	defer entryStoreMutex.Unlock()

	n := len(entryStore)
	result := make([]RequestEntry, n)

	for i := 0; i < n; i++ {
		if n < outgoingStoreMaxSize {
			result[i] = entryStore[n-1-i]
		} else {
			idx := (entryStoreCursor - 1 - i + outgoingStoreMaxSize) % outgoingStoreMaxSize
			result[i] = entryStore[idx]
		}
	}

	return result
}

func Reset() {
	outgoingStoreMutex.Lock()
	defer outgoingStoreMutex.Unlock()
	entryStoreMutex.Lock()
	defer entryStoreMutex.Unlock()
	createStore()
}
