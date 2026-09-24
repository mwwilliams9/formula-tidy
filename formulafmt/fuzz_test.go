package formulafmt

import "testing"

// FuzzTokenize checks that the tokenizer never panics and, whenever it
// succeeds, never hands back a token with no text: every branch in
// tokenize is supposed to consume at least one rune before appending, and
// a zero-length token would be a sign that some case broke that invariant
// and could loop forever without making progress through the input.
func FuzzTokenize(f *testing.F) {
	seeds := []string{
		"",
		"=SUM(A1:A10)",
		"sum(a1,a2,a3)",
		"=1+2",
		"=A1*-1",
		"=50%+1",
		`=if(a1>10,"big","small")`,
		"=$a$1+b2",
		"=sheet1!a1+1",
		"='My Sheet'!a1+1",
		"='O''Brien'!a1",
		"={1,2;3,4}",
		"=r[1]c[-1]",
		"=RC/RC[-1]",
		`="unterminated`,
		"='unterminated sheet name!a1",
		"=1+@2",
		`=SUM("a""b)`,
		"''''",
		`""`,
		"$$$",
		"R",
		"RC",
		"....",
		"1e",
		"1e+",
		"<=>=<>",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		toks, err := tokenize(s)
		if err != nil {
			return
		}
		for _, tok := range toks {
			if tok.text == "" {
				t.Fatalf("tokenize(%q) produced an empty %v token", s, tok.kind)
			}
		}
	})
}

// FuzzFormatIdempotent checks that formatting an already-formatted formula
// leaves it unchanged. The -diff flag depends on this: it treats a formula
// as settled once Format(f) == f, so any input that keeps changing under
// repeated formatting would make -diff loop forever on real input instead
// of converging.
func FuzzFormatIdempotent(f *testing.F) {
	seeds := []string{
		"=sum(a1:a10)",
		"sum(a1,a2,a3)",
		"=1+2",
		"=A1*-1",
		"=50%+1",
		`=if(a1>10,"big","small")`,
		"=$a$1+b2",
		"=sheet1!a1+1",
		"='My Sheet'!a1+1",
		"='O''Brien'!a1",
		"={1,2;3,4}",
		"=r[1]c[-1]",
		"=RC/RC[-1]",
		"=R2D2+1",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		once, err := Format(s)
		if err != nil {
			return
		}
		twice, err := Format(once)
		if err != nil {
			t.Fatalf("Format(%q) succeeded but re-formatting its own output %q failed: %v", s, once, err)
		}
		if once != twice {
			t.Fatalf("Format not idempotent for %q: first pass %q, second pass %q", s, once, twice)
		}
	})
}
