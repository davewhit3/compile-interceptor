package outgoing

import (
	"sync"
)

const outgoingStoreMaxSize = 100

var (
	outgoingStore       []string
	outgoingStoreMutex  sync.Mutex
	outgoingStoreCursor int
)

func init() {
	outgoingStore = make([]string, 0, outgoingStoreMaxSize)
	outgoingStoreCursor = 0
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
