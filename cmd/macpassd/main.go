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
		// read file for new entries
		file, err := os.Open(inputFile)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			mac, user, _ := strings.Cut(line, " ")
			insertEntry(currentEntries, mac, user)
			// fmt.Println("FILE:" + line)
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		file.Close()

		// fmt.Println("ARRAY:")
		// for i := 0; i < counter; i++ {
		// 	fmt.Println(currentEntries[i])
		// }

		time.Sleep(1 * time.Second)
	}
}

func insertEntry(currentEntries []macPerson, mac string, user string) {
	if searchMac(currentEntries, mac) {
		return
	}
	currentEntries = append(currentEntries, macPerson{mac, user, time.Now()})
}

func searchMac(currentEntries []macPerson, mac string) bool {
	for _, value := range currentEntries {
		if value.mac == mac {
			return true
		}
	}
	return false
}
