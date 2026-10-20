package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/takispanag/ergani-go-sdk/ergani"
)

func main() {
	// It's recommended to load credentials from environment variables or a secure source.
	username := os.Getenv("ERGANI_USERNAME")
	password := os.Getenv("ERGANI_PASSWORD")

	if username == "" || password == "" {
		log.Fatal("ERGANI_USERNAME and ERGANI_PASSWORD environment variables must be set")
	}

	// Create a context
	ctx := context.Background()

	// 1. Initialize the Ergani client
	// This will authenticate with the provided credentials.
	log.Println("Authenticating with Ergani...")
	client, err := ergani.NewClient(ctx, username, password)
	if err != nil {
		log.Fatalf("Failed to create Ergani client: %v", err)
	}
	log.Println("Authentication successful!")

	// 2. Prepare the data for submission.
	// This is an example of submitting a work card for an employee's arrival.
	powerOutageJustification := ergani.PowerOutage
	workCards := []ergani.CompanyWorkCard{
		{
			EmployerTaxID:        "999999999", // Company's Tax ID
			BusinessBranchNumber: 1,
			Comments:             "API submission from Go SDK.",
			CardDetails: []ergani.WorkCard{
				{
					EmployeeTaxID:            "123456789", // Employee's Tax ID
					EmployeeLastName:         "Papadopoulos",
					EmployeeFirstName:        "Giorgos",
					WorkCardMovementType:     ergani.Arrival,
					WorkCardSubmissionDate:   ergani.Date{Time: time.Now()},
					WorkCardMovementDateTime: ergani.DateTime{Time: time.Now()},
					// Example of including an optional field:
					LateDeclarationJustification: &powerOutageJustification,
				},
				{
					EmployeeTaxID:            "987654321", // Another Employee's Tax ID
					EmployeeLastName:         "Vassiliou",
					EmployeeFirstName:        "Maria",
					WorkCardMovementType:     ergani.Arrival,
					WorkCardSubmissionDate:   ergani.Date{Time: time.Now()},
					WorkCardMovementDateTime: ergani.DateTime{Time: time.Now()},
					// This entry has no late justification.
				},
			},
		},
	}

	// 3. Call the submission method
	log.Println("Submitting work cards...")
	submissionResponses, err := client.SubmitWorkCard(ctx, workCards)
	if err != nil {
		// The custom error types can be inspected for more details
		if apiErr, ok := err.(*ergani.APIError); ok {
			log.Fatalf("API Error occurred. Status: %d, Message: %s, Response: %s", apiErr.StatusCode, apiErr.Message, apiErr.Response)
		}
		log.Fatalf("Failed to submit work card: %v", err)
	}

	// 4. Process the response
	if len(submissionResponses) == 0 {
		log.Println("Submission was successful, but no submission response was returned (e.g., a 204 No Content response).")
		return
	}

	log.Println("Successfully submitted work cards!")
	for _, resp := range submissionResponses {
		log.Printf("-> Submission ID: %s, Protocol: %s, Date: %s\n", resp.ID, resp.Protocol, resp.SubmissionDate.Format(time.RFC3339))
	}
}
