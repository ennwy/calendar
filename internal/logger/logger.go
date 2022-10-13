package logger

import (
	"github.com/ennwy/calendar/internal/app"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	Logger *zap.SugaredLogger
}

var _ app.Logger = (*Logger)(nil)

func (l *Logger) Debug(args ...any) { l.Logger.Debugln(args...) }
func (l *Logger) Info(args ...any)  { l.Logger.Infoln(args...) }
func (l *Logger) Warn(args ...any)  { l.Logger.Warnln(args...) }
func (l *Logger) Error(args ...any) { l.Logger.Errorln(args...) }
func (l *Logger) Fatal(args ...any) { l.Logger.Fatalln(args...) }

func New(level, outputPath string) *Logger {
	logger, err := getLoggerConfig(level, outputPath).Build()
	if err != nil {
		panic(err)
	}

	sugaredLogger := logger.Sugar()

	return &Logger{
		Logger: sugaredLogger,
	}
}

func getLoggerConfig(level, outputPath string) zap.Config {
	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		panic(err)
	}

	return zap.Config{
		Level:            atomicLevel,
		Encoding:         "console",
		OutputPaths:      []string{outputPath},
		ErrorOutputPaths: []string{outputPath},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: level,
		},
	}
}
