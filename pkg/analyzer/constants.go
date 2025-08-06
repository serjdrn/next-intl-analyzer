package analyzer

// Common short UI words that should be translated
var ShortUIWords = []string{
	"OK", "No", "Yes", "Cancel", "Save", "Edit", "Delete", "Add", "New",
	"Back", "Next", "Prev", "Close", "Open", "Help", "Info", "Error",
	"Loading", "Done", "Submit", "Reset", "Clear", "Search", "Filter",
	"Sort", "View", "Hide", "Show", "More", "Less", "All", "None",
}

// Technical patterns that indicate code/technical content (should NOT be translated)
var TechnicalPatterns = []string{
	// HTML/JSX attributes
	"className", "id=", "href=", "src=", "alt=", "title=", "type=", "value=",
	"placeholder=", "aria-", "data-", "role=", "tabindex=", "disabled=",
	"readonly=", "required=", "maxlength=", "minlength=", "pattern=",
	
	// React/JSX patterns
	"onClick", "onChange", "onSubmit", "onLoad", "onBlur", "onFocus", "onKey",
	"useState", "useEffect", "useCallback", "useMemo", "useRef", "useContext",
	"import", "export", "const", "let", "var", "function", "return", "if", "else",
	"for", "while", "switch", "case", "default", "break", "continue",
	"true", "false", "null", "undefined", "NaN", "Infinity",
	
	// Common variable names and technical terms
	"props", "state", "ref", "key", "index", "item", "element", "component",
	"handler", "callback", "event", "e.target", "e.preventDefault",
	"children", "className=", "style=", "id=", "name=", "onClick=",
	"onChange=", "onSubmit=", "width=", "height=", "size=", "color=",
	
	// File extensions and paths
	".js", ".jsx", ".ts", ".tsx", ".css", ".scss", ".json", ".md",
	"http://", "https://", "www.", ".com", ".org", ".net",
	
	// Code patterns
	"console.log", "debugger", "throw", "catch", "finally", "try",
	"new ", "instanceof", "typeof", "delete", "in", "of",
	">", "</", "<", "/>", "={", "{}", "()", "[]",
	
	// JavaScript keywords
	"await", "async", "class", "super", "this", "prototype", "constructor",
	"extends", "implements", "interface", "enum", "public", "private", "protected",
	
	// Common code tokens
	"=>{", "()=>", "=>(", "=>{}", "...props", "...rest", "className={`", "className={",
}

// UI/UX patterns that indicate user-facing content (should be translated)
var UIPatterns = []string{
	"Welcome", "Hello", "Goodbye", "Thank you", "Please", "Sorry",
	"Success", "Error", "Warning", "Info", "Loading", "Processing",
	"Click", "Press", "Enter", "Select", "Choose", "Browse",
	"Download", "Upload", "Share", "Like", "Follow", "Subscribe",
	"Sign in", "Sign up", "Log in", "Log out", "Register", "Login",
	"Profile", "Settings", "Preferences", "Account", "Dashboard",
	"Home", "About", "Contact", "Help", "Support", "FAQ",
	"Terms", "Privacy", "Policy", "License", "Copyright",
	"Created by", "Developed by", "Powered by", "Made with", "© ", "Copyright",
	"All rights reserved", "Submit", "Save", "Cancel", "Delete", "Edit", "Remove",
	"Create", "Update", "Read more", "Learn more", "View details", "Continue",
	"Previous", "Next", "Back to", "Return to", "Go to", "Navigate to",
}

// Common punctuation marks that might appear in user-facing text
var CommonPunctuation = []string{
	"!", "?", ".", ",", ":", ";", "-", "—", "–", "…",
}

// Minimum length for text to be considered for translation
const MinTextLength = 3

// Alphabetic ratio thresholds for different text lengths
const (
	LongTextThreshold   = 10  // Characters
	MediumTextThreshold = 5   // Characters
	LongTextRatio       = 0.6  // 60% alphabetic (lowered to reduce false positives)
	MediumTextRatio     = 0.6  // 60% alphabetic (increased to be more strict)
	ShortTextRatio      = 0.75 // 75% alphabetic (increased to be more strict)
	MinWordsForSentence = 3    // Minimum words for a phrase to be considered a sentence
) 