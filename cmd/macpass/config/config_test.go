package config

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseConfigWrongPath(t *testing.T) {
	err := ParseConfig("/scjwei")
	assert.Assert(t, err != nil)
}

func TestParseConfigKerberos(t *testing.T) {
	err := ParseConfig("./config.yaml")
	assert.NilError(t, err)
	conf := Get()

	assert.Equal(t, conf.Login.KerberosDomains[0].Id, "name1")
	assert.Equal(t, conf.Login.KerberosDomains[0].Realm, "ATHENA.MIT.EDU")
	assert.Equal(t, conf.Login.KerberosDomains[0].DisablePAFXFAST, true)

	assert.Equal(t, conf.Login.KerberosDomains[1].Id, "name2")
	assert.Equal(t, conf.Login.KerberosDomains[1].Realm, "ZEUS.MIT.NOEDU")
	assert.Equal(t, conf.Login.KerberosDomains[1].DisablePAFXFAST, false)
}

func TestParseConfig(t *testing.T) {
	err := ParseConfig("./config.yaml")
	assert.NilError(t, err)
	conf := Get()
	assert.Equal(t, conf.SocketPath, "/tmp/macpass.sock")
	assert.Equal(t, conf.DBPath, "./db.sqlite")
	assert.Equal(t, conf.MaxConnectionTime, 8)
}
