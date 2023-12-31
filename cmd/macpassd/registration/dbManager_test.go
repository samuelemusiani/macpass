package registration

import (
	"net"
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TetsConnecting(t *testing.T) {
	test := "grfgQngnonfrJvguFrpergAnzr.sql"
	db := dbConnect(test)
	defer db.Close()
	os.Remove(test)
}

const MemoryDB = ":memory:"

func TestSetOutdated(t *testing.T) {
	db := dbConnect(MemoryDB)
	defer db.Close()

	hour, ms := time.Hour, time.Millisecond
	dbInsertRegistration(db, Registration{Id: 100, User: "user0",
		Mac: "08:7d:bb:7a:cb:d0", Ips: []net.IP{net.IPv4(1, 2, 3, 4)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})
	dbInsertRegistration(db, Registration{Id: 101, User: "user1",
		Mac: "80:57:61:7e:d1:dd", Ips: []net.IP{net.IPv4(5, 6, 7, 8)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})

	user2 := Registration{Id: 102, User: "user2",
		Mac: "fb:65:ee:13:76:af", Ips: []net.IP{net.IPv4(122, 245, 1, 75)},
		Start: time.Now(), End: time.Now().Add(ms), IsDown: false}

	time.Sleep(2 * ms)
	dbInsertRegistration(db, user2)

	r := dbGetOutdated(db)
	dbSetOutdated(db, r)
	assert.Equal(t, len(r), 1)

	user3out := r[0]
	checkEqualMacRegistration(t, user2, user3out)
}

func TestGetActive(t *testing.T) {
	db := dbConnect(MemoryDB)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")

	dbInsertRegistration(db, Registration{Id: 100, User: "user0",
		Mac: "08:7d:bb:7a:cb:d0", Ips: []net.IP{net.IPv4(1, 2, 3, 4)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})
	dbInsertRegistration(db, Registration{Id: 101, User: "user1",
		Mac: "80:57:61:7e:d1:dd", Ips: []net.IP{net.IPv4(5, 6, 7, 8)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})

	dbInsertRegistration(db, Registration{Id: 102, User: "user2",
		Mac: "fb:65:ee:13:76:af", Ips: []net.IP{net.IPv4(122, 245, 1, 75)},
		Start: time.Now(), End: time.Now().Add(ms), IsDown: false})

	time.Sleep(2 * ms)
	macs := dbGetActive(db)
	assert.Equal(t, len(macs), 2)
}

func TestIpsParsingEmpty(t *testing.T) {
	db := dbConnect(MemoryDB)
	defer db.Close()

	dbInsertRegistration(db, Registration{Id: 100, User: "user0",
		Mac: "08:7d:bb:7a:cb:d0", Ips: make([]net.IP, 0),
		Start: time.Now(), End: time.Now().Add(time.Hour), IsDown: false})

	a := dbGetActive(db)
	assert.Equal(t, len(a), 1)
	assert.Equal(t, len(a[0].Ips), 0)
}

func TestIpsParsing(t *testing.T) {
	db := dbConnect(MemoryDB)
	defer db.Close()

	ips := []net.IP{net.IPv4(192, 168, 1, 1), net.IPv4bcast, net.IPv4(1, 1, 1, 1), net.IPv6zero, net.IPv6loopback}

	dbInsertRegistration(db, Registration{Id: 100, User: "user0",
		Mac: "08:7d:bb:7a:cb:d0", Ips: ips,
		Start: time.Now(), End: time.Now().Add(time.Hour), IsDown: false})

	a := dbGetActive(db)
	assert.Equal(t, len(a), 1)
	assert.DeepEqual(t, a[0].Ips, ips)
}

func checkEqualMacRegistration(t *testing.T, macReg1, macReg2 Registration) {
	assert.Equal(t, macReg1.Mac, macReg2.Mac)
	assert.Equal(t, macReg1.User, macReg2.User)
	assert.Equal(t, macReg1.End.Round(time.Second).Unix(), macReg2.End.Round(time.Second).Unix())
	assert.Equal(t, macReg1.Start.Round(time.Second).Unix(), macReg2.Start.Round(time.Second).Unix())
}
