package main

import (
	"log/slog"
	"net"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
	"github.com/musianisamuele/macpass/pkg/macscan"
)

const timeout = time.Second
const threads = 50

func scanNetwork() {
	conf := config.Get()
	ip := net.ParseIP(conf.Network.Ip)
	mask := net.IPMask(net.ParseIP(conf.Network.Mask).To4())

	slog.With("ip", ip, "mask", mask).Debug("scanNetwork")

	r := macscan.ScanSubnet(net.IPNet{IP: ip, Mask: mask}, timeout, 50)

	slog.Debug("Adding ScanSubnet responses")
	for _, i := range r {
		registration.AddIpToMac(i.Ip, i.Mac)
	}
}
