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
	db := connectDB(test)
	defer db.Close()
	os.Remove(test)
}

const MemoryDB = ":memory:"

func TestSetOutdated(t *testing.T) {
	db := connectDB(MemoryDB)
	defer db.Close()

	hour, ms := time.Hour, time.Millisecond
	insertRegistration(db, Registration{Id: 100, User: "user0",
		Mac: "08:7d:bb:7a:cb:d0", Ips: []net.IP{net.IPv4(1, 2, 3, 4)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})
	insertRegistration(db, Registration{Id: 101, User: "user1",
		Mac: "80:57:61:7e:d1:dd", Ips: []net.IP{net.IPv4(5, 6, 7, 8)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})

	user2 := Registration{Id: 102, User: "user2",
		Mac: "fb:65:ee:13:76:af", Ips: []net.IP{net.IPv4(122, 245, 1, 75)},
		Start: time.Now(), End: time.Now().Add(ms), IsDown: false}

	time.Sleep(2 * ms)
	insertRegistration(db, user2)

	r := getOutdated(db)
	setOutdated(db, r)
	assert.Equal(t, len(r), 1)

	user3out := r[0]
	checkEqualMacRegistration(t, user2, user3out)
}

func TestGetActive(t *testing.T) {
	db := connectDB(MemoryDB)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")

	insertRegistration(db, Registration{Id: 100, User: "user0",
		Mac: "08:7d:bb:7a:cb:d0", Ips: []net.IP{net.IPv4(1, 2, 3, 4)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})
	insertRegistration(db, Registration{Id: 101, User: "user1",
		Mac: "80:57:61:7e:d1:dd", Ips: []net.IP{net.IPv4(5, 6, 7, 8)},
		Start: time.Now(), End: time.Now().Add(hour), IsDown: false})

	insertRegistration(db, Registration{Id: 102, User: "user2",
		Mac: "fb:65:ee:13:76:af", Ips: []net.IP{net.IPv4(122, 245, 1, 75)},
		Start: time.Now(), End: time.Now().Add(ms), IsDown: false})

	time.Sleep(2 * ms)
	macs := getActive(db)
	assert.Equal(t, len(macs), 2)
}

func checkEqualMacRegistration(t *testing.T, macReg1, macReg2 Registration) {
	assert.Equal(t, macReg1.Mac, macReg2.Mac)
	assert.Equal(t, macReg1.User, macReg2.User)
	assert.Equal(t, macReg1.End.Round(time.Second).Unix(), macReg2.End.Round(time.Second).Unix())
	assert.Equal(t, macReg1.Start.Round(time.Second).Unix(), macReg2.Start.Round(time.Second).Unix())
}
