package config

import (
	"os"
)

//func (db *DBConfig) SetEnv() (err error) {
//	dbInfo := map[string]string{
//		"DATABASE_USER":     db.User,
//		"DATABASE_PORT":     db.Port,
//		"DATABASE_HOST":     db.Host,
//		"DATABASE_PASSWORD": db.Password,
//		"DATABASE_NAME":     db.Name,
//	}
//
//	for k, v := range dbInfo {
//		if err = os.Setenv(k, v); err != nil {
//			return err
//		}
//	}
//
//	return err
//}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	OutputPath string `yaml:"outputPath"`
}

func (l *LoggerConfig) Set() {
	l.Level = os.Getenv("LOG_LEVEL")
	l.OutputPath = os.Getenv("LOG_OUTPUT_PATH")
}
