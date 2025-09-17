package utils

import (
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func baseAppLogger() zerolog.Logger {
	return log.With().Str("application", "stork-data-provider").Logger()
}

func MainLogger() zerolog.Logger {
	return baseAppLogger().With().Str("service", "main").Logger()
}

func WriterLogger() zerolog.Logger {
	return baseAppLogger().With().Str("service", "writer").Logger()
}

func DataSourceLogger(dataSourceId types.DataSourceID) zerolog.Logger {
	return baseAppLogger().With().Str("service", "data_source").Str("data_source_id", string(dataSourceId)).Logger()
}
