package cli

import (
	"fmt"
	"io"
)

const Version = "0.1.0-dev"

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "marvin %s\n", Version)
		return 0
	case "analyze":
		return runAnalyze(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "analyze requires a Cost Explorer CSV path")
		return 2
	}

	fmt.Fprintf(stdout, "analysis for %q is not implemented yet\n", args[0])
	return 1
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `Marvin reads exported AWS Cost Explorer CSV files and reports cost changes.

Usage:
  marvin analyze <cost-explorer.csv>
  marvin version
  marvin help

Status:
  CSV MVP scaffold. Cost analysis is planned for v0.1.
`)
}
