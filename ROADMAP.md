# Roadmap

The work that is still ahead for `ergani-go-sdk`. The items are tracked as
[GitHub issues](https://github.com/takispanag/ergani-go-sdk/issues), so the
progress of each one is visible.

The platform exposes about fifty declaration types. The SDK models the card
sector, which is where the traffic is, and submits the rest through the
generic path. The work below fills the gap.

## The lifecycle of the employment

- **Hiring (`WebE3N`)** with the shared core of the family: the identity of
  the employee, the insurance funds, the working time, the place of work and
  the attachments. The reference payload of the Ministry has ninety fields.
- **The change of the employment (`WebMA`)** and its variant for borrowed
  staff (`WebMAD`), with the code of the reason of the change.
- **The end of the employment**, the fourteen variants of the digital
  termination announcement: the voluntary resignation, the dismissal with and
  without notice, the retirement, the end of a fixed term contract, the
  transfer and the loan-out. They share one payload and a small
  discriminator, so one type and one enum cover all of them.
- **The special categories**: the construction workers, the internships, the
  beneficiaries of the public benefit programmes and the theoretical training.

## The working time organisation

- The retrospective variants of the daily schedule (`WTODailyA`) and of the
  overtime (`WTOOvA`), and the variants for the drivers (`WTODailyD`,
  `WTOOvD`). They share the payload of the forms that are already modelled.
- The change of the employment relationship for the working time
  (`WKChgWK`), the vehicle of the census of 2022.
- The exemption from the duty of advance notification (`ExProan`), which the
  branches of the card regime file every month.
- The declaration of the employment on the sixth day (`SixthDay`).

## Reading the platform back

- The monthly employment state (`EX_BASE_04`), the payload the labour
  inspector sees. It holds the counters of every kind of leave, of the
  overtime and of the work card, and it is what a declaration has to be
  reconciled against.
- The real working diary (`EX_BASE_07`) and the digital working time status
  (`EX_BASE_08`), which say what the platform recorded against what was
  declared.
- The acceptance status of the essential terms (`EX_BASE_06`).

## The tooling

- A command line tool that submits a declaration from a file and validates it
  against the schema of the API, next to the example under `cmd/`.
- A pre-flight validator, which checks a declaration against the declared
  schedule before it is sent: every card stamp paired, the overtime inside
  the legal cap, the leave days reconciled with the monthly state. The API
  offers no such check, and it is what the inspections look for.
- The XSD schemas of the Ministry embedded in the repository, so a payload
  can be validated against the schema of the API without a network call.

## The stability of the API

The platform publishes no version of its API. The interoperability guide
carries a document version and a date, and every change comes with an
announcement. The SDK therefore keeps the discovery endpoints first:
`ListSubmissions` and `DocumentSchema` are the authoritative source of the
codes and of the shapes, and the constants of the package are a convenience
that follows them.
