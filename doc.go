// Package ergani provides a client for the Ergani API, the platform of the
// Greek Ministry of Labour where employers declare the movements, the
// overtime and the working time of their employees.
//
// The package covers the whole API surface of the platform: the digital work
// card, the working time organisation, the overtime, the leaves, the
// interoperability services, and the declarations of the start and the end of
// an employment through the generic submission path.
//
// # Client
//
// A Client is created with New and is safe for concurrent use by multiple
// goroutines.
//
//	client, err := ergani.New(username, password)
//	if err != nil {
//		return err
//	}
//
//	client, err = ergani.New(username, password,
//		ergani.WithBaseURL(ergani.ProductionBaseURL),
//		ergani.WithMaxRetries(3),
//	)
//	if err != nil {
//		return err
//	}
//
// The client authenticates on its own before the first call, renews the
// access token when it expires, and retries the transient failures.
//
// # Declarations
//
// Each declaration type has its own method. A declaration groups the records
// of one business branch of one employer.
//
//	submissions, err := client.SubmitWorkCards(ctx, []ergani.WorkCardDeclaration{
//		{
//			EmployerTaxID: "094187530",
//			BranchNumber:  "0",
//			Cards: []ergani.WorkCard{
//				{
//					EmployeeTaxID:    "028233026",
//					EmployeeFirstName: "John",
//					EmployeeLastName:  "Doe",
//					MovementType:      ergani.MovementArrival,
//					ReferenceDate:    time.Now(),
//					MovementDateTime:  time.Now(),
//				},
//			},
//		},
//	})
//
// The models carry no knowledge of the wire format. The formats the API
// expects, which differ from one declaration type to the next, are applied by
// the SDK.
//
// # Environments
//
// The client talks to the trial environment by default, because a mistake
// against the production environment has legal consequences. Use
// [WithBaseURL] with [ProductionBaseURL] to submit against the production
// environment.
//
// # Errors
//
// Every failure is reported as an error. A failure of the API carries an
// [*APIError] with the HTTP status code and the message of the service. Use
// errors.Is with [ErrUnauthorized], [ErrInvalidRequest], [ErrNotFound],
// [ErrRateLimited] or [ErrServer] to classify a failure, and errors.As to
// inspect an [*APIError]. A declaration the SDK rejects before it reaches the
// API carries a [*ValidationError] with the path of the offending field.
package ergani
