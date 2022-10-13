package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	logFunc  func(*Logger, ...any)
	level    string
	text     string
	expected string
}

func TestLogger(t *testing.T) {
	file := createTmp()
	defer remove(file.Name())

	testTable := []testCase{
		{
			logFunc:  (*Logger).Error,
			level:    "info",
			text:     "text to log",
			expected: "text to log\n",
		},
		{
			logFunc:  (*Logger).Info,
			level:    "warn",
			text:     "some symbols: $!@#$",
			expected: "",
		},
		{
			logFunc:  (*Logger).Error,
			level:    "error",
			text:     "this text will be logged",
			expected: "this text will be logged\n",
		},
		{
			logFunc:  (*Logger).Warn,
			level:    "error",
			text:     "empty",
			expected: "",
		},
		{
			logFunc:  (*Logger).Info,
			level:    "error",
			text:     "l e t t e r s",
			expected: "",
		},
		{
			logFunc:  (*Logger).Debug,
			level:    "error",
			text:     "a few words",
			expected: "",
		},
		{
			logFunc:  (*Logger).Error,
			level:    "warn",
			text:     "some text",
			expected: "some text\n",
		},
		{
			logFunc:  (*Logger).Info,
			level:    "warn",
			text:     "this text won't be logged",
			expected: "",
		},
	}

	for i, c := range testTable {
		t.Run(fmt.Sprintf("test_N%d", i+1), func(t *testing.T) {
			l := createLogger(c.level, file.Name())
			c.logFunc(l, c.text)
			b, err := io.ReadAll(file)
			require.Nil(t, err)
			require.Equal(t, c.expected, string(b))
		})
	}
}

func createLogger(level, outputPath string) *Logger {
	config := getLoggerConfig(level, outputPath)

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	sugaredLogger := logger.Sugar()

	return &Logger{
		Logger: sugaredLogger,
	}
}

func remove(fileName string) {
	if err := os.RemoveAll(fileName); err != nil {
		fmt.Println(err)
	}
}

func createTmp() *os.File {
	file, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	}

	return file
}
