package server

import "os"

type GRPCConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (g *GRPCConfig) Set() {
	g.Host = os.Getenv("GRPC_HOST")
	g.Port = os.Getenv("GRPC_PORT")
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (h *HTTPConfig) Set() {
	h.Host = os.Getenv("HTTP_HOST")
	h.Port = os.Getenv("HTTP_PORT")
}
