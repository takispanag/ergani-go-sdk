# Ergani Go SDK

`ergani-go-sdk` is a Go SDK for interacting with the API of [Ergani](https://www.gov.gr/en/ipiresies/ergasia-kai-asphalise/apozemioseis-kai-parokhes/prosopopoiemene-plerophorese-misthotou-ergane).

## Inspired by
- [ergani-python](https://github.com/withlogicco/ergani-python-sdk) Python SDK by [LOGIC](https://withlogic.co/)
- [ergani-rust](https://github.com/pavlospt/ergani-rust-sdk) Rust SDK by [@pavlospt](https://github.com/pavlospt)

## Requirements

Go 1.19 or later

## Installation

```bash
go get github.com/takispanag/ergani-go-sdk
```

## Usage

### Create a client

To create a new Ergani client you have to set your Ergani username, password and optionally the Ergani API base URL, that defaults to https://trialeservices.yeka.gr/WebServicesAPI/api.

```go
import (
"context"
"os"
"time"
"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
ctx := context.Background()

config := ergani.Config{
Username: os.Getenv("ERGANI_USERNAME"),
Password: os.Getenv("ERGANI_PASSWORD"),
BaseURL:  os.Getenv("ERGANI_BASE_URL"), // Optional, defaults to trial URL
Timeout:  30 * time.Second,             // Optional, defaults to 30 seconds
}

client, err := ergani.NewClientWithConfig(ctx, config)
if err != nil {
panic(err)
}
}
```

If you intend to use this package for multiple company entities, it is necessary to create separate client instances for each entity with the appropriate credentials.

### Work card

Submit work card records to Ergani in order to declare an employee's movement (arrival, departure).

```go
func (c *Client) SubmitWorkCard(ctx context.Context, cards []CompanyWorkCard) (*Response, error)
```

#### Example

```go
package main

import (
	"context"
	"time"
	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	// Assume client is already initialized
	ctx := context.Background()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	workCard1 := ergani.WorkCard{
		EmployeeTaxIdentificationNumber: "123456789",
		EmployeeFirstName:              "first_name",
		EmployeeLastName:               "last_name",
		WorkCardMovementType:           "ARRIVAL",
		WorkCardSubmissionDate:         now,
		WorkCardMovementDateTime:       now,
	}

	workCard2 := ergani.WorkCard{
		EmployeeTaxIdentificationNumber: "123456789",
		EmployeeFirstName:              "first_name",
		EmployeeLastName:               "last_name",
		WorkCardMovementType:           "DEPARTURE",
		WorkCardSubmissionDate:         now,
		WorkCardMovementDateTime:       now,
	}

	companyWorkCard := ergani.CompanyWorkCard{
		EmployerTaxIdentificationNumber: "123456789",
		BusinessBranchNumber:           1,
		CardDetails:                    []ergani.WorkCard{workCard1, workCard2},
	}

	response, err := client.SubmitWorkCard(ctx, []ergani.CompanyWorkCard{companyWorkCard})
	if err != nil {
		panic(err)
	}
}
```

**Note:** You can submit work cards for various employees across multiple company branches simultaneously as shown above.

### Overtime

Submit overtime records to Ergani in order to declare employees overtimes.

```go
func (c *Client) SubmitOvertime(ctx context.Context, overtimes []CompanyOvertime) (*Response, error)
```

#### Example

```go
package main

import (
	"context"
	"time"
	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	// Assume client is already initialized
	ctx := context.Background()

	now := time.Now()

	overtime := ergani.Overtime{
		EmployeeTaxIdentificationNumber: "123456789",
		EmployeeSocialSecurityNumber:   "123456789",
		EmployeeFirstName:              "first_name",
		EmployeeLastName:               "last_name",
		OvertimeDate:                   now,
		OvertimeStartTime:              "18:00",
		OvertimeEndTime:                "20:00",
		OvertimeCancellation:           false,
		EmployeeProfessionCode:         "123456",
		OvertimeJustification:          "EXCEPTIONAL_WORKLOAD",
		WeeklyWorkdaysNumber:           5,
	}

	companyOvertime := ergani.CompanyOvertime{
		BusinessBranchNumber:                     1,
		SepeServiceCode:                         "12345",
		BusinessPrimaryActivityCode:             "1234",
		BusinessBranchActivityCode:              "1234",
		KallikreatisMunicipalCode:              "12345678",
		LegalRepresentativeTaxIdentificationNumber: "123456789",
		EmployeeOvertimes:                       []ergani.Overtime{overtime},
	}

	response, err := client.SubmitOvertime(ctx, []ergani.CompanyOvertime{companyOvertime})
	if err != nil {
		panic(err)
	}
}
```

### Daily schedule

Submit daily schedules to Ergani in order to declare schedules for employees that don't have a fixed schedule (e.g. shift workers).

```go
func (c *Client) SubmitDailySchedule(ctx context.Context, schedules []CompanyDailySchedule) (*Response, error)
```

#### Example

```go
package main

import (
	"context"
	"time"
	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	// Assume client is already initialized
	ctx := context.Background()

	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)

	workdayDetails := ergani.WorkdayDetails{
		WorkType:  "WORK_FROM_HOME",
		StartTime: "09:00",
		EndTime:   "17:00",
	}

	employeeSchedule := ergani.EmployeeDailySchedule{
		EmployeeTaxIdentificationNumber: "123456789",
		EmployeeFirstName:              "first_name",
		EmployeeLastName:               "last_name",
		ScheduleDate:                   tomorrow,
		WorkdayDetails:                 []ergani.WorkdayDetails{workdayDetails},
	}

	companySchedule := ergani.CompanyDailySchedule{
		BusinessBranchNumber: 1,
		EmployeeSchedules:    []ergani.EmployeeDailySchedule{employeeSchedule},
	}

	response, err := client.SubmitDailySchedule(ctx, []ergani.CompanyDailySchedule{companySchedule})
	if err != nil {
		panic(err)
	}
}
```

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

---
