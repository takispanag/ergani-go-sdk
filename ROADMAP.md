# Roadmap

The work that is still ahead for `ergani-go-sdk`. The items are tracked as
[GitHub issues](https://github.com/takispanag/ergani-go-sdk/issues), so the
progress of each one is visible.

## v0.2.0

### Wider API coverage

- Cover the Ergani document types that are not implemented yet, once their
  endpoint paths are confirmed against the official Yeka documentation.
- Support changing or cancelling a declaration that was already submitted,
  which the API exposes separately from the submission endpoints.

### A friendlier client

- Functional options for the client, so that new settings do not break the
  existing call sites.
- A typed error per endpoint, instead of a single `APIError` carrying a
  status code.
- A logging hook that reports the requests and the responses, with the
  credentials and the employee identifiers redacted.

### Tests

- Integration tests against the Ergani sandbox, enabled by an environment
  variable and skipped by default, so the CI never touches the real API.
- Golden files for every payload, so a change of the wire format is visible
  in the diff of a pull request.

## v0.3.0

- A small command line tool for one-shot submissions, for the employers that
  do not want to build their own binary.
- Read back the declarations that were submitted, so an employer can check
  what the platform recorded.

## Later

- Typed helpers that turn the Greek codes of the API into readable values,
  the other way around.
- A guide for the employers that migrate from the web form to the API.
