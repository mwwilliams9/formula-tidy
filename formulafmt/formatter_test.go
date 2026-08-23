package formulafmt

import "testing"

func TestFormat(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"=sum(a1:a10)", "=SUM(A1:A10)"},
		{"sum(a1,a2,a3)", "=SUM(A1, A2, A3)"},
		{"=1+2", "=1 + 2"},
		{"=1  +   2", "=1 + 2"},
		{"=A1*-1", "=A1 * -1"},
		{"=-1+2", "=-1 + 2"},
		{"=50%+1", "=50% + 1"},
		{`=if(a1>10,"big","small")`, `=IF(A1 > 10, "big", "small")`},
		{"=(1+2)*(3+4)", "=(1 + 2) * (3 + 4)"},
		{`=concatenate("a","b")`, `=CONCATENATE("a", "b")`},
		{"=$a$1+b2", "=$A$1 + B2"},
		{"=true", "=TRUE"},
		{"=MyNamedRange+1", "=MyNamedRange + 1"},
	}

	for _, c := range cases {
		got, err := Format(c.in)
		if err != nil {
			t.Errorf("Format(%q) returned error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Format(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatErrors(t *testing.T) {
	cases := []string{
		"",
		`="unterminated`,
		"=1+@2",
	}
	for _, in := range cases {
		if _, err := Format(in); err == nil {
			t.Errorf("Format(%q) expected an error, got none", in)
		}
	}
}

func TestUnterminatedStringDoesNotHang(t *testing.T) {
	if _, err := Format(`=SUM("a""b)`); err == nil {
		t.Errorf("expected unterminated string error")
	}
}
