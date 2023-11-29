// The registration package is responsible for the memorization of all the
// reqeusts. It does not interact with iptables or other part of the program.
// It is made to abstract the memorization process.
package registration

import (
	"internal/comunication"
	"log/slog"
	"net"
	"time"
)

// A Registration represent a pass that is binded to a user. The pass allow a
// Mac to exit the firewall
type Registration struct {
	Id       uint64
	User     string
	Mac      string
	Ips      []net.IP
	Start    time.Time
	End      time.Time
	LastPing time.Time
	IsDown   bool
}

var (
	// The map is used to store the current Registration that are active.
	current *safeMap = newSafeMap()
	ids     uint64   = 0
)

func Add(newRequest comunication.Request) (r Registration) {
	r = Registration{Id: ids, User: newRequest.User, Mac: newRequest.Mac,
		Start: time.Now(), End: time.Now().Add(newRequest.Duration),
		Ips: []net.IP{}, LastPing: time.Now(), IsDown: false}

	ids++

	slog.With("registration", r).Debug("New registration will be added")

	current.add(r)
	// Add to db

	return
}

func Remove(r Registration) {
	slog.With("registration", r).Debug("Removing registration")
	current.remove(r.Mac)
}

func GetOldEntries() (oldEntries []Registration) {
	slog.Debug("Getting old entries")

	// Get from map
	for _, reg := range current.v {
		if time.Now().Sub(reg.End) >= 0 {
			oldEntries = append(oldEntries, reg)
		}
	}

	// Get from db
	return
}

func AddIpToMac(ip net.IP, mac net.HardwareAddr) {
	slog.With("ip", ip, "mac", mac).Debug("Binding ip to mac")
	current.addIp(mac.String(), ip)

	// Add to db
}

func GetAllEntries() (entries []Registration) {
	slog.Debug("Getting all entries")

	// Get from map
	for _, reg := range current.v {
		entries = append(entries, reg)
	}

	// Get from db
	return
}

func UpdateLastPing(e Registration) {
	//update on map
	current.mu.Lock()
	e.LastPing = time.Now()
	current.v[e.Mac] = e
	current.mu.Unlock()
}
