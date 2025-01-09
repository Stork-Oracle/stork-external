package data_provider

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func baseAppLogger() zerolog.Logger {
	return log.With().Str("application", "stork-data-provider").Logger()
}

func MainLogger() zerolog.Logger {
	return baseAppLogger().With().Str("service", "main").Logger()
}

func writerLogger() zerolog.Logger {
	return baseAppLogger().With().Str("service", "writer").Logger()
}

func dataSourceLogger(dataSourceId DataSourceId) zerolog.Logger {
	return baseAppLogger().With().Str("service", "data_source").Str("data_source_id", string(dataSourceId)).Logger()
}
