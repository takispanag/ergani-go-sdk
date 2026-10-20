# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The versions below `v1.0.0` carry no compatibility guarantee: a minor release
may break the public API.

## [Unreleased]

## [0.2.0] - 2026-10-21

The release that makes the SDK usable, and that corrects the payloads it
sends. The wire format was checked against the official XSD schemas of the
Ministry of Labour, and the test suite now compares the payload of every
declaration type with the reference payloads the Ministry publishes.

### Fixed

- **The module was not importable.** The root of the module held a `package
  main`, so `go get github.com/takispanag/ergani-go-sdk` gave a command and
  not a library. The library is at the root now, and the example is at
  `cmd/ergani-example`.
- **The default base URL was dead.** The SDK talked to
  `trialeservices.yeka.gr`, which answers 404 on every route since the
  platform moved. The default is the current trial environment,
  `https://trialv2eservices.yeka.gr/WebServicesAPI/Api/`.
- **`f_aa`, `f_aa_pararthmatos` and `f_weekdates` were sent as JSON numbers.**
  The schemas declare them as strings, and the API expects `"0"` and `"5"`.
- **`f_day` was sent as a JSON number.** The schema declares it as a string.
- **`f_reference_date` was sent as `DD/MM/YYYY`.** The schema declares it as
  an `xs:date`, so the work card expects `YYYY-MM-DD`. The other declaration
  types do use `DD/MM/YYYY`, and the two are now distinct.
- **The nested payloads were not nested.** The documented field names use a
  `>` path syntax that `encoding/json` does not understand, so the keys of
  `Details`, `Ergazomenoi` and `ErgazomenosAnalytics` were emitted verbatim
  instead of being nested. Every declaration type was affected.
- **The models could not be read back.** The exported models carried JSON
  tags and a `MarshalJSON`, but no `UnmarshalJSON`, so decoding a payload
  silently dropped the nested records.
- **The access token never expired in the client.** The SDK read an
  `expiresIn` field that the API does not send, so the token was only ever
  renewed after a `401`, and a rejected token was never replaced.
- **The response of a submission failed as a whole** when the API reported a
  date in an unexpected format.

### Added

- The full route surface of the API: the renewal and the revocation of the
  credentials, the list of the declaration types, the schema of a
  declaration type, the recall of a submitted declaration, its retrieval as a
  PDF, the list of the interoperability services and their execution.
- The working time overtime declaration (`WTOOv`) and the leave declaration
  (`WTOLeave`), with the leave types of the official catalogue.
- The discovery of the declaration types the connected employer may submit,
  so the SDK does not have to hard-code a list the API changes.
- The catalogues of the codes the API accepts, through the `EX_BASE_03`
  service.
- Typed errors: `APIError` carries the status and the message, the sentinels
  `ErrUnauthorized`, `ErrInvalidRequest`, `ErrNotFound`, `ErrRateLimited` and
  `ErrServer` classify a failure, and `StatusCode` and `IsStatusCode` read
  the status.
- A `ValidationError` that points at the offending field, so a malformed
  declaration is reported before it reaches the API.
- A logger hook that reports the requests without ever writing the
  credentials or the identifiers of the employees to the log.
- The constants of the declaration types, of the working time types, of the
  leave types, of the justifications and of the weekly patterns.
- A test suite that compares the payload of every declaration type with the
  reference payloads of the Ministry, that exercises every route, that
  asserts the backoff sequence of the retries, and that checks the
  concurrency of the client.
- A fuzz target for the parser of the error bodies.

### Changed

- **Breaking** The constructor is `New(username, password, opts...)`. The
  `Config` struct, `NewClient` and `NewClientWithConfig` are gone.
- **Breaking** The models are plain data: they carry no JSON tags, and they
  use `time.Time`, `bool` and `time.Weekday` instead of the `Date`, `Time`,
  `Bool` and `Weekday` wrappers.
- **Breaking** The declaration types are named after what they are:
  `WorkCardDeclaration` replaces `CompanyWorkCard`, and so on.
- **Breaking** The submission methods are plural, and they take a slice of
  declarations.
- **Breaking** The errors are matched with `errors.Is` and `errors.As`, and
  the `HTTPClient` interface is replaced by `*http.Client`.
- The wire format moved to `internal/wire`, so it can be corrected without
  breaking the public API.
- The floor of the Go version is 1.21, which brings `log/slog` and the
  modern standard library.
- The linter set grew: `bodyclose`, `containedctx`, `errname`, `exhaustive`,
  `noctx` and `thelper` joined the others.

### Migration from v0.1.0

| v0.1.0 | v0.2.0 |
| --- | --- |
| `ergani.NewClient(u, p)` | `ergani.New(u, p)` |
| `ergani.NewClient(u, p, baseURL)` | `ergani.New(u, p, ergani.WithBaseURL(baseURL))` |
| `ergani.NewClientWithConfig(cfg)` | `ergani.New(cfg.Username, cfg.Password, ...)` |
| `cfg.HTTPClient` (an interface) | `ergani.WithHTTPClient(&http.Client{...})` |
| `cfg.MaxRetries = -1` | `ergani.WithMaxRetries(0)` |
| `cfg.Timeout` | `ergani.WithTimeout(d)` |
| `cfg.UserAgent` | `ergani.WithUserAgent(s)` |
| `ergani.CompanyWorkCard` | `ergani.WorkCardDeclaration` |
| `ergani.CompanyOvertime` | `ergani.OvertimeDeclaration` |
| `ergani.CompanyDailySchedule` | `ergani.DailyScheduleDeclaration` |
| `ergani.CompanyWeeklySchedule` | `ergani.WeeklyScheduleDeclaration` |
| `ergani.CompanyWorkCard.CardDetails` | `ergani.WorkCardDeclaration.Cards` |
| `ergani.Overtime` | `ergani.OvertimeRecord` |
| `ergani.SubmissionResponse` | `ergani.Submission` |
| `ergani.Date{Time: t}` | `t` |
| `ergani.Time{Time: t}` | `t` |
| `ergani.DateTime{Time: t}` | `t` |
| `ergani.Bool(b)` | `b` |
| `ergani.Weekday{Weekday: time.Monday}` | `time.Monday` |
| `ergani.Arrival` | `ergani.MovementArrival` |
| `ergani.Departure` | `ergani.MovementDeparture` |
| `ergani.PowerOutage` | `ergani.JustificationPowerOutage` |
| `ergani.WorkFromOffice` | `ergani.WorkTypeOffice` |
| `ergani.ExceptionalWorkload` | `ergani.OvertimeExceptionalWorkload` |
| `client.SubmitWorkCard(ctx, cards)` | `client.SubmitWorkCards(ctx, declarations)` |
| `client.SubmitOvertime(ctx, cards)` | `client.SubmitOvertime(ctx, declarations)` |
| `client.SubmitDailySchedule(ctx, s)` | `client.SubmitDailySchedules(ctx, declarations)` |
| `client.SubmitWeeklySchedule(ctx, s)` | `client.SubmitWeeklySchedules(ctx, declarations)` |
| `apiErr.Response` | `apiErr.RawBody` |
| `ergani.IsStatusCode(err, 400)` | unchanged |
| — | `ergani.StatusCode(err)` |
| — | `ergani.ErrInvalidRequest`, `ErrNotFound`, `ErrRateLimited`, `ErrServer` |
| — | `*ergani.ValidationError` |
| — | `client.ListSubmissions(ctx)` |
| — | `client.DocumentSchema(ctx, code)` |
| — | `client.CancelDocument(ctx, req)` |
| — | `client.DocumentPDF(ctx, code, protocol, date)` |
| — | `client.Services(ctx)`, `client.ExecuteService(ctx, req)`, `client.CodeList(ctx, name)` |
| — | `client.Authenticate(ctx)`, `client.Refresh(ctx)`, `client.Logout(ctx)` |

The import path changed as well: the package is `github.com/takispanag/ergani-go-sdk`
and the subdirectory `ergani/` is gone.

[Unreleased]: https://github.com/takispanag/ergani-go-sdk/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/takispanag/ergani-go-sdk/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/takispanag/ergani-go-sdk/releases/tag/v0.1.0
