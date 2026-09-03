// Command formula-tidy normalizes spreadsheet formulas so that pasting them
// into a review or a diff doesn't drown the actual change in noise from
// inconsistent spacing and casing.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mwwilliams9/formula-tidy/formulafmt"
)

func main() {
	diffOnly := flag.Bool("diff", false, "print only formulas whose formatting changed")
	flag.Parse()
	args := flag.Args()

	hadErr := false

	if len(args) == 0 {
		hadErr = processReader(os.Stdin, "stdin", *diffOnly)
	} else {
		for _, path := range args {
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				hadErr = true
				continue
			}
			if processReader(f, path, *diffOnly) {
				hadErr = true
			}
			f.Close()
		}
	}

	if hadErr {
		os.Exit(1)
	}
}

// processReader formats one formula per line, printing each result to
// stdout and any per-line error to stderr, and keeps going on failure so a
// single bad line in a large file doesn't stop the rest from being formatted.
// With diffOnly set, lines whose formatted form matches the input as given
// are dropped instead of printed, so the output is just what would change.
func processReader(r *os.File, name string, diffOnly bool) bool {
	scanner := bufio.NewScanner(r)
	hadErr := false
	line := 0

	for scanner.Scan() {
		line++
		text := scanner.Text()
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			if !diffOnly {
				fmt.Println()
			}
			continue
		}
		out, err := formulafmt.Format(text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s:%d: %v\n", name, line, err)
			hadErr = true
			continue
		}
		if diffOnly && out == trimmed {
			continue
		}
		fmt.Println(out)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		hadErr = true
	}
	return hadErr
}
