package main

import (
	c "github.com/ennwy/calendar/cmd"
	"os"
)

type Config struct {
	Logger  c.LoggerConfig `yaml:"logger"`
	HTTP    HTTPConfig     `yaml:"http"`
	GRPC    GRPCConfig     `yaml:"grpc"`
	Storage string         `yaml:"storage"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (h *HTTPConfig) Set() {
	h.Host = os.Getenv("HTTP_HOST")
	h.Port = os.Getenv("HTTP_PORT")
}

type GRPCConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (g *GRPCConfig) Set() {
	g.Host = os.Getenv("GRPC_HOST")
	g.Port = os.Getenv("GRPC_PORT")
}

func NewConfig() *Config {
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
	config.HTTP.Set()
	config.GRPC.Set()

	return config
}
