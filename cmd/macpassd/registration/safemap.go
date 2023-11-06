package registration

import (
	"sync"
)

// A safe map is a map that have a Mutex condition in order to support
// concurrency
type safeMap struct {
	mu sync.Mutex
	v  map[string]Registration
}

func newSafeMap() *safeMap {
	return &safeMap{v: make(map[string]Registration)}
}

func (m *safeMap) add(r Registration) {
	m.mu.Lock()
	m.v[r.Mac] = r
	m.mu.Unlock()
}

func (m *safeMap) remove(mac string) {
	m.mu.Lock()
	delete(m.v, mac)
	m.mu.Unlock()
}
