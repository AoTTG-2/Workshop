package migrator

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
}

func (l *Logger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *Logger) Verbose() bool {
	return zerolog.GlobalLevel() <= zerolog.DebugLevel
}
