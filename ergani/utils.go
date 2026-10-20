package ergani

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// Work card movement types
	ArrivalCode   = "0"
	DepartureCode = "1"

	// Late declaration justification codes
	PowerOutageCode                = "001"
	EmployerSystemsUnavailableCode = "002"
	ErganiSystemsUnavailableCode   = "003"

	// Overtime justification codes
	AccidentPreventionCode         = "001"
	UrgentSeasonalTasksCode        = "002"
	ExceptionalWorkloadCode        = "003"
	SupplementaryTasksCode         = "004"
	LostHoursSuddenCausesCode      = "005"
	LostHoursOfficialHolidaysCode  = "006"
	LostHoursWeatherConditionsCode = "007"
	EmergencyClosureDayCode        = "008"
	NonWorkdayTasksCode            = "009"

	// Schedule work type codes
	WorkFromOfficeCode = "ΕΡΓ"
	WorkFromHomeCode   = "ΤΗΛ"
	RestDayCode        = "ΑΝ"
	AbsentCode         = "ΜΕ"
)

// mapWorkCardMovementType converts a WorkCardMovementType to its string representation
// required by the Ergani API ("0" for Arrival, "1" for Departure).
func mapWorkCardMovementType(t WorkCardMovementType) (string, error) {
	switch t {
	case Arrival:
		return ArrivalCode, nil
	case Departure:
		return DepartureCode, nil
	default:
		return "", fmt.Errorf("invalid WorkCardMovementType: %v", t)
	}
}

// mapLateDeclarationJustification converts a LateDeclarationJustificationType to its
// API string code.
func mapLateDeclarationJustification(j LateDeclarationJustificationType) (string, error) {
	switch j {
	case PowerOutage:
		return PowerOutageCode, nil
	case EmployerSystemsUnavailable:
		return EmployerSystemsUnavailableCode, nil
	case ErganiSystemsUnavailable:
		return ErganiSystemsUnavailableCode, nil
	default:
		return "", fmt.Errorf("invalid LateDeclarationJustificationType: %v", j)
	}
}

// mapOvertimeJustification converts an OvertimeJustificationType to its API string code.
func mapOvertimeJustification(j OvertimeJustificationType) (string, error) {
	switch j {
	case AccidentPreventionOrDamageRestoration:
		return AccidentPreventionCode, nil
	case UrgentSeasonalTasks:
		return UrgentSeasonalTasksCode, nil
	case ExceptionalWorkload:
		return ExceptionalWorkloadCode, nil
	case SupplementaryTasks:
		return SupplementaryTasksCode, nil
	case LostHoursSuddenCauses:
		return LostHoursSuddenCausesCode, nil
	case LostHoursOfficialHolidays:
		return LostHoursOfficialHolidaysCode, nil
	case LostHoursWeatherConditions:
		return LostHoursWeatherConditionsCode, nil
	case EmergencyClosureDay:
		return EmergencyClosureDayCode, nil
	case NonWorkdayTasks:
		return NonWorkdayTasksCode, nil
	default:
		return "", fmt.Errorf("invalid OvertimeJustificationType: %v", j)
	}
}

// mapScheduleWorkType converts a ScheduleWorkType to its API string representation.
func mapScheduleWorkType(t ScheduleWorkType) (string, error) {
	switch t {
	case WorkFromOffice:
		return WorkFromOfficeCode, nil
	case WorkFromHome:
		return WorkFromHomeCode, nil
	case RestDay:
		return RestDayCode, nil
	case Absent:
		return AbsentCode, nil
	default:
		return "", fmt.Errorf("invalid ScheduleWorkType: %v", t)
	}
}

// MarshalJSON is a custom marshaller for the WorkCard struct.
// It ensures that enum types like WorkCardMovementType are converted to their
// correct API string representations before marshaling to JSON.
func (wc WorkCard) MarshalJSON() ([]byte, error) {
	type Alias WorkCard

	movementType, err := mapWorkCardMovementType(wc.WorkCardMovementType)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal WorkCard: %w", err)
	}

	var justification *string
	if wc.LateDeclarationJustification != nil {
		j, err := mapLateDeclarationJustification(*wc.LateDeclarationJustification)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal WorkCard: %w", err)
		}
		justification = &j
	}

	return json.Marshal(&struct {
		WorkCardMovementType string `json:"f_type"`
		*Alias
		LateDeclarationJustification *string `json:"f_aitiologia,omitempty"`
	}{
		WorkCardMovementType:         movementType,
		Alias:                        (*Alias)(&wc),
		LateDeclarationJustification: justification,
	})
}

// MarshalJSON is a custom marshaller for the Overtime struct.
// It converts the OvertimeJustification enum to its API string code.
func (o Overtime) MarshalJSON() ([]byte, error) {
	type Alias Overtime

	justification, err := mapOvertimeJustification(o.OvertimeJustification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Overtime: %w", err)
	}

	return json.Marshal(&struct {
		OvertimeJustification string `json:"f_reason"`
		*Alias
	}{
		OvertimeJustification: justification,
		Alias:                 (*Alias)(&o),
	})
}

// MarshalJSON is a custom marshaller for the WorkdayDetails struct.
// It converts the WorkType enum to its API string representation.
func (wd WorkdayDetails) MarshalJSON() ([]byte, error) {
	type Alias WorkdayDetails

	workType, err := mapScheduleWorkType(wd.WorkType)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal WorkdayDetails: %w", err)
	}

	return json.Marshal(&struct {
		WorkType string `json:"f_type"`
		*Alias
	}{
		WorkType: workType,
		Alias:    (*Alias)(&wd),
	})
}

// parseSubmissionResponse decodes the JSON body of a successful submission response
// from the API into a slice of SubmissionResponse structs.
func parseSubmissionResponse(resp *http.Response) ([]SubmissionResponse, error) {
	// A 204 No Content response is valid and means there's nothing to parse.
	if resp == nil || resp.StatusCode == http.StatusNoContent {
		return []SubmissionResponse{}, nil
	}

	var submissions []SubmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&submissions); err != nil {
		return nil, fmt.Errorf("failed to decode submission response: %w", err)
	}
	return submissions, nil
}
