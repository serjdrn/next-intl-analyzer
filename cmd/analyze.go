package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"next-intl-analyzer/pkg/analyzer"

	"github.com/spf13/cobra"
)

var AnalyzeCmd = &cobra.Command{
	Use:   "analyze [project-path]",
	Short: "Analyze next-intl translations in a project",
	Long: `Analyze a Next.js project to find unused and undeclared translations.
	
This command will:
- Scan for translation files strictly from messages/*.json
- Scan source files (.jsx and .tsx only) for translation usage
- Report unused translations (declared but not used)
- Report undeclared translations (used but not declared)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath := args[0]
		
		// Show progress indicator
		quiet, _ := cmd.Flags().GetBool("quiet")
		if !quiet {
			fmt.Println("🔍 Analyzing project...")
			fmt.Println("  ↳ Scanning files...")
		}
		
		analyzer := analyzer.NewAnalyzer(projectPath)
		
		// Add progress callback with spinner
		spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinIdx := 0
		lastStage := ""
		
		analyzer.SetProgressCallback(func(stage string, progress int, total int) {
			if !quiet {
				if stage != lastStage {
					if lastStage != "" {
						fmt.Println()
					}
					lastStage = stage
				}
				
				spinChar := spinChars[spinIdx%len(spinChars)]
				spinIdx++
				
				if total > 0 {
					percent := int((float64(progress) / float64(total)) * 100)
					fmt.Printf("\r  %s %s... %d/%d (%d%%)   ", spinChar, stage, progress, total, percent)
				} else {
					fmt.Printf("\r  %s %s...   ", spinChar, stage)
				}
			}
		})
		
		results, err := analyzer.Analyze()
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}
		
		if !quiet {
			fmt.Println("\r  ↳ Analysis complete!                      ")
			fmt.Println()
		}
		
		// Generate markdown report if flag is set (do this before display to avoid os.Exit)
		generateReport, _ := cmd.Flags().GetBool("report")
		if generateReport {
			reportFile, _ := cmd.Flags().GetString("report-file")
			if err := generateMarkdownReport(results, projectPath, reportFile, cmd); err != nil {
				return fmt.Errorf("failed to generate report: %w", err)
			}
		}
		
		// Display results unless quiet mode is enabled
		if !quiet {
			displayResults(results)
		}
		
		return nil
	},
}

func init() {
	AnalyzeCmd.Flags().Bool("report", false, "Generate a markdown report file")
	AnalyzeCmd.Flags().String("report-file", "translations-report.md", "Custom filename for the markdown report (will be placed in reports/ folder)")
	AnalyzeCmd.Flags().Bool("quiet", false, "Suppress console output (useful when generating reports)")
}



func displayResults(results *analyzer.AnalysisResult) {
	fmt.Println("=== Next-intl Translation Analysis ===")
	fmt.Println()
	
	fmt.Printf("📊 Overall Summary:\n")
	fmt.Printf("   Total translations: %d\n", results.TotalTranslations)
	fmt.Printf("   Used translations: %d\n", results.UsedTranslations)
	fmt.Printf("   Unused translations: %d\n", len(results.UnusedTranslations))
	fmt.Printf("   Undeclared translations: %d\n", len(results.UndeclaredTranslations))
	fmt.Printf("   Hardcoded strings: %d\n", len(results.HardcodedStrings))
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
			
			// Removed per-locale hardcoded strings section as they are now handled globally
			
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
	
	if len(results.HardcodedStrings) > 0 {
		fmt.Printf("🔤 Hardcoded strings (%d):\n", len(results.HardcodedStrings))
		for _, translation := range results.HardcodedStrings {
			fmt.Printf("   - %s (used in %s:%d)\n", translation.Key, translation.File, translation.Line)
		}
		fmt.Println()
	} else {
		fmt.Println("✅ No hardcoded strings found!")
		fmt.Println()
	}
	
	if len(results.UnusedTranslations) > 0 || len(results.UndeclaredTranslations) > 0 || len(results.HardcodedStrings) > 0 {
		os.Exit(1)
	}
}

// generateMarkdownReport creates a detailed markdown report of the analysis results
func generateMarkdownReport(results *analyzer.AnalysisResult, projectPath string, reportPath string, cmd *cobra.Command) error {
	// Create reports directory in the project path
	reportsDir := filepath.Join(projectPath, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}
	
	// Create the full path for the report file
	fullReportPath := filepath.Join(reportsDir, reportPath)
	
	content := fmt.Sprintf(`# Next-intl Translation Analysis Report

**Generated:** %s  
**Project:** %s

## 📊 Summary

| Metric | Count |
|--------|-------|
| Total Translations | %d |
| Used Translations | %d |
| Unused Translations | %d |
| Undeclared Translations | %d |
| Hardcoded Strings | %d |
| Locales Analyzed | %d |

## 🌍 Per-locale Analysis

`, time.Now().Format("2006-01-02 15:04:05"), projectPath, results.TotalTranslations, results.UsedTranslations, len(results.UnusedTranslations), len(results.UndeclaredTranslations), len(results.HardcodedStrings), len(results.LocaleResults))

	// Add per-locale results
	for locale, localeResult := range results.LocaleResults {
		content += fmt.Sprintf("### 📍 %s\n\n", strings.ToUpper(locale))
		content += fmt.Sprintf("| Metric | Count |\n")
		content += fmt.Sprintf("|--------|-------|\n")
		content += fmt.Sprintf("| Total Translations | %d |\n", localeResult.TotalTranslations)
		content += fmt.Sprintf("| Used Translations | %d |\n", localeResult.UsedTranslations)
		content += fmt.Sprintf("| Unused Translations | %d |\n", len(localeResult.UnusedTranslations))
		content += fmt.Sprintf("| Undeclared Translations | %d |\n\n", len(localeResult.UndeclaredTranslations))

		// Add unused translations for this locale
		if len(localeResult.UnusedTranslations) > 0 {
			content += fmt.Sprintf("#### ❌ Unused Translations in %s\n\n", strings.ToUpper(locale))
			content += "| Key | File |\n"
			content += "|-----|------|\n"
			for _, translation := range localeResult.UnusedTranslations {
				content += fmt.Sprintf("| `%s` | `%s` |\n", translation.Key, translation.File)
			}
			content += "\n"
		}

		// Add undeclared translations for this locale
		if len(localeResult.UndeclaredTranslations) > 0 {
			content += fmt.Sprintf("#### ⚠️ Undeclared Translations in %s\n\n", strings.ToUpper(locale))
			content += "| Key | File | Line |\n"
			content += "|-----|------|------|\n"
			for _, translation := range localeResult.UndeclaredTranslations {
				content += fmt.Sprintf("| `%s` | `%s` | %d |\n", translation.Key, translation.File, translation.Line)
			}
			content += "\n"
		}
		
		// Removed per-locale hardcoded strings section as they are now handled globally
	}

	// Add overall unused translations
	if len(results.UnusedTranslations) > 0 {
		content += "## ❌ Overall Unused Translations\n\n"
		content += "| Key | File | Locale |\n"
		content += "|-----|------|--------|\n"
		for _, translation := range results.UnusedTranslations {
			content += fmt.Sprintf("| `%s` | `%s` | %s |\n", translation.Key, translation.File, translation.Locale)
		}
		content += "\n"
	} else {
		content += "## ✅ No Unused Translations Found\n\n"
	}

	// Add overall undeclared translations
	if len(results.UndeclaredTranslations) > 0 {
		content += "## ⚠️ Overall Undeclared Translations\n\n"
		content += "| Key | File | Line | Locale |\n"
		content += "|-----|------|------|--------|\n"
		for _, translation := range results.UndeclaredTranslations {
			content += fmt.Sprintf("| `%s` | `%s` | %d | %s |\n", translation.Key, translation.File, translation.Line, translation.Locale)
		}
		content += "\n"
	} else {
		content += "## ✅ No Undeclared Translations Found\n\n"
	}
	
	// Add overall hardcoded strings
	if len(results.HardcodedStrings) > 0 {
		content += "## 🔤 Hardcoded Strings\n\n"
		content += "| Text | File | Line |\n"
		content += "|------|------|------|\n"
		for _, translation := range results.HardcodedStrings {
			content += fmt.Sprintf("| `%s` | `%s` | %d |\n", translation.Key, translation.File, translation.Line)
		}
		content += "\n"
	} else {
		content += "## ✅ No Hardcoded Strings Found\n\n"
	}

	// Add recommendations
	content += `## 💡 Recommendations

### For Unused Translations:
- Review and remove unused translation keys from your translation files
- Consider if these translations might be used in the future
- Use this list to clean up your translation files

### For Undeclared Translations:
- Add missing translation keys to your translation files
- Ensure all user-facing text is properly internationalized
- Consider using translation keys instead of hardcoded strings

### For Hardcoded Strings:
- Replace hardcoded strings with translation keys
- Create appropriate entries in your translation files
- Use the t() function or appropriate hooks to translate these strings

### Best Practices:
- Regularly run this analysis to maintain clean translation files
- Use consistent naming conventions for translation keys
- Consider implementing automated checks in your CI/CD pipeline
- Avoid hardcoding user-facing text directly in components

---

*Report generated by next-intl-analyzer*
`

	// Write the report to file
	if err := os.WriteFile(fullReportPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	// Only print success message if not in quiet mode
	quiet, _ := cmd.Flags().GetBool("quiet")
	if !quiet {
		fmt.Printf("📄 Markdown report generated: %s\n", fullReportPath)
	}
	return nil
} 