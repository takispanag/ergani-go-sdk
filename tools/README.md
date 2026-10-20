# The generators of the models

`generate_declarations.py` writes the wire structs, the public models, the
translation between them, the submission methods and the test that compares
every payload with the reference payload of the Ministry.

The field sets come from the reference payloads and the field types from the
official XSD schemas, both published by the Greek Ministry of Labour in the
sample archive of the interoperability guide.

## Running them

```sh
uv run --with pypdf python tools/extract_field_docs.py   # the descriptions
uv run python tools/infer_field_types.py                 # the types
uv run python tools/generate_declarations.py              # the Go code
gofmt -w .
```

The paths at the top of the generators point at a local copy of the sample
archive of the Ministry; adjust them before running them.

The code they write is committed, so the build never needs them.
