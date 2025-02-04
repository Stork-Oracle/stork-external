package generate

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func baseAppLogger() zerolog.Logger {
	return log.With().Str("application", "stork-generate").Logger()
}

func MainLogger() zerolog.Logger {
	return baseAppLogger().With().Str("service", "main").Logger()
}
