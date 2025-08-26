package pusher

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

func SolanaPusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return AppLogger("solana").With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}

func SuiPusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return AppLogger("sui").With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}

func CosmwasmPusherLogger(
	chainGrpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return AppLogger("cosmwasm").With().
		Str("chainGrpcUrl", chainGrpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}

func AptosPusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return AppLogger("aptos").With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}

func FuelPusherLogger(
	rpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return AppLogger("fuel").With().
		Str("chainRpcUrl", rpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
