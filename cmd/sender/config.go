package main

import (
	"fmt"
	c "github.com/ennwy/calendar/cmd"
	noti "github.com/ennwy/calendar/internal/notification"
)

type Config struct {
	Logger c.LoggerConfig `yaml:"logger"`
	MQ     noti.MQConsume `yaml:"mq"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	config.Logger.Set()

	if err := config.MQ.Set(); err != nil {
		return nil, fmt.Errorf("sender: new config: %w", err)
	}

	return config, nil
}
