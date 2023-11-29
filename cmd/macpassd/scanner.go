package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/j-keck/arping"
	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

func scanNetwork() {
	conf := config.Get()
	network := net.IPNet{IP: net.ParseIP(conf.Network.Ip), Mask: net.IPMask(net.ParseIP(conf.Network.Mask))}
	slog.With("ip", network.IP, "mask", network.Mask).Debug("Scanning network")

	// get arptable
	path := "/proc/net/arp"
	slog.With("path", path).Debug("Reading arp cache file")
	arpTable, err := os.ReadFile(path)
	if err != nil {
		slog.With("path", path, "err", err).Error("Can't read arp table. No logs can be provided for network devices")
	}

	emptyMac, _ := net.ParseMAC("00:00:00:00:00:00")
	line := []byte{}

	hwPos := -1
	isFirstLine := true
	for _, data := range arpTable {
		if !bytes.Equal([]byte{data}, []byte("\n")) {
			line = append(line, data)
		} else {
			slog.With("line", string(line)).Debug("Get arp file line")

			if isFirstLine {
				hwPos = strings.Index(string(line), "HW address")
				if hwPos == -1 {
					slog.With("line", string(line), "substring", "HW address").
						Error("Substring cannot be found. Arp tables is not right")
					break
				}
				slog.With("hwPos", hwPos).Debug("Found 'HW address' start")
				isFirstLine = false
			} else {

				ip, mac, err := parseArpLine(line, hwPos)
				slog.With("ip", ip, "mac", mac, "err", err).Debug("Line parsed")

				if err != nil {
					slog.With("line", string(line), "err", err).
						Error("Error parsing arp line")
				} else if !reflect.DeepEqual(mac.String(), emptyMac.String()) &&
					isInSubnet(ip, network) {
					slog.With("ip", ip, "mac", mac).Debug("Mac is not empty. Found ip in the subnet. Binding to mac")
					registration.AddIpToMac(ip, mac)
				}
			}
			line = line[:0]
		}
	}
}

// line is the all line of the arp table
// hwPos the the number of the first byte that contains the mac address
func parseArpLine(line []byte, hwPos int) (net.IP, net.HardwareAddr, error) {
	l := len(line)
	if hwPos >= l {
		return nil, nil, fmt.Errorf("hwPos is greater than the line length")
	}
	if hwPos <= 0 {
		return nil, nil, fmt.Errorf("hwPos is too small")
	}

	// forgive me for this
	ip := net.ParseIP(strings.TrimSpace(string(line[0:15])))
	mac, err := net.ParseMAC(string(line[hwPos : hwPos+17]))

	return ip, mac, err
}

func isInSubnet(ip net.IP, network net.IPNet) bool {
	return ip.Mask(network.Mask).Equal(network.IP.Mask(network.Mask))
}

func isStillConnected(e registration.Registration) bool {
	arping.SetTimeout(1 * time.Second) // should be put in config

	for _, ip := range e.Ips {
		mac, _, err := arping.Ping(ip)
		if err != nil {
			slog.With("ip", ip, "err", err).Error("error during arping")
		} else {

			if e.Mac != mac.String() {
				// in this case another host has reponded to the arping. It is possible
				// that multiples hosts have the same ip or that the previous host has
				// changed ip and another host has now his old ip

				// we need to delete old ips from the entries in order to perform this
				// check correctly
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
