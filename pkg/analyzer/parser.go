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
	
	// Track variable assignments of translation functions
	// Map of variable name -> namespace
	translationVars := make(map[string]string)
	
	for lineNum, line := range lines {
		lineNum++ // Convert to 1-based line numbers
		
		// Match direct useTranslations calls
		useTranslationsMatch := regexp.MustCompile(`useTranslations\(['"]([^'"]+)['"]\)`).FindStringSubmatch(line)
		if len(useTranslationsMatch) > 1 {
			currentNamespace = useTranslationsMatch[1]
			continue
		}
		
		// Match direct getTranslations calls
		getTranslationsMatch := regexp.MustCompile(`getTranslations\(['"]([^'"]+)['"]\)`).FindStringSubmatch(line)
		if len(getTranslationsMatch) > 1 {
			currentNamespace = getTranslationsMatch[1]
			continue
		}
		
		// Match variable assignments for translation functions
		// Example: const adminT = getTranslations("Admin")
		// Example: const t = useTranslations("Common")
		varAssignmentPattern := regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*(?:useTranslations|getTranslations)\(['"]([^'"]+)['"]\)`)
		varMatch := varAssignmentPattern.FindStringSubmatch(line)
		if len(varMatch) > 2 {
			varName := varMatch[1]
			namespace := varMatch[2]
			translationVars[varName] = namespace
			continue
		}
		
		// Match destructured assignments
		// Example: const { t } = useTranslations("Common")
		destructuredPattern := regexp.MustCompile(`(?:const|let|var)\s+\{\s*(\w+)[^\}]*\}\s*=\s*(?:useTranslations|getTranslations)\(['"]([^'"]+)['"]\)`)
		destructuredMatch := destructuredPattern.FindStringSubmatch(line)
		if len(destructuredMatch) > 2 {
			varName := destructuredMatch[1]
			namespace := destructuredMatch[2]
			translationVars[varName] = namespace
			continue
		}
		
		// Process generic translation calls (standard pattern: t("key"))
		genericCallsPattern := regexp.MustCompile(`(\w+)\s*\(['"]([^'"]+)['"]\)`)
		genericCalls := genericCallsPattern.FindAllStringSubmatch(line, -1)
		for _, match := range genericCalls {
			if len(match) > 2 {
				varName := match[1]
				key := match[2]
				
				if key == "" {
					continue
				}
				
				// Check if this is a call to a known translation variable
				namespace := ""
				if varName == "t" {
					namespace = currentNamespace
				} else if ns, exists := translationVars[varName]; exists {
					namespace = ns
				} else {
					// Skip if not a translation function call
					continue
				}
				
				fullKey := key
				if namespace != "" && !strings.Contains(key, ".") {
					fullKey = namespace + "." + key
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
		
		// Process extended translation API calls (t.rich(), t.markup(), etc.)
		// This now handles both standard t and variable-based translation functions
		extendedApiPattern := regexp.MustCompile(`(\w+)\.(?:rich|markup|raw|has)\s*\(['"]([^'"]+)['"]\)`)
		extendedCalls := extendedApiPattern.FindAllStringSubmatch(line, -1)
		for _, match := range extendedCalls {
			if len(match) > 2 {
				varName := match[1]
				key := match[2]
				
				if key == "" {
					continue
				}
				
				// Check if this is a call to a known translation variable
				namespace := ""
				if varName == "t" {
					namespace = currentNamespace
				} else if ns, exists := translationVars[varName]; exists {
					namespace = ns
				} else {
					// Skip if not a translation function call
					continue
				}
				
				fullKey := key
				if namespace != "" && !strings.Contains(key, ".") {
					fullKey = namespace + "." + key
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
		
		hardcodedPatterns := []*regexp.Regexp{
			// Text between JSX tags that's likely user-facing (between opening/closing tags)
			// Example: <h1>Welcome to our site</h1>
			regexp.MustCompile(`<(?:h[1-6]|p|li|span|div|button|a|label|td|th)\b[^>]*>([^<>{}\n]+[a-zA-Z][^<>{}\n]*)</(?:h[1-6]|p|li|span|div|button|a|label|td|th)>`),
			
			// Text in specific JSX attributes that are likely to contain user-facing content
			// Example: title="Click here to continue"
			regexp.MustCompile(`(?:title|alt|placeholder|aria-label|description)=["']([^"'<>]{3,}[a-zA-Z][^"'<>]*)["']`),
			
			// Text between closing tag and opening tag that's not just whitespace
			// Example: </Button>Click me<Button>
			regexp.MustCompile(`>([^<>{}\n]{3,}[a-zA-Z][^<>{}\n]{3,})<`),
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
				
				// Skip single words that are likely variable names
				if !strings.Contains(text, " ") && len(text) < 15 {
					continue
				}
				
				// Skip words that are common in code but not necessarily user-facing
				if text == "name" || text == "value" || text == "data" || text == "type" || 
				   text == "label" || text == "content" || text == "format" || text == "default" {
					continue
				}
				
				// Skip strings that look like component props or configurations
				if strings.Count(text, "=") > 1 || strings.Count(text, ":") > 1 {
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
	// Handle very short but common UI strings from our predefined list
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
	
	// Skip strings that look like code
	if strings.ContainsAny(text, "{}[]()<>=+*/") {
		return false
	}
	
	// Skip camelCase or snake_case identifiers which are likely code
	if !strings.Contains(text, " ") && 
	   (strings.ContainsRune(text, '_') || 
		(strings.ToLower(text) != text && text[0] != strings.ToUpper(text)[0])) {
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

	// Count words - strings with multiple words are more likely to be user-facing
	words := strings.Fields(text)
	if len(words) >= MinWordsForSentence {
		// Check if it's a proper sentence (starts with capital, ends with punctuation)
		if len(text) > 3 {
			firstChar := text[0]
			lastChar := text[len(text)-1]
			if (firstChar >= 'A' && firstChar <= 'Z') && 
			   (lastChar == '.' || lastChar == '!' || lastChar == '?') {
				return true
			}
		}
		
		// Even without proper sentence structure, multi-word phrases are often translatable
		return true
	}

	// Enhanced alphabetic ratio check with adjusted thresholds
	alphaCount := 0
	spaceCount := 0
	punctCount := 0
	
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			alphaCount++
		} else if char == ' ' {
			spaceCount++
		} else if strings.ContainsRune(".,!?:;-—–…", char) {
			punctCount++
		}
	}
	
	// Skip if mostly non-alphabetic (excluding spaces and common punctuation)
	nonAlphaCount := len(text) - alphaCount - spaceCount - punctCount
	
	// For longer text, require more alphabetic characters
	if len(text) > LongTextThreshold {
		return alphaCount >= int(float64(len(text))*LongTextRatio) && 
		       len(words) >= MinWordsForSentence
	}
	
	// For medium text, be more strict
	if len(text) > MediumTextThreshold {
		return alphaCount >= int(float64(len(text))*MediumTextRatio) &&
			   nonAlphaCount < len(text)/3
	}
	
	// For short text, be very strict to avoid false positives
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