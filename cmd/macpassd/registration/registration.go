// The registration package is responsible for the memorization of all the
// reqeusts. It does not interact with iptables or other part of the program.
// It is made to abstract the memorization process.
package registration

import (
	"database/sql"
	"fmt"
	"internal/comunication"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpass/config"
)

// A Registration represent a pass that is binded to a user. The pass allow a
// Mac to exit the firewall
type Registration struct {
	Id       uint64
	User     string
	Mac      string
	Ips      []net.IP
	OldIps   []net.IP
	Start    time.Time
	End      time.Time
	LastPing time.Time
	IsDown   bool
}

var (
	// The map is used to store the currentMap Registration that are active.
	currentMap *safeMap = nil
	ids        uint64   = 0
	currentDB  *sql.DB  = nil
)

func Init() {
	currentMap = newSafeMap()
	if currentMap == nil {
		log.Fatal("Can't create map in memory. Something wrong. Can't continue")
	}

	p := config.Get().DBPath
	if p != "" {
		currentDB = dbConnect(p)

		if currentDB == nil {
			slog.With("path", p).Error("Connection to db is nil. No database will be used")
		} else {
			slog.Info("Connected to db")
		}
	} else {
		slog.Info("Db path is empty, disabling db")
		currentDB = nil
	}
}

func AddRequest(newRequest comunication.Request) (r Registration) {
	r = Registration{Id: ids, User: newRequest.User, Mac: newRequest.Mac,
		Start: time.Now(), End: time.Now().Add(newRequest.Duration),
		Ips: []net.IP{}, OldIps: []net.IP{}, LastPing: time.Now(), IsDown: false}

	ids++

	slog.With("registration", r.String()).Debug("New registration will be added")

	currentMap.add(r)

	// Add to db
	if currentDB != nil {
		dbInsertRegistration(currentDB, r)
	}

	return
}

func AddRegistrationToMapFromDB(r Registration) {
	slog.With("registration", r.String()).Debug("Registration from DB will be added")
	currentMap.add(r)
	return
}

func Remove(r Registration) {
	slog.With("registration", r.String()).Debug("Removing registration")
	currentMap.remove(r.Mac)
}

// Old entries are registration on the hashmap that have the connection time
// expired.
func GetOldEntries() (oldEntries []Registration) {
	// Get from map
	currentMap.v.Range(func(key, value any) bool {
		val, ok := value.(Registration)
		if !ok {
			slog.With("value", value, "ok", ok).
				Error("Type assertion failed. Could not converto aval to type Registration")
			log.Fatal("Could not continue")
			return false
		}
		if time.Now().Sub(val.End) >= 0 {
			oldEntries = append(oldEntries, val)
		}
		return true
	})

	// Get from db ?
	return
}

// This function adds the ip to the registration associated with the mac address.
// If the ip is in the Ips or OldIps fields nothing is done
func AddIpToMac(ip net.IP, mac net.HardwareAddr) {
	slog.With("ip", ip, "mac", mac.String()).Debug("Binding ip to mac")
	currentMap.addIp(mac.String(), ip)

	// TODO: Add to db
}

func GetAllEntries() (entries []Registration) {
	// Get from map
	currentMap.v.Range(func(key, value any) bool {
		val, ok := value.(Registration)
		if !ok {
			slog.With("value", value, "ok", ok).
				Error("Type assertion failed. Could not converto aval to type Registration")
			log.Fatal("Could not continue")
			return false
		}
		entries = append(entries, val)
		return true
	})

	// Get from db
	return
}

func UpdateLastPing(e Registration) {
	//update on map
	e.LastPing = time.Now()
	currentMap.v.Store(e.Mac, e)
}

func SetHostDown(e Registration) {
	//update on map
	e.IsDown = true
	currentMap.v.Store(e.Mac, e)
}

func SetHostUp(e Registration) {
	//update on map
	e.IsDown = false
	currentMap.v.Store(e.Mac, e)
}

// This function removes the ip form the Ips field of a registration and save
// the new registration to the hashmap.
func RemoveIP(e Registration, ip net.IP) {
	newIps := make([]net.IP, 0)

	for _, i := range e.Ips {
		if !i.Equal(ip) {
			newIps = append(newIps, i)
		}
	}

	e.Ips = newIps
	currentMap.v.Store(e.Mac, e)
}

// This function removes the ip form the Ips field of a registration, move the
// ip to the OldIp field and save the new registration to the hashmap.
func SetOldIP(e Registration, ip net.IP) {
	newIps := make([]net.IP, 0)
	oldIps := make([]net.IP, 0)

	for _, i := range e.Ips {
		if !i.Equal(ip) {
			newIps = append(newIps, i)
		} else {
			oldIps = append(e.OldIps, i)
		}
	}

	e.Ips = newIps
	e.OldIps = oldIps
	currentMap.v.Store(e.Mac, e)
}

func GetOldStateFromDB() []Registration {
	if currentDB == nil {
		slog.Error("Database not initialized. Could no get old entries")
		return nil
	}

	return dbGetActive(currentDB)
}

func (e *Registration) String() string {
	f := time.TimeOnly
	return fmt.Sprintf("{Id: %d, User: %s, Mac: %s, IPs: %v, OldIps: %v, Start: %s, End: %s, LastPing: %s, IsDown: %t}",
		e.Id, e.User, e.Mac, e.Ips, e.OldIps, e.Start.Format(f), e.End.Format(f), e.LastPing.Format(f), e.IsDown)
}
