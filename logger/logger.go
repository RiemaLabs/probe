package logger

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func InitLogger(logLevel string, pretty bool, output io.Writer) {
	zerolog.TimeFieldFormat = time.RFC3339

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	if pretty {
		output = zerolog.ConsoleWriter{Out: output, TimeFormat: "2006-01-02 15:04:05"}
	}

	Logger = zerolog.New(output).With().Timestamp().Logger().Level(level)
}

func Debug(msg string, fields ...interface{}) {
	Logger.Debug().Fields(fields).Msg(msg)
}

func Info(msg string, fields ...interface{}) {
	Logger.Info().Fields(fields).Msg(msg)
}

func Warn(msg string, fields ...interface{}) {
	Logger.Warn().Fields(fields).Msg(msg)
}

func Error(msg string, err error, fields ...interface{}) {
	Logger.Error().Err(err).Fields(fields).Msg(msg)
}

func Fatal(msg string, err error, fields ...interface{}) {
	Logger.Fatal().Err(err).Fields(fields).Msg(msg)
}
