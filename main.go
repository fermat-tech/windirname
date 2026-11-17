// windirname strips the last component from file names.
//
// Usage:
//
//	windirname [OPTION] NAME...
//
// windirname is a Windows port of the GNU coreutils dirname(1) command. Each
// NAME is printed with its last non-separator component and any trailing
// separators removed; a NAME with no separator prints ".".
//
// Both "/" and "\" are separators, so a path written either way (or mixed) is
// handled, as brash and the Windows API both accept. The separators that
// remain are printed exactly as given; nothing is converted. Drive prefixes
// ("C:", "C:\") and "//" or "\\" roots (UNC) are kept as roots, following
// gnulib's rules for native Windows, which is what GNU dirname itself does
// when built for Windows.
//
// Options:
//
//	-z, --zero  end each output line with NUL, not newline
//	--help      display usage and exit
//	--version   output version information and exit
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

var progName string

// version is what --version reports. It is empty here so a release build can
// stamp it with -ldflags "-X main.version=vX.Y.Z"; otherwise buildVersion
// resolves it.
var version = ""

// defaultVersion is the version of this source: what --version reports for
// any build the toolchain cannot name a released version for (a source
// archive, an untagged commit, a modified working tree).
const defaultVersion = "v1.0.0"

func init() {
	progName = strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	if version == "" {
		version = buildVersion()
	}
}

// buildVersion prefers the version the Go toolchain stamped, but only when it
// names a real release (`go install …@vX.Y.Z`, or a clean checkout on a tag),
// so a later tag reports itself with no edit here.
func buildVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok && isReleaseVersion(bi.Main.Version) {
		return bi.Main.Version
	}
	return defaultVersion
}

// isReleaseVersion rejects "", "(devel)", a "+dirty" build and Go's
// pseudo-versions.
func isReleaseVersion(v string) bool {
	if len(v) < 2 || v[0] != 'v' || strings.ContainsRune(v, '+') {
		return false
	}
	return !isPseudoVersion(v)
}

// isPseudoVersion reports whether v ends the way every Go pseudo-version does:
// a 14-digit UTC timestamp, then "-" and a 12-hex-digit commit prefix. The
// separator before the timestamp is "-" in v0.0.0-… and "." in the forms that
// carry a base version (v1.2.3-0.…, v1.2.3-pre.0.…).
func isPseudoVersion(v string) bool {
	const stampLen, hashLen = 14, 12
	if len(v) < stampLen+hashLen+3 {
		return false
	}
	hash := v[len(v)-hashLen:]
	stamp := v[len(v)-hashLen-1-stampLen : len(v)-hashLen-1]
	if v[len(v)-hashLen-1] != '-' {
		return false
	}
	if sep := v[len(v)-hashLen-1-stampLen-1]; sep != '-' && sep != '.' {
		return false
	}
	for i := 0; i < len(stamp); i++ {
		if stamp[i] < '0' || stamp[i] > '9' {
			return false
		}
	}
	for i := 0; i < len(hash); i++ {
		c := hash[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "Usage: %s [OPTION] NAME...\n", progName)
	fmt.Fprint(w, `Output each NAME with its last non-slash component and trailing slashes
removed; if NAME contains no /'s, output '.' (meaning the current directory).
Both / and \ count as slashes; a drive prefix such as C: or C:\ is kept.

  -z, --zero     end each output line with NUL, not newline
      --help        display this help and exit
      --version     output version information and exit

Examples:
`)
	fmt.Fprintf(w, "  %s /usr/bin/          -> \"/usr\"\n", progName)
	fmt.Fprintf(w, "  %s C:\\Users\\me\\x.txt  -> \"C:\\Users\\me\"\n", progName)
	fmt.Fprintf(w, "  %s dir1/str dir2\\str -> \"dir1\" followed by \"dir2\"\n", progName)
	fmt.Fprintf(w, "  %s stdio.h            -> \".\"\n", progName)
}

func tryHelp(stderr io.Writer) {
	fmt.Fprintf(stderr, "Try '%s --help' for more information.\n", progName)
}

var longOpts = []string{"help", "version", "zero"}

// run parses args the way GNU getopt_long does (permuting unless
// POSIXLY_CORRECT is set, unique-prefix long options, "--" ends options) and
// returns the exit status.
func run(args []string, stdout, stderr io.Writer) int {
	zero := false
	var names []string
	permute := os.Getenv("POSIXLY_CORRECT") == ""

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			names = append(names, args[i+1:]...)
			i = len(args)
		case strings.HasPrefix(a, "--"):
			body := a[2:]
			name, hasArg := body, false
			if eq := strings.IndexByte(body, '='); eq >= 0 {
				name, hasArg = body[:eq], true
			}
			var matches []string
			for _, o := range longOpts {
				if o == name {
					matches = []string{o}
					break
				}
				if strings.HasPrefix(o, name) {
					matches = append(matches, o)
				}
			}
			switch len(matches) {
			case 0:
				fmt.Fprintf(stderr, "%s: unrecognized option '%s'\n", progName, a)
				tryHelp(stderr)
				return 1
			case 1:
			default:
				fmt.Fprintf(stderr, "%s: option '%s' is ambiguous; possibilities:", progName, a)
				for _, m := range matches {
					fmt.Fprintf(stderr, " '--%s'", m)
				}
				fmt.Fprintln(stderr)
				tryHelp(stderr)
				return 1
			}
			if hasArg {
				fmt.Fprintf(stderr, "%s: option '--%s' doesn't allow an argument\n", progName, matches[0])
				tryHelp(stderr)
				return 1
			}
			switch matches[0] {
			case "help":
				usage(stdout)
				return 0
			case "version":
				fmt.Fprintf(stdout, "%s %s\n", progName, version)
				return 0
			case "zero":
				zero = true
			}
		case len(a) > 1 && a[0] == '-':
			for _, c := range a[1:] {
				if c != 'z' {
					fmt.Fprintf(stderr, "%s: invalid option -- '%c'\n", progName, c)
					tryHelp(stderr)
					return 1
				}
				zero = true
			}
		default:
			if !permute {
				names = append(names, args[i:]...)
				i = len(args)
			} else {
				names = append(names, a)
			}
		}
	}

	if len(names) == 0 {
		fmt.Fprintf(stderr, "%s: missing operand\n", progName)
		tryHelp(stderr)
		return 1
	}

	end := byte('\n')
	if zero {
		end = 0
	}
	w := bufio.NewWriter(stdout)
	for _, n := range names {
		w.WriteString(dirname(n))
		w.WriteByte(end)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(stderr, "%s: write error: %v\n", progName, err)
		return 1
	}
	return 0
}

func isSlash(c byte) bool { return c == '/' || c == '\\' }

// prefixLen is gnulib's FILE_SYSTEM_PREFIX_LEN for native Windows: 2 for a
// drive letter followed by a colon, otherwise 0.
func prefixLen(s string) int {
	if len(s) >= 2 && s[1] == ':' && (s[0]|0x20) >= 'a' && (s[0]|0x20) <= 'z' {
		return 2
	}
	return 0
}

// lastComponent is gnulib's last_component: the index where the final
// non-separator component starts (len(s) if there is none).
func lastComponent(s string) int {
	base := prefixLen(s)
	for base < len(s) && isSlash(s[base]) {
		base++
	}
	lastWasSlash := false
	for p := base; p < len(s); p++ {
		if isSlash(s[p]) {
			lastWasSlash = true
		} else if lastWasSlash {
			base = p
			lastWasSlash = false
		}
	}
	return base
}

// dirLen is gnulib's dir_len with native-Windows settings
// (FILE_SYSTEM_DRIVE_PREFIX_CAN_BE_RELATIVE and DOUBLE_SLASH_IS_DISTINCT_ROOT
// both true).
func dirLen(s string) int {
	prefix := prefixLen(s)
	if prefix != 0 {
		if prefix < len(s) && isSlash(s[prefix]) {
			prefix++
		}
	} else if len(s) > 0 && isSlash(s[0]) {
		if len(s) > 1 && isSlash(s[1]) && (len(s) == 2 || !isSlash(s[2])) {
			prefix = 2
		} else {
			prefix = 1
		}
	}
	length := lastComponent(s)
	for ; prefix < length; length-- {
		if !isSlash(s[length-1]) {
			break
		}
	}
	return length
}

// dirname returns what GNU dirname prints for s (coreutils uses dir_len, and
// prints "." when that is zero).
func dirname(s string) string {
	n := dirLen(s)
	if n == 0 {
		return "."
	}
	return s[:n]
}
