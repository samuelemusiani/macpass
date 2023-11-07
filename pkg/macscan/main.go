package macscan

import (
	"encoding/binary"
	// "fmt"
	"log"
	"net"
	"time"

	"github.com/j-keck/arping"
)

type IpMac struct {
	Ip  net.IP
	Mac net.HardwareAddr
}

//	func main() {
//		r := ScanSubnet(net.IPNet{IP: net.IPv4(130, 136, 3, 1), Mask: net.IPv4Mask(255, 255, 255, 0)}, time.Second*4, 200)
//		for _, i := range r {
//			fmt.Println(i)
//		}
//	}

func ScanSubnet(subnet net.IPNet, timeout time.Duration, threads uint8) []IpMac {
	hosts := make(chan net.IP, 255)
	results := make(chan IpMac, 255)
	done := make(chan bool, threads)

	// Start workers
	log.Println("Starting workers")
	for worker := uint8(0); worker < threads; worker++ {
		go arpingHosts(hosts, results, done, timeout)
	}

	// Iterate over all the subnet and send hosts to hosts channel
	log.Println("Generating ips")
	for i, ip := bBu(subnet.Mask), bBu(subnet.IP.Mask(subnet.Mask))+1; i < ^uint32(0)-1; i, ip = i+1, ip+1 {
		var tmp []byte = make([]byte, 4)
		bBp(tmp, uint32(ip))
		// tmpip := make([]byte, 4)
		// copy(tmpip, tmp) // Need to copy because it's an array
		hosts <- tmp //Segmentation fault? Where does the scope of tmp ends?
	}
	close(hosts)

	// collect responses
	var responses []IpMac
	log.Println("Collecting responses")
	for {
		select {
		case found := <-results:
			responses = append(responses, found)
		case <-done:
			threads--
			if threads == 0 {
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
	// log.Println("Worker started")
	arping.SetTimeout(timeout)

	for host := range hosts {
		// log.Println("Make arping to ", host)
		mac, _, err := arping.Ping(host)
		if err != nil {
			// log.Println("Error during arping to host: ", host)
			// log.Println(err)
		} else {
			// log.Println("Host responded with mac: ", mac)
			repondingHosts <- IpMac{Ip: host, Mac: mac}
		}
	}

	done <- true
}

// Alias for binary.BigEndian.Uint32(a)
func bBu(a []byte) uint32 {
	return binary.BigEndian.Uint32(a)
}

// Alias for binary.BigEndian.PutUint32(a, v)
func bBp(a []byte, v uint32) {
	binary.BigEndian.PutUint32(a, v)
}
