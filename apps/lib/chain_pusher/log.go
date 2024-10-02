package chain_pusher

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func BaseLogger(application string) zerolog.Logger {
	return log.With().Str("application", application).
		Logger().
		Hook(zerolog.HookFunc(func(e *zerolog.Event, l zerolog.Level, msg string) {
			// for capture by cloudwatch log filter, should be in line with python logger levels
			e.Int32("levelno", (int32(l)+1)*10)
		}))
}

func AppLogger(command string) zerolog.Logger {
	return BaseLogger("stork-chain-pusher").With().Str("command", command).Logger()
}

func EvmPusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return AppLogger("evm").With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
