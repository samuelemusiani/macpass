package main

import (
	"os"
	"testing"
	"time"
)

func TetsConnecting(t *testing.T) {
	test := "test.sql"
	db := connectDB(test)
	defer db.Close()
	os.Remove(test)
}

func _checkEqualMacRegistration(t *testing.T, macReg1, macReg2 macRegistration) {
	if macReg1.mac != macReg2.mac || macReg1.reg.user != macReg2.reg.user || macReg1.reg.duration != macReg2.reg.duration {
		t.Errorf("Expected %v, got %v,", macReg1, macReg2)
	}
	// the object is slitly different, for the location of the time.Time object
	if macReg1.reg.start.Unix() != macReg2.reg.start.Unix() {
		t.Errorf("Expected %v, got %v,", macReg1.reg.start.Unix(), macReg2.reg.start.Unix())
	}

}
func TestSetOutdated(t *testing.T) {
	test := "test_outdated.sql"
	db := connectDB(test)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")
	insertMacRegistration(db, macRegistration{"100", registration{"user1", time.Now(), hour}})
	insertMacRegistration(db, macRegistration{"101", registration{"user2", time.Now(), hour}})
	user3 := macRegistration{"102", registration{"user3", time.Now(), ms}}
	insertMacRegistration(db, user3)
	macs := setOutdated(db)
	if len(macs) != 1 {
		t.Errorf("Expected 1, got %d", len(macs))
	}
	user3out := regInstanceToMacRegistration(macs[0])
	_checkEqualMacRegistration(t, user3, user3out)
	os.Remove(test)
}

func TestGetActive(t *testing.T) {
	test := "test_active.sql"
	db := connectDB(test)
	defer db.Close()
	hour, _ := time.ParseDuration("1h")
	ms, _ := time.ParseDuration("1ms")
	insertMacRegistration(db, macRegistration{"100", registration{"user1", time.Now(), hour}})
	insertMacRegistration(db, macRegistration{"101", registration{"user2", time.Now(), hour}})
	insertMacRegistration(db, macRegistration{"102", registration{"user3", time.Now(), ms}})
	macs := getActive(db)
	if len(macs) != 2 {
		t.Errorf("Expected 2, got %d", len(macs))
	}
	os.Remove(test)
}
