# The generator of the employment-lifecycle declarations

`generate_employment.py` writes the wire structs, the public models and the
translation between them, out of one table of the fields.

The table holds the type of every field and its name in the public API. The
field sets and the official descriptions come from the reference payloads and
the XSD schemas the Greek Ministry of Labour publishes.

## Running it

```sh
uv run --with pypdf python tools/extract_field_docs.py   # refresh the descriptions
uv run python tools/generate_employment.py               # write the Go code
gofmt -w .
```

The paths at the top of the generator point at the local copy of the sample
archive of the Ministry; adjust them before running it.

The code it writes is committed, so the build never needs the generator.
