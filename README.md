# windirname

A Windows port of the GNU coreutils [`dirname`](https://man7.org/linux/man-pages/man1/dirname.1.html)
command: strip the last component from each file name. Both `/` and `\` count
as separators, so it works the same on paths written either way, as typed in
brash, cmd.exe or PowerShell.

## Install

```powershell
go install github.com/fermat-tech/windirname@latest
```

Or build from source:

```powershell
git clone https://github.com/fermat-tech/windirname
cd windirname
go build -o windirname.exe .
```

## Usage

```
windirname [OPTION] NAME...
```

Each NAME is printed with its last non-separator component and any trailing
separators removed. A NAME with no separator prints `.`.

| Option | Description |
|--------|-------------|
| `-z`, `--zero` | end each output line with NUL, not newline |
| `--help` | display help and exit |
| `--version` | output version information and exit |

Options are parsed as GNU `getopt_long` parses them: they may follow operands
(unless `POSIXLY_CORRECT` is set), long options may be abbreviated to any
unique prefix, and `--` ends option processing.

## Examples

| Input | Output |
|-------|--------|
| `/usr/bin/` | `/usr` |
| `C:\Users\me\notes.txt` | `C:\Users\me` |
| `a\b/c` | `a\b` |
| `dir\` | `.` |
| `stdio.h` | `.` |
| `C:\x` | `C:\` |
| `C:x` | `C:` |
| `\\server\share\file` | `\\server\share` |

```sh
cd "$(windirname "$script")"
```

## Windows rules

The separators that remain are printed exactly as given; windirname never
converts `\` to `/` or back. Roots follow gnulib's rules for native Windows,
which is what GNU dirname itself does when built for Windows:

- A drive prefix is a root: `C:\x` gives `C:\`, `C:x` gives `C:`.
- A leading pair of separators is a distinct root (UNC), so `//` and `\\`
  print themselves; three or more collapse to one, so `///` gives `/`.

On Linux, GNU dirname treats `\` as an ordinary character and `//` as `/`.
Those are the only differences from it; everything else (output, exit status,
error messages) matches GNU dirname 8.32.

## License

MIT
