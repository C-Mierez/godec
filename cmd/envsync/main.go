package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/c-mierez/godec/internal/envsync"
)

type multiValue []string

func (m *multiValue) String() string {
	return fmt.Sprint([]string(*m))
}

func (m *multiValue) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "fix":
		os.Exit(runFix(os.Args[2:]))
	case "check":
		os.Exit(runCheck(os.Args[2:]))
	default:
		usage()
		os.Exit(2)
	}
}

func runFix(args []string) int {
	paths, err := parsePaths(args, []string{".env", ".env.example"})
	if err != nil {
		return 2
	}

	report, err := envsync.Fix(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := printFixReport(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func runCheck(args []string) int {
	paths, err := parsePaths(args, []string{".env", ".env.example"})
	if err != nil {
		return 2
	}

	if err := envsync.Check(paths); err != nil {
		printCheckError(err)
		return 1
	}

	if _, err := fmt.Fprintln(os.Stdout, "envsync check: ok"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func parsePaths(args []string, defaultFiles []string) (envsync.Paths, error) {
	flagSet := flag.NewFlagSet("envsync", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	configPath := flagSet.String("config", "internal/config/config.go", "path to the config source of truth")
	var files multiValue
	flagSet.Var(&files, "file", "env file to sync or check; repeat for multiple files")

	if err := flagSet.Parse(args); err != nil {
		return envsync.Paths{}, err
	}

	selected := []string(files)
	if len(selected) == 0 {
		selected = append(selected, defaultFiles...)
	}

	return envsync.Paths{
		ConfigPath: *configPath,
		Files:      selected,
	}, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: envsync <fix|check> [--config path] [--file path ...]")
}

func printFixReport(report *envsync.FixReport) error {
	if report == nil || len(report.Files) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "envsync fix: no changes")
		return err
	}

	for _, change := range report.Files {
		parts := make([]string, 0, 3)
		if change.Created {
			parts = append(parts, "created")
		}
		if len(change.Commented) > 0 {
			parts = append(parts, fmt.Sprintf("commented out %s", strings.Join(change.Commented, ", ")))
		}
		if len(change.Added) > 0 {
			parts = append(parts, fmt.Sprintf("added %s", strings.Join(change.Added, ", ")))
		}

		if len(parts) == 0 {
			parts = append(parts, "updated")
		}

		if _, err := fmt.Fprintf(os.Stdout, "envsync fix: %s: %s\n", change.Path, strings.Join(parts, "; ")); err != nil {
			return err
		}
	}

	return nil
}

func printCheckError(err error) {
	if checkErr, ok := err.(*envsync.CheckError); ok {
		fmt.Fprintln(os.Stderr, checkErr.Error())
		for _, issue := range checkErr.Issues {
			if len(issue.Missing) > 0 {
				fmt.Fprintf(os.Stderr, "envsync check: %s missing %s\n", issue.Path, strings.Join(issue.Missing, ", "))
			}
			if len(issue.Stale) > 0 {
				fmt.Fprintf(os.Stderr, "envsync check: %s stale %s\n", issue.Path, strings.Join(issue.Stale, ", "))
			}
		}

		return
	}

	fmt.Fprintln(os.Stderr, err)
}
