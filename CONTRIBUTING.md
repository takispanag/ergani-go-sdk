# Contributing

Thanks for taking the time to improve `ergani-go-sdk`.

## Getting started

```sh
git clone https://github.com/takispanag/ergani-go-sdk.git
cd ergani-go-sdk
make check
```

`make check` runs everything the CI runs: the formatting check, `go vet`, the
test suite with the race detector, golangci-lint, the compilation of every
Go example of the README and the check that the generated code is up to date.

The linter is skipped when golangci-lint is not installed. To install it:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

## The models are generated

The models of the declarations are written by the generator of the module,
which reads the reference payloads of the Ministry and the schemas that
describe them. Never edit `declarations.go`, `declarations_wire.go`,
`declarations_golden_test.go` or `internal/wire/declarations.go` by hand: the
next run of the generator overwrites them.

```sh
make generate        # write the code of the models
make generate-check  # fail when the committed code is not the code the generator writes
```

A new field of a declaration is a change of the tables in
`tools/generate/tables.go`, and the test that compares every payload with the
reference payload of the Ministry tells whether the change is right.

## Reporting a bug

Open an issue with:

- the version of the SDK and of Go you are using,
- a short snippet that reproduces the problem,
- what you expected and what happened instead.

Never paste real credentials, tax numbers or employee names into an issue.
The test suite uses the `httptest` package, so most of the time a bug can be
reproduced without touching the real API.

## Pull requests

- Keep a pull request focused on one change.
- Add a test that fails without your change.
- Update the README and the `CHANGELOG.md` when the public behaviour changes.
- Make sure `make check` passes.

## Style

- The code follows the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
  and is formatted with `gofmt`.
- Exported identifiers carry a doc comment that starts with the name of the
  identifier.
- The package has no third-party dependencies, and every change should keep it
  that way. The standard library is enough for a client of a JSON API.
- Commit messages are written in the imperative mood and explain why the
  change is made, not only what changed.

## Compatibility

The SDK supports Go 1.21 and later. A change that drops the support of an
older release is a breaking change and needs a major version bump.

## Security

Do not open a public issue for a security problem. Contact the maintainer
privately instead.
