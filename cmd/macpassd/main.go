package main

import (
	"bufio"
	"fmt"
	// "fmt"
	"log"
	"os"
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
	// ex, err := os.Executable()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// inputFile = filepath.Dir(ex) + "/mac_out"
	inputFile = "./mac_out" // very ugly
}

type registration struct {
	user  string
	start time.Time
}

type macRegistration struct {
	mac string
	reg registration
}

func startDaemon() {
	currentEntries := make(map[string]registration)

	ip4t, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		log.Fatal(err)
	}
	ip4t.ClearAll()
	denyAllMACs(ip4t)

	for true {
		fileEntries := scanFile()
		newEntries := findNewEntries(currentEntries, fileEntries)
		allowNewEntries(newEntries, ip4t)
		addNewEntriesToMap(currentEntries, newEntries)
		newEntries = nil
		// checkIfStilConnected
		deleteOldEntries(currentEntries, ip4t)
		fmt.Println("currentEntries DELETE: ", currentEntries)

		time.Sleep(1 * time.Second)
	}
}

func scanFile() (fileEntries []macRegistration) {
	// read file for new entries
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		mac, user, _ := strings.Cut(line, " ")
		fileEntries = append(fileEntries, macRegistration{mac, registration{user, time.Now()}})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file.Close()
	return
}

func findNewEntries(currenEntries map[string]registration, fileEntries []macRegistration) (newEntries []macRegistration) {
	for _, value := range fileEntries {
		if _, present := currenEntries[value.mac]; !present {
			newEntries = append(newEntries, value)
		}
	}
	return
}

func addNewEntriesToMap(m map[string]registration, n []macRegistration) {
	for _, val := range n {
		m[val.mac] = val.reg
	}
}

func denyAllMACs(t *iptables.IPTables) {
	t.Append("filter", "FORWARD", []string{"-i", "eth1", "-o", "eth0",
		"-j", "DROP"}...)
}

func allowNewEntries(entries []macRegistration, t *iptables.IPTables) {
	for _, value := range entries {
		err := t.InsertUnique("filter", "FORWARD", 1, []string{"-i", "eth1", "-o", "eth0",
			"-m", "mac", "--mac-source", value.mac, "-j", "ACCEPT"}...)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func deleteOldEntries(entries map[string]registration, t *iptables.IPTables) {
	for mac, value := range entries {

		fmt.Println("Checking: ", mac)
		fmt.Println("Time: ", time.Since(value.start))

		if time.Since(value.start) >= 5*time.Second {
			err := t.Delete("filter", "FORWARD", []string{"-i", "eth1", "-o", "eth0",
				"-m", "mac", "--mac-source", mac, "-j", "ACCEPT"}...)

			if err != nil {
				log.Fatal(err)
			}
			delete(entries, mac)
		}
	}
}
