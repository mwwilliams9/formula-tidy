// Package formulafmt normalizes the text of a spreadsheet formula: it fixes
// function name casing, cell reference casing, and whitespace around
// operators, commas and ranges, without changing what the formula computes.
package formulafmt

import (
	"fmt"
	"regexp"
	"strings"
)

type kind int

const (
	kString kind = iota
	kNumber
	kIdent
	kOp
	kLParen
	kRParen
	kComma
	kColon
	kSemicolon
)

type token struct {
	kind kind
	text string
}

var cellRefRe = regexp.MustCompile(`^\$?[A-Za-z]{1,3}\$?[0-9]+$`)

const opChars = "+-*/^&=<>%"

// Format takes a formula, with or without a leading "=", and returns it with
// consistent casing and spacing. The input is expected to be a single
// formula, not a range of cells.
func Format(input string) (string, error) {
	body := strings.TrimPrefix(strings.TrimSpace(input), "=")
	toks, err := tokenize(body)
	if err != nil {
		return "", err
	}
	if len(toks) == 0 {
		return "", fmt.Errorf("empty formula")
	}

	unary := make([]bool, len(toks))
	for i, t := range toks {
		if t.kind == kOp && (t.text == "+" || t.text == "-") {
			unary[i] = isUnaryContext(toks, i)
		}
	}

	var out strings.Builder
	out.WriteByte('=')
	for i, t := range toks {
		text := t.text
		if t.kind == kIdent {
			text = normalizeIdent(text, i, toks)
		}
		if i > 0 && needsSpace(toks, unary, i) {
			out.WriteByte(' ')
		}
		out.WriteString(text)
	}
	return out.String(), nil
}

// isUnaryContext reports whether a "+" or "-" at toks[i] is a sign rather
// than an arithmetic operator, based on what precedes it.
func isUnaryContext(toks []token, i int) bool {
	if i == 0 {
		return true
	}
	switch toks[i-1].kind {
	case kLParen, kComma, kSemicolon, kColon, kOp:
		return true
	}
	return false
}

// needsSpace decides whether a space belongs between toks[i-1] and toks[i].
func needsSpace(toks []token, unary []bool, i int) bool {
	prev, cur := toks[i-1], toks[i]

	switch cur.kind {
	case kRParen, kComma, kSemicolon, kColon:
		return false
	}
	if cur.kind == kOp && cur.text == "%" {
		return false
	}
	switch prev.kind {
	case kColon, kLParen:
		return false
	}
	if cur.kind == kLParen && (prev.kind == kIdent || prev.kind == kLParen) {
		return false
	}
	switch prev.kind {
	case kComma, kSemicolon:
		return true
	}
	if prev.kind == kOp && unary[i-1] {
		return false
	}
	return true
}

// normalizeIdent upper-cases function names, cell references and boolean
// literals, and leaves defined names and table names as the author wrote
// them, since their casing may be significant.
func normalizeIdent(text string, i int, toks []token) string {
	upper := strings.ToUpper(text)
	if i+1 < len(toks) && toks[i+1].kind == kLParen {
		return upper
	}
	if cellRefRe.MatchString(text) {
		return upper
	}
	if upper == "TRUE" || upper == "FALSE" {
		return upper
	}
	return text
}

func tokenize(s string) ([]token, error) {
	r := []rune(s)
	n := len(r)
	var toks []token

	for i := 0; i < n; {
		c := r[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++

		case c == '"':
			start := i
			i++
			closed := false
			for i < n {
				if r[i] == '"' {
					if i+1 < n && r[i+1] == '"' {
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated string literal starting at position %d", start)
			}
			toks = append(toks, token{kString, string(r[start:i])})

		case c == '(':
			toks = append(toks, token{kLParen, "("})
			i++
		case c == ')':
			toks = append(toks, token{kRParen, ")"})
			i++
		case c == ',':
			toks = append(toks, token{kComma, ","})
			i++
		case c == ':':
			toks = append(toks, token{kColon, ":"})
			i++
		case c == ';':
			toks = append(toks, token{kSemicolon, ";"})
			i++

		case strings.ContainsRune(opChars, c):
			start := i
			i++
			if (c == '<' && i < n && (r[i] == '=' || r[i] == '>')) ||
				(c == '>' && i < n && r[i] == '=') {
				i++
			}
			toks = append(toks, token{kOp, string(r[start:i])})

		case isIdentStart(c):
			start := i
			for i < n && isIdentPart(r[i]) {
				i++
			}
			toks = append(toks, token{kIdent, string(r[start:i])})

		case c >= '0' && c <= '9' || c == '.':
			start := i
			i++
			for i < n && (isDigit(r[i]) || r[i] == '.' || r[i] == 'e' || r[i] == 'E' ||
				((r[i] == '+' || r[i] == '-') && (r[i-1] == 'e' || r[i-1] == 'E'))) {
				i++
			}
			toks = append(toks, token{kNumber, string(r[start:i])})

		default:
			return nil, fmt.Errorf("unexpected character %q at position %d", c, i)
		}
	}
	return toks, nil
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isIdentStart(c rune) bool {
	return c == '$' || c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c rune) bool {
	return isIdentStart(c) || isDigit(c) || c == '.'
}
