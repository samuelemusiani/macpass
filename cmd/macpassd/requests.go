package main

import (
	"internal/comunication"
	"log/slog"

	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

func handleRequest(newEntry comunication.Request) {
	// Need too check if it's a new request
	// If it's new we add it on the registration and allow it on the firewall
	// If it's not new we semply update the old entry end time

	// For now we assume it's always a new request

	slog.With("requst", newEntry).Info("Handlig new request")

	reg := registration.AddRequest(newEntry)
	allowNewEntryOnFirewall(reg)
}
