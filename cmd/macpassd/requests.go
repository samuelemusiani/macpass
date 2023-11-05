package main

import (
	"internal/comunication"

	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

func handleRequest(newEntry comunication.Request) {
	// New too check if it's a new request
	// If it's new we add it on the registration and allow it on the firewall
	// If it's not new we semply update the old entry end time

	// For now we assume it's always a new request

	reg := registration.Add(newEntry)
	allowNewEntryOnFirewall(reg)
}
