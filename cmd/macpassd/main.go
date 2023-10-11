package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	for true {
		fileEntries := scanFile()
		newEntries := findNewEntries(currentEntries, fileEntries)
		allowNewEntries(newEntries)
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

func allowNewEntries(entries []macPerson) {
}
