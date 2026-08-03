package observability

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger sets up structured JSON logging to stdout with Unix timestamps
// and caller info — consistent format across all three services so logs
// aggregate cleanly regardless of which service emitted them.
func InitLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Caller().
		Str("service", "payment").
		Logger()
	log.Logger = logger
	return logger
}