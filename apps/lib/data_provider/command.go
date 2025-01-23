package data_provider

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

const (
	startupAnimationPath = "apps/lib/data_provider/configs/resources/frames"
)

var GenerateDataProviderCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate skeleton code for a new data source integration",
	RunE:  generateDataProvider,
}

var UpdateSharedCodeCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the shared code for the data provider sources",
	RunE:  runUpdateSharedCode,
}

var StartDataProviderCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a process to fetch prices from data sources",
	RunE:  runDataProvider,
}

// required
const (
	ConfigFilePathFlag   = "config-file-path"
	OutputAddressFlag    = "output-address"
	DataProviderNameFlag = "data-provider-name"
)

func init() {
	StartDataProviderCmd.Flags().StringP(ConfigFilePathFlag, "c", "", "the path of your config json file")
	StartDataProviderCmd.Flags().StringP(
		OutputAddressFlag, "o", "", "a string representing an output address (e.g. ws://localhost:5216/)",
	)
	StartDataProviderCmd.MarkFlagRequired(ConfigFilePathFlag)

	GenerateDataProviderCmd.Flags().StringP(
		DataProviderNameFlag, "n", "", "the name of your data provider in PascalCase (e.g. MyProvider)",
	)
	GenerateDataProviderCmd.MarkFlagRequired(DataProviderNameFlag)
}

func runDataProvider(cmd *cobra.Command, args []string) error {
	configFilePath, _ := cmd.Flags().GetString(ConfigFilePathFlag)
	outputAddress, _ := cmd.Flags().GetString(OutputAddressFlag)

	mainLogger := utils.MainLogger()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger.Info().Msg("Starting data provider")

	config, err := LoadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	if err := RunAnimation(); err != nil {
		mainLogger.Debug().Err(err).Msg("failed to run animation")
	}

	runner := NewDataProviderRunner(*config, outputAddress)
	runner.Run()

	return nil
}

func runUpdateSharedCode(cmd *cobra.Command, args []string) error {
	basePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	RunAnimation()

	return updateSharedCode(basePath)
}

func RunAnimation() error {
	basePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	frames, err := os.ReadDir(filepath.Join(basePath, startupAnimationPath))
	if err != nil {
		return fmt.Errorf("failed to read frames: %w", err)
	}

	for _, frame := range frames {
		frameContent, err := os.ReadFile(filepath.Join(basePath, startupAnimationPath, frame.Name()))
		if err != nil {
			return fmt.Errorf("failed to read frame %s: %w", frame.Name(), err)
		}

		// Clear the screen
		fmt.Print("\033[H\033[2J")

		// Print the frame
		fmt.Println(string(frameContent))
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Print("\033[H\033[2J")

	return nil
}
