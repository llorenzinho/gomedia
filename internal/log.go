package internal

import (
	"os"

	"github.com/phuslu/log"
)

func GetLogger() *log.Logger {
	return &log.Logger{
		Level: log.InfoLevel,
		Writer: &log.ConsoleWriter{
			Writer: os.Stdout,
		},
	}
}
