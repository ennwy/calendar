package main

import (
	"os"

	"github.com/ghodss/yaml"
)

type Config struct {
	Logger  LoggerConfig `yaml:"logger"`
	Storage string       `yaml:"storage"`
	HTTP    HTTPConfig   `yaml:"http"`
	DB      DBConfig     `yaml:"db"`
}

type DBConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

func (db *DBConfig) SetEnv() (err error) {
	dbInfo := map[string]string{
		"DATABASE_USER":     db.User,
		"DATABASE_PORT":     db.Port,
		"DATABASE_HOST":     db.Host,
		"DATABASE_PASSWORD": db.Password,
		"DATABASE_NAME":     db.Name,
	}

	for k, v := range dbInfo {
		if err = os.Setenv(k, v); err != nil {
			return err
		}
	}

	return err
}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	OutputPath string `yaml:"outputPath"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func NewConfig(path string) (*Config, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(configData, config)

	if err != nil {
		return nil, err
	}

	err = config.DB.SetEnv()

	return config, err
}
