// The following directive is necessary to make the package coherent:
//go:build ignore
// +build ignore

// This program generates (TODO fill in here). It can be invoked by running go generate

package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Get the current working directory to debug paths
	basePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}
	pascalName := getDataProviderName()
	generateAll(pascalName, basePath)
}

func getDataProviderName() string {
	for {
		fmt.Print("Please enter the name of your data provider in PascalCase: ")
		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		input = input[:len(input)-1]

		if validatePascalCase(input) {
			return input
		}

		fmt.Println("Invalid, name not in PascalCase. Please try again.")
	}
}

func generateAll(pascalName string, basePath string) {
	sourceDirPath := basePath + "/apps/lib/data_provider/sources/" + pascalToLower(pascalName)
	err := os.Mkdir(sourceDirPath, 0777) // what permissions should we have?
	if err != nil {
		log.Fatal(err)
	}

	generateConfigFile(sourceDirPath, pascalName)
	generateDataSourceFile(sourceDirPath, pascalName)
	generateDataSourceTestFile(sourceDirPath, pascalName)
	generateInitFile(sourceDirPath, pascalName)

	configSchemaPath := basePath + "/apps/lib/data_provider/configs/resources/source_config_schemas"
	generateSourceConfigSchema(configSchemaPath, pascalName)

	configTestPath := basePath + "/apps/lib/data_provider/configs/source_config_tests"
	generateSourceConfigTest(configTestPath, pascalName)
}

func generateConfigFile(sourceDirPath string, pascalName string) {
	configContent := `// Code generated by go generate.

package {{ .LowerStr }}

import "github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"

type {{ .PascalStr }}Config struct {
	DataSource types.DataSourceId ` + "`json:\"dataSource\"`" + `// required for all Data Provider Sources
	// TODO: Add any additional config parameters needed to pull a particular data feed
}

`
	generateFile(sourceDirPath + "/config.go", configContent, pascalName)
}

func generateDataSourceFile(sourceDirPath string, pascalName string) {
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
	generateFile(sourceDirPath + "/data_source.go", dataSourceContent, pascalName)	
}

func generateDataSourceTestFile(sourceDirPath string, pascalName string) {
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
	generateFile(sourceDirPath + "/data_source_test.go", dataSourceTestContent, pascalName)	
}

func generateInitFile(sourceDirPath string, pascalName string) {
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
	generateFile(sourceDirPath + "/init.go", initContent, pascalName)
}

func generateSourceConfigSchema(configSchemaPath string, pascalName string) {
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
	generateFile(fmt.Sprintf("%s/%s.json", configSchemaPath, pascalToLower(pascalName)), configSchemaContent, pascalName)
}

func generateSourceConfigTest(configTestPath string, pascalName string) {
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
	generateFile(fmt.Sprintf("%s/%s_test.go", configTestPath, pascalToSnake(pascalName)), configTestContent, pascalName)

}

func generateFile(filePath string, fileContent string, pascalName string) {
	template := template.Must(template.New("").Parse(fileContent))

	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = template.Execute(file, struct {
		PascalStr string
		LowerStr  string
		CamelStr  string
		SnakeStr  string
	}{
		PascalStr: pascalName,
		LowerStr:  pascalToLower(pascalName),
		CamelStr:  pascalToCamel(pascalName),
		SnakeStr:  pascalToSnake(pascalName),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func validatePascalCase(name string) bool {
	pascalCasePattern := regexp.MustCompile(`^[A-Z][A-Za-z]*$`)
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

func pascalToSnake(pascalName string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(pascalName, "${1}_${2}")

	return strings.ToLower(snake)
}
