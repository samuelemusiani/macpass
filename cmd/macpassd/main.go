package main

import (
	"encoding/json"
	"fmt"
	"internal/comunication"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/musianisamuele/macpass/cmd/macpassd/config"
)

var entriesLogger *log.Logger

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("Please provide a config path")
	} else if len(os.Args) > 2 {
		log.Fatal("Too many arguments provided")
	}

	config.ParseConfig(os.Args[1]) //tmp

	f, err := os.OpenFile(config.Get().LoggetPath,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	entriesLogger = log.New(f, "", 3)

	startDaemon()
}

type registration struct {
	user     string
	start    time.Time
	duration time.Duration
}

type macRegistration struct {
	mac string
	reg registration
}

// A safe map is a map that have a Mutex condition in order to support
// concurrency
type safeMap struct {
	mu sync.Mutex
	v  map[string]registration
}

func startDaemon() {
	// Hashmap were al the entries that are currently in use are stored
	currentEntries := safeMap{v: make(map[string]registration)}

	ip4t := initIptables()

	socket := initComunication()

	go handleComunication(&currentEntries, ip4t, socket)

	for {
		// checkIfStilConnected() TODO
		deleteOldEntries(&currentEntries, ip4t)

		time.Sleep(10 * time.Second)
	}
}

func initIptables() *iptables.IPTables {
	fmt.Println("Initializing iptables...")

	ip4t, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// We need to clear the iptable table in order to avoid previus entries
	err = ip4t.ClearAll()
	if err != nil {
		log.Println(err)
	}

	// The default rule of the firewall is deny all connection
	// Insert is used in case the iptables is not flush and there are still
	// entries that could compromise the security of the program
	err = ip4t.Insert("filter", "FORWARD", 1, []string{"-i", "eth1", "-o", "eth0",
		"-j", "DROP"}...)
	if err != nil {
		log.Println(err)
		os.Exit(3)
	}

	return ip4t
}

func initComunication() net.Listener {
	fmt.Println("Initializing comunications with macpass...")
	conf := config.Get()

	// Create a socket for comunication between macpass and macpassd
	socket, err := net.Listen("unix", conf.Socket.Path)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	// Get user owner group
	group, err := user.Lookup(conf.Socket.User)
	if err != nil {
		log.Fatal(err)
	}
	uid, _ := strconv.Atoi(group.Uid)
	gid, _ := strconv.Atoi(group.Gid)

	if err := os.Chown(conf.Socket.Path, uid, gid); err != nil {
		log.Fatal(err)
	}

	if err := os.Chmod(conf.LoggetPath, 0660); err != nil {
		log.Fatal(err)
	}

	// Cleanup the sockfile
	closeChannel := make(chan os.Signal, 1)
	signal.Notify(closeChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closeChannel
		os.Remove(conf.LoggetPath)
		os.Exit(0)
	}()

	return socket
}

func handleComunication(currentEntries *safeMap, ip4t *iptables.IPTables,
	socket net.Listener) {
	fmt.Println("Listening for new comunication on socket " +
		socket.Addr().String())
	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}

		buff := make([]byte, 4096)
		n, err := conn.Read(buff)

		if err != nil {
			if err == io.EOF {
				conn.Close()
				continue
			}

			log.Println(err)
		}
		var newEntry comunication.Request
		if err := json.Unmarshal(buff[:n], &newEntry); err != nil {
			log.Println(err)
		}

		// Check if the entry is really new
		if _, present := currentEntries.v[newEntry.Mac]; !present {
			allowNewEntry(newEntry, ip4t)
			addNewEntryToMap(currentEntries, newEntry)
		}
	}
}

func addNewEntryToMap(m *safeMap, n comunication.Request) {
	m.mu.Lock()
	m.v[n.Mac] = registration{user: n.User, start: time.Now(), duration: n.Duration}
	m.mu.Unlock()
}

func allowNewEntry(e comunication.Request, t *iptables.IPTables) {
	err := t.InsertUnique("filter", "FORWARD", 1, []string{"-i", "eth1", "-o", "eth0",
		"-m", "mac", "--mac-source", e.Mac, "-j", "ACCEPT"}...)

	if err != nil {
		log.Println(err)
	} else {
		entriesLogger.Println("ADDED: ", e)
	}
}

func deleteOldEntries(entries *safeMap, t *iptables.IPTables) {
	for mac, value := range entries.v {

		if time.Since(value.start) >= value.duration {
			err := t.Delete("filter", "FORWARD", []string{"-i", "eth1", "-o", "eth0",
				"-m", "mac", "--mac-source", mac, "-j", "ACCEPT"}...)

			if err != nil {
				log.Println(err)
			} else {
				// Delete entry on table only if deleted from iptables
				entries.mu.Lock()
				delete(entries.v, mac)
				entries.mu.Unlock()
				entriesLogger.Println("REMOVED: ", macRegistration{mac, value})
			}
		}
	}
}
