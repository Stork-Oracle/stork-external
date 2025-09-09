package main

import (
	"log"
	"os"
	"time"

	data_provider "github.com/Stork-Oracle/stork-external/apps/data_provider/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "stork-data-provider",
		Short: "Stork CLI tool for fetching prices from data sources",
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

	rootCmd.AddCommand(data_provider.StartDataProviderCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
