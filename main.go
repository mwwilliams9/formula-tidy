// Command formula-tidy normalizes spreadsheet formulas so that pasting them
// into a review or a diff doesn't drown the actual change in noise from
// inconsistent spacing and casing.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mwwilliams9/formula-tidy/formulafmt"
)

func main() {
	diffOnly := flag.Bool("diff", false, "print only formulas whose formatting changed")
	write := flag.Bool("w", false, "rewrite each file in place instead of printing to stdout")
	flag.Parse()
	args := flag.Args()

	if *write && *diffOnly {
		fmt.Fprintln(os.Stderr, "-w and -diff cannot be used together")
		os.Exit(1)
	}
	if *write && len(args) == 0 {
		fmt.Fprintln(os.Stderr, "-w requires at least one file argument")
		os.Exit(1)
	}

	hadErr := false

	if len(args) == 0 {
		hadErr = processReader(os.Stdin, os.Stdout, "stdin", *diffOnly)
	} else {
		for _, path := range args {
			if processFile(path, *diffOnly, *write) {
				hadErr = true
			}
		}
	}

	if hadErr {
		os.Exit(1)
	}
}

// processFile formats one file. With write set, the formatted result
// replaces the file's contents in place instead of going to stdout; a file
// that produced any per-line error is left untouched so a bad line can't
// cause a partially-formatted file to be written back.
func processFile(path string, diffOnly, write bool) bool {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return true
	}
	defer f.Close()

	var buf bytes.Buffer
	hadErr := processReader(f, &buf, path, diffOnly)
	if hadErr {
		return true
	}

	if !write {
		os.Stdout.Write(buf.Bytes())
		return false
	}

	info, err := f.Stat()
	mode := os.FileMode(0644)
	if err == nil {
		mode = info.Mode()
	}
	if err := os.WriteFile(path, buf.Bytes(), mode); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return true
	}
	return false
}

// processReader formats one formula per line, writing each result to w and
// any per-line error to stderr, and keeps going on failure so a single bad
// line in a large file doesn't stop the rest from being formatted. With
// diffOnly set, lines whose formatted form matches the input as given are
// dropped instead of written, so the output is just what would change.
func processReader(r io.Reader, w io.Writer, name string, diffOnly bool) bool {
	scanner := bufio.NewScanner(r)
	hadErr := false
	line := 0

	for scanner.Scan() {
		line++
		text := scanner.Text()
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			if !diffOnly {
				fmt.Fprintln(w)
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
		fmt.Fprintln(w, out)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		hadErr = true
	}
	return hadErr
}
