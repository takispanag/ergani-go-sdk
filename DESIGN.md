# Design: ergani-go-sdk v0.2.0

Grounded in the official sources fetched on 2026-10-21:

- **Interop guide v4** (13.02.2026), the current API specification.
- **`HELP_FILES_SAMPLES_trial2.zip`**: 40 official JSON examples, 96 XSD schemas, the official Postman collection.
- **Live probing** of `trialv2eservices.yeka.gr` and `eservices.yeka.gr`.

## 1. The official surface

Base URLs (verified live):

| Environment | Base URL |
|---|---|
| Trial (current) | `https://trialv2eservices.yeka.gr/WebServicesAPI/Api/` |
| Production | `https://eservices.yeka.gr/WebServicesAPI/Api/` |
| Trial (legacy) | `https://trialeservices.yeka.gr/WebServicesAPI/api/` — **dead, site-wide maintenance page** |

Routes (all verified live: 401/405 proves existence, 404 proves absence):

| Method | Path | Purpose |
|---|---|---|
| POST | `/Authentication` | JWT + refresh token |
| POST | `/Authentication/Refresh` | renew the access token |
| POST | `/Authentication/Logout` | revoke the refresh token |
| GET | `/Lookup/Submissions` | list the submission types available to the caller |
| GET | `/Documents/{code}` | JSON schema of a submission type |
| POST | `/Documents/{code}` | submit a declaration |
| POST | `/Documents/CancelSubmittedDocument` | recall a submitted declaration |
| GET | `/Documents/{code}?protocol=&submittedDate=` | the submitted declaration as a Base64 PDF |
| GET | `/WebServices/ServicesList` | interop services and their parameters |
| POST | `/WebServices/ExecuteService` | run an interop service |

`/Documents/{code}` is a catch-all: the auth filter runs before route binding, so
an unknown code answers 401, never 404. A document code cannot be validated by
probing — only `Lookup/Submissions` is authoritative.

## 2. Wire-format defects in v0.1.0

The official examples contradict the current SDK in four places. These are
production bugs: the payload the SDK sends is not the payload the API documents.

| Field | v0.1.0 sends | Official | Fix |
|---|---|---|---|
| `f_aa` (work card) | `1` (int) | `"0"` (string) | `BranchNumber string` |
| `f_aa_pararthmatos` (WTO, overtime) | `1` (int) | `"0"` (string) | `BranchNumber string` |
| `f_weekdates` (overtime) | `5` (int) | `"5"` (string) | `WeekDays string` |
| `f_reference_date` (work card) | `21/10/2026` | `"2022-05-04"` | dedicated ISO date type |

Date formats are **not** uniform across the API, which is why one `Date` type
cannot serve every form:

| Family | Date format | Time format |
|---|---|---|
| `WRKCardSE.f_reference_date` | `YYYY-MM-DD` | — |
| `WRKCardSE.f_date` | `YYYY-MM-DDTHH:MM:SS.fffffff±ZZ:ZZ` | — |
| all `WTO*`, `OvTime`, E3-family dates | `DD/MM/YYYY` | — |
| `f_from`, `f_to`, `f_proslipsitime` | — | `HH:MM` |
| response `submitDate` | `DD/MM/YYYY` | `HH:MM` |
| `submittedDate` query parameter | `yyyyMMdd` | — |
| `refreshTokenExpired` | ISO 8601 with offset | — |

## 3. Scope decision: what gets a typed model

Forty document types ship official examples; the E3/E5/E6/E7 family alone
carries 89–119 fields per form, with a large shared core.

- **Typed models** for the card-sector forms, which is where the volume is:
  work card, daily/weekly schedule, WTO overtime, leave, E8 overtime.
- **A shared employment core** for the E3/E5/E6/E7 family, so the common
  identity and employment fields are typed once and embedded by each form.
- **A generic path** for everything else: `Submit(ctx, code, payload)` over
  any of the forty codes, plus the schema endpoint. Nothing is unreachable.
- **Code constants** for all forty document codes, so callers never type a
  raw string.

This is the only design that reaches full coverage without several thousand
lines of struct tags, and it keeps the typed surface where the traffic is.

## 4. Public API

```go
// Construction
func NewClient(username, password string, opts ...Option) (*Client, error)
func NewClientWithConfig(cfg Config) (*Client, error)

// Auth
func (c *Client) Authenticate(ctx context.Context) (*Token, error)
func (c *Client) Refresh(ctx context.Context) (*Token, error)
func (c *Client) Logout(ctx context.Context) error

// Documents
func (c *Client) ListSubmissions(ctx context.Context) ([]SubmissionType, error)
func (c *Client) DocumentSchema(ctx context.Context, code string) (json.RawMessage, error)
func (c *Client) Submit(ctx context.Context, code string, payload any) ([]Submission, error)
func (c *Client) CancelDocument(ctx context.Context, req CancelRequest) error
func (c *Client) DocumentPDF(ctx context.Context, code, protocol string, submitted time.Time) ([]byte, error)

// Typed forms
func (c *Client) SubmitWorkCard(ctx context.Context, cards []CompanyWorkCard) ([]Submission, error)
func (c *Client) SubmitDailySchedule(...)
func (c *Client) SubmitWeeklySchedule(...)
func (c *Client) SubmitOvertimeSchedule(...)   // WTOOv
func (c *Client) SubmitLeave(...)              // wtoHoliday
func (c *Client) SubmitOvertime(...)           // OvTime (E8)

// Services
func (c *Client) Services(ctx context.Context) ([]Service, error)
func (c *Client) ExecuteService(ctx context.Context, req ServiceRequest) (json.RawMessage, error)
```

Design rules applied:

- `context.Context` is the first parameter of every method that performs I/O.
- The client is safe for concurrent use; the token is guarded by a mutex.
- Errors are typed: `*APIError` carries the status and the message, and
  `errors.Is(err, ErrUnauthorized)` works.
- The API answers with `{"message": ...}` **or** a bare JSON string; the parser
  handles both, and it refuses to parse an HTML error page.
- `Option` functions for the configuration, with `Config` kept for compatibility.
- The access token is refreshed on 401, and the refresh token is used when the
  access token has expired.

## 5. Testing

- Golden files per document type, compared byte for byte, so a change of the
  wire format shows up in the diff of a pull request.
- `httptest` servers for every route, including the 401-refresh path.
- Table tests for the date formats and the error parser.
- `Example` functions for the godoc, which the CI compiles.
- A fuzz target for the error-body parser, which the API keeps changing.
