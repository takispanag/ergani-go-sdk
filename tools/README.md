# The generator of the models

`generate` writes the Go code of the declarations of the Ergani API: the wire
structs, the public models, the translation between them, the submission
methods and the test that compares every payload with the reference payload
of the Ministry.

It reads the two things the Ministry publishes:

- the reference payloads of every declaration type, in `testdata/declarations`
- the schemas that describe them, in `internal/schema/xsds`

One table drives the four kinds of code, so they cannot drift apart. The
generator is written in Go and uses the standard library only, so the module
needs no other toolchain to be regenerated.

## Running it

```sh
go generate ./...
```

The directive lives in `generate.go`, in the root of the module, and the
generator writes the code with the formatter of the standard library, so the
result is the file `gofmt` would write.

The code it writes is committed, so the build of the module never needs it.

## What it reads

| What | Where |
| --- | --- |
| The reference payloads | `testdata/declarations` |
| The schemas of the Ministry | `internal/schema/xsds` |
| The fields the public API names itself | `tools/generate/tables.go` |
| The families of declarations | `tools/generate/tables.go` |

The paths can be set with the flags `-root`, `-samples` and `-schemas`.
