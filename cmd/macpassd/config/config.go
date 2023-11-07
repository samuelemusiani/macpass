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
	Ip   string `json:"ip"`
	Mask string `json:"subnetmask"`
}

type Config struct {
	Socket     Socket  `json:"socket"`
	Network    Network `json:"network"`
	LoggerPath string  `json:"loggerPath"`
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
