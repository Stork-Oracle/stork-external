package cosmwasm

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/rs/zerolog"
)

func PusherLogger(
	chainGrpcUrl string,
	contractAddress string,
) zerolog.Logger {
	return pusher.AppLogger("cosmwasm").With().
		Str("chainGrpcUrl", chainGrpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}
