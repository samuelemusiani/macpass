package registration

import (
	"net"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestAddRemove(t *testing.T) {
	m := newSafeMap()

	hour := time.Hour
	// hour, ms := time.Hour, time.Millisecond
	mac := "08:7d:bb:7a:cb:d0"
	r := Registration{Id: 100, User: "user0",
		Mac: mac, Ips: []net.IP{net.IPv4(1, 2, 3, 4)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false}

	m.add(r)
	assert.DeepEqual(t, m.v[mac], r)

	m.remove(mac)
	_, p := m.v[mac]
	assert.Equal(t, p, false)

}

func TestAddIP(t *testing.T) {
	m := newSafeMap()

	hour := time.Hour
	ip1 := net.IPv4(1, 1, 1, 1)
	ip2 := net.IPv4(2, 2, 2, 2)
	// hour, ms := time.Hour, time.Millisecond
	mac := "08:7d:bb:7a:cb:d0"
	r := Registration{Id: 100, User: "user0",
		Mac: mac, Ips: []net.IP{ip1},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false}

	m.add(r)

	m.addIp(mac, ip2)
	assert.DeepEqual(t, m.v[mac].Ips, []net.IP{ip1, ip2})

	m.addIp(mac, ip2)
	assert.DeepEqual(t, m.v[mac].Ips, []net.IP{ip1, ip2})
}
