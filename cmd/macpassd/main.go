package main

import (
	"log"
	"os"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

var entriesLogger *log.Logger

func main() {

	// Config
	if len(os.Args) <= 1 {
		log.Fatal("Please provide a config path")
	} else if len(os.Args) > 2 {
		log.Fatal("Too many arguments provided")
	}

	config.ParseConfig(os.Args[1]) //tmp

	// Logger
	f, err := os.OpenFile(config.Get().LoggerPath,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	entriesLogger = log.New(f, "", 3)

	startDaemon()
}

func startDaemon() {
	initIptables()

	initComunication()
	go handleComunication()

	for {
		// checkIfStilConnected() TODO
		deleteOldEntries()

		time.Sleep(10 * time.Second)
	}
}

func deleteOldEntries() {
	oldEntries := registration.GetOldEntries()
	for _, e := range oldEntries {
		deleteEntryFromFirewall(e)
		registration.Remove(e)
	}
}
