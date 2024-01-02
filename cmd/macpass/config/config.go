package config

import (
	"os"

	"github.com/ghodss/yaml"
)

type KerberosDomain struct {
	Id              string `json:"id"`
	Realm           string `json:"realm"`
	DisablePAFXFAST bool   `json:"disablePAFXFAST"`
}

type LdapDomain struct {
	Id                 string `json:"id"`
	Address            string `json:"address"`
	BindDN             string `json:"bindDN"`
	BindPW             string `json:"bindPW"`
	StartTLS           bool   `json:"startTLS"`
	InsecureNoSSL      bool   `json:"insecureNoSSL"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	UserDNType         string `json:"userDNType"`
	BaseDN             string `json:"baseDN"`
	RemoveMail         bool   `json:"removeMail"`
}

type Login struct {
	KerberosDomains []KerberosDomain `json:"kerberos"`
	LdapDomains     []LdapDomain     `json:"ldap"`
}

type Config struct {
	Login             Login  `json:"login"`
	SocketPath        string `json:"socketPath"`
	MaxConnectionTime int    `json:"maxConnectionTime"`
	DBPath            string `json:"databasePath"`
	Logs              bool   `json:"log"`
}

var config Config

func ParseConfig(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return err
	}
	return nil
}

func Get() *Config {
	return &config
}
