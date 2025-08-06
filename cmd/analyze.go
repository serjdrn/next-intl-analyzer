package cmd

import (
	"fmt"
	"os"
	"strings"

	"next-intl-analyzer/pkg/analyzer"

	"github.com/spf13/cobra"
)

var AnalyzeCmd = &cobra.Command{
	Use:   "analyze [project-path]",
	Short: "Analyze next-intl translations in a project",
	Long: `Analyze a Next.js project to find unused and undeclared translations.
	
This command will:
- Scan for translation files (messages/*.json, locales/*.json, etc.)
- Scan source files for translation usage
- Report unused translations (declared but not used)
- Report undeclared translations (used but not declared)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath := args[0]
		
		analyzer := analyzer.NewAnalyzer(projectPath)
		
		results, err := analyzer.Analyze()
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}
		
		displayResults(results)
		
		return nil
	},
}



func displayResults(results *analyzer.AnalysisResult) {
	fmt.Println("=== Next-intl Translation Analysis ===")
	fmt.Println()
	
	fmt.Printf("📊 Overall Summary:\n")
	fmt.Printf("   Total translations: %d\n", results.TotalTranslations)
	fmt.Printf("   Used translations: %d\n", results.UsedTranslations)
	fmt.Printf("   Unused translations: %d\n", len(results.UnusedTranslations))
	fmt.Printf("   Undeclared translations: %d\n", len(results.UndeclaredTranslations))
	fmt.Printf("   Locales analyzed: %d\n", len(results.LocaleResults))
	fmt.Println()
	
	if len(results.LocaleResults) > 0 {
		fmt.Println("🌍 Per-locale Analysis:")
		fmt.Println()
		
		for locale, localeResult := range results.LocaleResults {
			fmt.Printf("   📍 %s:\n", strings.ToUpper(locale))
			fmt.Printf("      Total translations: %d\n", localeResult.TotalTranslations)
			fmt.Printf("      Used translations: %d\n", localeResult.UsedTranslations)
			fmt.Printf("      Unused translations: %d\n", len(localeResult.UnusedTranslations))
			fmt.Printf("      Undeclared translations: %d\n", len(localeResult.UndeclaredTranslations))
			
			if len(localeResult.UnusedTranslations) > 0 {
				fmt.Printf("      ❌ Unused in %s:\n", strings.ToUpper(locale))
				for _, translation := range localeResult.UnusedTranslations {
					fmt.Printf("         - %s (in %s)\n", translation.Key, translation.File)
				}
			}
			
			if len(localeResult.UndeclaredTranslations) > 0 {
				fmt.Printf("      ⚠️  Undeclared in %s:\n", strings.ToUpper(locale))
				for _, translation := range localeResult.UndeclaredTranslations {
					fmt.Printf("         - %s (used in %s:%d)\n", translation.Key, translation.File, translation.Line)
				}
			}
			
			fmt.Println()
		}
	}
	
	if len(results.UnusedTranslations) > 0 {
		fmt.Printf("❌ Overall unused translations (%d):\n", len(results.UnusedTranslations))
		for _, translation := range results.UnusedTranslations {
			fmt.Printf("   - %s (in %s, locale: %s)\n", translation.Key, translation.File, translation.Locale)
		}
		fmt.Println()
	} else {
		fmt.Println("✅ No unused translations found!")
		fmt.Println()
	}
	
	if len(results.UndeclaredTranslations) > 0 {
		fmt.Printf("⚠️  Overall undeclared translations (%d):\n", len(results.UndeclaredTranslations))
		for _, translation := range results.UndeclaredTranslations {
			fmt.Printf("   - %s (used in %s:%d, locale: %s)\n", translation.Key, translation.File, translation.Line, translation.Locale)
		}
		fmt.Println()
	} else {
		fmt.Println("✅ No undeclared translations found!")
		fmt.Println()
	}
	
	if len(results.UnusedTranslations) > 0 || len(results.UndeclaredTranslations) > 0 {
		os.Exit(1)
	}
} 