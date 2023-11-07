package registration

import (
	"log"
	"net"
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

func (m *safeMap) addIp(mac string, ip net.IP) {
	_, present := current.v[mac]
	if present {
		current.mu.Lock()
		val := current.v[mac]

		if !isIpPrenset(val.Ips, ip) {
			val.Ips = append(val.Ips, ip) //Need to check for duplicates
		}

		current.v[mac] = val
		current.mu.Unlock()

		log.Println("The registration is been updated on the map: ", val)
	}
}

func isIpPrenset(set []net.IP, ip net.IP) bool {
	if len(set) == 0 {
		return false
	}

	// Likely because most devices have 1 ip
	if set[0].Equal(ip) {
		return true
	}

	for _, i := range set {
		if i.Equal(ip) {
			return true
		}
	}
	return false
}
