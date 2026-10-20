# Ergani Go SDK

[![CI](https://github.com/takispanag/ergani-go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/takispanag/ergani-go-sdk/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/takispanag/ergani-go-sdk.svg)](https://pkg.go.dev/github.com/takispanag/ergani-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/takispanag/ergani-go-sdk)](https://goreportcard.com/report/github.com/takispanag/ergani-go-sdk)
[![License: MIT](https://img.shields.io/github/license/takispanag/ergani-go-sdk)](https://opensource.org/licenses/MIT)

`ergani-go-sdk` is a Go SDK for interacting with the API of [Ergani](https://www.gov.gr/en/ipiresies/ergasia-kai-asphalise/apozemioseis-kai-parokhes/prosopopoiemene-plerophorese-misthotou-ergane),
the Greek Ministry of Labour's platform for declaring employee movements,
overtime and working schedules.

It is meant for Greek employers who need to submit those declarations
programmatically instead of filling in the web form by hand. The package has
no third-party dependencies: it only uses the Go standard library, so it drops
into an existing service without pulling anything else in.

## Features

- Submit work cards (arrival and departure declarations)
- Submit overtime declarations
- Submit daily work schedules
- Submit weekly work schedules
- Typed errors, so a rejected payload can be told apart from an expired token
- Automatic re-authentication when the access token expires
- Retries with exponential backoff for transient API failures
- A `HTTPClient` hook, so the client can be tested without touching the network

## Requirements

Go 1.19 or later.

## Installation

```sh
go get github.com/takispanag/ergani-go-sdk
```

## Usage

### Create a client

To create a new Ergani client you have to set your Ergani username, password
and optionally the Ergani API base URL, which defaults to
`https://trialeservices.yeka.gr/WebServicesAPI/api`.

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	config := ergani.Config{
		Username: os.Getenv("ERGANI_USERNAME"),
		Password: os.Getenv("ERGANI_PASSWORD"),
		BaseURL:  os.Getenv("ERGANI_BASE_URL"), // Optional, defaults to the trial URL
		Timeout:  30 * time.Second,             // Optional, defaults to 30 seconds
	}

	client, err := ergani.NewClientWithConfig(config)
	if err != nil {
		panic(err)
	}

	_ = client
	fmt.Println("client ready")
}
```

If you intend to use this package for multiple company entities, it is
necessary to create separate client instances for each entity with the
appropriate credentials.

The shorter form uses the default base URL and timeout:

```go
package main

import (
	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	client, err := ergani.NewClient("username", "password")
	if err != nil {
		panic(err)
	}

	_ = client
}
```

### Work card

Submit work card records to Ergani in order to declare an employee's movement
(arrival, departure).

<!-- readme:skip -->
```go
func (c *Client) SubmitWorkCard(ctx context.Context, cards []CompanyWorkCard) ([]SubmissionResponse, error)
```

#### Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	ctx := context.Background()

	client, err := ergani.NewClient("username", "password")
	if err != nil {
		panic(err)
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	workCard1 := ergani.WorkCard{
		EmployeeTaxID:            "123456789",
		EmployeeFirstName:        "John",
		EmployeeLastName:         "Doe",
		WorkCardMovementType:     ergani.Arrival,
		WorkCardSubmissionDate:   ergani.Date{Time: yesterday},
		WorkCardMovementDateTime: ergani.DateTime{Time: yesterday},
	}

	workCard2 := workCard1
	workCard2.WorkCardMovementType = ergani.Departure
	workCard2.WorkCardMovementDateTime = ergani.DateTime{Time: now}

	companyWorkCard := ergani.CompanyWorkCard{
		EmployerTaxID:        "999999999",
		BusinessBranchNumber: 1,
		CardDetails:          []ergani.WorkCard{workCard1, workCard2},
	}

	responses, err := client.SubmitWorkCard(ctx, []ergani.CompanyWorkCard{companyWorkCard})
	if err != nil {
		panic(err)
	}

	for _, response := range responses {
		fmt.Printf("submitted %s with protocol %s\n", response.ID, response.Protocol)
	}
}
```

**Note**: You can submit work cards for various employees across multiple
company branches simultaneously as shown above.

### Overtime

Submit overtime records to Ergani in order to declare employee overtimes.

<!-- readme:skip -->
```go
func (c *Client) SubmitOvertime(ctx context.Context, overtimes []CompanyOvertime) ([]SubmissionResponse, error)
```

#### Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	ctx := context.Background()

	client, err := ergani.NewClient("username", "password")
	if err != nil {
		panic(err)
	}

	overtimeDate := time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)
	start := time.Date(2025, 7, 10, 18, 0, 0, 0, time.UTC)
	end := time.Date(2025, 7, 10, 20, 0, 0, 0, time.UTC)

	overtime := ergani.Overtime{
		EmployeeTaxID:          "123456789",
		EmployeeSSN:            "987654321",
		EmployeeFirstName:      "John",
		EmployeeLastName:       "Doe",
		OvertimeDate:           ergani.Date{Time: overtimeDate},
		OvertimeStartTime:      ergani.Time{Time: start},
		OvertimeEndTime:        ergani.Time{Time: end},
		OvertimeCancellation:   false,
		EmployeeProfessionCode: "123456",
		OvertimeJustification:  ergani.ExceptionalWorkload,
		WeeklyWorkdaysNumber:   5,
	}

	companyOvertime := ergani.CompanyOvertime{
		BusinessBranchNumber: 1,
		SEPEServiceCode:      "12345",
		PrimaryActivityCode:  "1234",
		BranchActivityCode:   "1234",
		KallikratisCode:      "12345678",
		LegalRepTaxID:        "999999999",
		EmployeeOvertimes:    []ergani.Overtime{overtime},
	}

	responses, err := client.SubmitOvertime(ctx, []ergani.CompanyOvertime{companyOvertime})
	if err != nil {
		panic(err)
	}

	for _, response := range responses {
		fmt.Printf("submitted %s with protocol %s\n", response.ID, response.Protocol)
	}
}
```

### Daily schedule

Submit daily schedules to Ergani in order to declare schedules for employees
that don't have a fixed schedule (e.g. shift workers).

<!-- readme:skip -->
```go
func (c *Client) SubmitDailySchedule(ctx context.Context, schedules []CompanyDailySchedule) ([]SubmissionResponse, error)
```

#### Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	ctx := context.Background()

	client, err := ergani.NewClient("username", "password")
	if err != nil {
		panic(err)
	}

	scheduleDate := time.Date(2025, 7, 11, 0, 0, 0, 0, time.UTC)
	start := time.Date(2025, 7, 11, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 7, 11, 17, 0, 0, 0, time.UTC)

	workdayDetails := ergani.WorkdayDetails{
		WorkType:  ergani.WorkFromOffice,
		StartTime: ergani.Time{Time: start},
		EndTime:   ergani.Time{Time: end},
	}

	employeeSchedule := ergani.EmployeeDailySchedule{
		EmployeeTaxID:     "123456789",
		EmployeeFirstName: "John",
		EmployeeLastName:  "Doe",
		ScheduleDate:      ergani.Date{Time: scheduleDate},
		WorkdayDetails:    []ergani.WorkdayDetails{workdayDetails},
	}

	companySchedule := ergani.CompanyDailySchedule{
		BusinessBranchNumber: 1,
		EmployeeSchedules:    []ergani.EmployeeDailySchedule{employeeSchedule},
	}

	responses, err := client.SubmitDailySchedule(ctx, []ergani.CompanyDailySchedule{companySchedule})
	if err != nil {
		panic(err)
	}

	for _, response := range responses {
		fmt.Printf("submitted %s with protocol %s\n", response.ID, response.Protocol)
	}
}
```

### Weekly schedule

Submit weekly schedules to Ergani in order to declare the working pattern of
an employee for every day of a week.

<!-- readme:skip -->
```go
func (c *Client) SubmitWeeklySchedule(ctx context.Context, schedules []CompanyWeeklySchedule) ([]SubmissionResponse, error)
```

#### Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	ctx := context.Background()

	client, err := ergani.NewClient("username", "password")
	if err != nil {
		panic(err)
	}

	weekStart := time.Date(2025, 7, 7, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2025, 7, 13, 0, 0, 0, 0, time.UTC)
	start := time.Date(2025, 7, 7, 9, 0, 0, 0, time.UTC)
	end := time.Date(2025, 7, 7, 17, 0, 0, 0, time.UTC)

	workdayDetails := ergani.WorkdayDetails{
		WorkType:  ergani.WorkFromOffice,
		StartTime: ergani.Time{Time: start},
		EndTime:   ergani.Time{Time: end},
	}

	employeeSchedule := ergani.EmployeeWeeklySchedule{
		EmployeeTaxID:     "123456789",
		EmployeeFirstName: "John",
		EmployeeLastName:  "Doe",
		ScheduleDay:       ergani.Weekday{Weekday: time.Monday},
		WorkdayDetails:    []ergani.WorkdayDetails{workdayDetails},
	}

	companySchedule := ergani.CompanyWeeklySchedule{
		BusinessBranchNumber: 1,
		StartDate:            ergani.Date{Time: weekStart},
		EndDate:              ergani.Date{Time: weekEnd},
		EmployeeSchedules:    []ergani.EmployeeWeeklySchedule{employeeSchedule},
	}

	responses, err := client.SubmitWeeklySchedule(ctx, []ergani.CompanyWeeklySchedule{companySchedule})
	if err != nil {
		panic(err)
	}

	for _, response := range responses {
		fmt.Printf("submitted %s with protocol %s\n", response.ID, response.Protocol)
	}
}
```

### Error handling

Every failure of the API is reported as an `*ergani.APIError`, which carries the
HTTP status code and the message returned by the service. A rejected token or
a rejected set of credentials additionally matches `ergani.ErrUnauthorized`.

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	ctx := context.Background()

	client, err := ergani.NewClient("username", "password")
	if err != nil {
		panic(err)
	}

	_, err = client.SubmitWorkCard(ctx, nil)

	switch {
	case errors.Is(err, ergani.ErrUnauthorized):
		fmt.Println("the credentials or the access token were rejected")
	case ergani.IsStatusCode(err, 400):
		fmt.Println("the API rejected the payload")
	case err != nil:
		fmt.Println("unexpected error:", err)
	}
}
```

### Retries and timeouts

A request is retried when the API answers with `429` or with a `5xx` status, or
when the connection fails. A `401` triggers a fresh authentication and a new
attempt. The delay between two attempts grows exponentially and honours the
`Retry-After` header.

```go
package main

import (
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	client, err := ergani.NewClientWithConfig(ergani.Config{
		Username:     "username",
		Password:     "password",
		Timeout:      30 * time.Second,
		MaxRetries:   3,           // Zero means ergani.DefaultMaxRetries
		RetryBackoff: time.Second, // Zero means ergani.DefaultRetryBackoff
		UserAgent:    "my-app/1.0",
	})
	if err != nil {
		panic(err)
	}

	_ = client
}
```

### Testing against the SDK

The client accepts any `HTTPClient`, which makes it easy to test the code that
submits declarations without reaching the real API.

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	client, err := ergani.NewClientWithConfig(ergani.Config{
		Username:   "username",
		Password:   "password",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})
	if err != nil {
		panic(err)
	}

	_ = client
	fmt.Println("client ready")
}
```

## Development

```sh
git clone https://github.com/takispanag/ergani-go-sdk.git
cd ergani-go-sdk

make test        # run the test suite
make test-race   # run the test suite with the race detector
make lint        # run golangci-lint, skipped when it is not installed
make check       # everything the CI runs
```

`make check` also compiles every Go example of this file, so the documentation
cannot drift away from the code.

The roadmap of the work still ahead lives in [ROADMAP.md](ROADMAP.md).

## Glossary

The glossary might help you if you're taking a look at the official documentation of the Ergani
API (https://eservices.yeka.gr/(S(ayldvlj35eukgvmzrr055oe5))/Announcements.aspx?id=257).

### Work card

| **Original**       | **Original help text** (in Greek)              | **Translated**                        |
|--------------------|------------------------------------------------|---------------------------------------|
| `f_afm_ergodoti`   | Α.Φ.Μ Εργοδότη (Για επαλήθευση)                | `employer_tax_identification_number`  |
| `f_aa`             | Α/Α Παραρτήματος                               | `business_branch_number`              |
| `f_comments`       | ΣΧΟΛΙΑ                                         | `comments`                            |
| `f_afm`            | ΑΡΙΘΜΟΣ ΦΟΡΟΛΟΓΙΚΟΥ ΜΗΤΡΩΟΥ (Α.Φ.Μ.)           | `employee_tax_indentification_number` |
| `f_eponymo`        | ΕΠΩΝΥΜΟ                                        | `employee_last_name`                  |
| `f_onoma`          | ΟΝΟΜΑ                                          | `employee_first_name`                 |
| `f_type`           | Τύπος Κίνησης                                  | `work_card_movement_type`             |
| `f_reference_date` | ΗΜ/ΝΙΑ Αναφοράς                                | `work_card_submission_date`           |
| `f_date`           | ΗΜ/ΝΙΑ Κίνησης                                 | `work_card_movement_datetime`         |
| `f_aitiologia`     | ΚΩΔΙΚΟΣ ΑΙΤΙΟΛΟΓΙΑΣ (Σε περίπτωση Εκπρόθεσμου) | `late_declaration_justification`      |

#### Work card movement types

| **Original API code** | **Original help text** (in Greek) | **Translated** |
|-----------------------|-----------------------------------|----------------|
| `0`                   | ΠΡΟΣΕΛΕΥΣΗ                        | `ARRIVAL`      |
| `1`                   | ΑΠΟΧΩΡΗΣΗ                         | `DEPARTURE`    |

#### Work card justifications

| **Original API code** | **Original help text** (in Greek)           | **Translated**                 |
|-----------------------|---------------------------------------------|--------------------------------|
| `001`                 | ΠΡΟΒΛΗΜΑ ΣΤΗΝ ΗΛΕΚΤΡΟΔΟΤΗΣΗ/ΤΗΛΕΠΙΚΟΙΝΩΝΙΕΣ | `POWER_OUTAGE`                 |
| `002`                 | ΠΡΟΒΛΗΜΑ ΣΤΑ ΣΥΣΤΗΜΑΤΑ ΤΟΥ ΕΡΓΟΔΟΤΗ         | `EMPLOYER_SYSTEMS_UNAVAILABLE` |
| `003`                 | ΠΡΟΒΛΗΜΑ ΣΥΝΔΕΣΗΣ ΜΕ ΤΟ ΠΣ ΕΡΓΑΝΗ           | `ERGANI_SYSTEMS_UNAVAILABLE`   |

### Overtime

| **Original**                 | **Original help text** (in Greek)                 | **Translated**                                   |
|------------------------------|---------------------------------------------------|--------------------------------------------------|
| `f_aa`                       | Α/Α Παραρτήματος                                  | `business_branch_number`                         |
| `f_rel_protocol`             | ΣΧΕΤΙΚΟ ΕΝΤΥΠΟ ΑΡΙΘ. ΠΡΩΤ.	                       | `related_protocol_id`                            |
| `f_rel_date`                 | ΣΧΕΤΙΚΟ ΕΝΤΥΠΟ ΗΜΕΡΟΜΗΝΙΑ	                        | `related_protocol_date`                          |
| `f_ypiresia_sepe`            | ΚΩΔΙΚΟΣ ΥΠΗΡΕΣΙΑΣ ΣΕΠΕ	                           | `sepe_service_code`                              |
| `f_ergodotikh_organwsh`      | ΕΡΓΟΔΟΤΙΚΗ ΟΡΓΑΝΩΣΗ	                              | `employer_organization`                          |
| `f_kad_kyria`                | Κ.Α.Δ. - ΚΥΡΙΑ ΔΡΑΣΤΗΡΙΟΤΗΤΑ	                     | `business_primary_activity_code`                 |
| `f_kad_deyt_1`               | Κ.Α.Δ. - ΚΥΡΙΑ ΔΡΑΣΤΗΡΙΟΤΗΤΑ	1                    | `business_secondary_activity_code_1`             |
| `f_kad_deyt_2`               | Κ.Α.Δ. - ΚΥΡΙΑ ΔΡΑΣΤΗΡΙΟΤΗΤΑ	2                    | `business_secondary_activity_code_2`             |
| `f_kad_deyt_3`               | Κ.Α.Δ. - ΚΥΡΙΑ ΔΡΑΣΤΗΡΙΟΤΗΤΑ	3                    | `business_secondary_activity_code_3`             |
| `f_kad_deyt_4`               | Κ.Α.Δ. - ΚΥΡΙΑ ΔΡΑΣΤΗΡΙΟΤΗΤΑ	4                    | `business_secondary_activity_code_4`             |
| `f_kad_pararthmatos`         | Κ.Α.Δ. ΠΑΡΑΡΤΗΜΑΤΟΣ	                              | `business_brach_activity_code`                   |
| `f_kallikratis_pararthmatos` | ΔΗΜΟΤΙΚΗ / ΤΟΠΙΚΗ ΚΟΙΝΟΤΗΤΑ	                      | `kallikratis_municipal_code`                     |
| `f_comments`                 | ΠΑΡΑΤΗΡΗΣΕΙΣ                                      | `comments`                                       |
| `f_afm_proswpoy`             | Νόμιμος Εκπρόσωπος(Α.Φ.Μ.)                        | `legal_representative_tax_identification_number` |
| `f_afm`                      | ΑΡΙΘΜΟΣ ΦΟΡΟΛΟΓΙΚΟΥ ΜΗΤΡΩΟΥ (Α.Φ.Μ.)              | `employee_tax_indentification_number`            |
| `f_amka`                     | ΑΡΙΘΜΟΣ ΜΗΤΡΩΟΥ ΚΟΙΝΩΝΙΚΗΣ ΑΣΦΑΛΙΣΗΣ (Α.Μ.Κ.Α.)   | `employee_social_security_number`                |
| `f_eponymo`                  | ΕΠΩΝΥΜΟ                                           | `employee_last_name`                             |
| `f_onoma`                    | ΟΝΟΜΑ                                             | `employee_first_name`                            |
| `f_date`                     | ΗΜΕΡΟΜΗΝΙΑ ΥΠΕΡΩΡΙΑΣ	                             | `overtime_date`                                  |
| `f_from`                     | ΩΡΑ ΕΝΑΡΞΗΣ ΥΠΕΡΩΡΙΑΣ (HH24:MM)	                  | `overtime_start_time`                            |
| `f_to`                       | ΩΡΑ ΛΗΞΗΣ ΥΠΕΡΩΡΙΑΣ (HH24:MM)	                    | `overtime_end_time`                              |
| `f_cancellation`             | ΑΚΥΡΩΣΗ ΥΠΕΡΩΡΙΑΣ	                                | `overtime_cancellation`                          |
| `f_step`                     | ΕΙΔΙΚΟΤΗΤΑ ΚΩΔΙΚΟΣ	                               | `employee_profession_code`                       |
| `f_reason`                   | ΑΙΤΙΟΛΟΓΙΑ ΚΩΔΙΚΟΣ	                               | `overtime_justification`                         |
| `f_weekdates`                | ΕΒΔΟΜΑΔΙΑΙΑ ΑΠΑΣΧΟΛΗΣΗ (5) ΠΕΝΘΗΜΕΡΟ (6) ΕΞΑΗΜΕΡΟ | `weekly_workdays_number`                         |
| `f_asee`                     | ΕΓΚΡΙΣΗ ΑΣΕΕ	                                     | `asee_approval`                                  |

#### Overtime justfications

| **Original API code** | **Original help text** (in Greek)                                                 | Translation                                 |
|-----------------------|-----------------------------------------------------------------------------------|---------------------------------------------|
| `001`                 | ΠΡΟΛΗΨΗ ΑΤΥΧΗΜΑΤΩΝ Η ΑΠΟΚΑΤΑΣΤΑΣΗ ΖΗΜΙΩΝ                                          | `ACCIDENT_PREVENTION_OR_DAMAGE_RESTORATION` |
| `002`                 | ΕΠΕΙΓΟΥΣΕΣ ΕΡΓΑΣΙΕΣ ΕΠΟΧΙΑΚΟΥ ΧΑΡΑΚΤΗΡΑ                                           | `URGENT_SEASONAL_TASKS`                     |
| `003`                 | ΕΞΑΙΡΕΤΙΚΗ ΣΩΡΕΥΣΗ ΕΡΓΑΣΙΑΣ – ΦΟΡΤΟΣ ΕΡΓΑΣΙΑΣ                                     | `EXCEPTIONAL_WORKLOAD`                      |
| `004`                 | ΠΡΟΕΠΙΣΚΕΥΑΣΤΙΚΕΣ Η ΣΥΜΠΛΗΡΩΜΑΤΙΚΕΣ ΕΡΓΑΣΙΕΣ                                      | `SUPPLEMENTARY_TASKS`                       |
| `005`                 | ΑΝΑΠΛΗΡΩΣΗ ΧΑΜΕΝΩΝ ΩΡΩΝ ΛΟΓΩ ΞΑΦΝΙΚΩΝ ΑΙΤΙΩΝ Η ΑΝΩΤΕΡΑΣ ΒΙΑΣ                      | `LOST_HOURS_SUDDEN_CAUSES`                  |
| `006`                 | ΑΝΑΠΛΗΡΩΣΗ ΧΑΜΕΝΩΝ ΩΡΩΝ ΛΟΓΩ ΕΠΙΣΗΜΩΝ ΑΡΓΙΩΝ                                      | `LOST_HOURS_OFFICIAL_HOLIDAYS`              |
| `007`                 | ΑΝΑΠΛΗΡΩΣΗ ΧΑΜΕΝΩΝ ΩΡΩΝ ΛΟΓΩ ΚΑΙΡΙΚΩΝ ΣΥΝΘΗΚΩΝ                                    | `LOST_HOURS_WEATHER_CONDITIONS`             |
| `008`                 | ΈΚΤΑΚΤΕΣ ΕΡΓΑΣΙΕΣ ΚΛΕΙΣΙΜΑΤΟΣ ΗΜΕΡΑΣ Η ΜΗΝΑ                                       | `EMERGENCY_CLOSURE_DAY`                     |
| `009`                 | ΛΟΙΠΕΣ ΕΡΓΑΣΙΕΣ ΟΙ ΟΠΟΙΕΣ ΔΕΝ ΜΠΟΡΟΥΝ ΝΑ ΠΡΑΓΜΑΤΟΠΟΙΗΘΟΥΝ ΚΑΤΑ ΤΙΣ ΕΡΓΑΣΙΜΕΣ ΩΡΕΣ | `NON_WORKDAY_TASKS`                         |

### Daily schedule

| **Original**        | **Original help text** (in Greek)    | **Translated**                        |
|---------------------|--------------------------------------|---------------------------------------|
| `f_aa_pararthmatos` | Α/Α ΠΑΡΑΡΤΗΜΑΤΟΣ                     | `business_branch_number`              |
| `f_rel_protocol`    | ΣΧΕΤΙΚΟ ΕΝΤΥΠΟ ΑΡΙΘ. ΠΡΩΤ.           | `related_protocol_id`                 |
| `f_rel_date`        | ΣΧΕΤΙΚΟ ΕΝΤΥΠΟ ΗΜΕΡΟΜΗΝΙΑ            | `related_protocol_date`               |
| `f_comments`        | ΠΑΡΑΤΗΡΗΣΕΙΣ                         | `comments`                            |
| `f_from_date`       | ΗΜΕΡΟΜΗΝΙΑ ΑΠΟ                       | `start_date`                          |
| `f_to_date`         | ΗΜΕΡΟΜΗΝΙΑ ΕΩΣ                       | `end_date`                            |
| `f_afm`             | ΑΡΙΘΜΟΣ ΦΟΡΟΛΟΓΙΚΟΥ ΜΗΤΡΩΟΥ (Α.Φ.Μ.) | `employee_tax_indentification_number` |
| `f_eponymo`         | ΕΠΩΝΥΜΟ                              | `employee_last_name`                  |
| `f_onoma`           | ΟΝΟΜΑ                                | `employee_first_name`                 |
| `f_day`             | ΗΜΕΡΑ                                | `schedule_date`                       |
| `f_type`            | ΤΥΠΟΣ ΑΝΑΛΥΤΙΚΗΣ ΕΓΓΡΑΦΗΣ - ΚΩΔΙΚΟΣ  | `work_type`                           |
| `f_from`            | ΩΡΑ ΑΠΟ (HH24:MM)                    | `start_time`                          |
| `f_to`              | ΩΡΑ ΕΩΣ (HH24:MM)                    | `end_time`                            |

### Weekly schedule

| **Original**        | **Original help text** (in Greek)    | **Translated**                        |
|---------------------|--------------------------------------|---------------------------------------|
| `f_aa_pararthmatos` | Α/Α ΠΑΡΑΡΤΗΜΑΤΟΣ                     | `business_brach_number`               |
| `f_rel_protocol`    | ΣΧΕΤΙΚΟ ΕΝΤΥΠΟ ΑΡΙΘ. ΠΡΩΤ.           | `related_protocol_id`                 |
| `f_rel_date`        | ΣΧΕΤΙΚΟ ΕΝΤΥΠΟ ΗΜΕΡΟΜΗΝΙΑ            | `related_protocol_date`               |
| `f_comments`        | ΠΑΡΑΤΗΡΗΣΕΙΣ                         | `comments`                            |
| `f_from_date`       | ΗΜΕΡΟΜΗΝΙΑ ΑΠΟ                       | `start_date`                          |
| `f_to_date`         | ΗΜΕΡΟΜΗΝΙΑ ΕΩΣ                       | `end_date`                            |
| `f_afm`             | ΑΡΙΘΜΟΣ ΦΟΡΟΛΟΓΙΚΟΥ ΜΗΤΡΩΟΥ (Α.Φ.Μ.) | `employee_tax_indentification_number` |
| `f_eponymo`         | ΕΠΩΝΥΜΟ                              | `employee_last_name`                  |
| `f_onoma`           | ΟΝΟΜΑ                                | `employee_first_name`                 |
| `f_date`            | ΗΜΕΡΟΜΗΝΙΑ                           | `schedule_date`                       |
| `f_type`            | ΤΥΠΟΣ ΑΝΑΛΥΤΙΚΗΣ ΕΓΓΡΑΦΗΣ - ΚΩΔΙΚΟΣ  | `work_type`                           |
| `f_from`            | ΩΡΑ ΑΠΟ (HH24:MM)                    | `start_time`                          |
| `f_to`              | ΩΡΑ ΕΩΣ (HH24:MM)                    | `end_time`                            |

### Schedule work types

| **Original API code** | **Original help text** (in Greek) | **Translated**     |
|-----------------------|-----------------------------------|--------------------|
| `ΜΕ`                  | ΜΗ ΕΡΓΑΣΙΑ                        | `ABSENT`           |
| `ΑΝ`                  | ΑΝΑΠΑΥΣΗ/ΡΕΠΟ                     | `REST_DAY`         |
| `ΤΗΛ`                 | ΤΗΛΕΡΓΑΣΙΑ                        | `WORK_FROM_HOME`   |
| `ΕΡΓ`                 | ΕΡΓΑΣΙΑ                           | `WORK_FROM_OFFICE` |


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
