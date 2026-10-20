# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- The badge of the sunset Go Report Card service was replaced by the badge of
  golangci-lint, the linter the CI actually runs.

## [0.1.0] - 2026-10-21

The first tagged release. It publishes the SDK, the documentation and the
tooling that keeps both in sync.

### Added

- Support for the work card, overtime, daily schedule and weekly schedule
  document types.
- Automatic re-authentication when the access token is rejected with a `401`.
- Retries with an exponential backoff for `429`, `5xx` and transport failures,
  honouring the `Retry-After` header. The number of retries and the base
  delay are configurable through `Config.MaxRetries` and
  `Config.RetryBackoff`.
- `ergani.ErrUnauthorized`, matched through `errors.Is`, and
  `ergani.IsStatusCode` to inspect the status of a failed call.
- A `UserAgent` option, sent with every request.
- A GitHub Actions pipeline that builds and tests the package on Go 1.19 up
  to Go 1.24 with the race detector, checks the formatting, runs
  golangci-lint, compiles the README examples and builds the Docker image.
- A Makefile that runs the same checks locally.
- A tool that compiles every Go example of the README, so the documentation
  cannot drift away from the code.
- `CHANGELOG.md`, `CONTRIBUTING.md` and `ROADMAP.md`.

### Fixed

- The nested payloads are now sent as nested objects. The documented field
  names use a `>` path syntax that `encoding/json` does not understand, so
  the keys of `Details`, `Ergazomenoi` and `ErgazomenosAnalytics` used to be
  emitted verbatim (`Details>CardDetails`) instead of being nested. Every
  document type was affected.
- Every Go example of the README compiles again. The client constructor was
  called with a context it does not take, the model fields were named
  differently than in the code and the submission methods were documented as
  returning a single response instead of a slice.
- `SubmitWeeklySchedule` is documented.
- The client is safe for concurrent use by multiple goroutines.
- A `.gitattributes` file normalises the line endings, so a Windows checkout
  no longer makes every file look unformatted to `gofmt`.
