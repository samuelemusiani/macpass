package macscan

import (
	"encoding/binary"
	"log/slog"
	"net"
	"time"

	"github.com/j-keck/arping"
)

type IpMac struct {
	Ip  net.IP
	Mac net.HardwareAddr
}

func ScanSubnet(subnet net.IPNet, timeout time.Duration, threads uint8) []IpMac {
	slog.With("subnet", subnet, "timeout", timeout, "threads", threads).Debug("Scanning subnet")

	if threads == 0 {
		slog.With("thread", threads).
			Error("0 Worker specified. A network scan is impossible")
		return []IpMac{}
	}

	hosts := make(chan net.IP, 255)
	results := make(chan IpMac, 255)
	done := make(chan bool, threads)

	// Start workers
	slog.Debug("Starting workers")
	for worker := uint8(0); worker < threads; worker++ {
		go arpingHosts(hosts, results, done, timeout)
	}

	// Iterate over all the subnet and send hosts to hosts channel
	slog.Debug("Generating ips")
	for i, ip := bBu(subnet.Mask), bBu(subnet.IP.Mask(subnet.Mask))+1; i < ^uint32(0)-1; i, ip = i+1, ip+1 {
		var tmp []byte = make([]byte, 4)
		bBp(tmp, uint32(ip))
		// tmpip := make([]byte, 4)
		// copy(tmpip, tmp) // Need to copy because it's an array
		slog.With("ip", net.IP(tmp).String()).Debug("Generated ip")
		hosts <- tmp //Segmentation fault? Where does the scope of tmp ends?
	}
	close(hosts)

	slog.Debug("Collecting responses from workers")
	var responses []IpMac
	for {
		select {
		case found := <-results:
			responses = append(responses, found)
		case <-done:
			threads--
			if threads == 0 {
				slog.Debug("All worker have finished")
				close(results)
				for i := range results {
					responses = append(responses, i)
				}
				return responses
			}
		}
	}
}

func arpingHosts(hosts chan net.IP, repondingHosts chan IpMac, done chan bool, timeout time.Duration) {
	slog.Debug("Worker started")
	arping.SetTimeout(timeout)

	for host := range hosts {
		mac, _, err := arping.Ping(host)
		if err != nil {
			slog.With("error", err, "host", host).Debug("Error during arping to host")
		} else {
			slog.With("host", host, "mac", mac).Debug("Host responded")
			repondingHosts <- IpMac{Ip: host, Mac: mac}
		}
	}

	done <- true
	slog.Debug("Worker have finished")
}

// Alias for binary.BigEndian.Uint32(a)
func bBu(a []byte) uint32 {
	return binary.BigEndian.Uint32(a)
}

// Alias for binary.BigEndian.PutUint32(a, v)
func bBp(a []byte, v uint32) {
	binary.BigEndian.PutUint32(a, v)
}
