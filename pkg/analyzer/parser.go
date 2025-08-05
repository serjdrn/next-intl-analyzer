package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TranslationParser handles parsing of translation files and source code
type TranslationParser struct{}

// NewTranslationParser creates a new parser instance
func NewTranslationParser() *TranslationParser {
	return &TranslationParser{}
}

// ParseTranslationFile parses a JSON translation file and extracts all translation keys
func (p *TranslationParser) ParseTranslationFile(filePath string) (map[string]Translation, error) {
	declared := make(map[string]Translation)
	
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	
	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("error parsing JSON in %s: %w", filePath, err)
	}
	
	// Extract all translation keys recursively
	keys := p.extractKeys(data, "")
	
	// Create Translation objects
	for _, key := range keys {
		declared[key] = Translation{
			Key:      key,
			File:     filePath,
			Line:     0, // We'll need to implement line number detection if needed
			Used:     false,
			Declared: true,
		}
	}
	
	return declared, nil
}

// extractKeys recursively extracts all translation keys from a nested JSON object
func (p *TranslationParser) extractKeys(data interface{}, prefix string) []string {
	var keys []string
	
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentKey := key
			if prefix != "" {
				currentKey = prefix + "." + key
			}
			
			// Add the current key
			keys = append(keys, currentKey)
			
			// Recursively extract nested keys
			if nested, ok := value.(map[string]interface{}); ok {
				keys = append(keys, p.extractKeys(nested, currentKey)...)
			}
		}
	case []interface{}:
		// Handle arrays - we might want to skip these or handle them differently
		for i, item := range v {
			if nested, ok := item.(map[string]interface{}); ok {
				currentKey := fmt.Sprintf("%s[%d]", prefix, i)
				keys = append(keys, p.extractKeys(nested, currentKey)...)
			}
		}
	}
	
	return keys
}

// ParseSourceFile parses a source file to find translation usage patterns
func (p *TranslationParser) ParseSourceFile(filePath string) (map[string]Translation, error) {
	used := make(map[string]Translation)
	untranslated := make(map[string]Translation)
	
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	
	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")
	
	// Track current namespace context
	currentNamespace := ""
	
	// Find all translation usage patterns
	for lineNum, line := range lines {
		lineNum++ // Convert to 1-based line numbers
		
		// Check for useTranslations hook to set namespace context
		useTranslationsMatch := regexp.MustCompile(`useTranslations\(['"]([^'"]+)['"]\)`).FindStringSubmatch(line)
		if len(useTranslationsMatch) > 1 {
			currentNamespace = useTranslationsMatch[1]
			continue
		}
		
		// Check for getTranslations to set namespace context
		getTranslationsMatch := regexp.MustCompile(`getTranslations\(['"]([^'"]+)['"]\)`).FindStringSubmatch(line)
		if len(getTranslationsMatch) > 1 {
			currentNamespace = getTranslationsMatch[1]
			continue
		}
		
		// Find t() function calls with translation keys
		tCalls := regexp.MustCompile(`t\(['"]([^'"]+)['"]`).FindAllStringSubmatch(line, -1)
		for _, match := range tCalls {
			if len(match) > 1 {
				key := match[1]
				
				// Skip empty keys
				if key == "" {
					continue
				}
				
				// Build full key path if we have a namespace
				fullKey := key
				if currentNamespace != "" && !strings.Contains(key, ".") {
					fullKey = currentNamespace + "." + key
				}
				
				// Add the translation key
				used[fullKey] = Translation{
					Key:      fullKey,
					File:     filePath,
					Line:     lineNum,
					Used:     true,
					Declared: false,
				}
			}
		}
		
		// Find other t.*() function calls
		otherPatterns := []struct {
			name string
			regex *regexp.Regexp
		}{
			{"t.rich()", regexp.MustCompile(`t\.rich\(['"]([^'"]+)['"]`)},
			{"t.markup()", regexp.MustCompile(`t\.markup\(['"]([^'"]+)['"]`)},
			{"t.raw()", regexp.MustCompile(`t\.raw\(['"]([^'"]+)['"]`)},
			{"t.has()", regexp.MustCompile(`t\.has\(['"]([^'"]+)['"]`)},
		}
		
		for _, pattern := range otherPatterns {
			matches := pattern.regex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					key := match[1]
					if key != "" {
						// Build full key path if we have a namespace
						fullKey := key
						if currentNamespace != "" && !strings.Contains(key, ".") {
							fullKey = currentNamespace + "." + key
						}
						
						used[fullKey] = Translation{
							Key:      fullKey,
							File:     filePath,
							Line:     lineNum,
							Used:     true,
							Declared: false,
						}
					}
				}
			}
		}
		
		// Find hardcoded untranslated strings in JSX
		// Look for text content between JSX tags that might need translation
		hardcodedPatterns := []*regexp.Regexp{
			// Text between tags like <h1>text</h1>
			regexp.MustCompile(`<[^>]+>([^<>{}\n]+[a-zA-Z][^<>{}\n]*)</[^>]+>`),
			// Text in JSX expressions like {text}
			regexp.MustCompile(`\{([^}]*[a-zA-Z][^}]*)\}`),
		}
		
		for _, pattern := range hardcodedPatterns {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					text := strings.TrimSpace(match[1])
					
					// Skip if it's already a translation call or contains JSX
					if strings.Contains(text, "t(") || strings.Contains(text, "<") || strings.Contains(text, ">") {
						continue
					}
					
					// Skip if it's just whitespace or very short
					if len(text) < 3 || strings.TrimSpace(text) == "" {
						continue
					}
					
					// Skip if it's a variable or expression
					if strings.Contains(text, "=") || strings.Contains(text, "+") || strings.Contains(text, "-") {
						continue
					}
					
					// Only consider text that looks like user-facing content
					if p.isUserFacingText(text) {
						untranslated[text] = Translation{
							Key:      text,
							File:     filePath,
							Line:     lineNum,
							Used:     true,
							Declared: false,
						}
					}
				}
			}
		}
	}
	
	// Merge used and untranslated translations
	for key, translation := range untranslated {
		used[key] = translation
	}
	
	return used, nil
}

// isUserFacingText checks if a string looks like user-facing content that should be translated
func (p *TranslationParser) isUserFacingText(text string) bool {
	// Skip if it's just a variable name or technical content
	if len(text) < 3 {
		return false
	}
	
	// Skip if it contains technical patterns
	technicalPatterns := []string{
		"className", "id=", "href=", "src=", "alt=", "title=",
		"onClick", "onChange", "onSubmit", "onLoad",
		"useState", "useEffect", "useCallback", "useMemo",
		"import", "export", "const", "let", "var", "function",
		"return", "if", "else", "for", "while", "switch",
		"true", "false", "null", "undefined",
	}
	
	for _, pattern := range technicalPatterns {
		if strings.Contains(text, pattern) {
			return false
		}
	}
	
	// Skip if it's mostly symbols or numbers
	alphaCount := 0
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			alphaCount++
		}
	}
	
	if alphaCount < len(text)/2 {
		return false
	}
	
	// Consider it user-facing if it contains words and isn't technical
	return true
}

// MergeTranslationMaps merges multiple translation maps into one
func (p *TranslationParser) MergeTranslationMaps(maps ...map[string]Translation) map[string]Translation {
	result := make(map[string]Translation)
	
	for _, m := range maps {
		for key, translation := range m {
			result[key] = translation
		}
	}
	
	return result
} 