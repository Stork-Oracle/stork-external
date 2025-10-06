//go:build initia

package main

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/initia_minimove"
	"github.com/spf13/cobra"
)

func addCosmwasmCmd(_ *cobra.Command) {
	// Not available when building with initia tag
}

func addInitiaMiniMoveCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(initia_minimove.NewPushCmd())
}
