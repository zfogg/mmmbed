# mmmbed

Generate source code containing binary files and their filenames. The generated
code exposes the original bytes to the target language; no original file is
needed at runtime. mmmbed is written in Go and supports multiple assets per
output. It stores each filename and its immediate parent directory name, but not the full path.

## Install

```sh
go install github.com/zfogg/mmmbed/cmd/mmmbed@latest
```

Or clone this repository and run `go build ./cmd/mmmbed`.

## Usage

```sh
# One input and an output file (omit the output to write to stdout).
mmmbed -l c font.ttf font_data.c
mmmbed -l go --go-package media movie.mp4 movie_data.go

# Multiple files in one output. Order is preserved.
mmmbed -l python -i fonts/regular.ttf -i fonts/bold.ttf -o fonts.py

# Set one variable name per input (in the same order as the inputs).
mmmbed -l python -i fonts/regular.ttf -i fonts/bold.ttf --variable regular --variable bold -o fonts.py

# Standard input must have a name; output defaults to stdout.
cat font.ttf | mmmbed -l rust --name font.ttf > font.rs
mmmbed -l go --name video.mp4 -o video.go < video.mp4
```

`--language` (`-l`) is required. Use `--go-package` to choose the package for Go output (default `main`). Repeat `--input` (`-i`) to add files; a
positional input may also be supplied. `--output` (`-o`) selects an output
file, as does the second positional argument. Do not specify both. With no
file inputs, mmmbed reads stdin and requires `--name`; stdin cannot be combined
with file inputs. Filenames must be valid UTF-8 basenames, and duplicate names
are rejected. Input files are read before the output is written. The output
must not refer to an input file.

## Per-file variables

For each input, mmmbed exposes its data under `mmmbed_file_` followed by the
first six hexadecimal digits of the SHA-256 hash of its bytes. It also exposes
`<name>_filename` and `<name>_path_parent`. The latter contains only the immediate
parent directory name; for stdin it is empty. Use `--omit-parent` for reproducible output across checkout directories, or `--parent NAME` to set it explicitly. C and C++ also expose `<name>_size`.

Use repeatable `--variable` to supply a name for every input, in input order.
Identifiers must use ASCII letters, digits and underscores, start with a letter
or underscore, and avoid reserved keywords. Elixir names must start with a
lowercase letter. Default names that collide are disambiguated with `_2`, `_3`, and so on.
Custom names that collide are rejected.
Go names starting with an uppercase letter can be used from other packages.

Java and C# expose these names as `Mmmbed` fields, Scala as `Mmmbed` fields,
Elixir as zero-argument `Mmmbed` functions, Lua as keys on the returned table,
and JavaScript as CommonJS exports. Other languages expose them as top-level
variables (Perl uses package variables). Ruby top-level local variables are
available in the generated file's scope. The collection interfaces below remain
available as well.

## Generated interfaces

| Language (`-l`) | Access |
| --- | --- |
| `c`, `cpp` | `assets` array of `struct mmmbed_asset` with `name`, `data`, `size`; `assets_count` |
| `go` | exported `Assets` map from string to `[]byte` |
| `javascript` | CommonJS `require('./file').assets`; values are Node.js `Buffer`s |
| `typescript` | exported `assets: Record<string, Buffer>`; requires Node.js types/runtime |
| `python` | `assets` dictionary of `bytes` |
| `rust` | `ASSETS` slice of `( &str, &[u8] )` pairs |
| `java` | package-private `Mmmbed.assets`, a map of byte arrays |
| `csharp` | `Mmmbed.Assets`, a dictionary of byte arrays |
| `php` | `$assets` array of binary strings |
| `ruby` | `ASSETS` hash of binary strings |
| `swift` | `assets` dictionary of `Data` |
| `kotlin` | `assets` map of byte arrays |
| `scala` | `Mmmbed.assets` map of byte arrays |
| `dart` | `assets` map of `Uint8List` |
| `lua` | module returns an `assets` table of binary strings |
| `perl` | `%assets` package variable of byte strings |
| `r` | `assets` named list of raw vectors |
| `julia` | `assets` dictionary of byte vectors |
| `elixir` | `Mmmbed.assets()` map of binaries |

Generated code uses the target language's standard library; JavaScript and
TypeScript target Node.js. Include the generated C/C++ source in the translation
unit that consumes it, since its symbols are `static`. The generated Go file uses `package main` by default; pass `--go-package` to select another package. For Java, place the file in the same package as its consumer. Names
are UTF-8. Binary data is not text: use the provided length or byte container.

Most generators encode bytes as base64 and decode them during initialization;
C, C++, Rust, Lua and R emit byte literals instead. The generated source and
compilation can be substantially larger and slower than the original data, and
some target compilers or runtimes impose size limits. mmmbed holds input and
output in memory, so very large videos may exceed available memory or compiler
limits. It does not compress or encrypt data; anyone with the compiled program
can recover it.

## Development

Run `go test ./...`. Unit tests check the generated declarations and asset entries
for all 20 languages without executing generated code. The full suite also compiles
or runs generated programs for several languages when their toolchains are installed.
To run only the source-level generator checks, use `go test ./embed -run
TestGeneratedEntries`. Format Go code with `gofmt`.

