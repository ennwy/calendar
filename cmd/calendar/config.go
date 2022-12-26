package main

import (
	"github.com/ennwy/calendar/internal/logger"
	"github.com/ennwy/calendar/internal/server"
	"os"
)

type Config struct {
	Logger  logger.Config     `yaml:"logger"`
	HTTP    server.HTTPConfig `yaml:"http"`
	GRPC    server.GRPCConfig `yaml:"grpc"`
	Storage string            `yaml:"storage"`
	Server  string            `yaml:"server"`
}

func NewConfig() *Config {
	config := &Config{}
	config.Server = os.Getenv("SERVER_TYPE")
	config.Logger.Set()
	config.HTTP.Set()
	config.GRPC.Set()

	return config
}
