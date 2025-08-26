package solana

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/rs/zerolog"
)

func PusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return pusher.AppLogger("solana").With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}