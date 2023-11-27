package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"os"
	"reflect"
	"strings"

	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

func scanNetwork() {
	// get arptable
	path := "/proc/net/arp"
	arpTable, err := os.ReadFile(path)
	if err != nil {
		slog.With("path", path, "err", err).Error("Can't read arp table. No logs can be provided for network devices")
	}

	emptyMac, _ := net.ParseMAC("00:00:00:00:00:00")
	line := []byte{}

	hwPos := -1
	for _, data := range arpTable {
		if !bytes.Equal([]byte{data}, []byte("\n")) {
			line = append(line, data)
		} else {
			if hwPos == -1 {
				hwPos = strings.Index(string(line), "HW address")
				if hwPos == -1 {
					slog.With("line", string(line), "substring", "HW address").
						Error("Substring cannot be found. Arp tables is not right")
					break
				}
			} else {
				ip, mac, err := parseArpLine(line, hwPos)
				if err != nil {
					slog.With("line", string(line), "err", err).
						Error("Error parsing arp line")
				} else if !reflect.DeepEqual(mac.String(), emptyMac.String()) {
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

	var ip net.IP
	if line[11] == byte(' ') {
		ip = net.ParseIP(string(line[0:11]))
	} else {
		ip = net.ParseIP(string(line[0:12]))
	}

	mac, err := net.ParseMAC(string(line[hwPos : hwPos+17]))

	return ip, mac, err
}
