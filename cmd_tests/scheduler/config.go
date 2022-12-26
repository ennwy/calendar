package main

import (
	"fmt"
	"github.com/ennwy/calendar/internal/logger"
	noti "github.com/ennwy/calendar/internal/notification"
)

type Config struct {
	Logger    logger.Config  `yaml:"logger"`
	MQProduce noti.MQProduce `yaml:"MQProduce"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	config.Logger.Set()

	if err := config.MQProduce.Set(); err != nil {
		return nil, fmt.Errorf("scheduler: new configs: %w", err)
	}

	return config, nil
}
