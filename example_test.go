package ergani_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/takispanag/ergani-go-sdk"
)

// ExampleNew shows how to create a client.
func ExampleNew() {
	client, err := ergani.New("username", "password")
	if err != nil {
		log.Fatal(err)
	}

	_ = client
}

// ExampleNew_options shows how to configure a client.
func ExampleNew_options() {
	client, err := ergani.New("username", "password",
		// The declarations have legal consequences, so the trial
		// environment is the default. The production environment has to be
		// asked for.
		ergani.WithBaseURL(ergani.ProductionBaseURL),
		ergani.WithTimeout(15*time.Second),
		ergani.WithMaxRetries(3),
		ergani.WithRetryBackoff(time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client
}

// ExampleClient_SubmitWorkCards shows how to declare the arrival and the
// departure of an employee.
func ExampleClient_SubmitWorkCards() {
	client, err := ergani.New("username", "password")
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()

	submissions, err := client.SubmitWorkCards(context.Background(), []ergani.WorkCardDeclaration{
		{
			EmployerTaxID: "094187530",
			BranchNumber:  "0",
			Comments:      "Daily declaration",
			Cards: []ergani.WorkCard{
				{
					EmployeeTaxID:     "028233026",
					EmployeeFirstName: "John",
					EmployeeLastName:  "Doe",
					MovementType:      ergani.MovementArrival,
					ReferenceDate:     now,
					MovementDateTime:  now,
				},
				{
					EmployeeTaxID:     "028233026",
					EmployeeFirstName: "John",
					EmployeeLastName:  "Doe",
					MovementType:      ergani.MovementDeparture,
					ReferenceDate:     now,
					MovementDateTime:  now,
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, submission := range submissions {
		log.Printf("submitted %s with protocol %s", submission.ID, submission.Protocol)
	}
}

// ExampleClient_SubmitDailySchedules shows how to declare the working time
// of an employee for a day.
func ExampleClient_SubmitDailySchedules() {
	client, err := ergani.New("username", "password")
	if err != nil {
		log.Fatal(err)
	}

	day := time.Date(2025, 7, 11, 0, 0, 0, 0, time.UTC)

	_, err = client.SubmitDailySchedules(context.Background(), []ergani.DailyScheduleDeclaration{
		{
			BranchNumber: "0",
			Schedules: []ergani.EmployeeDailySchedule{
				{
					EmployeeTaxID:     "028233026",
					EmployeeFirstName: "John",
					EmployeeLastName:  "Doe",
					Date:              day,
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
	})
	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_ListSubmissions shows how to ask the API which declarations
// the connected employer may submit.
func ExampleClient_ListSubmissions() {
	client, err := ergani.New("username", "password")
	if err != nil {
		log.Fatal(err)
	}

	types, err := client.ListSubmissions(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, declarationType := range types {
		log.Printf("%s: %s", declarationType.Code, declarationType.Description)
	}
}

// ExampleStatusCode shows how to tell the failures apart.
func ExampleStatusCode() {
	err := &ergani.APIError{StatusCode: http.StatusBadRequest, Message: "the AFM does not match"}

	switch {
	case errors.Is(err, ergani.ErrUnauthorized):
		fmt.Println("the credentials were rejected")
	case errors.Is(err, ergani.ErrInvalidRequest):
		fmt.Println("the API rejected the payload")
	default:
		fmt.Println("status", ergani.StatusCode(err))
	}

	// Output: the API rejected the payload
}

// ExampleMovementType_Valid shows how to check a value before it is sent.
func ExampleMovementType_Valid() {
	var movement ergani.MovementType = "SLEEP"

	fmt.Println(movement.Valid())

	// Output: false
}

// ExampleWithHTTPClient shows how to use a custom HTTP client, for a proxy
// or for a transport with its own settings.
func ExampleWithHTTPClient() {
	client, err := ergani.New("username", "password",
		ergani.WithHTTPClient(&http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client
}

// ExampleWithLogger shows how to report the requests of the client. The
// credentials and the identifiers of the employees never reach the log.
func ExampleWithLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := ergani.New("username", "password", ergani.WithLogger(logger))
	if err != nil {
		log.Fatal(err)
	}

	_ = client
}
