package initia_minimove

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/rs/zerolog"
)

func PusherLogger(
	chainRpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return pusher.AppLogger("initia_minimove").With().
		Str("chainRpcUrl", chainRpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
