package registration

import (
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TetsConnecting(t *testing.T) {
	test := "test.sql"
	db := connectDB(test)
	defer db.Close()
	os.Remove(test)
}

const MemoryDB = ":memory:"

func checkEqualMacRegistration(t *testing.T, macReg1, macReg2 Registration) {
	assert.Equal(t, macReg1.Mac, macReg2.Mac)
	assert.Equal(t, macReg1.User, macReg2.User)
	assert.Equal(t, macReg1.End.In(time.UTC), macReg2.End.In(time.UTC))
	assert.Equal(t, macReg1.Start.Unix(), macReg2.Start.Unix())
}

func TestSetOutdated(t *testing.T) {
	db := connectDB(MemoryDB)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")
	insertRegistration(db, Registration{100, "user1", "08:7d:bb:7a:cb:d0",
		time.Now(), time.Now().Add(hour), false})
	insertRegistration(db, Registration{101, "user2", "80:57:61:7e:d1:dd",
		time.Now(), time.Now().Add(hour), false})

	user3 := Registration{102, "user3", "fb:65:ee:13:76:af",
		time.Now(), time.Now().Add(ms), false}

	time.Sleep(2 * ms)
	insertRegistration(db, user3)

	macs := getOutdated(db)
	setOutdated(db, macs)
	assert.Equal(t, len(macs), 1)
	user3out := macs[0]
	checkEqualMacRegistration(t, user3, user3out)
}

func TestGetActive(t *testing.T) {
	db := connectDB(MemoryDB)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")

	insertRegistration(db, Registration{100, "user1", "08:7d:bb:7a:cb:d0",
		time.Now(), time.Now().Add(hour), false})
	insertRegistration(db, Registration{101, "user2", "80:57:61:7e:d1:dd",
		time.Now(), time.Now().Add(hour), false})
	insertRegistration(db, Registration{102, "user3", "fb:65:ee:13:76:af",
		time.Now(), time.Now().Add(ms), false})

	time.Sleep(2 * ms)
	macs := getActive(db)
	assert.Equal(t, len(macs), 2)
}
