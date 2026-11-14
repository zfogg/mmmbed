# mmmbed documentation

`mmmbed` embeds text, SQL, and source assets into Go, C, and native build
workflows. This searchable site brings the repository README, package guides,
CLI behavior, and examples into one documentation home.

```{toctree}
:maxdepth: 2
:caption: Start here

README
../README
getting-started
```

```{toctree}
:maxdepth: 2
:caption: Guides and reference

go-embedding
sql-queries
cli
api/index
```

```{toctree}
:maxdepth: 2
:caption: Examples and development

examples/go-generate
examples/sql
examples/build
contributing
```

The Markdown pages remain in `docs/` beside the code. The site is built with
Sphinx and MyST, with Hawkmoth included for the repository’s C-facing examples.
