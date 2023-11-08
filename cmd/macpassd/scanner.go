package main

import (
	"log/slog"
	"net"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
	"github.com/musianisamuele/macpass/pkg/macscan"
)

func scanNetwork() {
	conf := config.Get()
	ip := net.ParseIP(conf.Network.Ip)
	mask := net.IPMask(net.ParseIP(conf.Network.Mask).To4())
	timeout := time.Millisecond * time.Duration(time.Duration(conf.Network.Timeout).Milliseconds())
	workers := conf.Network.Workers

	slog.With("ip", ip, "mask", mask).Debug("scanNetwork")

	r := macscan.ScanSubnet(net.IPNet{IP: ip, Mask: mask}, timeout, workers)

	slog.Debug("Adding ScanSubnet responses")
	for _, i := range r {
		registration.AddIpToMac(i.Ip, i.Mac)
	}
}
