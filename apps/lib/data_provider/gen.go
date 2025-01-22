package data_provider

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

const (
	dirMode = 0o777
)

func generateDataProvider(cmd *cobra.Command, args []string) error {
	dataProviderName, _ := cmd.Flags().GetString(DataProviderNameFlag)

	mainLogger := utils.MainLogger()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	basePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := validateDataProviderName(dataProviderName, basePath); err != nil {
		return fmt.Errorf("failed to validate data provider name: %w", err)
	}

	mainLogger.Info().Msg("Generating data provider")

	if err := generateAll(dataProviderName, basePath); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	if err := runUpdateSharedCodeScript(basePath); err != nil {
		return fmt.Errorf("failed to run Python script: %w", err)
	}

	return nil
}

func validateDataProviderName(dataProviderName string, basePath string) error {
	if !validatePascalCase(dataProviderName) {
		return fmt.Errorf("data provider name must be in PascalCase. Please try again.")
	}

	dataSourcesDir := basePath + "/apps/lib/data_provider/sources"
	dirEntries, err := os.ReadDir(dataSourcesDir)
	if err != nil {
		return fmt.Errorf("failed to read data sources directory: %w", err)
	}

	existingDataNames := []string{}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			existingDataNames = append(existingDataNames, dirEntry.Name())
		}
	}

	if slices.Contains(existingDataNames, pascalToLower(dataProviderName)) {
		return fmt.Errorf("data provider name already taken. Please try again.")
	}

	return nil
}

func generateAll(pascalName string, basePath string) error {
	sourceDirPath := basePath + "/apps/lib/data_provider/sources/" + pascalToLower(pascalName)
	if err := os.Mkdir(sourceDirPath, dirMode); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	if err := generateConfigFile(sourceDirPath, pascalName); err != nil {
		return fmt.Errorf("failed to generate config file: %w", err)
	}

	if err := generateDataSourceFile(sourceDirPath, pascalName); err != nil {
		return fmt.Errorf("failed to generate data source file: %w", err)
	}

	if err := generateDataSourceTestFile(sourceDirPath, pascalName); err != nil {
		return fmt.Errorf("failed to generate data source test file: %w", err)
	}

	if err := generateInitFile(sourceDirPath, pascalName); err != nil {
		return fmt.Errorf("failed to generate init file: %w", err)
	}

	configSchemaPath := basePath + "/apps/lib/data_provider/configs/resources/source_config_schemas"
	if err := generateSourceConfigSchema(configSchemaPath, pascalName); err != nil {
		return fmt.Errorf("failed to generate source config schema: %w", err)
	}

	configTestPath := basePath + "/apps/lib/data_provider/configs/source_config_tests"
	if err := generateSourceConfigTest(configTestPath, pascalName); err != nil {
		return fmt.Errorf("failed to generate source config test: %w", err)
	}

	return nil
}

func runUpdateSharedCodeScript(basePath string) error {
	scriptPath := basePath + "/apps/scripts/update_shared_data_provider_code.py"
	cmd := exec.Command("python3", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func generateConfigFile(sourceDirPath string, pascalName string) error {
	configContent := `// Code generated by go generate.

package {{ .LowerStr }}

import "github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"

type {{ .PascalStr }}Config struct {
	DataSource types.DataSourceId ` + "`json:\"dataSource\"`" + `// required for all Data Provider Sources
	// TODO: Add any additional config parameters needed to pull a particular data feed
}

`
	return generateFile(sourceDirPath+"/config.go", configContent, pascalName)
}

func generateDataSourceFile(sourceDirPath string, pascalName string) error {
	dataSourceContent := `// Code generated by go generate.

package {{ .LowerStr }}

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
)

type {{ .CamelStr }}DataSource struct {
	{{ .PascalStr }}Config {{ .PascalStr }}Config
	// TODO: set any necessary parameters
}

func new{{ .PascalStr }}DataSource(sourceConfig types.DataProviderSourceConfig) *{{ .CamelStr }}DataSource {
	{{ .CamelStr }}Config, err := GetSourceSpecificConfig(sourceConfig)
	if err != nil {
		panic("unable to decode config: " + err.Error())
	}

	// TODO: add any necessary initialization code
	return &{{ .CamelStr }}DataSource{
		{{ .PascalStr }}Config: {{ .CamelStr }}Config,
	}
}

func (r {{ .CamelStr }}DataSource) RunDataSource(updatesCh chan types.DataSourceUpdateMap) {
	// TODO: Write all logic to fetch data points and report them to updatesCh
	panic("implement me")
}

`
	return generateFile(sourceDirPath+"/data_source.go", dataSourceContent, pascalName)
}

func generateDataSourceTestFile(sourceDirPath string, pascalName string) error {
	dataSourceTestContent := `// Code generated by go generate.

package {{ .LowerStr }}

import (
	"testing"
)

func Test{{ .PascalStr }}DataSource(t *testing.T) {
	// TODO: write some unit tests for your data source
	t.Fatalf("implement me")
}
	
`
	return generateFile(sourceDirPath+"/data_source_test.go", dataSourceTestContent, pascalName)
}

func generateInitFile(sourceDirPath string, pascalName string) error {
	initContent := `// Code generated by go generate.

package {{ .LowerStr }}

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/mitchellh/mapstructure"
)

var {{ .PascalStr }}DataSourceId types.DataSourceId = types.DataSourceId(utils.GetCurrentDirName())

type {{ .CamelStr }}DataSourceFactory struct{}

func (f *{{ .CamelStr }}DataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return new{{ .PascalStr }}DataSource(sourceConfig)
}

func init() {
	sources.RegisterDataSourceFactory({{ .PascalStr }}DataSourceId, &{{ .CamelStr }}DataSourceFactory{})
}

// assert we're satisfying our interfaces
var (
	_ types.DataSource        = (*{{ .CamelStr }}DataSource)(nil)
	_ types.DataSourceFactory = (*{{ .CamelStr }}DataSourceFactory)(nil)
)

func GetSourceSpecificConfig(sourceConfig types.DataProviderSourceConfig) ({{ .PascalStr }}Config, error) {
	var config {{ .PascalStr }}Config
	err := mapstructure.Decode(sourceConfig.Config, &config)

	return config, err
}

`
	return generateFile(sourceDirPath+"/init.go", initContent, pascalName)
}

func generateSourceConfigSchema(configSchemaPath string, pascalName string) error {
	configSchemaContent := `{
  "$id": "/resources/source_config_schemas/{{ .LowerStr }}",
  "type": "object",
  "properties": {
    "dataSource": {
      "type": "string",
      "const": "{{ .LowerStr }}"
    }
  },
  "required": ["dataSource"],
  "additionalProperties": false
}`
	return generateFile(
		fmt.Sprintf("%s/%s.json", configSchemaPath, pascalToLower(pascalName)), configSchemaContent, pascalName,
	)
}

func generateSourceConfigTest(configTestPath string, pascalName string) error {
	configTestContent := `// Code generated by go generate.
package config

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/configs"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources/{{ .LowerStr }}"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestValid{{ .PascalStr }}Config(t *testing.T) {

	// TODO: set this to a valid config string using a feed from your new source
	validConfig := ` + "`" + `
	{
	  "sources": [
		{
		  "id": "MY_VALUE",
		  "config": {
			"dataSource": "{{ .LowerStr }}"
		  }
		}
	  ]
	}` + "`" + `

	config, err := configs.LoadConfigFromBytes([]byte(validConfig))
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Sources))

	sourceConfig := config.Sources[0]

	dataSourceId, err := utils.GetDataSourceId(sourceConfig.Config)
	assert.NoError(t, err)
	assert.Equal(t, {{ .LowerStr }}.{{ .PascalStr }}DataSourceId, dataSourceId)

	sourceSpecificConfig, err := {{ .LowerStr }}.GetSourceSpecificConfig(sourceConfig)
	assert.NoError(t, err)
	assert.NotNil(t, sourceSpecificConfig)

	// TODO: write some asserts to check that the fields on sourceSpecificConfig have the values you'd expect
	t.Fatalf("implement me")
}`
	return generateFile(
		fmt.Sprintf("%s/%s_test.go", configTestPath, pascalToLower(pascalName)), configTestContent, pascalName,
	)
}

func generateFile(filePath string, fileContent string, pascalName string) error {
	tmpl, err := template.New("").Parse(fileContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	inputData := struct {
		PascalStr string
		LowerStr  string
		CamelStr  string
	}{
		PascalStr: pascalName,
		LowerStr:  pascalToLower(pascalName),
		CamelStr:  pascalToCamel(pascalName),
	}

	if err := tmpl.Execute(file, inputData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func validatePascalCase(name string) bool {
	pascalCasePattern := regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
	if !pascalCasePattern.MatchString(name) {
		return false
	}

	return true
}

func pascalToLower(pascalName string) string {
	return strings.ToLower(pascalName)
}

func pascalToCamel(pascalName string) string {
	return strings.ToLower(pascalName[:1]) + pascalName[1:]
}
