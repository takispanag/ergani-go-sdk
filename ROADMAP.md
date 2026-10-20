# Roadmap

The work that is still ahead for `ergani-go-sdk`. The items are tracked as
[GitHub issues](https://github.com/takispanag/ergani-go-sdk/issues), so the
progress of each one is visible.

The platform exposes about fifty declaration types. The SDK models every one
of them the Ministry publishes a reference payload for, which is 31 of them,
and submits the rest through the generic path.

## The declarations that are not modelled yet

- The correction variants, such as `WebE14C` and `FastAnaC`. They share the
  payload of the form they correct, so they are a variant of it rather than a
  new type. The official reference implementation follows the same rule: only
  six families have a correction code, and a correction of the other families
  is expressed with the same code and the fields `f_rel_protocol` and
  `f_rel_date`.
- The beneficiaries of the public benefit programmes (`WebE31`) and the
  theoretical training (`WebE32`). The Ministry publishes no reference
  payload for them.

Both are submitted with `Client.Submit`, and `Client.DocumentSchema` returns
the shape the API expects.

## The tooling

- A command line tool that submits a declaration from a file and validates
  it against the schema of the API, next to the example under `cmd/`.
- A validator that checks a declaration against the schedule that was
  declared before: every card stamp paired, the overtime inside the legal
  cap of the working pattern, the leave days reconciled with the monthly
  state. The API offers no such check, and it is what the inspections look
  for.
- The generator keeps the models in step with the schemas of the API: a new
  field of a declaration is a change of one table, and the test that compares
  every payload with the reference payload of the Ministry tells whether the
  change is right.

## The stability of the API

The platform publishes no version of its API. The interoperability guide
carries a document version and a date, and every change comes with an
announcement. The SDK therefore keeps the discovery endpoints first:
`ListSubmissions` and `DocumentSchema` are the authoritative source of the
codes and of the shapes, and the constants of the package are a convenience
that follows them.
