package main

import (
	c "github.com/ennwy/calendar/cmd"
	"github.com/ghodss/yaml"
	"os"
)

type Config struct {
	Logger  c.LoggerConfig `yaml:"logger"`
	DB      c.DBConfig     `yaml:"db"`
	HTTP    HTTPConfig     `yaml:"http"`
	GRPC    GRPCConfig     `yaml:"grpc"`
	Storage string         `yaml:"storage"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type GRPCConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func NewConfig(path string) (*Config, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(configData, config)

	if err != nil {
		return nil, err
	}

	err = config.DB.SetEnv()

	return config, err
}
