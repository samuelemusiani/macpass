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
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/spf13/viper"
)

var socketPath string
var loggerPath string
var entriesLogger *log.Logger

func main() {
	parseConfig()

	f, err := os.OpenFile(loggerPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	entriesLogger = log.New(f, "", 3)

	startDaemon()
}

func parseConfig() {
	fmt.Println("Reading config file...")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := filepath.Dir(ex)
	viper.AddConfigPath(exPath) // Config in the same directory
	viper.AddConfigPath("/etc/macpassd")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found")
			log.Fatal(err)
		} else {
			log.Println("Config file was found but another error was produced")
			log.Fatal(err)
		}
	}

	socketPath = viper.GetString("socketPath")
	loggerPath = viper.GetString("loggerPath")

	fmt.Println("Config parsed successfully")
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

	// Create a socket for comunication between macpass and macpassd
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	// For now everyone can write to the socket
	if err := os.Chmod(socketPath, 0777); err != nil {
		log.Fatal(err)
	}

	// Cleanup the sockfile
	closeChannel := make(chan os.Signal, 1)
	signal.Notify(closeChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closeChannel
		os.Remove(socketPath)
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
