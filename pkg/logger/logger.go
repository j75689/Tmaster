package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// NewLogger returns a zerolog.Logger
func NewLogger(logLevel, logFormat string) (zerolog.Logger, error) {
	level := zerolog.InfoLevel
	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		return zerolog.Logger{}, err
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	switch strings.ToLower(logFormat) {
	case "json":
		return zerolog.New(os.Stdout).With().Caller().Timestamp().Logger(), nil
	default: // default console
		return zerolog.New(os.Stdout).With().Caller().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout}), nil
	}
}
