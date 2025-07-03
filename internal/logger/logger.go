package logger

import (
	"github.com/gin-gonic/gin"
	zerolog "github.com/rs/zerolog"
	"log"
	"os"
	"time"
)

func InitLogger(level string, pretty bool) *zerolog.Logger {
	var logger zerolog.Logger

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Printf("Invalid log level '%s', defaulting to 'info'", level)
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	if pretty {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        zerolog.ConsoleWriter{Out: log.Writer()},
			TimeFormat: time.RFC3339,
		}).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		logger = zerolog.New(os.Stderr).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	log.SetFlags(0)
	log.SetOutput(logger)

	logger.Info().
		Str("logLevel", level).
		Msg("Initialized logger")

	return &logger
}

func FromContext(c *gin.Context) *zerolog.Logger {
	logger, exists := c.Get("logger")
	if !exists {
		panic("logger not found in context")
	}
	return logger.(*zerolog.Logger)
}
