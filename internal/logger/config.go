package logger

import "os"

type Config struct {
	Level      string `yaml:"level"`
	OutputPath string `yaml:"outputPath"`
}

func (l *Config) Set() {
	l.Level = os.Getenv("LOG_LEVEL")
	l.OutputPath = os.Getenv("LOG_OUTPUT_PATH")
}
