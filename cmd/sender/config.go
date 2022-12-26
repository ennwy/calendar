package main

import (
	"fmt"
	"github.com/ennwy/calendar/internal/logger"
	"github.com/ennwy/calendar/internal/notification"
)

type Config struct {
	Logger logger.Config          `yaml:"logger"`
	MQ     notification.MQConsume `yaml:"MQConsume"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	config.Logger.Set()

	if err := config.MQ.Set(); err != nil {
		return nil, fmt.Errorf("sender: new configs: %w", err)
	}

	return config, nil
}
