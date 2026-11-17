package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDirname(t *testing.T) {
	cases := []struct{ in, want string }{
		// POSIX / GNU behavior with forward slashes.
		{"/usr/bin/", "/usr"},
		{"/usr/bin", "/usr"},
		{"dir1/str", "dir1"},
		{"stdio.h", "."},
		{"", "."},
		{"/", "/"},
		{"///", "/"},
		{"/a", "/"},
		{"a/", "."},
		{"a//b//", "a"},
		{".", "."},
		{"..", "."},
		{"../x", ".."},
		// Backslash is a separator too.
		{`\`, `\`},
		{`a\b`, `a`},
		{`a\b\`, `a`},
		{`dir\`, `.`},
		{`a\b/c`, `a\b`},
		{`a/b\c`, `a/b`},
		{`a\\b\\`, `a`},
		// Drive prefixes.
		{`C:\Users\me\x.txt`, `C:\Users\me`},
		{`C:\x`, `C:\`},
		{`C:\`, `C:\`},
		{`C:\\`, `C:\`},
		{`c:/a/b`, `c:/a`},
		{`C:x`, `C:`},
		{`C:`, `C:`},
		{`C:a\b`, `C:a`},
		// Double-separator (UNC) roots stay distinct; three or more collapse.
		{`//`, `//`},
		{`\\`, `\\`},
		{`\\server`, `\\`},
		{`\\server\share\f`, `\\server\share`},
		{`//server/share/`, `//server`},
		{`\\\x`, `\`},
		// Non-ASCII passes through untouched.
		{"Ünïcode/файл", "Ünïcode"},
	}
	for _, c := range cases {
		if got := dirname(c.in); got != c.want {
			t.Errorf("dirname(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func runArgs(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return out.String(), errb.String(), code
}

func TestRun(t *testing.T) {
	progName = "dirname"
	t.Setenv("POSIXLY_CORRECT", "")

	if out, _, code := runArgs(t, "a/b", `c\d`); out != "a\nc\n" || code != 0 {
		t.Errorf("multiple: %q %d", out, code)
	}
	if out, _, _ := runArgs(t, "-z", "a/b", "c"); out != "a\x00.\x00" {
		t.Errorf("-z: %q", out)
	}
	if out, _, _ := runArgs(t, "a/b", "--ze"); out != "a\x00" {
		t.Errorf("permuted abbreviated --zero: %q", out)
	}
	if out, _, _ := runArgs(t, "--", "-z/x"); out != "-z\n" {
		t.Errorf("after --: %q", out)
	}
	if out, _, _ := runArgs(t, "-"); out != ".\n" {
		t.Errorf("lone dash: %q", out)
	}
	if _, e, code := runArgs(t); code != 1 || !strings.HasPrefix(e, "dirname: missing operand\n") {
		t.Errorf("missing operand: %q %d", e, code)
	}
	if _, e, code := runArgs(t, "-x", "a"); code != 1 || !strings.HasPrefix(e, "dirname: invalid option -- 'x'\n") {
		t.Errorf("invalid: %q %d", e, code)
	}
	if _, e, code := runArgs(t, "--bogus"); code != 1 || !strings.HasPrefix(e, "dirname: unrecognized option '--bogus'\n") {
		t.Errorf("unrecognized: %q %d", e, code)
	}
	if _, e, code := runArgs(t, "--zero=1", "a"); code != 1 || !strings.Contains(e, "doesn't allow an argument") {
		t.Errorf("argument to --zero: %q %d", e, code)
	}
	if out, _, code := runArgs(t, "--he"); code != 0 || !strings.HasPrefix(out, "Usage: dirname") {
		t.Errorf("--help: %q %d", out, code)
	}
}

func TestIsReleaseVersion(t *testing.T) {
	cases := map[string]bool{
		"v1.0.0":                                   true,
		"v1.2.3-rc.1":                              true,
		"":                                         false,
		"(devel)":                                  false,
		"v1.0.0+dirty":                             false,
		"v0.0.0-20261117120000-0123456789ab":       false,
		"v1.0.1-0.20261117120000-0123456789ab":     false,
		"v1.0.1-pre.0.20261117120000-0123456789ab": false,
	}
	for v, want := range cases {
		if got := isReleaseVersion(v); got != want {
			t.Errorf("isReleaseVersion(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestPosixlyCorrect(t *testing.T) {
	t.Setenv("POSIXLY_CORRECT", "1")
	if out, _, _ := runArgs(t, "a/b", "-z"); out != "a\n.\n" {
		t.Errorf("POSIXLY_CORRECT: %q", out)
	}
}
