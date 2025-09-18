package main

import (
	"log"
	"time"

	self_serve_evm "github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/evm"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "self-serve-chain-pusher",
		Short: "Self-serve chain pusher for receiving publisher messages and pushing to configured contracts",
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

			zerolog.SetGlobalLevel(logLevel)
		},
	}
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	rootCmd.AddCommand(self_serve_evm.NewPushCmd())

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
