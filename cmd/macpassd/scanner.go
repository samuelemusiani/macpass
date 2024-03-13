package main

import (
	"log/slog"
	"net"
	"time"

	"github.com/j-keck/arping"
	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
	"github.com/vishvananda/netlink"
)

func scanNetwork() {
	conf := config.Get()

	network := net.IPNet{IP: net.ParseIP(conf.Network.Ip), Mask: net.IPMask(net.ParseIP(conf.Network.Mask))}

	slog.With("ip", network.IP.String(), "mask", network.Mask.String()).
		Debug("Listening for neighbor updates")

	nUpdate := make(chan netlink.NeighUpdate)
	done := make(chan struct{})

	netlink.NeighSubscribe(nUpdate, done)

	for {
		nu := <-nUpdate
		n := nu.Neigh

		if !isInSubnet(n.IP, network) {
			continue
		}

		// If the state is Reachable or Stale we can assume that the MAC address
		// is not empy

		if n.State == netlink.NUD_REACHABLE || n.State == netlink.NUD_STALE {
			slog.With("IP", n.IP, "MAC", n.HardwareAddr).
				Debug("Received a REACHABLE or STALE update from neighbor")
			registration.AddIpToMac(n.IP, n.HardwareAddr)
		} else {
			slog.With("IP", n.IP, "MAC", n.HardwareAddr).
				Debug("Received an update from neighbor that will not be hanled")
		}
	}
}

func isInSubnet(ip net.IP, network net.IPNet) bool {
	return ip.Mask(network.Mask).Equal(network.IP.Mask(network.Mask))
}

func isStillConnected(e registration.Registration) bool {
	arping.SetTimeout(1 * time.Second) // should be put in config

	for _, ip := range e.Ips {
		mac, _, err := arping.Ping(ip)
		if err != nil {
			slog.With("ip", ip, "err", err).Debug("error during arping")
		} else {

			if e.Mac != mac.String() {
				// in this case another host has reponded to the arping. It is possible
				// that multiples hosts have the same ip or that the previous host has
				// changed ip and another host has now his old ip. We assume the first
				// one

				// we need to delete old ips from the entries in order to perform this
				// check correctly
				slog.With("registration", e, "new mac", mac.String()).Debug("Different mac responded to arping")

				// we delete the newest registration with the ip
				entries := registration.GetAllEntries()
				valid := -1
				for i, e := range entries {
					if isIpPrenset(e.Ips, ip) {
						if valid == -1 {
							valid = i
						} else if e.Start.Compare(entries[valid].Start) == -1 {
							valid = i
						}
					}
				}

				deleteEntryFromFirewall(entries[valid])
				registration.Remove(entries[valid])

				// If we set up a mail server we can send a mail explaining why the
				// connection is dropped
				return false
			}

			return true
		}
	}

	if len(e.Ips) > 0 {
		return false
	} else {
		//if we do not have ips yet it's probably that is not connected yet
		return true
	}
}

func isIpPrenset(set []net.IP, ip net.IP) bool {
	if len(set) == 0 {
		return false
	}

	// Likely because most devices have 1 ip
	if set[0].Equal(ip) {
		return true
	}

	for _, i := range set {
		if i.Equal(ip) {
			return true
		}
	}
	return false
}
