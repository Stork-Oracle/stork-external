package main

import (
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "stork",
		Short: "Stork CLI tool for EVM contract updates based on WebSocket prices",
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	rootCmd.AddCommand(evmpushCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
