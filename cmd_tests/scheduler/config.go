package main

import (
	"fmt"
	c "github.com/ennwy/calendar/cmd"
	noti "github.com/ennwy/calendar/internal/notification"
)

type Config struct {
	Logger c.LoggerConfig `yaml:"logger"`
	MQ     noti.MQProduce `yaml:"mq"`
}

func NewConfig() (*Config, error) {
	//configData, err := os.ReadFile(path)
	//if err != nil {
	//	return nil, err
	//}
	//
	//config := &Config{}
	//err = yaml.Unmarshal(configData, config)
	//
	//if err != nil {
	//	return nil, err
	//}

	config := &Config{}
	config.Logger.Set()
	if err := config.MQ.Set(); err != nil {
		return nil, fmt.Errorf("scheduler: new config: %w", err)
	}

	return config, nil
}
