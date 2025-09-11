package fuel

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/rs/zerolog"
)

func PusherLogger(
	rpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return pusher.AppLogger("fuel").With().
		Str("chainRpcUrl", rpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
