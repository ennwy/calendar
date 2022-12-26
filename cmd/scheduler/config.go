package main

import (
	"fmt"
	"github.com/ennwy/calendar/internal/logger"
	noti "github.com/ennwy/calendar/internal/notification"
)

type Config struct {
	Logger logger.Config  `yaml:"logger"`
	MQ     noti.MQProduce `yaml:"mq"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	config.Logger.Set()
	if err := config.MQ.Set(); err != nil {
		return nil, fmt.Errorf("scheduler: new configs: %w", err)
	}

	return config, nil
}
