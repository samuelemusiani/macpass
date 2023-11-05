// The registration package is responsible for the memorization of all the
// reqeusts. It does not interact with iptables or other part of the program.
// It is made to abstract the memorization process.
package registration

import (
	"internal/comunication"
	"time"
)

// A Registration represent a pass that is binded to a user. The pass allow a
// Mac to exit the firewall
type Registration struct {
	Id   uint64
	User string
	Mac  string
	// ip(s)
	Start  time.Time
	End    time.Time
	IsDown bool
}

// The map is used to store the current Registration that are active.
var current safeMap
var ids uint64 = 0

func Add(newRequest comunication.Request) (r Registration) {
	r = Registration{Id: ids, User: newRequest.User, Mac: newRequest.Mac,
		Start: time.Now(), End: time.Now().Add(newRequest.Duration)}
	ids++

	current.add(r)
	// Add to db

	return
}

func Remove(r Registration) {
	current.remove(r.Mac)
}

func GetOldEntries() (oldEntries []Registration) {
	// Get from map

	for _, reg := range current.v {
		if time.Now().Sub(reg.End) >= 0 {
			oldEntries = append(oldEntries, reg)
		}
	}

	// Get from db
	return
}
