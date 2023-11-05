package config

import (
	"os"

	"github.com/ghodss/yaml"
)

type Socket struct {
	Path string `json:"path"`
	User string `json:"user"`
}

type Config struct {
	Socket     Socket `json:"socket"`
	LoggerPath string `json:"loggerPath"`
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
