package main

import (
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/j-keck/arping"
	"github.com/mehrdadrad/ping"
	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
	"github.com/vishvananda/netlink"
)

func listenForNeighbourUpdates() {
	conf := config.Get()

	_, network4, err := net.ParseCIDR(conf.Network.IP4)
	if err != nil {
		slog.With("net", network4, "err", err).
			Error("Could not parse CIDR for IPv4 into a network")
		log.Fatal("Could not continue")
	}

	_, network6, err := net.ParseCIDR(conf.Network.IP6)
	if err != nil {
		slog.With("net", network4, "err", err).
			Error("Could not parse CIDR for IPv6 into a network")
		log.Fatal("Could not continue")
	}

	slog.With("IPv4", network4.String(), "IPv6", network6.String()).
		Debug("Listening for neighbor updates")

	nUpdate := make(chan netlink.NeighUpdate)
	done := make(chan struct{})

	netlink.NeighSubscribe(nUpdate, done)

	for {
		nu := <-nUpdate
		n := nu.Neigh

		if !isInSubnet(n.IP, network4) && !isInSubnet(n.IP, network6) {
			slog.With("ip", n.IP.String(), "net4", network4.String(), "net6", network6.String()).Debug("Ip not in subnet, ignoring")
			continue
		}

		// If the state is Reachable or Stale we can assume that the MAC address
		// is not empy
		if n.State == netlink.NUD_REACHABLE || n.State == netlink.NUD_STALE {
			slog.With("IP", n.IP.String(), "MAC", n.HardwareAddr.String()).
				Debug("Received a REACHABLE or STALE update from neighbor")
			registration.AddIpToMac(n.IP, n.HardwareAddr)
		} else {
			slog.With("IP", n.IP.String(), "MAC", n.HardwareAddr.String()).
				Debug("Received an update from neighbor that will not be hanled")
		}
	}
}

func isInSubnet(ip net.IP, network *net.IPNet) bool {
	return ip.Mask(network.Mask).Equal(network.IP.Mask(network.Mask))
}

func isRegistrationStillConnected(e registration.Registration) bool {
	arping.SetTimeout(1 * time.Second) // should be put in config

	slog.With("Registration", e.String()).Debug("Checking registration")
	for _, ip := range e.Ips {
		if isIPStillConnected(ip) {
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

func isIPStillConnected(ip net.IP) bool {
	slog.With("ip", ip).Debug("Checking ip")

	if ip.To4() != nil { // Is an IPv4
		_, _, err := arping.Ping(ip)
		if err != nil {
			slog.With("ip", ip, "err", err).Debug("error during arping")
			return false
		}
		return true
	}

	// Is an IPv6
	// To check IPv6 we try to ping the host
	p, err := ping.New(ip.String())
	if err != nil {
		slog.With("ip", ip.String(), "err", err).Error("Could not construct ping object")
		return false
	}
	p.SetCount(1)

	r, err := p.Run()
	if err != nil {
		slog.With("ip", ip.String(), "err", err).Error("Could not run ping to IPv6")
		return false
	}

	for pr := range r {
		if pr.Err != nil {
			return true
		}
	}
	return false
}
