// Command formula-tidy normalizes spreadsheet formulas so that pasting them
// into a review or a diff doesn't drown the actual change in noise from
// inconsistent spacing and casing.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mwwilliams9/formula-tidy/formulafmt"
)

func main() {
	args := os.Args[1:]
	hadErr := false

	if len(args) == 0 {
		hadErr = processReader(os.Stdin, "stdin")
	} else {
		for _, path := range args {
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				hadErr = true
				continue
			}
			if processReader(f, path) {
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
func processReader(r *os.File, name string) bool {
	scanner := bufio.NewScanner(r)
	hadErr := false
	line := 0

	for scanner.Scan() {
		line++
		text := scanner.Text()
		if strings.TrimSpace(text) == "" {
			fmt.Println()
			continue
		}
		out, err := formulafmt.Format(text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s:%d: %v\n", name, line, err)
			hadErr = true
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
