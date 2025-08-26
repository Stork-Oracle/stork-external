package main

import (
	"log"
	"os"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/aptos"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/evm"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/fuel"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/solana"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/sui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "stork-chain-push",
		Short: "Stork CLI tool for pushing prices to contracts on multiple chains",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			zerolog.TimeFieldFormat = time.RFC3339Nano
			zerolog.DurationFieldUnit = time.Nanosecond
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

	rootCmd.AddCommand(evm.PushCmd)
	rootCmd.AddCommand(solana.PushCmd)
	rootCmd.AddCommand(sui.PushCmd)
	rootCmd.AddCommand(cosmwasm.PushCmd)
	rootCmd.AddCommand(aptos.PushCmd)
	rootCmd.AddCommand(fuel.PushCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
