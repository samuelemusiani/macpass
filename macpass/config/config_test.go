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
	assert.Equal(t, conf.Kerberos.Realm, "ATHENA.MIT.EDU")
	assert.Equal(t, conf.Kerberos.DisablePAFXFAST, true)
}

func TestParseConfig(t *testing.T) {
	err := ParseConfig("./config.yaml")
	assert.NilError(t, err)
	conf := Get()
	assert.Equal(t, conf.SocketPath, "/tmp/macpass.sock")
	assert.Equal(t, conf.MaxConnectionTime, 8)
}
