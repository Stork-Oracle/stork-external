package pusher

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// BaseLogger is a base logger for the application.
func BaseLogger(application string) zerolog.Logger {
	return log.With().Str("application", application).
		Logger().
		Hook(zerolog.HookFunc(func(e *zerolog.Event, l zerolog.Level, msg string) {
			// for capture by cloudwatch log filter, should be in line with python logger levels
			//nolint:mnd // Magic number is a one-off used to map log levels to cloudwatch log levels
			e.Int32("levelno", (int32(l)+1)*10)
		}))
}

// AppLogger is a base logger configured for the chain pusher.
func AppLogger(command string) zerolog.Logger {
	return BaseLogger("first-party-chain-pusher").With().Str("command", command).Logger()
}

func PusherLogger(
	command string,
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return AppLogger(command).With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
