package aptos

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/rs/zerolog"
)

func PusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return pusher.AppLogger("aptos").With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
