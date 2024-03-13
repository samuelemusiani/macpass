package config

import (
	"os"

	"github.com/ghodss/yaml"
)

type Socket struct {
	Path string `json:"path"`
	User string `json:"user"`
}

type Network struct {
	IP4     string `json:"ip4"`
	IP6     string `json:"ip6"`
	Timeout uint64 `json:"timeout"`
	IFace   string `json:"inInterface"`
}

type Config struct {
	Socket            Socket  `json:"socket"`
	Network           Network `json:"network"`
	LoggerPath        string  `json:"loggerPath"`
	DbPath            string  `json:"dbPath"`
	LogLevel          string  `json:"logLevel"`
	IterationTime     int     `json:"iterationTime"`
	DisconnectionTime int     `json:"disconnectionTime"`
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
