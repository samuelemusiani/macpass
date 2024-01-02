package config

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseConfigWrongPath(t *testing.T) {
	err := ParseConfig("/scjwei")
	assert.Assert(t, err != nil)
}

func TestParseConfigLdap(t *testing.T) {
	err := ParseConfig("./config.yaml")
	assert.NilError(t, err)
	conf := Get()

	assert.Equal(t, conf.Login.LdapDomains[0].Id, "name3")
	assert.Equal(t, conf.Login.LdapDomains[0].Address, "ldaps://ldap.example.com:636")
	assert.Equal(t, conf.Login.LdapDomains[0].BindDN, "cn=admin,dc=example,dc=com")
	assert.Equal(t, conf.Login.LdapDomains[0].BindPW, "password")
	assert.Equal(t, conf.Login.LdapDomains[0].StartTLS, true)
	assert.Equal(t, conf.Login.LdapDomains[0].InsecureNoSSL, false)
	assert.Equal(t, conf.Login.LdapDomains[0].InsecureSkipVerify, false)
	assert.Equal(t, conf.Login.LdapDomains[0].UserDNType, "uid")
	assert.Equal(t, conf.Login.LdapDomains[0].BaseDN, "dc=example,dc=com")

	assert.Equal(t, conf.Login.LdapDomains[1].Id, "name4")
	assert.Equal(t, conf.Login.LdapDomains[1].Address, "ldap://ldap.google.com:389")
	assert.Equal(t, conf.Login.LdapDomains[1].BindDN, "cn=admin,dc=google,dc=com")
	assert.Equal(t, conf.Login.LdapDomains[1].BindPW, "1234")
	assert.Equal(t, conf.Login.LdapDomains[1].StartTLS, false)
	assert.Equal(t, conf.Login.LdapDomains[1].InsecureNoSSL, true)
	assert.Equal(t, conf.Login.LdapDomains[1].InsecureSkipVerify, true)
	assert.Equal(t, conf.Login.LdapDomains[1].UserDNType, "uid")
	assert.Equal(t, conf.Login.LdapDomains[1].BaseDN, "ou=users,dc=google,dc=com")
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
