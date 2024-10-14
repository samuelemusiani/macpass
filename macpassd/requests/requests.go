package requests

import (
	"internal/comunication"
	"log/slog"

	"github.com/musianisamuele/macpass/macpassd/fw"
	"github.com/musianisamuele/macpass/macpassd/registration"
)

func Handle(newEntry comunication.Request, fw fw.Firewall) {
	// If it's new we add it on the registration and allow it on the firewall
	// If it's not new we do nothing

	e := registration.GetAllEntries()
	for i := range e {
		if e[i].Mac == newEntry.Mac {
			// Registration already present. Do nothing
			return
		}
	}

	// It's always a new request

	slog.With("requst", newEntry).Info("Handlig new request")

	reg := registration.AddRequest(newEntry)
	fw.Allow(reg)
}
