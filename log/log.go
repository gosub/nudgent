package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func Init(humanReadable bool) {
	if humanReadable {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		Logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	zerolog.DefaultContextLogger = &Logger
}
