package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type TranslationParser struct{}

func NewTranslationParser() *TranslationParser {
	return &TranslationParser{}
}

func (p *TranslationParser) ParseTranslationFile(filePath string) (map[string]Translation, error) {
	declared := make(map[string]Translation)
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("error parsing JSON in %s: %w", filePath, err)
	}
	
	keys := p.extractKeys(data, "")
	
	for _, key := range keys {
		declared[key] = Translation{
			Key:      key,
			File:     filePath,
			Line:     0,
			Used:     false,
			Declared: true,
		}
	}
	
	return declared, nil
}

func (p *TranslationParser) extractKeys(data interface{}, prefix string) []string {
	var keys []string
	
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentKey := key
			if prefix != "" {
				currentKey = prefix + "." + key
			}
			
			keys = append(keys, currentKey)
			
			if nested, ok := value.(map[string]interface{}); ok {
				keys = append(keys, p.extractKeys(nested, currentKey)...)
			}
		}
	case []interface{}:
		for i, item := range v {
			if nested, ok := item.(map[string]interface{}); ok {
				currentKey := fmt.Sprintf("%s[%d]", prefix, i)
				keys = append(keys, p.extractKeys(nested, currentKey)...)
			}
		}
	}
	
	return keys
}

func (p *TranslationParser) ParseSourceFile(filePath string) (map[string]Translation, error) {
	used := make(map[string]Translation)
	untranslated := make(map[string]Translation)
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	
	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")
	
	currentNamespace := ""
	
	for lineNum, line := range lines {
		lineNum++ // Convert to 1-based line numbers
		
		useTranslationsMatch := regexp.MustCompile(`useTranslations\(['"]([^'"]+)['"]\)`).FindStringSubmatch(line)
		if len(useTranslationsMatch) > 1 {
			currentNamespace = useTranslationsMatch[1]
			continue
		}
		
		getTranslationsMatch := regexp.MustCompile(`getTranslations\(['"]([^'"]+)['"]\)`).FindStringSubmatch(line)
		if len(getTranslationsMatch) > 1 {
			currentNamespace = getTranslationsMatch[1]
			continue
		}
		
		tCalls := regexp.MustCompile(`t\(['"]([^'"]+)['"]`).FindAllStringSubmatch(line, -1)
		for _, match := range tCalls {
			if len(match) > 1 {
				key := match[1]
				
				if key == "" {
					continue
				}
				
				fullKey := key
				if currentNamespace != "" {
					fullKey = currentNamespace + "." + key
				}
				
				used[fullKey] = Translation{
					Key:      fullKey,
					File:     filePath,
					Line:     lineNum,
					Used:     true,
					Declared: false,
					Type:     "translation_call",
				}
			}
		}
		
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
				fullKey := key
				if currentNamespace != "" {
					fullKey = currentNamespace + "." + key
				}
				
				used[fullKey] = Translation{
							Key:      fullKey,
							File:     filePath,
							Line:     lineNum,
							Used:     true,
							Declared: false,
							Type:     "translation_call",
						}
					}
				}
			}
		}
		
		hardcodedPatterns := []*regexp.Regexp{
			// Text between JSX tags like <h1>text</h1> (but not translation calls)
			regexp.MustCompile(`<[^>]+>([^<>{}\n]+[a-zA-Z][^<>{}\n]*)</[^>]+>`),
			// Text in JSX expressions like {text} (but not translation calls)
			regexp.MustCompile(`\{([^}]*[a-zA-Z][^}]*)\}`),
			// Text in JSX attributes like title="text" (but not href, src, etc.)
			regexp.MustCompile(`["']([^"']*[a-zA-Z][^"']*)["']`),
			// Text after JSX tags like <button>text
			regexp.MustCompile(`>([^<>{}\n]+[a-zA-Z][^<>{}\n]*)<`),
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
					
					// Skip if this looks like a translation call pattern (contains quotes and parentheses)
					if strings.Contains(text, "(") && strings.Contains(text, "'") {
						continue
					}
					
					// Skip if this looks like a translation key (contains dots and is short, or matches common key patterns)
					if (strings.Contains(text, ".") && len(text) < 20) || 
					   (len(text) < 20 && !strings.Contains(text, " ") && 
					    (strings.Contains(text, "button") || strings.Contains(text, "navigation") || 
					     strings.Contains(text, "title") || strings.Contains(text, "welcome") || 
					     strings.Contains(text, "about") || strings.Contains(text, "description") ||
					     strings.Contains(text, "undeclaredKey"))) {
						continue
					}
					
					// Skip imports and function names (including destructured imports)
					if strings.Contains(text, "import") || strings.Contains(text, "export") || 
					   strings.Contains(text, "function") || strings.Contains(text, "const") {
						continue
					}
					
					// Skip any destructured import pattern (e.g. { useState, useEffect, useTranslations })
					if strings.HasPrefix(strings.TrimSpace(text), "{") && strings.HasSuffix(strings.TrimSpace(text), "}") {
						continue
					}
					
					// Skip any identifier that looks like a React hook or translation function
					if strings.HasPrefix(text, "use") || strings.HasPrefix(text, "get") {
						continue
					}
					
					// Skip URLs and paths
					if strings.HasPrefix(text, "/") || strings.Contains(text, "http") || 
					   strings.Contains(text, "www.") || strings.Contains(text, ".com") {
						continue
					}
					
					// Skip technical patterns using the constants.go definitions
					isTechnical := false
					for _, pattern := range TechnicalPatterns {
						if strings.Contains(text, pattern) {
							isTechnical = true
							break
						}
					}
					if isTechnical {
						continue
					}
					
					// Skip comments
					if strings.HasPrefix(text, "/*") || strings.HasPrefix(text, "//") || 
					   strings.Contains(text, "*/") {
						continue
					}
					
					// Skip translation key patterns (common patterns that look like translation keys)
					if (strings.Contains(text, "button.") || strings.Contains(text, "navigation.") || 
						strings.Contains(text, "title") || strings.Contains(text, "welcome") || 
						strings.Contains(text, "about") || strings.Contains(text, "description")) && 
						len(text) < 30 && !strings.Contains(text, " ") {
						continue
					}
					
					// Skip if it's just whitespace or very short
					if len(text) < MinTextLength || strings.TrimSpace(text) == "" {
						continue
					}
					
					// Skip if it's a variable or expression
					if strings.Contains(text, "=") || strings.Contains(text, "+") || strings.Contains(text, "-") {
						continue
					}
					
					// Skip if it's a number or mostly numeric
					if p.isNumeric(text) {
						continue
					}
					
					// Skip if it's a single character (except common punctuation)
					if len(text) == 1 {
						isPunctuation := false
						for _, punct := range CommonPunctuation {
							if text == punct {
								isPunctuation = true
								break
							}
						}
						if !isPunctuation {
							continue
						}
					}
					
					if p.isUserFacingText(text) {
						untranslated[text] = Translation{
							Key:      text,
							File:     filePath,
							Line:     lineNum,
							Used:     true,
							Declared: false,
							Type:     "hardcoded_string",
						}
					}
				}
			}
		}
	}
	
	for key, translation := range untranslated {
		used[key] = translation
	}
	
	return used, nil
}

func (p *TranslationParser) isUserFacingText(text string) bool {
	// Handle very short but common UI strings
	if len(text) >= 2 && len(text) <= 4 {
		for _, word := range ShortUIWords {
			if strings.EqualFold(text, word) {
				return true
			}
		}
	}

	// Minimum length check
	if len(text) < MinTextLength {
		return false
	}

	// Check for technical patterns
	for _, pattern := range TechnicalPatterns {
		if strings.Contains(text, pattern) {
			return false
		}
	}

	// Check for common UI/UX patterns that indicate user-facing content
	for _, pattern := range UIPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}

	// Check for sentence-like patterns (starts with capital, ends with punctuation)
	if len(text) > 3 {
		firstChar := text[0]
		lastChar := text[len(text)-1]
		if (firstChar >= 'A' && firstChar <= 'Z') && 
		   (lastChar == '.' || lastChar == '!' || lastChar == '?') {
			return true
		}
	}

	// Enhanced alphabetic ratio check with better thresholds
	alphaCount := 0
	spaceCount := 0
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			alphaCount++
		} else if char == ' ' {
			spaceCount++
		}
	}

	// For longer text, require more alphabetic characters
	if len(text) > LongTextThreshold {
		return alphaCount >= int(float64(len(text))*LongTextRatio)
	}
	
	// For medium text, standard ratio
	if len(text) > MediumTextThreshold {
		return alphaCount >= int(float64(len(text))*MediumTextRatio)
	}
	
	// For short text, be more lenient but still require majority alphabetic
	return alphaCount >= int(float64(len(text))*ShortTextRatio)
}

// isNumeric checks if a string is mostly numeric
func (p *TranslationParser) isNumeric(text string) bool {
	digitCount := 0
	for _, char := range text {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}
	return digitCount >= len(text)/2
}

func (p *TranslationParser) MergeTranslationMaps(maps ...map[string]Translation) map[string]Translation {
	result := make(map[string]Translation)
	
	for _, m := range maps {
		for key, translation := range m {
			result[key] = translation
		}
	}
	
	return result
} 