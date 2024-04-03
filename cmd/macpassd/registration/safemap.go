package registration

import (
	"log"
	"log/slog"
	"net"
	"sync"
)

// A safe map is a map that have a Mutex condition in order to support
// concurrency
type safeMap struct {
	v sync.Map
}

func newSafeMap() *safeMap {
	return &safeMap{}
}

func (m *safeMap) add(r Registration) {
	m.v.Store(r.Mac, r)
}

func (m *safeMap) remove(mac string) {
	m.v.Delete(mac)
}

// This function adds the ip to the registration associated with the mac address
// in the m map. If the ip is in the OldIPs files nothing is done
func (m *safeMap) addIp(mac string, ip net.IP) {
	aval, present := m.v.Load(mac)
	if present {
		val, ok := aval.(Registration)

		if !ok {
			slog.With("ok", ok, "aval", aval).
				Error("Type assertion failed. Could not converto aval to type Registration")
			log.Fatal("Could not continue")
		}

		if !isIpPrenset(val.Ips, ip) && !isIpPrenset(val.OldIps, ip) {
			val.Ips = append(val.Ips, ip) //Need to check for duplicates
			slog.With("registration", val.String()).
				Info("The registration is been updated on the map")
		}

		m.v.Store(mac, val)
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
