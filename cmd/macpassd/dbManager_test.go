package main

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

func checkEqualMacRegistration(t *testing.T, macReg1, macReg2 macRegistration) {
	assert.Equal(t, macReg1.mac, macReg2.mac)
	assert.Equal(t, macReg1.reg.user, macReg2.reg.user)
	assert.Equal(t, macReg1.reg.duration, macReg2.reg.duration)
	assert.Equal(t, macReg1.reg.start.Unix(), macReg2.reg.start.Unix())
}

func TestSetOutdated(t *testing.T) {
	db := connectDB(MemoryDB)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")
	insertMacRegistration(db, macRegistration{"100", registration{"user1", time.Now(), hour}})
	insertMacRegistration(db, macRegistration{"101", registration{"user2", time.Now(), hour}})
	user3 := macRegistration{"102", registration{"user3", time.Now(), ms}}
	time.Sleep(2 * ms)
	insertMacRegistration(db, user3)
	macs := getOutdated(db)
	setOutdated(db, macs)
	assert.Equal(t, len(macs), 1)
	user3out := regInstanceToMacRegistration(macs[0])
	checkEqualMacRegistration(t, user3, user3out)
}

func TestGetActive(t *testing.T) {
	db := connectDB(MemoryDB)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")
	insertMacRegistration(db, macRegistration{"100", registration{"user1", time.Now(), hour}})
	insertMacRegistration(db, macRegistration{"101", registration{"user2", time.Now(), hour}})
	insertMacRegistration(db, macRegistration{"102", registration{"user3", time.Now(), ms}})
	time.Sleep(2 * ms)
	macs := getActive(db)
	assert.Equal(t, len(macs), 2)
}
