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
		
		hardcodedPatterns := []*regexp.Regexp{
			regexp.MustCompile(`<[^>]+>([^<>{}\n]+[a-zA-Z][^<>{}\n]*)</[^>]+>`),
			regexp.MustCompile(`\{([^}]*[a-zA-Z][^}]*)\}`),
		}
		
		for _, pattern := range hardcodedPatterns {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					text := strings.TrimSpace(match[1])
					
					if strings.Contains(text, "t(") || strings.Contains(text, "<") || strings.Contains(text, ">") {
						continue
					}
					
					if len(text) < 3 || strings.TrimSpace(text) == "" {
						continue
					}
					
					if strings.Contains(text, "=") || strings.Contains(text, "+") || strings.Contains(text, "-") {
						continue
					}
					
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
	
	for key, translation := range untranslated {
		used[key] = translation
	}
	
	return used, nil
}

func (p *TranslationParser) isUserFacingText(text string) bool {
	if len(text) < 3 {
		return false
	}
	
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
	
	alphaCount := 0
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			alphaCount++
		}
	}
	
	return alphaCount >= len(text)/2
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