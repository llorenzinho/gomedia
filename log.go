package gomedia

import (
	"os"

	"github.com/phuslu/log"
)

func getLogger() *log.Logger {
	return &log.Logger{
		Level: log.InfoLevel,
		Writer: &log.ConsoleWriter{
			Writer: os.Stdout,
		},
	}
}
