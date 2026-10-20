package ergani_test

import (
	"errors"
	"testing"
	"time"

	"github.com/takispanag/ergani-go-sdk"
)

func TestValidateWorkCards(t *testing.T) {
	valid := []ergani.WorkCardDeclaration{
		{
			EmployerTaxID: "094187530",
			BranchNumber:  "0",
			Cards: []ergani.WorkCard{
				{
					EmployeeTaxID:    "028233026",
					MovementType:     ergani.MovementArrival,
					ReferenceDate:    time.Now().AddDate(0, 0, -1),
					MovementDateTime: time.Now().AddDate(0, 0, -1),
				},
			},
		},
	}

	if problems := ergani.ValidateWorkCards(valid); len(problems) != 0 {
		t.Errorf("ValidateWorkCards() found %d problems in a valid declaration: %v", len(problems), problems)
	}

	broken := []ergani.WorkCardDeclaration{
		{
			EmployerTaxID: "0941875",
			BranchNumber:  "",
			Cards: []ergani.WorkCard{
				{
					EmployeeTaxID:    "028233026",
					MovementType:     ergani.MovementType("SLEEP"),
					ReferenceDate:    time.Now().AddDate(0, 0, 1),
					MovementDateTime: time.Now().AddDate(0, 0, -2),
				},
			},
		},
	}

	problems := ergani.ValidateWorkCards(broken)
	if len(problems) < 4 {
		t.Errorf("ValidateWorkCards() found %d problems, want at least 4: %v", len(problems), problems)
	}

	for _, problem := range problems {
		var validationErr *ergani.ValidationError
		if !errors.As(problem, &validationErr) {
			t.Errorf("the problem %v is a %T, want a *ValidationError", problem, problem)
		}
	}
}

func TestValidateDailySchedules(t *testing.T) {
	day := time.Date(2025, 7, 11, 0, 0, 0, 0, time.UTC)

	valid := []ergani.DailyScheduleDeclaration{
		{
			BranchNumber: "0",
			Schedules: []ergani.EmployeeDailySchedule{
				{
					EmployeeTaxID: "028233026",
					Date:          day,
					Workday: []ergani.Workday{
						{
							WorkType:  ergani.WorkTypeOffice,
							StartTime: day.Add(9 * time.Hour),
							EndTime:   day.Add(18 * time.Hour),
						},
					},
				},
			},
		},
	}

	if problems := ergani.ValidateDailySchedules(valid); len(problems) != 0 {
		t.Errorf("ValidateDailySchedules() found %d problems in a valid declaration: %v", len(problems), problems)
	}

	broken := []ergani.DailyScheduleDeclaration{
		{
			BranchNumber: "0",
			Schedules: []ergani.EmployeeDailySchedule{
				{
					EmployeeTaxID: "028233026",
					Workday: []ergani.Workday{
						{
							WorkType:  ergani.WorkTypeOffice,
							StartTime: day.Add(18 * time.Hour),
							EndTime:   day.Add(9 * time.Hour),
						},
					},
				},
			},
		},
	}

	if problems := ergani.ValidateDailySchedules(broken); len(problems) < 2 {
		t.Errorf("ValidateDailySchedules() found %d problems, want at least 2: %v", len(problems), problems)
	}
}

func TestCheckSchema(t *testing.T) {
	if len(ergani.Schemas()) == 0 {
		t.Fatal("the SDK embeds no schema")
	}

	fields := map[string]any{
		"f_afm_ergodoti": "094187530",
		"f_aa":           "0",
		"f_comments":     "",
		"f_not_a_field":  "oops",
	}

	problems := ergani.CheckSchema("Card_v1", fields)
	if len(problems) == 0 {
		t.Error("CheckSchema() found no problem in a payload with an unknown field")
	}
}

func TestDeclarationError(t *testing.T) {
	err := &ergani.DeclarationError{
		Code: ergani.CodeWorkCard,
		Err:  ergani.ErrUnauthorized,
	}

	if !ergani.IsStatusCode(err, 0) {
		t.Error("IsStatusCode() did not accept a declaration error")
	}
	if got := ergani.CodeOf(err); got != ergani.CodeWorkCard {
		t.Errorf("CodeOf() = %q, want %q", got, ergani.CodeWorkCard)
	}
}
