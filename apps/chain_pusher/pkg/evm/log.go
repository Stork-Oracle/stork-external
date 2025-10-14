package evm

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/rs/zerolog"
)

func PusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return pusher.AppLogger("evm").With().
		Str("chainHttpRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
