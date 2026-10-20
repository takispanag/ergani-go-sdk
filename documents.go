package ergani

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/takispanag/ergani-go-sdk/internal/wire"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SubmitWorkCards submits work card declarations, which report the arrival
// and the departure of the employees of one business branch.
//
// The returned slice holds one entry per accepted declaration. It is nil when
// the API accepted the declaration but returned no record.
//
// It returns a [*ValidationError] when a declaration is malformed, and an
// [*APIError] when the API rejects it.
//
// The submission is not idempotent. When the API answers with a transient
// failure after it has recorded the declaration, the retry may record it a
// second time.
func (c *Client) SubmitWorkCards(ctx context.Context, declarations []WorkCardDeclaration) ([]Submission, error) {
	payload, err := buildWorkCards(declarations)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, CodeWorkCard, payload)
}

// SubmitDailySchedules submits daily schedule declarations, which report the
// working hours of the employees of one business branch for one day.
func (c *Client) SubmitDailySchedules(ctx context.Context, declarations []DailyScheduleDeclaration) ([]Submission, error) {
	payload, err := buildDailySchedules(declarations)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, CodeScheduleDaily, payload)
}

// SubmitWeeklySchedules submits weekly schedule declarations, which report
// the working pattern of the employees of one business branch for every day
// of a week.
func (c *Client) SubmitWeeklySchedules(ctx context.Context, declarations []WeeklyScheduleDeclaration) ([]Submission, error) {
	payload, err := buildWeeklySchedules(declarations)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, CodeScheduleWeekly, payload)
}

// SubmitOvertimeSchedules submits overtime working time declarations, which
// report the overtime of the employees of one business branch as working
// periods.
func (c *Client) SubmitOvertimeSchedules(ctx context.Context, declarations []OvertimeScheduleDeclaration) ([]Submission, error) {
	payload, err := buildOvertimeSchedules(declarations)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, CodeScheduleOvertime, payload)
}

// SubmitLeaves submits leave declarations, which report the leave of the
// employees of one business branch.
func (c *Client) SubmitLeaves(ctx context.Context, declarations []LeaveDeclaration) ([]Submission, error) {
	payload, err := buildLeaves(declarations)
	if err != nil {
		return nil, err
	}

	return c.submit(ctx, CodeLeave, payload)
}

// Submit submits a declaration of any type.
//
// The payload is sent as it is, so it must have the shape the API expects for
// the given code. The schema of a code is available from
// [Client.DocumentSchema].
//
// The typed methods of the client are the better choice when they cover the
// declaration type. This method exists for the types the SDK does not model
// yet, such as the E3, E5, E6 and E7 families.
func (c *Client) Submit(ctx context.Context, code DocumentCode, payload any) ([]Submission, error) {
	return c.submit(ctx, code, payload)
}

// ListSubmissions returns the declaration types the connected employer may
// submit.
//
// The API decides which types are available, and it adds them over time.
func (c *Client) ListSubmissions(ctx context.Context) ([]SubmissionType, error) {
	data, err := c.call(ctx, http.MethodGet, "/Lookup/Submissions", nil, nil, true)
	if err != nil {
		return nil, err
	}

	var response []struct {
		ID          int    `json:"id"`
		Code        string `json:"code"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the submission types: %w", err)
	}

	types := make([]SubmissionType, 0, len(response))
	for _, item := range response {
		types = append(types, SubmissionType{
			ID:          item.ID,
			Code:        DocumentCode(item.Code),
			Description: item.Description,
		})
	}

	return types, nil
}

// DocumentSchema returns the JSON schema of a declaration type, as the API
// describes it.
//
// The schema is a template: the values of the fields are the names of the
// fields themselves.
func (c *Client) DocumentSchema(ctx context.Context, code DocumentCode) (json.RawMessage, error) {
	data, err := c.call(ctx, http.MethodGet, "/Documents/"+string(code), nil, nil, true)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

// CancelRequest identifies a declaration to recall.
type CancelRequest struct {
	// TypeOfDocument identifies the declaration type.
	//
	// The official example of the API uses the numeric identifier of the
	// declaration type, zero padded to five digits, such as "00009". The
	// code of the type, such as "WTOLeave", is accepted as well.
	// [SubmissionType.DocumentID] produces the numeric form of the type
	// that [Client.ListSubmissions] reports.
	TypeOfDocument string

	// Protocol is the protocol number of the declaration.
	Protocol string

	// SubmittedDate is the day the declaration was submitted on.
	SubmittedDate time.Time
}

// CancelDocument recalls a submitted declaration.
//
// The API supports it for the leave declarations and for the correction of a
// leave declaration.
func (c *Client) CancelDocument(ctx context.Context, req CancelRequest) error {
	if req.TypeOfDocument == "" {
		return &ValidationError{Field: "typeOfDocument", Reason: "the identifier of the declaration type is required"}
	}
	if req.Protocol == "" {
		return &ValidationError{Field: "protocol", Reason: "the protocol number is required"}
	}
	if req.SubmittedDate.IsZero() {
		return &ValidationError{Field: "submittedDate", Reason: "the date of the submission is required"}
	}

	body, err := encode(cancelRequest{
		TypeOfDocument: req.TypeOfDocument,
		Protocol:       req.Protocol,
		SubmittedDate:  req.SubmittedDate.Format("20060102"),
	})
	if err != nil {
		return err
	}

	_, err = c.call(ctx, http.MethodPost, "/Documents/CancelSubmittedDocument", nil, body, true)

	return err
}

// cancelRequest is the body of the recall call.
type cancelRequest struct {
	TypeOfDocument string `json:"TypeOfDocument"`
	Protocol       string `json:"Protocol"`
	SubmittedDate  string `json:"SubmittedDate"`
}

// DocumentPDF returns the submitted declaration as a PDF.
//
// The API returns the document Base64 encoded, and the SDK decodes it.
func (c *Client) DocumentPDF(ctx context.Context, code DocumentCode, protocol string, submitted time.Time) ([]byte, error) {
	if protocol == "" {
		return nil, &ValidationError{Field: "protocol", Reason: "the protocol number is required"}
	}
	if submitted.IsZero() {
		return nil, &ValidationError{Field: "submitted", Reason: "the date of the submission is required"}
	}

	query := url.Values{}
	query.Set("protocol", protocol)
	query.Set("submittedDate", submitted.Format("20060102"))

	data, err := c.call(ctx, http.MethodGet, "/Documents/"+string(code), query, nil, true)
	if err != nil {
		return nil, err
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the document of the declaration: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the document of the declaration: %w", err)
	}

	return decoded, nil
}

// submit sends a payload to the API and decodes the acknowledgement.
//
// A failure of the API is wrapped in a [DeclarationError], so that the caller
// knows which type of declaration was rejected.
func (c *Client) submit(ctx context.Context, code DocumentCode, payload any) ([]Submission, error) {
	body, err := encode(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.call(ctx, http.MethodPost, "/Documents/"+string(code), nil, body, true)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			return nil, &DeclarationError{Code: code, Err: apiErr}
		}
		return nil, err
	}

	return parseSubmissions(data)
}

// submissionResponse is the acknowledgement of a submission.
type submissionResponse struct {
	ID         string `json:"id"`
	Protocol   string `json:"protocol"`
	SubmitDate string `json:"submitDate"`
}

// parseSubmissions decodes the acknowledgement of a submission.
//
// A date the API does not report, or reports in an unexpected format, leaves
// the submission without one rather than failing the whole call.
func parseSubmissions(data []byte) ([]Submission, error) {
	var responses []submissionResponse
	if err := json.Unmarshal(data, &responses); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the response of the submission: %w", err)
	}

	if len(responses) == 0 {
		return nil, nil
	}

	submissions := make([]Submission, 0, len(responses))
	for _, response := range responses {
		submission := Submission{ID: response.ID, Protocol: response.Protocol}
		if submitted, err := time.Parse("02/01/2006 15:04", response.SubmitDate); err == nil {
			submission.SubmittedAt = submitted
		}
		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// buildWorkCards turns the public declarations into the payload of the API.
func buildWorkCards(declarations []WorkCardDeclaration) (*wire.WorkCardDeclarations, error) {
	payload := &wire.WorkCardDeclarations{}

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireAFM(field+".employerTaxID", declaration.EmployerTaxID); err != nil {
			return nil, err
		}
		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			return nil, err
		}

		card := wire.CompanyWorkCard{
			EmployerTaxID: declaration.EmployerTaxID,
			BranchNumber:  declaration.BranchNumber,
			Comments:      declaration.Comments,
		}

		for j, movement := range declaration.Cards {
			path := fmt.Sprintf("%s.cards[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", movement.EmployeeTaxID); err != nil {
				return nil, err
			}
			if movement.ReferenceDate.IsZero() {
				return nil, &ValidationError{Field: path + ".referenceDate", Reason: "the reference date is required"}
			}
			if movement.MovementDateTime.IsZero() {
				return nil, &ValidationError{Field: path + ".movementDateTime", Reason: "the moment of the movement is required"}
			}

			code, ok := movement.MovementType.code()
			if !ok {
				return nil, &ValidationError{
					Field:  path + ".movementType",
					Value:  movement.MovementType,
					Reason: "the movement type is not one the API accepts",
				}
			}

			entry := wire.WorkCard{
				EmployeeTaxID:     movement.EmployeeTaxID,
				EmployeeLastName:  movement.EmployeeLastName,
				EmployeeFirstName: movement.EmployeeFirstName,
				MovementType:      code,
				ReferenceDate:     wire.ISODate{Time: movement.ReferenceDate},
				MovementDateTime:  wire.DateTime{Time: movement.MovementDateTime},
			}

			if movement.LateJustification != nil {
				justification, ok := movement.LateJustification.code()
				if !ok {
					return nil, &ValidationError{
						Field:  path + ".lateJustification",
						Value:  *movement.LateJustification,
						Reason: "the justification is not one the API accepts",
					}
				}
				entry.LateJustification = &justification
			}

			card.Details.CardDetails = append(card.Details.CardDetails, entry)
		}

		payload.Cards.Card = append(payload.Cards.Card, card)
	}

	return payload, nil
}

// buildDailySchedules turns the public declarations into the payload of the
// API.
func buildDailySchedules(declarations []DailyScheduleDeclaration) (*wire.ScheduleDeclarations, error) {
	payload := &wire.ScheduleDeclarations{}

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			return nil, err
		}

		branch := wire.CompanySchedule{
			BranchNumber:    declaration.BranchNumber,
			RelatedProtocol: declaration.RelatedProtocol,
			RelatedDate:     wire.Date{Time: declaration.RelatedDate},
			Comments:        declaration.Comments,
			FromDate:        wire.Date{Time: declaration.StartDate},
			ToDate:          wire.Date{Time: declaration.EndDate},
		}

		for j, schedule := range declaration.Schedules {
			path := fmt.Sprintf("%s.schedules[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", schedule.EmployeeTaxID); err != nil {
				return nil, err
			}
			if schedule.Date.IsZero() {
				return nil, &ValidationError{Field: path + ".date", Reason: "the date of the schedule is required"}
			}

			analyses, err := buildWorkdays(path, schedule.Workday)
			if err != nil {
				return nil, err
			}

			date := wire.Date{Time: schedule.Date}
			branch.Employees.WTO = append(branch.Employees.WTO, wire.EmployeeSchedule{
				EmployeeTaxID:     schedule.EmployeeTaxID,
				EmployeeLastName:  schedule.EmployeeLastName,
				EmployeeFirstName: schedule.EmployeeFirstName,
				Date:              &date,
				Analytics:         wire.WorkdayAnalytics{Analyses: analyses},
			})
		}

		payload.WTOS.WTO = append(payload.WTOS.WTO, branch)
	}

	return payload, nil
}

// buildWeeklySchedules turns the public declarations into the payload of the
// API.
func buildWeeklySchedules(declarations []WeeklyScheduleDeclaration) (*wire.ScheduleDeclarations, error) {
	payload := &wire.ScheduleDeclarations{}

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			return nil, err
		}

		branch := wire.CompanySchedule{
			BranchNumber:    declaration.BranchNumber,
			RelatedProtocol: declaration.RelatedProtocol,
			RelatedDate:     wire.Date{Time: declaration.RelatedDate},
			Comments:        declaration.Comments,
			FromDate:        wire.Date{Time: declaration.StartDate},
			ToDate:          wire.Date{Time: declaration.EndDate},
		}

		for j, schedule := range declaration.Schedules {
			path := fmt.Sprintf("%s.schedules[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", schedule.EmployeeTaxID); err != nil {
				return nil, err
			}

			analyses, err := buildWorkdays(path, schedule.Workday)
			if err != nil {
				return nil, err
			}

			day := wire.Day(strconv.Itoa(int(schedule.Day)))
			branch.Employees.WTO = append(branch.Employees.WTO, wire.EmployeeSchedule{
				EmployeeTaxID:     schedule.EmployeeTaxID,
				EmployeeLastName:  schedule.EmployeeLastName,
				EmployeeFirstName: schedule.EmployeeFirstName,
				Day:               &day,
				Analytics:         wire.WorkdayAnalytics{Analyses: analyses},
			})
		}

		payload.WTOS.WTO = append(payload.WTOS.WTO, branch)
	}

	return payload, nil
}

// buildOvertimeSchedules turns the public declarations into the payload of the
// API.
func buildOvertimeSchedules(declarations []OvertimeScheduleDeclaration) (*wire.ScheduleDeclarations, error) {
	payload := &wire.ScheduleDeclarations{}

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			return nil, err
		}

		branch := wire.CompanySchedule{
			BranchNumber: declaration.BranchNumber,
			Comments:     declaration.Comments,
			FromDate:     wire.Date{Time: declaration.StartDate},
			ToDate:       wire.Date{Time: declaration.EndDate},
		}

		for j, schedule := range declaration.Schedules {
			path := fmt.Sprintf("%s.schedules[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", schedule.EmployeeTaxID); err != nil {
				return nil, err
			}
			if schedule.Date.IsZero() {
				return nil, &ValidationError{Field: path + ".date", Reason: "the date of the overtime is required"}
			}

			analyses, err := buildWorkdays(path, schedule.Workday)
			if err != nil {
				return nil, err
			}

			date := wire.Date{Time: schedule.Date}
			branch.Employees.WTO = append(branch.Employees.WTO, wire.EmployeeSchedule{
				EmployeeTaxID:     schedule.EmployeeTaxID,
				EmployeeLastName:  schedule.EmployeeLastName,
				EmployeeFirstName: schedule.EmployeeFirstName,
				Date:              &date,
				Analytics:         wire.WorkdayAnalytics{Analyses: analyses},
			})
		}

		payload.WTOS.WTO = append(payload.WTOS.WTO, branch)
	}

	return payload, nil
}

// buildLeaves turns the public declarations into the payload of the API.
func buildLeaves(declarations []LeaveDeclaration) (*wire.ScheduleDeclarations, error) {
	payload := &wire.ScheduleDeclarations{}

	for i, declaration := range declarations {
		field := fmt.Sprintf("declarations[%d]", i)

		if err := requireBranch(field+".branchNumber", declaration.BranchNumber); err != nil {
			return nil, err
		}

		branch := wire.CompanySchedule{
			BranchNumber: declaration.BranchNumber,
			Comments:     declaration.Comments,
			FromDate:     wire.Date{Time: declaration.StartDate},
			ToDate:       wire.Date{Time: declaration.EndDate},
		}

		for j, leave := range declaration.Schedules {
			path := fmt.Sprintf("%s.schedules[%d]", field, j)

			if err := requireAFM(path+".employeeTaxID", leave.EmployeeTaxID); err != nil {
				return nil, err
			}
			if leave.Date.IsZero() {
				return nil, &ValidationError{Field: path + ".date", Reason: "the date of the leave is required"}
			}

			code, ok := leave.LeaveType.code()
			if !ok {
				return nil, &ValidationError{
					Field:  path + ".leaveType",
					Value:  leave.LeaveType,
					Reason: "the leave type is not one the API accepts",
				}
			}

			analysis := wire.WorkdayAnalysis{WorkType: code}
			if leave.Year > 0 {
				analysis.Year = strconv.Itoa(leave.Year)
			}
			if leave.EntitledDays > 0 {
				analysis.EntitledDays = fmt.Sprintf("%03d", leave.EntitledDays)
			}

			date := wire.Date{Time: leave.Date}
			branch.Employees.WTO = append(branch.Employees.WTO, wire.EmployeeSchedule{
				EmployeeTaxID:     leave.EmployeeTaxID,
				EmployeeLastName:  leave.EmployeeLastName,
				EmployeeFirstName: leave.EmployeeFirstName,
				Date:              &date,
				Analytics:         wire.WorkdayAnalytics{Analyses: []wire.WorkdayAnalysis{analysis}},
			})
		}

		payload.WTOS.WTO = append(payload.WTOS.WTO, branch)
	}

	return payload, nil
}

// buildWorkdays turns the working periods of a day into the analyses of the
// API.
func buildWorkdays(field string, workdays []Workday) ([]wire.WorkdayAnalysis, error) {
	analyses := make([]wire.WorkdayAnalysis, 0, len(workdays))

	for i, workday := range workdays {
		path := fmt.Sprintf("%s.workday[%d]", field, i)

		code, ok := workday.WorkType.code()
		if !ok {
			return nil, &ValidationError{
				Field:  path + ".workType",
				Value:  workday.WorkType,
				Reason: "the working time type is not one the API accepts",
			}
		}

		analyses = append(analyses, wire.WorkdayAnalysis{
			WorkType: code,
			FromTime: wire.Time{Time: workday.StartTime},
			ToTime:   wire.Time{Time: workday.EndTime},
		})
	}

	return analyses, nil
}

// requireAFM checks that a tax identification number has the nine digits the
// schema asks for.
func requireAFM(field, value string) error {
	if len(value) != 9 {
		return &ValidationError{
			Field:  field,
			Value:  value,
			Reason: "the tax identification number must have 9 digits",
		}
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return &ValidationError{
				Field:  field,
				Value:  value,
				Reason: "the tax identification number must have 9 digits",
			}
		}
	}

	return nil
}

// requireBranch checks that a branch number is set.
func requireBranch(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Reason: "the branch number is required"}
	}

	return nil
}
