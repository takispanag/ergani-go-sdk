// Package ergani provides a client for interacting with the Greek government's
// Ergani API for labor-related declarations.
package ergani

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// defaultBaseURL is the default base URL for the Ergani API.
	defaultBaseURL   = "https://trialeservices.yeka.gr/WebServicesAPI/api"
	UserTypeEmployer = "01"
	DefaultTimeout   = 30 * time.Second
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Config struct {
	Username   string
	Password   string
	BaseURL    string
	Timeout    time.Duration
	HTTPClient HTTPClient
}

// Client is a client for interacting with the Ergani API.
// It handles authentication and provides methods for submitting various documents.
type Client struct {
	baseURL    *url.URL
	httpClient HTTPClient
	token      string
}

func NewClientWithConfig(ctx context.Context, config Config) (*Client, error) {
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	c := &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}

	if err := c.authenticate(ctx, config.Username, config.Password); err != nil {
		return nil, err
	}

	return c, nil
}

// NewClient creates and configures a new Ergani API client.
// It authenticates with the provided credentials and returns a client instance
// ready to make API calls. An optional customBaseURL can be provided for testing
// or to target a different API version/environment.
func NewClient(ctx context.Context, username, password string, customBaseURL ...string) (*Client, error) {
	baseURL := defaultBaseURL
	if len(customBaseURL) > 0 && customBaseURL[0] != "" {
		baseURL = customBaseURL[0]
	}

	config := Config{
		Username: username,
		Password: password,
		BaseURL:  baseURL,
		Timeout:  DefaultTimeout,
	}

	return NewClientWithConfig(ctx, config)
}

// authenticate performs the initial authentication against the API to retrieve an access token.
// The token is stored in the client for subsequent requests.
func (c *Client) authenticate(ctx context.Context, username, password string) error {
	authPayload := map[string]string{
		"Username": username,
		"Password": password,
		"UserType": UserTypeEmployer, // UserType for employers
	}

	bodyBytes, err := json.Marshal(authPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	endpoint := c.baseURL.JoinPath("/Authentication")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create authentication request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return newAPIError(resp)
	}

	var authResponse struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResponse.AccessToken == "" {
		return &AuthenticationError{Message: "authentication successful but no token was returned"}
	}

	c.token = authResponse.AccessToken
	return nil
}

// request is a helper function to create, execute, and handle a generic API request.
// It marshals the payload, sets necessary headers (including the auth token),
// and handles non-successful status codes.
func (c *Client) request(ctx context.Context, method, path string, payload interface{}) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	endpoint := c.baseURL.JoinPath(path)
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// For 204 No Content, the response is successful but has no body.
		if resp.StatusCode == http.StatusNoContent {
			return resp, nil
		}
		return nil, newAPIError(resp)
	}

	return resp, nil
}

// SubmitWorkCard submits work card records (check-in/check-out) for employees.
// It takes a slice of CompanyWorkCard, each representing the records for a specific
// business branch.
func (c *Client) SubmitWorkCard(ctx context.Context, companyWorkCards []CompanyWorkCard) ([]SubmissionResponse, error) {
	// The API expects the payload to be nested within "Cards" and "Card" keys.
	payload := map[string]map[string][]CompanyWorkCard{
		"Cards": {"Card": companyWorkCards},
	}
	resp, err := c.request(ctx, http.MethodPost, "/Documents/WRKCardSE", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseSubmissionResponse(resp)
}

// SubmitOvertime submits overtime records for employees.
// It takes a slice of CompanyOvertime, each representing the records for a specific
// business branch.
func (c *Client) SubmitOvertime(ctx context.Context, companyOvertimes []CompanyOvertime) ([]SubmissionResponse, error) {
	// The API expects the payload to be nested within "Overtimes" and "Overtime" keys.
	payload := map[string]map[string][]CompanyOvertime{
		"Overtimes": {"Overtime": companyOvertimes},
	}
	resp, err := c.request(ctx, http.MethodPost, "/Documents/OvTime", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseSubmissionResponse(resp)
}

// SubmitDailySchedule submits daily work schedules for employees.
// It takes a slice of CompanyDailySchedule, each representing the schedules for
// a specific business branch.
func (c *Client) SubmitDailySchedule(ctx context.Context, companyDailySchedules []CompanyDailySchedule) ([]SubmissionResponse, error) {
	// The API expects the payload to be nested within "WTOS" and "WTO" keys.
	payload := map[string]map[string][]CompanyDailySchedule{
		"WTOS": {"WTO": companyDailySchedules},
	}
	resp, err := c.request(ctx, http.MethodPost, "/Documents/WTODaily", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseSubmissionResponse(resp)
}

// SubmitWeeklySchedule submits weekly work schedules for employees.
// It takes a slice of CompanyWeeklySchedule, each representing the schedules for
// a specific business branch.
func (c *Client) SubmitWeeklySchedule(ctx context.Context, companyWeeklySchedules []CompanyWeeklySchedule) ([]SubmissionResponse, error) {
	// The API expects the payload to be nested within "WTOS" and "WTO" keys.
	payload := map[string]map[string][]CompanyWeeklySchedule{
		"WTOS": {"WTO": companyWeeklySchedules},
	}
	resp, err := c.request(ctx, http.MethodPost, "/Documents/WTOWeek", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseSubmissionResponse(resp)
}
