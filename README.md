# Next-intl Analyzer

A command-line tool to analyze next-intl translations in Next.js projects. This tool helps you find unused and undeclared translations to keep your internationalization files clean and maintainable.

> **Disclaimer**: This tool is **not** officially affiliated with or endorsed by the `next-intl` team. It is an independent analysis tool designed to help developers working with the `next-intl` library. See [DISCLAIMER.md](DISCLAIMER.md) for more details.

## Features

- 🔍 **Find unused translations**: Identify translation keys that are declared but never used in your code
- ⚠️ **Find undeclared translations**: Detect translation keys that are used in code but not declared in translation files
- 🔤 **Detect hardcoded strings**: Find user-facing text that should be translated
- 📊 **Comprehensive reporting**: Get detailed reports with file locations and line numbers
- 🚀 **Fast analysis**: Efficient scanning of your entire project
- 🎯 **Next.js optimized**: Specifically designed for Next.js projects using next-intl

## Installation

### Prerequisites

- Go 1.21 or higher

### Option 1: Run directly with Go (Recommended for development)

```bash
git clone <repository-url>
cd next-intl-analyzer
go mod tidy

# Run directly without building
go run main.go analyze /path/to/your/project
```

### Option 2: Build and run the binary

```bash
git clone <repository-url>
cd next-intl-analyzer
go mod tidy
go build -o next-intl-analyzer

# Run the binary
./next-intl-analyzer analyze /path/to/your/project
```

### Option 3: Install globally

```bash
git clone <repository-url>
cd next-intl-analyzer
go mod tidy
go install .

# Now you can run it from anywhere
next-intl-analyzer analyze /path/to/your/project
```

## Usage

### Basic usage

```bash
# If running directly with Go
go run main.go analyze /path/to/your/nextjs-project

# If you built the binary
./next-intl-analyzer analyze /path/to/your/nextjs-project

# If installed globally
next-intl-analyzer analyze /path/to/your/nextjs-project
```

### Examples

```bash
# Analyze current directory
go run main.go analyze .

# Analyze specific project
go run main.go analyze /Users/username/projects/my-app

# Analyze with relative path
go run main.go analyze ../my-nextjs-app

# Test with the included test data
go run main.go analyze test-data

# Generate a markdown report
go run main.go analyze test-data --report

# Generate a markdown report with custom filename
go run main.go analyze test-data --report --report-file my-analysis-report.md

# Generate a report without console output
go run main.go analyze test-data --report --quiet

## CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--report` | Generate a markdown report file | `false` |
| `--report-file` | Custom filename for the markdown report | `next-intl-analysis-report.md` |
| `--quiet` | Suppress console output (useful when generating reports) | `false` |

## Report Generation

The tool can generate a detailed markdown report of the analysis results:

```bash
# Generate a report with default filename
go run main.go analyze /path/to/your/project --report

# Generate a report with custom filename
go run main.go analyze /path/to/your/project --report --report-file my-report.md
```

The report includes:
- 📊 **Summary statistics** with counts of total, used, unused, undeclared translations, and hardcoded strings
- 🌍 **Per-locale analysis** showing results for each language
- ❌ **Unused translations** with file locations
- ⚠️ **Undeclared translations** with file locations and line numbers
- 🔤 **Hardcoded strings** with file locations and line numbers
- 💡 **Recommendations** for maintaining clean translation files

See [sample-report.md](sample-report.md) for an example of the generated report format.

## How it works

The CLI tool performs the following analysis:

1. **Scans translation files**: Looks for translation files in common locations:
   - `messages/*.json`
   - `locales/*.json`
   - `i18n/*.json`
   - Any `.json`, `.yaml`, or `.yml` files in translation directories

2. **Scans source files**: Analyzes JavaScript/TypeScript files for translation usage:
   - **Client-side patterns**: `useTranslations('namespace')` and `t('key')` calls
   - **Server-side patterns**: `getTranslations('namespace')` calls
   - **Advanced patterns**: `t.rich()`, `t.markup()`, `t.raw()`, `t.has()` calls
   - **Namespace context**: Automatically builds full key paths (e.g., `HomePage.title`)

3. **Compares and reports**: 
   - Identifies unused translations (declared but not used)
   - Finds undeclared translations (used but not declared)
   - Detects hardcoded strings (user-facing text that should be translated)
   - Provides detailed reports with file locations and line numbers

## Output

The tool provides a comprehensive report including:

- 📊 **Summary statistics**: Total, used, unused, undeclared translations, and hardcoded string counts
- ❌ **Unused translations**: List of translation keys that are declared but never used
- ⚠️ **Undeclared translations**: List of translation keys used in code but not declared
- 🔤 **Hardcoded strings**: List of user-facing text that should be translated
- 📁 **File locations**: Exact file paths and line numbers for each issue

### Example output

```
=== Next-intl Translation Analysis ===

📊 Overall Summary:
   Total translations: 28
   Used translations: 12
   Unused translations: 16
   Undeclared translations: 1
   Hardcoded strings: 2
   Locales analyzed: 2

🌍 Per-locale Analysis:

   📍 EN:
      Total translations: 28
      Used translations: 12
      Unused translations: 16
      Undeclared translations: 1
      Hardcoded strings: 2

❌ Unused translations (16):
   - Common.button.delete (in messages/en.json)
   - Common.navigation.contact (in messages/en.json)
   - Errors.notFound (in messages/en.json)
   - Errors.serverError (in messages/en.json)
   - Metadata.title (in messages/en.json)
   - Metadata.description (in messages/en.json)
   - Layout.language (in messages/en.json)
   - Layout.switchLocale (in messages/en.json)

⚠️  Undeclared translations (1):
   - About.undeclaredKey (used in src/components/ServerComponent.tsx:11)

🔤 Hardcoded strings (2):
   - Welcome to our application (used in src/components/UntranslatedComponent.tsx:13)
   - About Us (used in src/components/UntranslatedComponent.tsx:20)
```

## Exit codes

- `0`: Analysis completed successfully with no issues found
- `1`: Analysis completed but found unused translations, undeclared translations, or hardcoded strings
- `2`: Error occurred during analysis

## Supported file types

### Translation files
- `.json` files in translation directories (`messages/`, `locales/`, `i18n/`)
- `.yaml` and `.yml` files in translation directories

### Source files
- `.js` and `.jsx` files
- `.ts` and `.tsx` files
- Excludes `node_modules` and `.next` directories

## Supported translation patterns

### Client-side patterns
```typescript
// Basic usage
const t = useTranslations('HomePage');
t('title')                    // Detected as HomePage.title
t('welcome')                  // Detected as HomePage.welcome

// Advanced usage
t.rich('message', {...})      // Rich text formatting
t.markup('content', {...})    // HTML markup
t.raw('content')              // Raw content
t.has('key')                  // Optional message checks
```

### Server-side patterns
```typescript
// Server components
const t = await getTranslations('About');
t('title')                    // Detected as About.title
t('description')              // Detected as About.description
```

### Nested key patterns
```typescript
// Dot notation (fully qualified keys)
t('Common.button.save')       // Detected as Common.button.save
t('About.title')              // Detected as About.title
```

## Development

### Project structure

```
next-intl-analyzer/
├── main.go                   # CLI entry point
├── cmd/
│   └── analyze.go           # Analyze command implementation
├── pkg/
│   └── analyzer/
│       ├── analyzer.go      # Core analysis logic
│       ├── parser.go        # Translation file and source code parsing
│       └── constants.go     # Constants for text analysis
├── test-data/               # Test files for development
├── reports/                 # Generated reports directory
├── go.mod                   # Go module file
└── README.md                # This file
```

### Building

```bash
go build -o next-intl-analyzer
```

### Running tests

```bash
go test ./...
```

### Testing with sample data

```bash
go run main.go analyze test-data
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Related

- [next-intl](https://next-intl.dev/) - The internationalization library this tool analyzes (created by [Creator](https://github.com/amannn))
- [Next.js](https://nextjs.org/) - The React framework this tool is designed for

## Legal

- [DISCLAIMER.md](DISCLAIMER.md) - Important legal disclaimers and third-party attributions
- [LICENSE](LICENSE) - This project's license 