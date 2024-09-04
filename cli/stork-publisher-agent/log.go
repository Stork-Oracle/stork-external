package stork_publisher_agent

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func BaseAppLogger() zerolog.Logger {
	return log.With().Str("application", "stork-publisher-agent").Logger()
}

func MainLogger() zerolog.Logger {
	return BaseAppLogger().With().Str("service", "main").Logger()
}

func RunnerLogger(signatureType SignatureType) zerolog.Logger {
	return BaseAppLogger().With().Str("service", "runner").Str("signature_type", string(signatureType)).Logger()
}

func IncomingLogger() zerolog.Logger {
	return BaseAppLogger().With().Str("service", "incoming").Logger()
}

type HttpHeaders http.Header

func (hdrs HttpHeaders) MarshalZerologObject(e *zerolog.Event) {
	if hdrs == nil {
		return
	}
	m := (map[string][]string)(hdrs)
	for key, val := range m {
		e.Strs(key, val)
	}
}
