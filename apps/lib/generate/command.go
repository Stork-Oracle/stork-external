package generate

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate [name]",
	Short: "Generate a new data provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateDataProvider(cmd, args)
	},
}

var RemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove code related toa data source integration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeDataProvider(cmd, args)
	},
}

var UpdateSharedCodeCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the shared code for all data provider sources",
	RunE:  runUpdateSharedCode,
}

func runUpdateSharedCode(cmd *cobra.Command, args []string) error {
	basePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	return updateSharedCode(basePath)
}
