# mmmbed

Generate source code containing binary files and their filenames, or named SQL
queries with execution metadata. The generated
code exposes the original bytes to the target language; no original file is
needed at runtime. mmmbed is written in Go and supports multiple assets per
output. It stores each filename and its immediate parent directory name, but not the full path.

## Install

```sh
go install github.com/zfogg/mmmbed/cmd/mmmbed@latest
```

Or clone this repository and run `go build ./cmd/mmmbed`. Published releases
include prebuilt Linux, macOS, and Windows binaries and SHA-256 checksums.
A pushed `v*` tag triggers the release workflow.

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
must not refer to an input file. Use `--check` with an output file to compare
without writing; it exits nonzero if the output is missing or stale.

## Named SQL queries

`mmmbed sql` embeds named SQL statements and their declared metadata. Write
regular `.sql` files, with one or more queries per file. A query starts with
`-- name:` and must declare its access and row count before its SQL:

```sql
-- name: LookupItem
-- access: read
-- rows: zero-or-one
SELECT name
FROM items
WHERE id = :id;

-- name: InsertItem
-- access: write
-- rows: one
INSERT INTO items (name) VALUES (:name)
RETURNING id;
```

`access` is `read` or `write`; `rows` is `zero`, `zero-or-one`, `one`, or
`many` (zero or more). These are independent declarations, so a write can
return rows. They describe expectations for consumers; mmmbed neither checks
SQL against a database nor enforces access or cardinality. Use `write` for
transaction, schema and modifying pragma statements. Names are unique across
all inputs and use ASCII letters, digits and underscores, starting with a
letter or underscore. `COUNT` is reserved for C and C++ output.

```sh
mmmbed sql -l c -i sql -i extra/one.sql --prefix database -o queries.h
mmmbed sql -l c -i sql -i extra/one.sql --prefix database -o queries.h --check
mmmbed sql -l go --go-package store sql -o queries.go
mmmbed sql -l go --go-package store --dialect mysql mysql-queries -o mysql_queries.go
mmmbed sql -l go --go-package store --dialect postgres postgres-queries -o postgres_queries.go
mmmbed sql -l python --dialect duckdb duckdb-queries -o duckdb_queries.py
```

For C and C++, `--single-line` joins SQL lines with spaces in the generated
catalog. It rejects comments, multiline literals or identifiers, and continued
string literals, where removing a newline could change SQL meaning. Without
this option the SQL text is preserved, with newlines rendered as `\n` in C.

Inputs may be repeated files or directories, using `-i` or positional paths.
Directories are scanned recursively for `.sql` files; symlinks are skipped
and explicit symlink inputs are rejected. Missing inputs, overlapping files,
duplicate names, and incomplete metadata are errors. Files and queries are
sorted for reproducible output. The output defaults to stdout. `--check`
compares an existing output without modifying it. Generated C/C++ headers
contain `database_queries`, `DATABASE_QUERY_LookupItem`-style indices, and
access and row enums when `--prefix database` is used. The table and its
strings have static lifetime. Numeric indices can change as queries are added.
Go generates a `Queries` map of `Query` records. All languages supported by
asset mode also support SQL mode, with these generated collection interfaces:

| Language | Query collection |
| --- | --- |
| C, C++ | `<prefix>_queries` array and `<PREFIX>_QUERY_<Name>` indices |
| Go | `Queries` map of `Query` records |
| JavaScript, TypeScript, Python, PHP, Swift, Kotlin, Dart, R | `queries` map |
| Rust | `QUERIES` slice of `Query` records |
| Java, Scala | `MmmbedQueries.queries` map |
| C# | `MmmbedQueries.Queries` dictionary |
| Ruby, Julia | `QUERIES` map |
| Lua | `queries` field of the returned module table |
| Perl | `%queries` |
| Elixir | `MmmbedQueries.queries()` map |

Each record supplies the query name, SQL text, access and rows; C# and Go use
capitalized field names. Some languages decode SQL from base64 during
initialization using their standard libraries. These collections are intended
for ordinary query catalogs; very large catalogs can encounter target compiler
or runtime limits, as with embedded assets.

Query bodies extend to the next top-level `-- name:` directive or EOF; SQL
semicolons do not split them. Leading and trailing whitespace around each body
is removed, and other content is preserved. Directive-like lines inside SQL
quoted strings, identifiers, block comments and PostgreSQL dollar-quoted
strings do not start queries. SQL syntax and dialect are otherwise not parsed.
The default `standard` directive lexer recognizes SQL-standard quoting and
PostgreSQL escape and dollar-quoted strings. Use `--dialect postgres` (or
`postgresql`) or `--dialect duckdb` for array/list brackets and their SQL
quoting rules. PostgreSQL `$1` positional parameters and DuckDB `$1`, `$name`,
and `?` parameters pass through unchanged; supported binding styles depend on
the driver. Use `--dialect mysql` for MySQL's
backslash-escaped strings, `#` comments, and whitespace-sensitive `--` comments.
The dialect setting only selects directive recognition rules; it never translates SQL.
Other dialect-specific quoting forms and modes, including PostgreSQL with
`standard_conforming_strings` disabled and MySQL `ANSI_QUOTES` or
`NO_BACKSLASH_ESCAPES`, are not recognized and may be rejected or misread.
Named placeholders are passed through unchanged, including SQLite's `:name`,
and the database driver binds values. Never interpolate untrusted values into
SQL. See `examples/sql` for a CMake/SQLite example, including named binding and
`RETURNING`; it requires SQLite 3.35 or later and an installed `mmmbed` binary.

The Go API is independent of the CLI:

```go
queries, err := sqlquery.LoadPaths([]string{"sql", "extra/one.sql"})
if err != nil { return err }
err = sqlquery.RenderTo(output, "c", queries, sqlquery.Options{Prefix: "database"})
```

Import `github.com/zfogg/mmmbed/sqlquery`. `sqlquery.Parse` also accepts a
source name and `io.Reader`, so callers can use their own filesystem discovery.
`ParseWithOptions` and `LoadPathsWithOptions` accept
`ParseOptions{Dialect: "postgres"}` (or `"duckdb"` / `"mysql"`) when needed.
Both APIs retain query data in memory; source locations are kept for diagnostics
and are not included in generated output.

## Per-file variables

For each input, mmmbed exposes its data under `mmmbed_file_` followed by the
first six hexadecimal digits of the SHA-256 hash of its bytes. It also exposes
`<name>_filename` and `<name>_path_parent`. The latter contains only the immediate
parent directory name; for stdin it is empty. Use `--omit-parent` for reproducible output across checkout directories, or `--parent NAME` to set it explicitly. C and C++ also expose `<name>_size`.

Use repeatable `--variable` to supply a name for every input, in input order.
Identifiers must use ASCII letters, digits and underscores, start with a letter
or underscore, and avoid keywords reserved by the selected language and mmmbed helper names. Elixir names must start with a
lowercase letter; R names must start with a letter. Default names that collide are disambiguated with `_2`, `_3`, and so on.
Custom names that collide are rejected.

Use `--collection NAME` to add a stable alias for the collection without
removing the default interface. Use `--sha256` to expose `<name>_sha256` as
a lowercase hexadecimal SHA-256 checksum of the original bytes. Lua exposes
these as keys of the returned table; its collection alias refers to that same
table. Use `--visibility public` or `--visibility private` for Go, Rust,
JavaScript, or TypeScript output. In Go, this adjusts the first letter of the
generated per-file and collection names; for the other three languages it
controls exports. Visibility is rejected for other languages.
Go names starting with an uppercase letter can be used from other packages.

Java and C# expose these names as `Mmmbed` fields, Scala as `Mmmbed` fields,
Elixir as zero-argument `Mmmbed` functions, Lua as keys on the returned table,
and JavaScript as CommonJS exports. Other languages expose them as top-level
variables (Perl uses package variables). Ruby top-level local variables are
available in the generated file's scope. The collection interfaces below remain
available as well.

Run `mmmbed languages` to list supported languages, suggested extensions, and runtime requirements.

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
some target compilers or runtimes impose size limits. The CLI and `embed.RenderTo` spool inputs to temporary disk and stream the output,
using bounded working memory; allow temporary disk space equal to the total input
size and output disk space for the generated source. Very large assets can still
exceed target compiler, source format, object-file, or runtime limits. The
convenience `embed.Render` and `embed.RenderWithOptions` APIs retain data in memory. It does not compress or encrypt data; anyone with the compiled program
can recover it.

## Development

Run `go test ./...`. Unit tests check the generated declarations and asset entries
for all 20 languages without executing generated code. The full suite also compiles
or runs generated programs for several languages when their toolchains are installed.
To run only the source-level generator checks, use `go test ./embed -run
TestGeneratedEntries`. Format Go code with `gofmt`.


## Go API and build integration

Import `github.com/zfogg/mmmbed/embed` and call `embed.RenderWithOptions` with
`[]embed.Asset` and `embed.Options`. The returned source is held in memory.
For large inputs use `embed.RenderTo` with `[]embed.StreamAsset` and an `io.Writer`:

```go
input := embed.StreamAsset{Name: "font.ttf", Reader: fontFile}
err := embed.RenderTo(outputFile, "go", []embed.StreamAsset{input}, embed.Options{
    GoPackage: "assets", Variables: []string{"Font"},
    Collection: "Files", SHA256: true,
})
```

`errors.Is` recognizes `embed.ErrInvalidInput` and `embed.ErrUnsupportedLanguage`;
underlying I/O errors remain wrapped. See `embed/example_test.go` for a runnable example.
See `examples/go-generate` for a `//go:generate` example and `examples/build`
for Make and CMake examples (install `mmmbed` first). Run `make generate`
to refresh it, `make build` to build the CLI, `make test` to test, `make lint`
to run `go vet`, and `make fmt` or `make fmt-check` for formatting.

With CGO enabled, unit tests parse generated source with Tree-sitter grammars
for 19 languages. A pure-Go Tree-sitter parser checks Perl.
These syntax tests do not run the generated source or replace target compiler
and runtime tests. CI runs the tests with CGO enabled on Linux.

For the optional streaming stress test, set `MMMBED_LARGE_TEST=1` and run
`go test ./embed -run TestRenderToMultiGiB`. It consumes 2 GiB of temporary
storage and writes generated Go source to a counting sink; it does not compile
that source or validate target compiler limits.
