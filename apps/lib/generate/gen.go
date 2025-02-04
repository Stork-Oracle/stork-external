package generate

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

//go:embed resources
var resourcesFS embed.FS

const (
	dirMode      = 0o777
	templatesDir = "resources/templates"
	sourceDir    = "apps/lib/data_provider/sources"
	pathPrefix   = "// @path: "
)

type templateStrings struct {
	PascalStr    string
	LowerStr     string
	CamelStr     string
	PackageNames []string
}

type templateFile struct {
	content  string
	destPath string
}

func generateDataProvider(cmd *cobra.Command, args []string) error {
	dataProviderName := args[0]

	mainLogger := MainLogger()

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

	if err := generateSourceCode(dataProviderName, basePath, mainLogger); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	if err := updateSharedCode(basePath); err != nil {
		return fmt.Errorf("failed to run Python script: %w", err)
	}

	return nil
}

func validateDataProviderName(dataProviderName string, basePath string) error {
	if !validatePascalCase(dataProviderName) {
		return fmt.Errorf("data provider name must be in PascalCase. Please try again.")
	}

	if err := validateUniqueDataSourceName(dataProviderName, basePath); err != nil {
		return err
	}

	return nil
}

func generateSourceCode(pascalName string, basePath string, mainLogger zerolog.Logger) error {
	stringData := templateStrings{
		PascalStr: pascalName,
		LowerStr:  pascalToLower(pascalName),
		CamelStr:  pascalToCamel(pascalName),
	}

	templates, err := processTemplateFiles(basePath, templatesDir+"/source", stringData)
	if err != nil {
		return fmt.Errorf("failed to process templates: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(basePath, sourceDir, stringData.LowerStr), dirMode); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	for _, template := range templates {
		if err := generateFileFromContent(template.destPath, template.content, stringData); err != nil {
			return fmt.Errorf("failed to generate file %s: %w", template.destPath, err)
		}
		mainLogger.Info().Msgf("Generated file %s", template.destPath)
	}

	return nil
}

func updateSharedCode(basePath string) error {
	packageNames, err := getPackageNames(filepath.Join(basePath, sourceDir))
	if err != nil {
		return fmt.Errorf("failed to get sources metadata: %w", err)
	}

	stringData := templateStrings{
		PackageNames: packageNames,
	}

	templates, err := processTemplateFiles(basePath, templatesDir+"/shared", stringData)
	if err != nil {
		return fmt.Errorf("failed to process templates: %w", err)
	}

	for _, template := range templates {
		if err := generateFileFromContent(template.destPath, template.content, stringData); err != nil {
			return fmt.Errorf("failed to generate file %s: %w", template.destPath, err)
		}
	}

	return nil
}

func processTemplateFiles(basePath string, templateDir string, stringData templateStrings) ([]templateFile, error) {
	dirEntries, err := resourcesFS.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	templates := make([]templateFile, 0, len(dirEntries))

	for _, entry := range dirEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gotmpl") {
			continue
		}

		// Entry is a template file, load into templateFile struct
		template, err := loadTemplate(basePath, templateDir+"/"+entry.Name(), stringData)
		if err != nil {
			return nil, fmt.Errorf("failed to load template %s: %w", entry.Name(), err)
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func generateFileFromContent(filePath string, content string, stringData templateStrings) error {
	tmpl, err := template.New("").Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, stringData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func loadTemplate(basePath string, templatePath string, stringData templateStrings) (templateFile, error) {
	contentBytes, err := resourcesFS.ReadFile(templatePath)
	if err != nil {
		return templateFile{}, fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	lines := strings.Split(string(contentBytes), "\n")
	if len(lines) == 0 {
		return templateFile{}, fmt.Errorf("empty template file: %s", templatePath)
	}

	// Extract path from first line comment
	pathComment := lines[0]
	if !strings.HasPrefix(pathComment, pathPrefix) {
		return templateFile{}, fmt.Errorf("template missing @path comment: %s", templatePath)
	}

	genericPath := strings.TrimPrefix(pathComment, pathPrefix)

	// Parse the destination path as a template
	pathTemplate, err := template.New("").Parse(genericPath)
	if err != nil {
		return templateFile{}, fmt.Errorf("failed to parse path template: %w", err)
	}

	var destPath strings.Builder
	if err := pathTemplate.Execute(&destPath, stringData); err != nil {
		return templateFile{}, fmt.Errorf("failed to execute path template: %w", err)
	}

	// Remove the path comment from content
	content := strings.Join(lines[1:], "\n")

	return templateFile{
		content:  content,
		destPath: filepath.Join(basePath, destPath.String()),
	}, nil
}

func getPackageNames(sourcesDir string) ([]string, error) {
	entries, err := os.ReadDir(sourcesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sources directory: %w", err)
	}

	packageNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			packageNames = append(packageNames, entry.Name())
		}
	}
	slices.Sort(packageNames)

	return packageNames, nil
}

func validatePascalCase(name string) bool {
	pascalCasePattern := regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
	if !pascalCasePattern.MatchString(name) {
		return false
	}

	return true
}

func validateUniqueDataSourceName(dataProviderName string, basePath string) error {
	dataSourcesDir := basePath + "/" + sourceDir
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

func pascalToLower(pascalName string) string {
	return strings.ToLower(pascalName)
}

func pascalToCamel(pascalName string) string {
	// Find end of first word by first lowercase or last uppercase in sequence of uppercase
	var endOfFirstWord int
	for i := 1; i < len(pascalName); i++ {
		if pascalName[i] >= 'a' && pascalName[i] <= 'z' {
			if i == 1 {
				endOfFirstWord = i
			} else {
				endOfFirstWord = i - 1
			}

			break
		}
	}

	return strings.ToLower(pascalName[:endOfFirstWord]) + pascalName[endOfFirstWord:]
}

func removeDataProvider(cmd *cobra.Command, args []string) error {
	dataProviderName := args[0]

	mainLogger := MainLogger()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	basePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	mainLogger.Info().Msg("Removing data provider")

	if err := deleteSourceCode(dataProviderName, basePath, mainLogger); err != nil {
		return fmt.Errorf("failed to delete source code: %w", err)
	}

	if err := updateSharedCode(basePath); err != nil {
		return fmt.Errorf("failed to run Python script: %w", err)
	}

	return nil
}

func deleteSourceCode(pascalName string, basePath string, mainLogger zerolog.Logger) error {
	stringData := templateStrings{
		PascalStr: pascalName,
		LowerStr:  pascalToLower(pascalName),
		CamelStr:  pascalToCamel(pascalName),
	}

	templates, err := processTemplateFiles(basePath, templatesDir+"/source", stringData)
	if err != nil {
		return fmt.Errorf("failed to process templates: %w", err)
	}

	if err := os.RemoveAll(filepath.Join(basePath, sourceDir, stringData.LowerStr)); err != nil {
		return fmt.Errorf("failed to delete source directory: %w", err)
	}

	for _, template := range templates {
		if err := os.RemoveAll(template.destPath); err != nil {
			return fmt.Errorf("failed to delete file %s: %w", template.destPath, err)
		}
		mainLogger.Info().Msgf("Deleted file %s", template.destPath)
	}

	return nil
}
