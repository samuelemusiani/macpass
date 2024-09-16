package config

import (
	"os"

	"github.com/ghodss/yaml"
)

type KerberosConfig struct {
	Realm           string `json:"realm"`
	DisablePAFXFAST bool   `json:"disablePAFXFAST"`
}

type Config struct {
	Kerberos          KerberosConfig `json:"kerberos"`
	DummyLogin        bool           `json:"dummyLogin"`
	SocketPath        string         `json:"socketPath"`
	MaxConnectionTime int            `json:"maxConnectionTime"`
	DBPath            string         `json:"databasePath"`
	Debug             bool           `json:"debug"`
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
