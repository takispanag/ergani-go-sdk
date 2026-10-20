# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The versions below `v1.0.0` carry no compatibility guarantee: a minor release
may break the public API.

## [Unreleased]

## [0.6.0] - 2026-10-21

The release that brings the generation of the models into the module, in Go.

### Changed

- **The generator of the models is written in Go.** It lived in Python, which
  meant that the module needed a second toolchain to be regenerated, and
  that the generator was outside the reach of the compiler, of the linter and
  of the tests. It is now a program of the module, under `tools/generate`, and
  it uses the standard library only.
- **The generation is wired with `go generate`.** The directive lives in
  `generate.go`, in the root of the module, so `go generate ./...` writes the
  models and `make generate-check` fails when the committed code is not the
  code the generator writes.
- **The generator reads the module, not a copy of it.** The reference payloads
  come from `testdata/declarations` and the schemas from
  `internal/schema/xsds`, both of which the module already carries, so the
  code can be regenerated from a fresh clone.
- The generator writes the code with the formatter of the standard library,
  so the file it writes is the file `gofmt` would write.

### Added

- The CI checks that the generated code is up to date.

## [0.5.0] - 2026-10-21

The release that types the nested objects of the declarations, and that puts
the files of the package in the place their concern belongs to.

### Fixed

- **The nested objects of a declaration are typed.** The beneficiary fund
  selections, the types of a change, the workers of a construction site, the
  insurance numbers, the vehicles, the leaves and the details of the reports
  were carried as raw JSON. They have their own type now, and the fields the
  API may leave out are pointers.
- The inner object of a nested object is named after the declaration type by
  the API. The renaming happens in the wire layer, so the public type is the
  same for every variant of a family and the caller never has to know the
  name the API expects.
- A field that only some of the variants, or some of the entries of a list,
  carry is left out, the way the API leaves it out.
- A number the API writes with three digits, such as the age of an employee
  of an employment office, is written with three digits again.

### Changed

- **Breaking** The models of the declarations are generated now, out of the
  reference payloads and the schemas of the Ministry. The generator is
  committed under `tools/`, and it writes the wire structs, the public
  models, the translation between them and the test that compares every
  payload with the reference payload of the Ministry.
- **Breaking** The files of the package are organised by concern: the client,
  the options, the request plumbing, the credentials, the errors, the codes,
  the enumerated values, the models, the declarations, the services, the
  validator and the schemas. The wire format and the schemas of the Ministry
  live under `internal/`.

## [0.4.0] - 2026-10-21

The release that covers the declaration types of the platform. Every one of
them is now modelled and verified against the reference payloads the Ministry
publishes.

### Added

- **The change of the employment** (`WebMA`, `WebMAD`), the notification an
  employer files when the details of an employment change.
- **The construction workers** (`WebE12`, `WebE12Ap`), the notification of
  the personnel of a construction site.
- **The internships** (`WebE35`, `WebE35D`, `WebE35T`), the start, the end
  and the change of the schedule of an internship.
- **The parental leave grant** (`WebE14`), the notification of the employer.
- **The urgent hiring** (`FastAna`), the hire that covers an emergency need.
- **The unified hiring notification** (`E3`), the form of Ergani II.
- The annual leave notification (`WebEA`) and the reconciling reports of the
  private employment offices (`WebIgee`) and of the temporary employment
  agencies (`WebEpa`).
- The models of the declarations are generated now. The generator is committed
  under `tools/`, and it writes the wire structs, the public models, the
  translation between them and the test that compares every payload with the
  reference payload of the Ministry.

### Changed

- **Breaking** The models of the employment lifecycle moved to the generated
  files. The names of the fields and of the variants did not change.
- **Breaking** The nested objects of a declaration, the selections of the
  beneficiary funds and the types of a change, are carried as raw JSON. The
  API names their inner object after the declaration type, so the caller has
  to build it.

### Fixed

- The API writes a day and a month with and without their leading zero. The
  dates are read in both forms.
- An amount that is not given is left empty, the way the API leaves it, and
  not written as a zero.

## [0.3.0] - 2026-10-21

The release that completes the API surface. The family of the employment
lifecycle, the variants of the working time organisation, the services that
read the platform back, a validator and the official schemas.

### Added

- **The employment lifecycle declarations.** The four notifications of the
  start of an employment, a hire, a transfer, borrowed staff and a hire for
  a loan-out, and the thirteen notifications of the end of an employment, the
  resignations, the dismissals, the retirements, the death, the end of a
  probation period, the transfer and the end of a fixed term contract.
  Every field carries the official description of the API, and the
  enumerated fields are typed.
- The declaration of a change of the employment relationship for the working
  time organisation, and the declaration of an exemption from the duty of
  advance notification.
- A validator that checks a declaration before it is sent: a movement
  without a date, a movement dated in the future, a working period that
  ends before it starts, a tax identification number that is not nine
  digits.
- The official schemas of the API, embedded in the SDK, and a check of a
  payload against the schema of its type.
- The services that read the platform back: the diary of the actual hours,
  the digital working time status, the current workforce and the monthly
  employment state.
- A `DeclarationError`, which carries the type of the declaration that was
  rejected, and `CodeOf` to read it.
- Integration tests, which run against the trial environment when the
  credentials are in the environment and are skipped otherwise.

### Changed

- **Breaking** The legacy E8 overtime declaration is removed. The API
  rejects its payload on both the trial and the production environment, and
  the action code it used belongs to the first generation of the platform.
  The overtime of a branch is declared with the working time declaration.
- The declarations of the employment lifecycle are a list, because the
  schemas of the API declare the inner element with `maxOccurs="unbounded"`.
- The optional booleans of the employment lifecycle are a pointer, because
  the API leaves them empty when they are not answered.

### Fixed

- The values of the numbers of the employment lifecycle follow the format of
  the API: the thousands are separated by a dot, the decimals by a comma, and
  the hours carry a single decimal.

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

[Unreleased]: https://github.com/takispanag/ergani-go-sdk/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/takispanag/ergani-go-sdk/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/takispanag/ergani-go-sdk/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/takispanag/ergani-go-sdk/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/takispanag/ergani-go-sdk/releases/tag/v0.1.0
