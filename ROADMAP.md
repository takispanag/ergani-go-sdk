# Roadmap

The work that is still ahead for `ergani-go-sdk`. The items are tracked as
[GitHub issues](https://github.com/takispanag/ergani-go-sdk/issues), so the
progress of each one is visible.

The platform exposes about fifty declaration types. The SDK models the card
sector and the lifecycle of the employment, and submits the rest through the
generic path.

## The declarations that are not modelled yet

- The special categories: the construction workers (`WebE12`, `WebE12Ap`),
  the internships (`WebE35`), the beneficiaries of the public benefit
  programmes (`WebE31`) and the theoretical training (`WebE32`). Their
  payloads are small, and the schemas of the Ministry describe them.
- The change of the employment (`WebMA`, `WebMAD`), which shares its core
  with the hiring declaration and adds the reason of the change and the
  types of the change.
- The remaining working time variants: the retrospective daily schedule
  (`WTODailyA`) and the one of the drivers (`WTODailyD`), the retrospective
  overtime (`WTOOvA`) and the one of the drivers (`WTOOvD`).

## Reading the platform back

- A typed model of the monthly employment state (`EX_BASE_04`), the payload
  with the counters of every kind of leave, of the overtime and of the work
  card. It is what the labour inspector sees, and it is what a declaration
  has to be reconciled against.
- The employer and the branch master data (`EX_BASE_01`, `EX_BASE_02`), with
  a typed model rather than the raw payload.
- The acceptance status of the essential terms (`EX_BASE_06`).

## The tooling

- A command line tool that submits a declaration from a file and validates
  it against the schema of the API, next to the example under `cmd/`.
- A validator that checks a declaration against the schedule that was
  declared before: every card stamp paired, the overtime inside the legal
  cap of the working pattern, the leave days reconciled with the monthly
  state. The API offers no such check, and it is what the inspections look
  for.
- A generator that keeps the models of the declarations in step with the
  schemas, so that a new field of the API is a change of one table rather
  than of a hundred lines.

## The stability of the API

The platform publishes no version of its API. The interoperability guide
carries a document version and a date, and every change comes with an
announcement. The SDK therefore keeps the discovery endpoints first:
`ListSubmissions` and `DocumentSchema` are the authoritative source of the
codes and of the shapes, and the constants of the package are a convenience
that follows them.
