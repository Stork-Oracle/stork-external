package main

import (
	"log"
	"time"

	first_party_evm "github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/evm"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

func main() {
	var verbose bool

	rootCmd := &cobra.Command{
		Use:   "first-party-chain-pusher",
		Short: "First party chain pusher for receiving publisher messages and pushing to configured contracts",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
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

			zerolog.SetGlobalLevel(logLevel)
		},
	}
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	rootCmd.AddCommand(first_party_evm.NewPushCmd())

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
