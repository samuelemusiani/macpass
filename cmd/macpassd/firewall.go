package main

import (
	"log"

	"github.com/coreos/go-iptables/iptables"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

var ipTable *iptables.IPTables

func initIptables() {
	log.Println("Initializing iptables")

	var err error
	ipTable, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		log.Fatal(err)
	}

	// We need to clear the iptable table in order to avoid previus entries
	err = ipTable.ClearAll()
	if err != nil {
		log.Fatal(err)
	}

	// The default rule of the firewall is deny all connections
	// Insert is used in case the iptables is not flush and there are still
	// entries that could compromise the security of the program
	err = ipTable.Insert("filter", "FORWARD", 1, []string{"-i", "eth1", "-o", "eth0",
		"-j", "DROP"}...)
	if err != nil {
		log.Fatal(err)
	}
}

func allowNewEntryOnFirewall(r registration.Registration) {
	err := ipTable.InsertUnique("filter", "FORWARD", 1, []string{"-i", "eth1", "-o", "eth0",
		"-m", "mac", "--mac-source", r.Mac, "-j", "ACCEPT"}...)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("ADDED: ", r)
		entriesLogger.Println("ADDED: ", r)
	}
}

func deleteEntryFromFirewall(r registration.Registration) {
	err := ipTable.Delete("filter", "FORWARD", []string{"-i", "eth1", "-o",
		"eth0", "-m", "mac", "--mac-source", r.Mac, "-j", "ACCEPT"}...)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("REMOVED: ", r)
		entriesLogger.Println("REMOVED: ", r)
	}
}
