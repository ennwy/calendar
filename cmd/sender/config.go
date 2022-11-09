package main

import (
	c "github.com/ennwy/calendar/cmd"
	noti "github.com/ennwy/calendar/internal/notification"
	"github.com/ghodss/yaml"
	"os"
)

type Config struct {
	Logger c.LoggerConfig `yaml:"logger"`
	MQ     noti.MQConsume `yaml:"mq"`
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

	return config, err
}
