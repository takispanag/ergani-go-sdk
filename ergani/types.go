package ergani

import (
	"encoding/json"
	"time"
)

// WorkCardMovementType defines the type of work card movement (arrival or departure).
type WorkCardMovementType string

const (
	// Arrival signifies an employee clocking in.
	Arrival WorkCardMovementType = "ARRIVAL"
	// Departure signifies an employee clocking out.
	Departure WorkCardMovementType = "DEPARTURE"
)

// LateDeclarationJustificationType defines the official reason for a late submission
// of a work card entry.
type LateDeclarationJustificationType string

const (
	// PowerOutage signifies a power outage as the reason.
	PowerOutage LateDeclarationJustificationType = "POWER_OUTAGE"
	// EmployerSystemsUnavailable signifies a failure of the employer's IT systems.
	EmployerSystemsUnavailable LateDeclarationJustificationType = "EMPLOYER_SYSTEMS_UNAVAILABLE"
	// ErganiSystemsUnavailable signifies a failure of the Ergani IT systems.
	ErganiSystemsUnavailable LateDeclarationJustificationType = "ERGANI_SYSTEMS_UNAVAILABLE"
)

// OvertimeJustificationType defines the official reason for an employee working overtime.
type OvertimeJustificationType string

const (
	AccidentPreventionOrDamageRestoration OvertimeJustificationType = "ACCIDENT_PREVENTION_OR_DAMAGE_RESTORATION"
	UrgentSeasonalTasks                   OvertimeJustificationType = "URGENT_SEASONAL_TASKS"
	ExceptionalWorkload                   OvertimeJustificationType = "EXCEPTIONAL_WORKLOAD"
	SupplementaryTasks                    OvertimeJustificationType = "SUPPLEMENTARY_TASKS"
	LostHoursSuddenCauses                 OvertimeJustificationType = "LOST_HOURS_SUDDEN_CAUSES"
	LostHoursOfficialHolidays             OvertimeJustificationType = "LOST_HOURS_OFFICIAL_HOLIDAYS"
	LostHoursWeatherConditions            OvertimeJustificationType = "LOST_HOURS_WEATHER_CONDITIONS"
	EmergencyClosureDay                   OvertimeJustificationType = "EMERGENCY_CLOSURE_DAY"
	NonWorkdayTasks                       OvertimeJustificationType = "NON_WORKDAY_TASKS"
)

// ScheduleWorkType defines the type of work activity in a schedule (e.g., office, remote).
type ScheduleWorkType string

const (
	// WorkFromOffice indicates work performed at the employer's premises.
	WorkFromOffice ScheduleWorkType = "WORK_FROM_OFFICE"
	// WorkFromHome indicates remote work (teleworking).
	WorkFromHome ScheduleWorkType = "WORK_FROM_HOME"
	// RestDay indicates a scheduled day off.
	RestDay ScheduleWorkType = "REST_DAY"
	// Absent indicates a planned absence (e.g., leave).
	Absent ScheduleWorkType = "ABSENT"
)

// Custom time/date types for correct JSON formatting as required by the Ergani API.

// Time wraps time.Time to format as "15:04" (HH:MM) for JSON marshaling.
type Time struct{ time.Time }

// MarshalJSON implements the json.Marshaler interface for the Time type.
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format("15:04"))
}

// Date wraps time.Time to format as "02/01/2006" (DD/MM/YYYY) for JSON marshaling.
type Date struct{ time.Time }

// MarshalJSON implements the json.Marshaler interface for the Date type.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("02/01/2006"))
}

// DateTime wraps time.Time to format as ISO 8601 for JSON marshaling.
type DateTime struct{ time.Time }

// MarshalJSON implements the json.Marshaler interface for the DateTime type.
func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02T15:04:05.999Z07:00"))
}

// Bool wraps bool to format as "0" (false) or "1" (true) for JSON marshaling.
type Bool bool

// MarshalJSON implements the json.Marshaler interface for the Bool type.
func (b Bool) MarshalJSON() ([]byte, error) {
	if b {
		return json.Marshal("1")
	}
	return json.Marshal("0")
}

// Weekday wraps time.Weekday for custom JSON marshaling.
type Weekday struct{ time.Weekday }

// MarshalJSON implements the json.Marshaler interface for the Weekday type.
// The Ergani API expects Sunday as 0, Monday as 1, ..., Saturday as 6, which
// matches Go's time.Weekday integer representation.
func (w Weekday) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(w.Weekday))
}
