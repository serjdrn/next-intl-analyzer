package main

import (
	"fmt"
	"os"

	"next-intl-analyzer/cmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "next-intl-analyzer",
	Short: "A CLI tool to analyze next-intl translations",
	Long: `A command line tool to find unused and undeclared translations 
in Next.js projects using next-intl for internationalization.

Examples:
  next-intl-analyzer analyze ./my-nextjs-project
  next-intl-analyzer analyze /path/to/your/project`,
}

func main() {
	rootCmd.AddCommand(cmd.AnalyzeCmd)
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
} 