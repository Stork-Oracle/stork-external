package generate

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var GenerateDataProviderCmd = &cobra.Command{
	Use:   "data-provider [name]",
	Short: "Generate a new data provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return generateDataProvider(cmd, args)
	},
}

var UpdateSharedCodeCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the shared code for the data provider sources",
	RunE:  runUpdateSharedCode,
}

// required
const (
	DataProviderNameFlag = "data-provider-name"
)

func runUpdateSharedCode(cmd *cobra.Command, args []string) error {
	basePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	return updateSharedCode(basePath)
}
