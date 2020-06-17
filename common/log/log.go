package log

import (
	"os"

	"github.com/rs/zerolog"
)

var defaultLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()

func SetLogger(logger zerolog.Logger) {
	defaultLogger = logger
}

func SetLevel(level zerolog.Level) {
	defaultLogger = defaultLogger.Level(level)
}

func Info(format string, args ...interface{}) {
	defaultLogger.WithLevel(zerolog.InfoLevel).Msgf(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.WithLevel(zerolog.WarnLevel).Msgf(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.WithLevel(zerolog.ErrorLevel).Msgf(format, args...)
}
