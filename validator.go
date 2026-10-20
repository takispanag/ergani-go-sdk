package ergani

import (
	"fmt"
	"time"
)

// A validator checks a declaration before it is sent.
//
// The API records a declaration as it is and does not check it against the
// schedule that was declared before, which is what the labour inspection
// looks at. The rules below are the ones that can be checked without asking
// the platform anything.
//
// A validator reports every problem it finds, so that a caller can show them
// all at once.

// Validation problems are reported as [*ValidationError], so that they are
// told apart from the failures of the API.

// ValidateWorkCards checks the work card declarations.
//
// It reports a movement without a reference date, a movement without a
// moment, a movement that is dated in the future, a movement that happened
// before its reference date, a tax identification number that is not nine
// digits, and a movement of an unknown type.
func ValidateWorkCards(declarations []WorkCardDeclaration) []error {
	var problems []error

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireAFM(field+".employerTaxID", declaration.EmployerTaxID); err != nil {
			problems = append(problems, err)
		}
		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			problems = append(problems, err)
		}

		for j, movement := range declaration.Cards {
			path := fmt.Sprintf("%s.cards[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", movement.EmployeeTaxID); err != nil {
				problems = append(problems, err)
			}
			if !movement.MovementType.Valid() {
				problems = append(problems, &ValidationError{
					Field:  path + ".movementType",
					Value:  movement.MovementType,
					Reason: "the movement type is not one the API accepts",
				})
			}
			if movement.ReferenceDate.IsZero() {
				problems = append(problems, &ValidationError{
					Field:  path + ".referenceDate",
					Reason: "the reference date is required",
				})
			}
			if movement.MovementDateTime.IsZero() {
				problems = append(problems, &ValidationError{
					Field:  path + ".movementDateTime",
					Reason: "the moment of the movement is required",
				})
			}
			if !movement.ReferenceDate.IsZero() && !movement.MovementDateTime.IsZero() {
				problems = append(problems, checkDates(path, movement.ReferenceDate, movement.MovementDateTime)...)
			}
		}
	}

	return problems
}

// ValidateDailySchedules checks the daily schedule declarations.
func ValidateDailySchedules(declarations []DailyScheduleDeclaration) []error {
	var problems []error

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			problems = append(problems, err)
		}

		for j, schedule := range declaration.Schedules {
			path := fmt.Sprintf("%s.schedules[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", schedule.EmployeeTaxID); err != nil {
				problems = append(problems, err)
			}
			if schedule.Date.IsZero() {
				problems = append(problems, &ValidationError{
					Field:  path + ".date",
					Reason: "the date of the schedule is required",
				})
			}
			if len(schedule.Workday) == 0 {
				problems = append(problems, &ValidationError{
					Field:  path + ".workday",
					Reason: "the schedule has no working period",
				})
			}
			problems = append(problems, checkWorkdays(path, schedule.Workday)...)
		}
	}

	return problems
}

// ValidateWeeklySchedules checks the weekly schedule declarations.
func ValidateWeeklySchedules(declarations []WeeklyScheduleDeclaration) []error {
	var problems []error

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			problems = append(problems, err)
		}
		if declaration.StartDate.IsZero() {
			problems = append(problems, &ValidationError{
				Field:  field + ".startDate",
				Reason: "the first day of the week is required",
			})
		}
		if declaration.EndDate.IsZero() {
			problems = append(problems, &ValidationError{
				Field:  field + ".endDate",
				Reason: "the last day of the week is required",
			})
		}

		for j, schedule := range declaration.Schedules {
			path := fmt.Sprintf("%s.schedules[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", schedule.EmployeeTaxID); err != nil {
				problems = append(problems, err)
			}
			if len(schedule.Workday) == 0 {
				problems = append(problems, &ValidationError{
					Field:  path + ".workday",
					Reason: "the schedule has no working period",
				})
			}
			problems = append(problems, checkWorkdays(path, schedule.Workday)...)
		}
	}

	return problems
}

// checkWorkdays checks the working periods of a day.
func checkWorkdays(field string, workdays []Workday) []error {
	var problems []error

	for i, workday := range workdays {
		path := fmt.Sprintf("%s.workday[%d]", field, i)

		if !workday.WorkType.Valid() {
			problems = append(problems, &ValidationError{
				Field:  path + ".workType",
				Value:  workday.WorkType,
				Reason: "the working time type is not one the API accepts",
			})
		}
		if workday.StartTime.IsZero() {
			problems = append(problems, &ValidationError{
				Field:  path + ".startTime",
				Reason: "the start of the working period is required",
			})
		}
		if workday.EndTime.IsZero() {
			problems = append(problems, &ValidationError{
				Field:  path + ".endTime",
				Reason: "the end of the working period is required",
			})
		}
		if !workday.StartTime.IsZero() && !workday.EndTime.IsZero() && !workday.EndTime.After(workday.StartTime) {
			problems = append(problems, &ValidationError{
				Field:  path + ".endTime",
				Reason: "the working period ends before it starts",
			})
		}
	}

	return problems
}

// checkDates checks the dates of a movement.
func checkDates(field string, reference time.Time, movement time.Time) []error {
	var problems []error

	if reference.After(time.Now()) {
		problems = append(problems, &ValidationError{
			Field:  field + ".referenceDate",
			Value:  reference,
			Reason: "the reference date is in the future",
		})
	}
	if movement.Before(reference) {
		problems = append(problems, &ValidationError{
			Field:  field + ".movementDateTime",
			Value:  movement,
			Reason: "the movement happened before the day it refers to",
		})
	}

	return problems
}
