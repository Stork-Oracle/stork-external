// Package main provides the CLI entrypoint for the chain pusher.
package main

import (
	"log"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/aptos"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/evm"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/fuel"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/initia_minimove"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/solana"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/sui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "stork-chain-push",
		Short: "Stork CLI tool for pushing prices to contracts on multiple chains",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			//nolint:reassign
			zerolog.TimeFieldFormat = time.RFC3339Nano
			//nolint:reassign
			zerolog.DurationFieldUnit = time.Nanosecond
			//nolint:reassign
			zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

			var logLevel zerolog.Level
			if verbose {
				logLevel = zerolog.DebugLevel
			} else {
				logLevel = zerolog.InfoLevel
			}

			// set global log level
			zerolog.SetGlobalLevel(logLevel)
		},
	}
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	rootCmd.AddCommand(evm.NewPushCmd())
	rootCmd.AddCommand(solana.NewPushCmd())
	rootCmd.AddCommand(sui.NewPushCmd())
	rootCmd.AddCommand(cosmwasm.NewPushCmd())
	rootCmd.AddCommand(aptos.NewPushCmd())
	rootCmd.AddCommand(fuel.NewPushCmd())
	rootCmd.AddCommand(initia_minimove.NewPushCmd())

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
