package main

import (
	"bytes"
	"log"
	"log/slog"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/j-keck/arping"
	"github.com/musianisamuele/macpass/macpassd/config"
	"github.com/musianisamuele/macpass/macpassd/registration"
	"github.com/vishvananda/netlink"
)

func scanNeighbours() {
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

	link, err := netlink.LinkByName(conf.Network.IFace)
	if err != nil {
		slog.With("link", conf.Network.IFace, "err", err).
			Error("Could not get link by name")
		log.Fatal("Could not continue")
	}

	for {
		nl, err := netlink.NeighList(link.Attrs().Index, netlink.FAMILY_ALL)
		if err != nil {
			slog.With("link", conf.Network.IFace, "err", err).
				Error("Could not get neighbor list")
		}

		for _, n := range nl {

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
		time.Sleep(time.Duration(conf.IterationTime) * time.Second)
	}
}

func isInSubnet(ip net.IP, network *net.IPNet) bool {
	return ip.Mask(network.Mask).Equal(network.IP.Mask(network.Mask))
}

func isRegistrationStillConnected(e registration.Registration) bool {
	arping.SetTimeout(1 * time.Second) // should be put in config
	slog.With("Registration", e.String()).Debug("Checking registration")

	if len(e.Ips) == 0 {
		return true
	}

	connected := false

	for _, ip := range e.Ips {
		m, connected := isIPStillConnectedWithMac(ip)
		if connected {
			slog.With("ip", ip, "mac", m).Debug("IP is still connected")
			if m != nil && m.String() != e.Mac {
				// The mac address changed
				slog.With("oldMac", e.Mac, "newMac", m.String(), "ip", ip).
					Warn("The mac address associated with the ip changed")
			}
			connected = true
		} else {
			slog.With("ip", ip).Debug("IP is not connected")
		}
	}
	if !connected {
		slog.With("registration", e.String()).Debug("The registration is not connected")
	}
	return connected
}

func isIPStillConnected(ip net.IP) bool {
	_, connected := isIPStillConnectedWithMac(ip)
	return connected
}

func isIPStillConnectedWithMac(ip net.IP) (net.HardwareAddr, bool) {
	slog.With("ip", ip).Debug("Checking ip")

	if ip.To4() != nil { // Is an IPv4
		m, _, err := arping.Ping(ip)
		if err != nil {
			slog.With("ip", ip, "err", err).Debug("error during arping")
			return nil, false
		}

		slog.With("ip", ip).Debug("IP responded to arping")
		return m, true
	}

	// Is an IPv6
	// To check IPv6 we try to ping the host

	/* From man page on ping:
	 * If ping does not receive any reply packets at all it will exit with code 1.
	 * If a packet count and deadline are both specified, and fewer than count
	 * packets are received by the time the deadline has arrived, it will also
	 * exit with code 1. On other error it exits with code 2. Otherwise it exits
	 * with code 0. This makes it possible to use the exit code to see if a host
	 * is alive or not.
	 */

	cmd := exec.Command("ping", "-c", "1", "-w", "1", ip.String())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		slog.With("ip", ip).Debug("Host reachable")
		return nil, true
	} else if strings.Contains(err.Error(), "1") {
		slog.With("ip", ip).Debug("Host unreachable")
		return nil, false
	} else {
		slog.With("ip", ip, "err", err.Error()+": "+stderr.String()).Error("During ping")
		return nil, false
	}
}
