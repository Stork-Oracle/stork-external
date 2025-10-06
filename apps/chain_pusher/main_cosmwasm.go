//go:build cosmwasm

package main

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm"
	"github.com/spf13/cobra"
)

func addCosmwasmCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(cosmwasm.NewPushCmd())
}

func addInitiaMiniMoveCmd(_ *cobra.Command) {
	// Not available when building with cosmwasm tag
}
