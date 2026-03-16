package outgoing

import "sync"

var outgoingStore []string
var outgoingStoreMutex sync.Mutex

func init() {
	outgoingStore = make([]string, 0)
}

func Add(url string) {
	outgoingStoreMutex.Lock()
	defer outgoingStoreMutex.Unlock()
	outgoingStore = append(outgoingStore, url)
}

func Get() []string {
	return outgoingStore
}
