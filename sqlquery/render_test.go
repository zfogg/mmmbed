package sqlquery

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGeneratedCatalogsRun(t *testing.T) {
	queries := []Query{{Name: "Lookup", SQL: "SELECT :name, 'café', '\\';\n-- comment", Access: "read", Rows: "zero-or-one"}}
	for _, language := range []string{"c", "cpp", "go", "python", "javascript", "typescript", "rust", "lua", "perl", "java", "csharp", "php", "ruby", "swift", "kotlin", "scala", "dart", "r", "julia", "elixir"} {
		t.Run(language, func(t *testing.T) {
			var source bytes.Buffer
			if err := RenderTo(&source, language, queries, Options{}); err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			exeName := "catalog-test"
			if runtime.GOOS == "windows" {
				exeName += ".exe"
			}
			base64SQL := base64.StdEncoding.EncodeToString([]byte(queries[0].SQL))
			write := func(name, data string) {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run := func(tool string, args ...string) {
				t.Helper()
				if filepath.IsAbs(tool) {
					if _, err := os.Stat(tool); err != nil {
						t.Fatalf("compiled executable missing: %v", err)
					}
				} else if _, err := exec.LookPath(tool); err != nil {
					t.Skipf("%s unavailable", tool)
				}
				cmd := exec.Command(tool, args...)
				cmd.Dir = dir
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s: %v\n%s", tool, err, out)
				}
			}
			switch language {
			case "c", "cpp":
				write("queries.h", source.String())
				write("main.c", fmt.Sprintf("#include <string.h>\n#include \"queries.h\"\nint main(void) { return MMMBED_QUERY_COUNT != 1 || strcmp(mmmbed_queries[MMMBED_QUERY_Lookup].sql, %s) || mmmbed_queries[0].access != MMMBED_READ || mmmbed_queries[0].rows != MMMBED_ZERO_OR_ONE; }\n", cString(queries[0].SQL)))
				compiler := "cc"
				if language == "cpp" {
					compiler = "c++"
				}
				run(compiler, "main.c", "-o", exeName)
				run(filepath.Join(dir, exeName))
			case "go":
				write("queries.go", source.String())
				write("main.go", fmt.Sprintf("package main\nfunc main() { q := Queries[\"Lookup\"]; if q.SQL != %q || q.Access != \"read\" || q.Rows != \"zero-or-one\" { panic(q) } }\n", queries[0].SQL))
				run("go", "run", "queries.go", "main.go")
			case "python":
				write("queries.py", source.String())
				run("python", "-c", fmt.Sprintf("from queries import queries\nassert queries['Lookup']['sql'] == %q\nassert queries['Lookup']['rows'] == 'zero-or-one'", queries[0].SQL))
			case "javascript":
				write("queries.js", source.String())
				run("node", "-e", fmt.Sprintf("const q = require('./queries.js').queries.Lookup; if (q.sql !== %q || q.rows !== 'zero-or-one') process.exit(1)", queries[0].SQL))
			case "typescript":
				write("queries.ts", source.String())
				run("tsc", "--target", "ES2020", "--module", "commonjs", "queries.ts")
				run("node", "-e", fmt.Sprintf("const q = require('./queries.js').queries.Lookup; if (q.sql !== %q || q.rows !== 'zero-or-one') process.exit(1)", queries[0].SQL))
			case "rust":
				write("queries.rs", source.String())
				write("main.rs", fmt.Sprintf("mod queries; fn main() { let q = &queries::QUERIES[0]; assert_eq!(q.sql, %s); assert_eq!(q.rows, \"zero-or-one\"); }\n", rustString(queries[0].SQL)))
				run("rustc", "main.rs", "-o", exeName)
				run(filepath.Join(dir, exeName))
			case "lua":
				write("queries.lua", source.String())
				run("lua", "-e", fmt.Sprintf("local q = require('queries').queries.Lookup; local h = q.sql:gsub('.', function(c) return string.format('%%02x', string.byte(c)) end); assert(h == '%x' and q.rows == 'zero-or-one')", []byte(queries[0].SQL)))
			case "perl":
				write("queries.pl", source.String())
				run("perl", "-e", fmt.Sprintf("require './queries.pl'; use Encode; die unless unpack('H*', Encode::encode('UTF-8', $queries{'Lookup'}{'sql'})) eq '%x' and $queries{'Lookup'}{'rows'} eq 'zero-or-one'", []byte(queries[0].SQL)))
			case "java":
				write("Queries.java", source.String())
				write("Main.java", fmt.Sprintf("class Main { public static void main(String[] a) { MmmbedQueries.Query q = MmmbedQueries.queries.get(\"Lookup\"); if (!java.util.Base64.getEncoder().encodeToString(q.sql.getBytes(java.nio.charset.StandardCharsets.UTF_8)).equals(\"%s\") || !q.rows.equals(\"zero-or-one\")) throw new AssertionError(); } }", base64SQL))
				run("javac", "Queries.java", "Main.java")
				run("java", "Main")
			case "csharp":
				version, err := exec.Command("dotnet", "--list-sdks").Output()
				if err != nil || len(strings.Fields(string(version))) == 0 {
					t.Skip(".NET SDK unavailable")
				}
				major := strings.Split(strings.Fields(string(version))[0], ".")[0]
				write("Example.csproj", fmt.Sprintf("<Project Sdk=\"Microsoft.NET.Sdk\"><PropertyGroup><OutputType>Exe</OutputType><TargetFramework>net%s.0</TargetFramework></PropertyGroup></Project>", major))
				write("Queries.cs", source.String())
				write("Program.cs", fmt.Sprintf("class Program { static void Main() { var q=MmmbedQueries.Queries[\"Lookup\"]; if(System.Convert.ToBase64String(System.Text.Encoding.UTF8.GetBytes(q.Sql)) != \"%s\" || q.Rows != \"zero-or-one\") throw new System.Exception(); } }", base64SQL))
				run("dotnet", "run", "--project", "Example.csproj")
			case "php":
				write("queries.php", source.String())
				run("php", "-r", fmt.Sprintf("require 'queries.php'; $q=$queries['Lookup']; if (base64_encode($q['sql']) !== '%s' || $q['rows'] !== 'zero-or-one') exit(1);", base64SQL))
			case "ruby":
				write("queries.rb", source.String())
				run("ruby", "-r", "./queries.rb", "-e", fmt.Sprintf("q=QUERIES['Lookup']; abort unless [q[:sql]].pack('m0') == '%s' && q[:rows] == 'zero-or-one'", base64SQL))
			case "swift":
				write("queries.swift", source.String())
				write("main.swift", fmt.Sprintf("import Foundation\nlet q=queries[\"Lookup\"]!\nprecondition(q.sql.data(using: .utf8)!.base64EncodedString() == \"%s\" && q.rows == \"zero-or-one\")\n", base64SQL))
				run("swiftc", "queries.swift", "main.swift", "-o", exeName)
				run(filepath.Join(dir, exeName))
			case "kotlin":
				write("queries.kt", source.String())
				write("main.kt", fmt.Sprintf("fun main() { val q=queries.getValue(\"Lookup\"); check(java.util.Base64.getEncoder().encodeToString(q.sql.toByteArray(Charsets.UTF_8)) == \"%s\" && q.rows == \"zero-or-one\") }\n", base64SQL))
				run("kotlinc", "queries.kt", "main.kt", "-include-runtime", "-d", "example.jar")
				run("java", "-jar", "example.jar")
			case "scala":
				write("queries.scala", source.String())
				write("Main.scala", fmt.Sprintf("object Main { def main(args: Array[String]): Unit = { val q=MmmbedQueries.queries(\"Lookup\"); assert(java.util.Base64.getEncoder.encodeToString(q.sql.getBytes(java.nio.charset.StandardCharsets.UTF_8)) == \"%s\" && q.rows == \"zero-or-one\") } }\n", base64SQL))
				run("scalac", "queries.scala", "Main.scala")
				run("scala", "Main")
			case "dart":
				write("queries.dart", source.String())
				write("main.dart", fmt.Sprintf("import 'dart:convert';\nimport 'queries.dart';\nvoid main() { final q=queries['Lookup']!; if (base64Encode(utf8.encode(q['sql']!)) != '%s' || q['rows'] != 'zero-or-one') throw StateError('bad query'); }\n", base64SQL))
				run("dart", "run", "main.dart")
			case "r":
				write("queries.R", source.String())
				run("Rscript", "-e", fmt.Sprintf("source('queries.R'); q <- queries[['Lookup']]; stopifnot(paste(sprintf('%%02x', as.integer(charToRaw(q$sql))), collapse='') == '%x', q$rows == 'zero-or-one')", []byte(queries[0].SQL)))
			case "julia":
				write("queries.jl", source.String())
				run("julia", "-e", fmt.Sprintf("include(\"queries.jl\"); using Base64; q=QUERIES[\"Lookup\"]; @assert base64encode(q.sql) == \"%s\" && q.rows == \"zero-or-one\"", base64SQL))
			case "elixir":
				write("queries.ex", source.String())
				run("elixir", "-r", "queries.ex", "-e", fmt.Sprintf("q=MmmbedQueries.queries()[\"Lookup\"]; if Base.encode64(q.sql) != \"%s\" or q.rows != \"zero-or-one\", do: raise(\"bad query\")", base64SQL))
			}
		})
	}
}

func TestSQLiteExample(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("pkg-config integration is exercised on Unix; CMake example supports Windows")
	}
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("C compiler unavailable")
	}
	if err := exec.Command("pkg-config", "--atleast-version=3.35", "sqlite3").Run(); err != nil {
		t.Skip("SQLite 3.35 development files unavailable")
	}
	flags, err := exec.Command("pkg-config", "--cflags", "--libs", "sqlite3").Output()
	if err != nil {
		t.Fatal(err)
	}
	queries, err := LoadPaths([]string{"../examples/sql/queries"})
	if err != nil {
		t.Fatal(err)
	}
	var header bytes.Buffer
	if err = RenderTo(&header, "c", queries, Options{}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.WriteFile(filepath.Join(dir, "queries_generated.h"), header.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	main, err := os.ReadFile("../examples/sql/main.c")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "main.c"), main, 0600); err != nil {
		t.Fatal(err)
	}
	args := append([]string{"-std=c11", "-Wall", "-Wextra", "-Werror", "main.c", "-o", "example"}, strings.Fields(string(flags))...)
	cmd := exec.Command("cc", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	cmd = exec.Command(filepath.Join(dir, "example"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
}

func TestJVMLargeSQL(t *testing.T) {
	for _, lang := range []string{"java", "kotlin", "scala"} {
		t.Run(lang, func(t *testing.T) {
			tools := map[string][]string{"java": {"javac", "java"}, "kotlin": {"kotlinc", "java"}, "scala": {"scalac", "scala"}}
			for _, name := range tools[lang] {
				if _, err := exec.LookPath(name); err != nil {
					t.Skipf("%s unavailable", name)
				}
			}
			queries := []Query{{Name: "Large", SQL: "SELECT '" + strings.Repeat("x", 55000) + "'", Access: "read", Rows: "one"}}
			var output bytes.Buffer
			if err := RenderTo(&output, lang, queries, Options{}); err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			write := func(name, content string) {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run := func(name string, args ...string) {
				t.Helper()
				cmd := exec.Command(name, args...)
				cmd.Dir = dir
				if result, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s: %v\n%s", name, err, result)
				}
			}
			switch lang {
			case "java":
				write("Queries.java", output.String())
				write("Main.java", fmt.Sprintf("class Main { public static void main(String[] a) { if (MmmbedQueries.queries.get(\"Large\").sql.length() != %d) throw new AssertionError(); } }", len(queries[0].SQL)))
				run("javac", "Queries.java", "Main.java")
				run("java", "Main")
			case "kotlin":
				write("queries.kt", output.String())
				write("main.kt", fmt.Sprintf("fun main() { check(queries.getValue(\"Large\").sql.length == %d) }", len(queries[0].SQL)))
				run("kotlinc", "queries.kt", "main.kt", "-include-runtime", "-d", "example.jar")
				run("java", "-jar", "example.jar")
			case "scala":
				write("queries.scala", output.String())
				write("Main.scala", fmt.Sprintf("object Main { def main(args: Array[String]): Unit = assert(MmmbedQueries.queries(\"Large\").sql.length == %d) }", len(queries[0].SQL)))
				run("scalac", "queries.scala", "Main.scala")
				run("scala", "Main")
			}
		})
	}
}
