package config

import (
	"os"

	"github.com/ghodss/yaml"
)

type KerberosConfig struct {
	Realm           string `json:"realm"`
	DisablePAFXFAST bool   `json:"disablePAFXFAST"`
}

type Socket struct {
	Path string `json:"path"`
}

type HttpServer struct {
	Url  string `json:"url"`
	Port uint16 `json:"port"`
}

type Server struct {
	Type   string     `json:"type"`
	Socket Socket     `json:"socket"`
	Http   HttpServer `json:"http"`
}

type Config struct {
	Kerberos          KerberosConfig `json:"kerberos"`
	DummyLogin        bool           `json:"dummyLogin"`
	Server            Server         `json:"server"`
	MaxConnectionTime int            `json:"maxConnectionTime"`
	DBPath            string         `json:"databasePath"`
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
