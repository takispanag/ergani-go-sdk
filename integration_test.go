package ergani_test

import (
	"context"
	"github.com/takispanag/ergani-go-sdk"
	"os"
	"testing"
	"time"
)

// The integration tests talk to the Ergani platform, so they need
// credentials and they are skipped without them.
//
// The credentials come from the environment:
//
//	ERGANI_USERNAME    the user of the employer
//	ERGANI_PASSWORD    the password of the user
//	ERGANI_BASE_URL    the environment, the trial one by default
//	ERGANI_AFM         the tax identification number of the employer
//
// The declarations of the trial environment are marked as void and have no
// legal effect, so the tests may run against it freely.
func integrationClient(t *testing.T) *ergani.Client {
	t.Helper()

	username := os.Getenv("ERGANI_USERNAME")
	password := os.Getenv("ERGANI_PASSWORD")
	if username == "" || password == "" {
		t.Skip("ERGANI_USERNAME and ERGANI_PASSWORD are not set, the integration test is skipped")
	}

	options := []ergani.Option{}
	if baseURL := os.Getenv("ERGANI_BASE_URL"); baseURL != "" {
		options = append(options, ergani.WithBaseURL(baseURL))
	}

	client, err := ergani.New(username, password, options...)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return client
}

// TestIntegration_Authenticate checks that the credentials work and that the
// token can be renewed.
func TestIntegration_Authenticate(t *testing.T) {
	client := integrationClient(t)

	token, err := client.Authenticate(context.Background())
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if token.AccessToken == "" {
		t.Fatal("Authenticate() returned an empty access token")
	}
	if token.Expired(time.Now()) {
		t.Error("Authenticate() returned a token that is already expired")
	}

	renewed, err := client.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if renewed.AccessToken == "" {
		t.Fatal("Refresh() returned an empty access token")
	}
}

// TestIntegration_ListSubmissions checks that the declaration types of the
// connected employer can be read.
func TestIntegration_ListSubmissions(t *testing.T) {
	client := integrationClient(t)

	types, err := client.ListSubmissions(context.Background())
	if err != nil {
		t.Fatalf("ListSubmissions() error = %v", err)
	}
	if len(types) == 0 {
		t.Fatal("ListSubmissions() returned no declaration type")
	}

	for _, declarationType := range types {
		if declarationType.Code == "" {
			t.Errorf("the declaration type %d has no code", declarationType.ID)
		}
	}
}

// TestIntegration_DocumentSchema checks that the schema of a declaration type
// can be read.
func TestIntegration_DocumentSchema(t *testing.T) {
	client := integrationClient(t)

	types, err := client.ListSubmissions(context.Background())
	if err != nil {
		t.Fatalf("ListSubmissions() error = %v", err)
	}
	if len(types) == 0 {
		t.Skip("the account has no declaration type")
	}

	schema, err := client.DocumentSchema(context.Background(), types[0].Code)
	if err != nil {
		t.Fatalf("DocumentSchema() error = %v", err)
	}
	if len(schema) == 0 {
		t.Error("DocumentSchema() returned an empty schema")
	}
}

// TestIntegration_Services checks that the interoperability services of the
// account can be read.
func TestIntegration_Services(t *testing.T) {
	client := integrationClient(t)

	services, err := client.Services(context.Background())
	if err != nil {
		t.Fatalf("Services() error = %v", err)
	}
	if len(services) == 0 {
		t.Fatal("Services() returned no service")
	}
}
