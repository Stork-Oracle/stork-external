//go:build !cosmwasm && !initia

package main

import (
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/initia_minimove"
	"github.com/spf13/cobra"
)

func addCosmwasmCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(cosmwasm.NewPushCmd())
}

func addInitiaMiniMoveCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(initia_minimove.NewPushCmd())
}
