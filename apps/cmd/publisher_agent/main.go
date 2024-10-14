package main

import (
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"

	"github.com/Stork-Oracle/stork-external/apps/lib/publisher_agent"
)

var verbose bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "stork-publisher-agent",
		Short: "Stork CLI tool for publishing prices to the Stork network",
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

	rootCmd.AddCommand(publisher_agent.PublisherAgentCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
