package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Translation represents a translation key and its usage
type Translation struct {
	Key      string
	File     string
	Line     int
	Used     bool
	Declared bool
	Locale   string // Add locale field
}

// AnalysisResult contains the results of the translation analysis
type AnalysisResult struct {
	UnusedTranslations    []Translation
	UndeclaredTranslations []Translation
	TotalTranslations     int
	UsedTranslations      int
	LocaleResults         map[string]*LocaleAnalysisResult // Add per-locale results
}

// LocaleAnalysisResult contains analysis results for a specific locale
type LocaleAnalysisResult struct {
	Locale                string
	UnusedTranslations    []Translation
	UndeclaredTranslations []Translation
	TotalTranslations     int
	UsedTranslations      int
}

// Analyzer handles the analysis of next-intl translations
type Analyzer struct {
	projectPath string
	results     *AnalysisResult
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(projectPath string) *Analyzer {
	return &Analyzer{
		projectPath: projectPath,
		results: &AnalysisResult{
			UnusedTranslations:    make([]Translation, 0),
			UndeclaredTranslations: make([]Translation, 0),
			LocaleResults:         make(map[string]*LocaleAnalysisResult),
		},
	}
}

// Analyze performs the translation analysis
func (a *Analyzer) Analyze() (*AnalysisResult, error) {
	if err := a.validateProjectPath(); err != nil {
		return nil, fmt.Errorf("invalid project path: %w", err)
	}

	// Find all translation files
	translationFiles, err := a.findTranslationFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding translation files: %w", err)
	}

	// Find all source files that might use translations
	sourceFiles, err := a.findSourceFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding source files: %w", err)
	}

	// Group translation files by locale
	localeFiles := a.groupTranslationFilesByLocale(translationFiles)

	// Analyze used translations (same across all locales)
	usedTranslations, err := a.analyzeUsedTranslations(sourceFiles)
	if err != nil {
		return nil, fmt.Errorf("error analyzing used translations: %w", err)
	}

	// Analyze each locale separately
	for locale, files := range localeFiles {
		localeResult, err := a.analyzeLocale(locale, files, usedTranslations)
		if err != nil {
			fmt.Printf("Warning: Error analyzing locale %s: %v\n", locale, err)
			continue
		}
		a.results.LocaleResults[locale] = localeResult
	}

	// Generate overall results
	a.generateOverallResults()

	return a.results, nil
}

// validateProjectPath checks if the project path exists and is a directory
func (a *Analyzer) validateProjectPath() error {
	info, err := os.Stat(a.projectPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory")
	}
	return nil
}

// findTranslationFiles finds all translation files in the project
func (a *Analyzer) findTranslationFiles() ([]string, error) {
	var files []string
	
	err := filepath.Walk(a.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			// Look for common translation file patterns
			ext := filepath.Ext(path)
			if ext == ".json" || ext == ".yaml" || ext == ".yml" {
				// Check if it's in a messages, locales, or i18n directory
				if strings.Contains(path, "messages") || strings.Contains(path, "locales") || strings.Contains(path, "i18n") {
					files = append(files, path)
				}
			}
		}
		return nil
	})
	
	return files, err
}

// findSourceFiles finds all source files that might use translations
func (a *Analyzer) findSourceFiles() ([]string, error) {
	var files []string
	
	err := filepath.Walk(a.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			ext := filepath.Ext(path)
			// Include common source file extensions
			if ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" {
				// Exclude node_modules and build directories
				if !strings.Contains(path, "node_modules") && !strings.Contains(path, ".next") {
					files = append(files, path)
				}
			}
		}
		return nil
	})
	
	return files, err
}

// analyzeDeclaredTranslations analyzes translation files to find declared translations
func (a *Analyzer) analyzeDeclaredTranslations(files []string) (map[string]Translation, error) {
	parser := NewTranslationParser()
	allDeclared := make(map[string]Translation)
	
	for _, file := range files {
		declared, err := parser.ParseTranslationFile(file)
		if err != nil {
			fmt.Printf("Warning: Could not parse translation file %s: %v\n", file, err)
			continue
		}
		
		// Merge with existing translations
		for key, translation := range declared {
			allDeclared[key] = translation
		}
	}
	
	return allDeclared, nil
}

// analyzeUsedTranslations analyzes source files to find used translations
func (a *Analyzer) analyzeUsedTranslations(files []string) (map[string]Translation, error) {
	parser := NewTranslationParser()
	allUsed := make(map[string]Translation)
	
	for _, file := range files {
		used, err := parser.ParseSourceFile(file)
		if err != nil {
			fmt.Printf("Warning: Could not parse source file %s: %v\n", file, err)
			continue
		}
		
		// Merge with existing translations
		for key, translation := range used {
			allUsed[key] = translation
		}
	}
	
	return allUsed, nil
}

// groupTranslationFilesByLocale groups translation files by their locale
func (a *Analyzer) groupTranslationFilesByLocale(files []string) map[string][]string {
	localeFiles := make(map[string][]string)
	
	for _, file := range files {
		locale := a.extractLocaleFromPath(file)
		if locale != "" {
			localeFiles[locale] = append(localeFiles[locale], file)
		}
	}
	
	return localeFiles
}

// extractLocaleFromPath extracts the locale from a file path
func (a *Analyzer) extractLocaleFromPath(filePath string) string {
	// Extract locale from filename (e.g., en.json -> en, de.json -> de)
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	if ext == ".json" || ext == ".yaml" || ext == ".yml" {
		locale := strings.TrimSuffix(fileName, ext)
		// Validate locale format (2-3 character language code)
		if len(locale) >= 2 && len(locale) <= 3 {
			return locale
		}
	}
	
	// Try to extract from directory structure (e.g., messages/en.json)
	dir := filepath.Dir(filePath)
	dirName := filepath.Base(dir)
	if len(dirName) >= 2 && len(dirName) <= 3 {
		return dirName
	}
	
	return ""
}

// analyzeLocale analyzes translations for a specific locale
func (a *Analyzer) analyzeLocale(locale string, files []string, usedTranslations map[string]Translation) (*LocaleAnalysisResult, error) {
	// Analyze declared translations for this locale
	declaredTranslations, err := a.analyzeDeclaredTranslations(files)
	if err != nil {
		return nil, fmt.Errorf("error analyzing declared translations for locale %s: %w", locale, err)
	}

	// Add locale information to declared translations
	for key, translation := range declaredTranslations {
		translation.Locale = locale
		declaredTranslations[key] = translation
	}

	// Create locale result
	localeResult := &LocaleAnalysisResult{
		Locale:                locale,
		UnusedTranslations:    make([]Translation, 0),
		UndeclaredTranslations: make([]Translation, 0),
	}

	// Find unused translations for this locale
	for key, translation := range declaredTranslations {
		if _, exists := usedTranslations[key]; !exists {
			localeResult.UnusedTranslations = append(localeResult.UnusedTranslations, translation)
		}
	}

	// Find undeclared translations for this locale
	for key, translation := range usedTranslations {
		if _, exists := declaredTranslations[key]; !exists {
			translation.Locale = locale
			translation.Declared = false
			localeResult.UndeclaredTranslations = append(localeResult.UndeclaredTranslations, translation)
		}
	}

	localeResult.TotalTranslations = len(declaredTranslations)
	localeResult.UsedTranslations = len(usedTranslations)

	return localeResult, nil
}

// generateOverallResults aggregates results from all locales
func (a *Analyzer) generateOverallResults() {
	allUnused := make([]Translation, 0)
	allUndeclared := make([]Translation, 0)
	totalTranslations := 0
	usedTranslations := 0

	for _, localeResult := range a.results.LocaleResults {
		allUnused = append(allUnused, localeResult.UnusedTranslations...)
		allUndeclared = append(allUndeclared, localeResult.UndeclaredTranslations...)
		totalTranslations += localeResult.TotalTranslations
		usedTranslations += localeResult.UsedTranslations
	}

	a.results.UnusedTranslations = allUnused
	a.results.UndeclaredTranslations = allUndeclared
	a.results.TotalTranslations = totalTranslations
	a.results.UsedTranslations = usedTranslations
} 