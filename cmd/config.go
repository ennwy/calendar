package config

import (
	"os"
)

type LoggerConfig struct {
	Level      string `yaml:"level"`
	OutputPath string `yaml:"outputPath"`
}

func (l *LoggerConfig) Set() {
	l.Level = os.Getenv("LOG_LEVEL")
	l.OutputPath = os.Getenv("LOG_OUTPUT_PATH")
}
