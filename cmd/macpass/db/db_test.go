package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const DB_PATH = ":memory:"

func TestConnection(t *testing.T) {
	Connect(DB_PATH)
}

func TestInsertUser(t *testing.T) {
	Connect(DB_PATH)

	user := "alice"
	mac := "11:22:33:44:55:66"
	InsertUser(user, mac)
	mac_db, present := GetMac("alice")
	assert.True(t, present)
	assert.Equal(t, mac, mac_db)
}

func TestReplaceMac(t *testing.T) {
	Connect(DB_PATH)

	user := "alice"
	mac := "11:22:33:44:55:66"
	InsertUser(user, mac)

	mac = "22:33:44:55:66:11"
	InsertUser(user, mac)

	mac_db, present := GetMac("alice")
	assert.True(t, present)
	assert.Equal(t, mac, mac_db)
}

func TestNotPresent(t *testing.T) {
	Connect(DB_PATH)

	user := "bob"
	_, present := GetMac(user)
	assert.False(t, present)
}

func TestAddKeyToUser(t *testing.T) {
	Connect(DB_PATH)

	user := "bob"
	mac := "22:33:44:55:66:11"
	InsertUser(user, mac)

	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO/gWVcWT5GHwDWRVtFWGkUpQhCycrH7yrMjmfwUjYLp ali@pc"

	err := AddKeyToUser(user, key)
	assert.Nil(t, err)
}

func TestGetKeysFromUser(t *testing.T) {
	Connect(DB_PATH)

	user := "alice"
	mac := "22:33:44:55:66:11"
	InsertUser(user, mac)

	key1 := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO/gWVcWT5GHwDWRVtFWGkUpQhCycrH7yrMjmfwUjYLp ali@pc"
	key2 := "ssh-ed25519 AAAAC3Nza1NTEC1lZDI5AAAAIO/gWVwDWRVtFcWT5GHWGkycrHUpQhC7yrMjmfwUjYLp ali@pc1"

	keys := []string{key1, key2}

	err := AddKeyToUser(user, key1)
	assert.Nil(t, err)

	err = AddKeyToUser(user, key2)
	assert.Nil(t, err)

	keys_db, err := GetKeysFromUser(user)
	assert.Nil(t, err)

	assert.Equal(t, len(keys_db), 2)

	found := func(keys []string, key string) bool {
		for k := range keys {
			if keys[k] == key {
				return true
			}
		}
		return false
	}

	assert.True(t, true, found(keys, key1))
	assert.True(t, true, found(keys, key2))
}
