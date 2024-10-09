package fw

import (
	"log/slog"
	"os"

	"github.com/coreos/go-iptables/iptables"
	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

type Iptables struct {
	ip4Table *iptables.IPTables
	ip6Table *iptables.IPTables

	conf *config.Config
}

func (i Iptables) Init() {
	slog.Info("Initializing iptables")

	i.conf = config.Get()

	var err error
	i.ip4Table, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		slog.With("error", err).Error("Failing creating iptables object for IPv4")
		os.Exit(3)
	}

	i.ip6Table, err = iptables.NewWithProtocol(iptables.ProtocolIPv6)
	if err != nil {
		slog.With("error", err).Error("Failing creating iptables object for IPv6")
		os.Exit(3)
	}

	// We need to clear the iptable table in order to avoid previus entries
	err = i.ip4Table.ClearChain("filter", "FORWARD")
	if err != nil {
		slog.With("error", err).Error("Failing clearing iptables rules for IPv4")
		os.Exit(3)
	}

	err = i.ip6Table.ClearChain("filter", "FORWARD")
	if err != nil {
		slog.With("error", err).Error("Failing clearing iptables rules for IPv6")
		os.Exit(3)
	}

	// The default rule of the firewall is deny all connections
	// Insert is used in case the iptables is not flush and there are still
	// entries that could compromise the security of the program
	err = i.ip4Table.Insert("filter", "FORWARD", 1, []string{"-i",
		i.conf.Network.IFace, "-j", "DROP"}...)

	if err != nil {
		slog.With("error", err).Error("Inserting default deny rule on iptable for IPv4")
		os.Exit(3)
	}

	err = i.ip6Table.Insert("filter", "FORWARD", 1, []string{"-i",
		i.conf.Network.IFace, "-j", "DROP"}...)

	if err != nil {
		slog.With("error", err).Error("Inserting default deny rule on iptable for IPv6")
		os.Exit(3)
	}
}

func (i Iptables) Allow(r registration.Registration) {
	err0 := i.ip4Table.InsertUnique("filter", "FORWARD", 1, []string{"-i",
		i.conf.Network.IFace, "-m", "mac", "--mac-source",
		r.Mac, "-j", "ACCEPT"}...)

	err1 := i.ip6Table.InsertUnique("filter", "FORWARD", 1, []string{"-i",
		i.conf.Network.IFace, "-m", "mac", "--mac-source",
		r.Mac, "-j", "ACCEPT"}...)

	if err0 != nil || err1 != nil {
		slog.With("registration", r.String(), "error ipv4", err0, "error IPv6", err1).
			Error("Inserting registration")
	} else {
		slog.With("registration", r.String()).Info("ADDED")
		// entriesLogger.Println("ADDED: ", r)
	}
}

func (i Iptables) Delete(r registration.Registration) {
	err0 := i.ip4Table.Delete("filter", "FORWARD", []string{"-i", i.conf.Network.IFace,
		"-m", "mac", "--mac-source", r.Mac, "-j", "ACCEPT"}...)

	err1 := i.ip6Table.Delete("filter", "FORWARD", []string{"-i", i.conf.Network.IFace,
		"-m", "mac", "--mac-source", r.Mac, "-j", "ACCEPT"}...)

	if err0 != nil || err1 != nil {
		slog.With("registration", r.String(), "error ipv4", err0, "error IPv6", err1).
			Error("Removing registration")

	} else {
		slog.With("registration", r.String()).Info("REMOVED")
		// entriesLogger.Println("REMOVED: ", r)
	}
}
