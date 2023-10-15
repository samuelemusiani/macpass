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
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-iptables/iptables"
)

var socketPath string

func main() {
	parseConfig()
	startDaemon()
}

func parseConfig() {
	socketPath = "/tmp/macpass.sock" // very ugly
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

type safeMap struct {
	mu sync.Mutex
	v  map[string]registration
}

func startDaemon() {
	// Hashmap were al the entries that are currently in use are stored
	// currentEntries := make(map[string]registration)
	currentEntries := safeMap{v: make(map[string]registration)}

	ip4t, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		log.Fatal(err)
	}
	// We need to clear the iptable table in order to avoid previus entries
	ip4t.ClearAll()

	// The default rule of the firewall is deny all connection
	denyAllMACs(ip4t)

	// Create a socket for comunication between macpass and macpassd
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
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
		os.Exit(1)
	}()

	// Accept connection and read from the socket
	go func(currentEntries *safeMap) {
		for {
			conn, err := socket.Accept()
			if err != nil {
				log.Fatal(err)
			}

			buff := make([]byte, 4096)
			n, err := conn.Read(buff)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// Do i still need this?
			if err != nil {
				if err == io.EOF {
					fmt.Println("Read EOF")
					conn.Close()
					continue
				}

				log.Fatal(err)
			}
			var newEntry comunication.Request
			if err := json.Unmarshal(buff[:n], &newEntry); err != nil {
				log.Fatal(err)
			}

			// Check if the entry is really new
			currentEntries.mu.Lock()
			if _, present := currentEntries.v[newEntry.Mac]; !present {
				allowNewEntry(newEntry, ip4t)
				addNewEntryToMap(currentEntries, newEntry)
			}
			currentEntries.mu.Unlock()
		}
	}(&currentEntries)

	for {
		// checkIfStilConnected
		deleteOldEntries(&currentEntries, ip4t)

		time.Sleep(1 * time.Second)
	}
}

func addNewEntryToMap(m *safeMap, n comunication.Request) {
	m.v[n.Mac] = registration{user: n.User, start: time.Now(), duration: n.Duration}
}

func denyAllMACs(t *iptables.IPTables) {
	t.Append("filter", "FORWARD", []string{"-i", "eth1", "-o", "eth0",
		"-j", "DROP"}...)
}

func allowNewEntry(e comunication.Request, t *iptables.IPTables) {
	err := t.InsertUnique("filter", "FORWARD", 1, []string{"-i", "eth1", "-o", "eth0",
		"-m", "mac", "--mac-source", e.Mac, "-j", "ACCEPT"}...)

	if err != nil {
		log.Fatal(err)
	}
}

func deleteOldEntries(entries *safeMap, t *iptables.IPTables) {
	entries.mu.Lock()
	for mac, value := range entries.v {

		fmt.Println("Checking: ", mac)
		fmt.Println("Time: ", time.Since(value.start))

		if time.Since(value.start) >= value.duration {
			err := t.Delete("filter", "FORWARD", []string{"-i", "eth1", "-o", "eth0",
				"-m", "mac", "--mac-source", mac, "-j", "ACCEPT"}...)

			if err != nil {
				log.Fatal(err)
			}
			delete(entries.v, mac)
		}
	}
	entries.mu.Unlock()
}
