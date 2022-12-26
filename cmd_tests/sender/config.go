package main

import (
	"fmt"
	"github.com/ennwy/calendar/internal/logger"
	"github.com/ennwy/calendar/internal/notification"
)

type Config struct {
	Logger    logger.Config          `yaml:"logger"`
	MQConsume notification.MQConsume `yaml:"MQConsume"`
	MQProduce notification.MQProduce `yaml:"MQProduce"`
}

func NewConfig() (config *Config, err error) {
	config = &Config{}
	config.Logger.Set()

	if err = config.MQConsume.Set(); err != nil {
		return nil, fmt.Errorf("sender: new configs: %w", err)
	} else if err = config.MQProduce.Set(); err != nil {
		return nil, fmt.Errorf("scheduler: new configs: %w", err)
	}

	return config, nil
}
