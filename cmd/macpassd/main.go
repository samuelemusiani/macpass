package main

import (
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

// var entriesLogger *log.Logger

func main() {

	// Config
	if len(os.Args) <= 1 {
		slog.Error("Please provide a config path")
		os.Exit(1)
	} else if len(os.Args) > 2 {
		slog.Error("Too many arguments provided")
		os.Exit(1)
	}

	config.ParseConfig(os.Args[1]) //tmp

	conf := config.Get()

	var logLevel slog.Level
	switch conf.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelWarn
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))

	// // Logger
	// f, err := os.OpenFile(config.Get().LoggerPath,
	// 	os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	// if err != nil {
	// 	slog.With("path", config.Get().LoggerPath, "error", err).
	// 		Error("Failed to open logger")
	// 	os.Exit(2)
	// }
	// entriesLogger = log.New(f, "", 3)

	startDaemon()
}

func startDaemon() {
	slog.Info("Starting macpassd daemon")

	initIptables()

	initComunication()
	registration.Init()

	// Reload state from db to avoid dropping connection on restart
	old := registration.GetOldStateFromDB()
	for i := range old {
		registration.AddRegistrationToMapFromDB(old[i])
		allowNewEntryOnFirewall(old[i])
	}

	go handleComunication()
	conf := config.Get()

	go scanNeighbours()

	for {
		deleteOldEntries()
		deleteOldIps()
		deleteDisconnected()

		time.Sleep(time.Duration(conf.IterationTime) * time.Second)
	}
}

// This functions delete all the old entries. An old entry is a registration
// that have the connection time expired.
func deleteOldEntries() {
	oldEntries := registration.GetOldEntries()

	slog.With("oldEntries", oldEntries).Debug("Removing old entries")
	for _, e := range oldEntries {
		deleteEntryFromFirewall(e)
		registration.Remove(e)
	}
}

// If a host change the ip in the registration there will be 2 ips. But
// one of them does not respond and could be take by another host. So if a host
// have multiples ips we check every ip and if at least one respond we remove
// the others.
//
// We can't simply removed them from the map because the arp cache still has
// the old ips binded to the mac address of the host if they are not taken by
// others. So if we remove the ip, the function scanNetwork() insert the ip
// again in the map causing a lot of logs. We move the ips in the old_ip field
//
// This function does not remove the entry and does not remove the ips if none
// of them respond
func deleteOldIps() {
	entries := registration.GetAllEntries()

	for _, e := range entries {
		if len(e.Ips) > 1 {
			ipsThatDidNotAnswered := make([]net.IP, 0)

			for _, ip := range e.Ips {
				if !isIPStillConnected(ip) {
					ipsThatDidNotAnswered = append(ipsThatDidNotAnswered, ip)
				}
			}

			if len(e.Ips) == len(ipsThatDidNotAnswered) {
				// it's not my job if none of the ips answered
				break
			}

			for _, ip := range ipsThatDidNotAnswered {
				registration.SetOldIP(e, ip)
			}
		}
	}
}

// If a host goes offline we wait for a period of DisconnectionTime and if it
// does not respond we delete him
func deleteDisconnected() {
	entries := registration.GetAllEntries()

	discTime := config.Get().DisconnectionTime

	for _, e := range entries {
		if !isRegistrationStillConnected(e) {
			if !e.IsDown {
				slog.Info(e.User + " disconnected")
				registration.SetHostDown(e)
			}

			if time.Now().Sub(e.LastPing) > time.Duration(discTime)*time.Minute {
				deleteEntryFromFirewall(e)
				registration.Remove(e)
			}
		} else {
			registration.UpdateLastPing(e)
			if e.IsDown {
				slog.Info(e.User + " reconnected")
				registration.SetHostUp(e)
			}
		}
	}
}
