package stork_publisher_agent

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

func BaseAppLogger() zerolog.Logger {

	return log.With().Str("application", "stork-publisher-agent").
		Logger().
		Hook(zerolog.HookFunc(func(e *zerolog.Event, l zerolog.Level, msg string) {
			// for capture by cloudwatch log filter, should be in line with python logger levels
			e.Int32("levelno", (int32(l)+1)*10)
		}))
}

func MainLogger() zerolog.Logger {
	return BaseAppLogger().With().Str("service", "main").Logger()
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
