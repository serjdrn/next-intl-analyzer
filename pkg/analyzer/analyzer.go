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
	Locale   string
}

// AnalysisResult contains the results of the translation analysis
type AnalysisResult struct {
	UnusedTranslations    []Translation
	UndeclaredTranslations []Translation
	TotalTranslations     int
	UsedTranslations      int
	LocaleResults         map[string]*LocaleAnalysisResult
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

func (a *Analyzer) Analyze() (*AnalysisResult, error) {
	if err := a.validateProjectPath(); err != nil {
		return nil, fmt.Errorf("invalid project path: %w", err)
	}

	translationFiles, err := a.findTranslationFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding translation files: %w", err)
	}

	sourceFiles, err := a.findSourceFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding source files: %w", err)
	}

	localeFiles := a.groupTranslationFilesByLocale(translationFiles)

	usedTranslations, err := a.analyzeUsedTranslations(sourceFiles)
	if err != nil {
		return nil, fmt.Errorf("error analyzing used translations: %w", err)
	}

	for locale, files := range localeFiles {
		localeResult, err := a.analyzeLocale(locale, files, usedTranslations)
		if err != nil {
			fmt.Printf("Warning: Error analyzing locale %s: %v\n", locale, err)
			continue
		}
		a.results.LocaleResults[locale] = localeResult
	}

	a.generateOverallResults()

	return a.results, nil
}

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

func (a *Analyzer) findTranslationFiles() ([]string, error) {
	var files []string
	
	err := filepath.Walk(a.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".json" || ext == ".yaml" || ext == ".yml" {
				if strings.Contains(path, "messages") || strings.Contains(path, "locales") || strings.Contains(path, "i18n") {
					files = append(files, path)
				}
			}
		}
		return nil
	})
	
	return files, err
}

func (a *Analyzer) findSourceFiles() ([]string, error) {
	var files []string
	
	err := filepath.Walk(a.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" {
				if !strings.Contains(path, "node_modules") && !strings.Contains(path, ".next") {
					files = append(files, path)
				}
			}
		}
		return nil
	})
	
	return files, err
}

func (a *Analyzer) analyzeDeclaredTranslations(files []string) (map[string]Translation, error) {
	parser := NewTranslationParser()
	allDeclared := make(map[string]Translation)
	
	for _, file := range files {
		declared, err := parser.ParseTranslationFile(file)
		if err != nil {
			fmt.Printf("Warning: Could not parse translation file %s: %v\n", file, err)
			continue
		}
		
		for key, translation := range declared {
			allDeclared[key] = translation
		}
	}
	
	return allDeclared, nil
}

func (a *Analyzer) analyzeUsedTranslations(files []string) (map[string]Translation, error) {
	parser := NewTranslationParser()
	allUsed := make(map[string]Translation)
	
	for _, file := range files {
		used, err := parser.ParseSourceFile(file)
		if err != nil {
			fmt.Printf("Warning: Could not parse source file %s: %v\n", file, err)
			continue
		}
		
		for key, translation := range used {
			allUsed[key] = translation
		}
	}
	
	return allUsed, nil
}

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

func (a *Analyzer) extractLocaleFromPath(filePath string) string {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	if ext == ".json" || ext == ".yaml" || ext == ".yml" {
		locale := strings.TrimSuffix(fileName, ext)
		if len(locale) >= 2 && len(locale) <= 3 {
			return locale
		}
	}
	
	dir := filepath.Dir(filePath)
	dirName := filepath.Base(dir)
	if len(dirName) >= 2 && len(dirName) <= 3 {
		return dirName
	}
	
	return ""
}

func (a *Analyzer) analyzeLocale(locale string, files []string, usedTranslations map[string]Translation) (*LocaleAnalysisResult, error) {
	declaredTranslations, err := a.analyzeDeclaredTranslations(files)
	if err != nil {
		return nil, fmt.Errorf("error analyzing declared translations for locale %s: %w", locale, err)
	}

	for key, translation := range declaredTranslations {
		translation.Locale = locale
		declaredTranslations[key] = translation
	}

	localeResult := &LocaleAnalysisResult{
		Locale:                locale,
		UnusedTranslations:    make([]Translation, 0),
		UndeclaredTranslations: make([]Translation, 0),
	}

	for key, translation := range declaredTranslations {
		if _, exists := usedTranslations[key]; !exists {
			localeResult.UnusedTranslations = append(localeResult.UnusedTranslations, translation)
		}
	}

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