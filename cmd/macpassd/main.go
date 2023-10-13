package main

import (
	"bufio"
	// "fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/go-iptables/iptables"
)

var inputFile string

func main() {
	parseConfig()
	startDaemon()
}

func parseConfig() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	inputFile = filepath.Dir(ex) + "/mac_out"
}

type macPerson struct {
	mac   string
	user  string
	start time.Time
}

func startDaemon() {
	currentEntries := make([]macPerson, 0)

	ip4t, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		log.Fatal(err)
	}
	denyAllMACs(ip4t)

	for true {
		fileEntries := scanFile()
		newEntries := findNewEntries(currentEntries, fileEntries)
		allowNewEntries(newEntries, ip4t)
		// checkIfStilConnected
		// deleteOldEntries

		time.Sleep(1 * time.Second)
	}
}

func scanFile() (fileEntries []macPerson) {
	// read file for new entries
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		mac, user, _ := strings.Cut(line, " ")
		fileEntries = append(fileEntries, macPerson{mac, user, time.Now()})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file.Close()
	return
}

func findNewEntries(currenEntries []macPerson, fileEntries []macPerson) (newEntries []macPerson) {
	for _, value := range fileEntries {
		if !searchMac(currenEntries, value.mac) {
			newEntries = append(newEntries, value)
		}
	}
	return
}

func insertEntry(entries []macPerson, mac string, user string) {
	if searchMac(entries, mac) {
		return
	}
	entries = append(entries, macPerson{mac, user, time.Now()})
}

func searchMac(entries []macPerson, mac string) bool {
	for _, value := range entries {
		if value.mac == mac {
			return true
		}
	}
	return false
}

func denyAllMACs(t *iptables.IPTables) {
	t.Append("filter", "FORWARD", []string{"-i", "eth1", "-o", "eth0",
		"-j", "DROP"}...)
}

func allowNewEntries(entries []macPerson, t *iptables.IPTables) {
	for _, value := range entries {
		t.InsertUnique("filter", "FORWARD", 1, []string{"-i", "eth1", "-o", "eth0",
			"-m", "mac", "--mac-source", value.mac, "-j", "ACCEPT"}...)
	}
}
