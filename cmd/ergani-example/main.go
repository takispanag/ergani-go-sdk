// Command ergani-example submits a work card declaration to the Ergani API.
//
// It is a runnable example of the SDK. The credentials come from the
// environment, and the declaration from the flags.
//
//	export ERGANI_USERNAME=...
//	export ERGANI_PASSWORD=...
//	go run ./cmd/ergani-example \
//		-employer 094187530 \
//		-employee 028233026 \
//		-first-name "John" \
//		-last-name "Doe" \
//		-movement arrival
//
// The client talks to the trial environment by default. Add
// -production to submit against the production environment.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/takispanag/ergani-go-sdk"
)

func main() {
	var (
		employer   = flag.String("employer", "", "tax identification number of the employer")
		employee   = flag.String("employee", "", "tax identification number of the employee")
		firstName  = flag.String("first-name", "", "first name of the employee")
		lastName   = flag.String("last-name", "", "last name of the employee")
		branch     = flag.String("branch", "0", "number of the branch of the employer")
		movement   = flag.String("movement", "arrival", "arrival or departure")
		production = flag.Bool("production", false, "submit against the production environment")
	)
	flag.Parse()

	if err := run(*employer, *employee, *firstName, *lastName, *branch, *movement, *production); err != nil {
		log.Fatalf("ergani-example: %v", err)
	}
}

func run(employer, employee, firstName, lastName, branch, movement string, production bool) error {
	if employer == "" || employee == "" {
		return fmt.Errorf("the employer and the employee tax identification numbers are required")
	}

	username := os.Getenv("ERGANI_USERNAME")
	password := os.Getenv("ERGANI_PASSWORD")
	if username == "" || password == "" {
		return fmt.Errorf("ERGANI_USERNAME and ERGANI_PASSWORD must be set")
	}

	options := []ergani.Option{}
	if production {
		options = append(options, ergani.WithBaseURL(ergani.ProductionBaseURL))
	}

	client, err := ergani.New(username, password, options...)
	if err != nil {
		return fmt.Errorf("failed to create the client: %w", err)
	}

	kind := ergani.MovementArrival
	if movement == "departure" {
		kind = ergani.MovementDeparture
	}

	now := time.Now()

	submissions, err := client.SubmitWorkCards(context.Background(), []ergani.WorkCardDeclaration{
		{
			EmployerTaxID: employer,
			BranchNumber:  branch,
			Comments:      "Submitted by ergani-example",
			Cards: []ergani.WorkCard{
				{
					EmployeeTaxID:     employee,
					EmployeeFirstName: firstName,
					EmployeeLastName:  lastName,
					MovementType:      kind,
					ReferenceDate:     now,
					MovementDateTime:  now,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to submit the work card: %w", err)
	}

	for _, submission := range submissions {
		fmt.Printf("submitted %s with protocol %s on %s\n",
			submission.ID, submission.Protocol, submission.SubmittedAt.Format(time.RFC3339))
	}

	return nil
}
