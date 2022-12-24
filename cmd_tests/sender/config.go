package main

import (
	"fmt"
	c "github.com/ennwy/calendar/cmd"
	noti "github.com/ennwy/calendar/internal/notification"
)

type Config struct {
	Logger    c.LoggerConfig `yaml:"logger"`
	MQConsume noti.MQConsume `yaml:"MQConsume"`
	MQProduce noti.MQProduce `yaml:"MQProduce"`
}

func NewConfig() (config *Config, err error) {
	config = &Config{}
	config.Logger.Set()

	if err = config.MQConsume.Set(); err != nil {
		return nil, fmt.Errorf("sender: new config: %w", err)
	} else if err = config.MQProduce.Set(); err != nil {
		return nil, fmt.Errorf("scheduler: new config: %w", err)
	}

	return config, nil
}
