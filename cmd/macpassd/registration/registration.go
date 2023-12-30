// The registration package is responsible for the memorization of all the
// reqeusts. It does not interact with iptables or other part of the program.
// It is made to abstract the memorization process.
package registration

import (
	"database/sql"
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
		Ips: []net.IP{}, LastPing: time.Now(), IsDown: false}

	ids++

	slog.With("registration", r).Debug("New registration will be added")

	currentMap.add(r)

	// Add to db
	if currentDB != nil {
		dbInsertRegistration(currentDB, r)
	}

	return
}

func AddRegistrationToMapFromDB(r Registration) {
	slog.With("registration", r).Debug("Registration from DB will be added")
	currentMap.add(r)
	return
}

func Remove(r Registration) {
	slog.With("registration", r).Debug("Removing registration")
	currentMap.remove(r.Mac)
}

func GetOldEntries() (oldEntries []Registration) {
	slog.Debug("Getting old entries")

	// Get from map
	for _, reg := range currentMap.v {
		if time.Now().Sub(reg.End) >= 0 {
			oldEntries = append(oldEntries, reg)
		}
	}

	// Get from db ?
	return
}

func AddIpToMac(ip net.IP, mac net.HardwareAddr) {
	slog.With("ip", ip, "mac", mac).Debug("Binding ip to mac")
	currentMap.addIp(mac.String(), ip)

	// TODO: Add to db
}

func GetAllEntries() (entries []Registration) {
	slog.Debug("Getting all entries")

	// Get from map
	for _, reg := range currentMap.v {
		entries = append(entries, reg)
	}

	// Get from db
	return
}

func UpdateLastPing(e Registration) {
	//update on map
	currentMap.mu.Lock()
	e.LastPing = time.Now()
	currentMap.v[e.Mac] = e
	currentMap.mu.Unlock()
}

func SetHostDown(e Registration) {
	//update on map
	currentMap.mu.Lock()
	e.IsDown = true
	currentMap.v[e.Mac] = e
	currentMap.mu.Unlock()
}

func SetHostUp(e Registration) {
	//update on map
	currentMap.mu.Lock()
	e.IsDown = false
	currentMap.v[e.Mac] = e
	currentMap.mu.Unlock()
}

func RemoveIP(e Registration, ip net.IP) {
	newIps := make([]net.IP, 0)

	for _, i := range e.Ips {
		if !i.Equal(ip) {
			newIps = append(newIps, i)
		}
	}

	currentMap.mu.Lock()
	e.Ips = newIps
	currentMap.v[e.Mac] = e
	currentMap.mu.Unlock()
}

func GetOldStateFromDB() []Registration {
	if currentDB == nil {
		slog.Error("Database not initialized. Could no get old entries")
		return nil
	}

	return dbGetActive(currentDB)
}
