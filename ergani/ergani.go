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
	"strconv"
	"sync"
	"time"
)

const (
	// defaultBaseURL is the default base URL for the Ergani API.
	defaultBaseURL   = "https://trialeservices.yeka.gr/WebServicesAPI/api"
	UserTypeEmployer = "01"
	DefaultTimeout   = 30 * time.Second

	// DefaultMaxRetries is the number of times a request is retried after a
	// transient failure when Config.MaxRetries is left at zero.
	DefaultMaxRetries = 2

	// DefaultRetryBackoff is the base delay between two retries. The delay
	// grows exponentially with every attempt.
	DefaultRetryBackoff = 500 * time.Millisecond

	// maxRetryBackoff caps the exponential backoff.
	maxRetryBackoff = 30 * time.Second

	// defaultUserAgent identifies this SDK to the API.
	defaultUserAgent = "ergani-go-sdk"
)

// HTTPClient is the subset of *http.Client used to talk to the Ergani API.
// It exists so that tests can inject a stub transport.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds the settings used to create a Client.
type Config struct {
	// Username and Password are the Ergani credentials of the employer.
	Username string
	Password string

	// BaseURL is the API endpoint. It defaults to the Ergani trial service.
	BaseURL string

	// Timeout is the timeout of the default HTTP client. It defaults to
	// DefaultTimeout. It is ignored when HTTPClient is set.
	Timeout time.Duration

	// HTTPClient is the client used to perform the requests. When nil, an
	// *http.Client with the configured timeout is used.
	HTTPClient HTTPClient

	// MaxRetries is the maximum number of retries after the initial attempt.
	// Retries happen when the API answers with 401 (the access token expired),
	// with 429 or with a 5xx status, and when the request fails at the
	// transport level. Zero means DefaultMaxRetries, a negative value disables
	// retries entirely.
	MaxRetries int

	// RetryBackoff is the base delay between two retries. It defaults to
	// DefaultRetryBackoff and grows exponentially with every attempt.
	RetryBackoff time.Duration

	// UserAgent is sent with every request. It defaults to "ergani-go-sdk".
	UserAgent string
}

// Client is a client for interacting with the Ergani API.
// It handles authentication and provides methods for submitting various documents.
//
// A Client is safe for concurrent use by multiple goroutines.
type Client struct {
	baseURL      *url.URL
	httpClient   HTTPClient
	username     string
	password     string
	userAgent    string
	maxRetries   int
	retryBackoff time.Duration

	// authMu serializes authentication so that concurrent callers do not
	// trigger a burst of login requests.
	authMu sync.Mutex

	// mu guards the cached access token.
	mu          sync.RWMutex
	token       string
	tokenExpiry time.Time
}

// NewClientWithConfig creates a client from the given configuration.
func NewClientWithConfig(config Config) (*Client, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	parsed, err := url.Parse(baseURL)
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

	maxRetries := config.MaxRetries
	switch {
	case maxRetries == 0:
		maxRetries = DefaultMaxRetries
	case maxRetries < 0:
		maxRetries = 0
	}

	retryBackoff := config.RetryBackoff
	if retryBackoff <= 0 {
		retryBackoff = DefaultRetryBackoff
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	return &Client{
		baseURL:      parsed,
		httpClient:   httpClient,
		username:     config.Username,
		password:     config.Password,
		userAgent:    userAgent,
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
	}, nil
}

// NewClient creates and configures a new Ergani API client.
// It authenticates with the provided credentials and returns a client instance
// ready to make API calls. An optional customBaseURL can be provided for testing
// or to target a different API version/environment.
func NewClient(username, password string, customBaseURL ...string) (*Client, error) {
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

	return NewClientWithConfig(config)
}

// authenticate performs the initial authentication against the API to retrieve an access token.
// The token is cached on the client and reused until the API rejects it.
func (c *Client) authenticate(ctx context.Context) (string, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	return c.authenticateLocked(ctx)
}

// authenticateLocked performs the authentication request. The caller must hold
// authMu.
//
// Transient failures (transport errors, 429 and 5xx answers) are retried;
// rejected credentials are not.
func (c *Client) authenticateLocked(ctx context.Context) (string, error) {
	for attempt := 0; ; attempt++ {
		token, retryable, err := c.authenticateOnce(ctx)
		if err == nil {
			return token, nil
		}
		if !retryable || attempt >= c.maxRetries {
			return "", err
		}
		if waitErr := sleep(ctx, c.backoff(attempt)); waitErr != nil {
			return "", waitErr
		}
	}
}

// authenticateOnce performs a single authentication request. It reports
// whether the failure is worth retrying.
func (c *Client) authenticateOnce(ctx context.Context) (string, bool, error) {
	authPayload := map[string]string{
		"Username": c.username,
		"Password": c.password,
		"UserType": UserTypeEmployer, // UserType for employers
	}

	bodyBytes, err := json.Marshal(authPayload)
	if err != nil {
		return "", false, fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	endpoint := c.baseURL.JoinPath("/Authentication")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return "", false, fmt.Errorf("failed to create authentication request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", true, fmt.Errorf("authentication request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		return "", retryable, newAPIError(resp)
	}
	defer func() { _ = resp.Body.Close() }()

	var authResponse struct {
		AccessToken string `json:"accessToken"`
		// ExpiresIn is optional. When the API reports a lifetime, the token is
		// refreshed proactively; otherwise it is kept until the API rejects it.
		ExpiresIn int64 `json:"expiresIn"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", false, fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResponse.AccessToken == "" {
		return "", false, &AuthenticationError{Message: "authentication successful but no token was returned"}
	}

	c.storeToken(authResponse.AccessToken, authResponse.ExpiresIn)

	return authResponse.AccessToken, false, nil
}

// storeToken caches the access token together with its expiry.
func (c *Client) storeToken(token string, expiresIn int64) {
	var expiry time.Time

	switch {
	case expiresIn <= 0:
		// The API did not report a lifetime. The token stays valid until the
		// API answers with 401.
	case expiresIn > 1e9:
		// An absolute expiry, expressed as a Unix timestamp.
		expiry = time.Unix(expiresIn, 0)
	default:
		expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.token = token
	c.tokenExpiry = expiry
}

// invalidateToken drops the cached access token, forcing the next request to
// authenticate again.
func (c *Client) invalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.token = ""
	c.tokenExpiry = time.Time{}
}

// currentToken returns the cached access token, authenticating first when
// there is none or when the known expiry has passed.
func (c *Client) currentToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.token
	expiry := c.tokenExpiry
	c.mu.RUnlock()

	if token != "" && (expiry.IsZero() || time.Now().Before(expiry)) {
		return token, nil
	}

	return c.authenticate(ctx)
}

// request is a helper function to create, execute, and handle a generic API request.
// It marshals the payload, sets necessary headers (including the auth token),
// and handles non-successful status codes.
//
// Transient failures are retried: a 401 triggers a fresh authentication, while
// 429, 5xx and transport errors are retried with an exponential backoff.
func (c *Client) request(ctx context.Context, method, path string, payload interface{}) (*http.Response, error) {
	body, err := encodePayload(payload)
	if err != nil {
		return nil, err
	}

	for attempt := 0; ; attempt++ {
		token, err := c.currentToken(ctx)
		if err != nil {
			return nil, err
		}

		req, err := c.newRequest(ctx, method, path, body, token)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				if waitErr := sleep(ctx, c.backoff(attempt)); waitErr != nil {
					return nil, waitErr
				}
				continue
			}
			return nil, fmt.Errorf("request to %s failed: %w", path, err)
		}

		switch {
		case resp.StatusCode == http.StatusUnauthorized:
			// The access token was rejected, most likely because it expired.
			apiErr := newAPIError(resp)
			c.invalidateToken()

			if attempt < c.maxRetries {
				continue
			}
			return nil, apiErr

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			delay := c.retryDelay(resp, attempt)
			apiErr := newAPIError(resp)

			if attempt < c.maxRetries {
				if waitErr := sleep(ctx, delay); waitErr != nil {
					return nil, waitErr
				}
				continue
			}
			return nil, apiErr

		case resp.StatusCode < 200 || resp.StatusCode >= 300:
			return nil, newAPIError(resp)

		default:
			return resp, nil
		}
	}
}

// newRequest builds a request for the given API path.
func (c *Client) newRequest(ctx context.Context, method, path string, body []byte, token string) (*http.Request, error) {
	endpoint := c.baseURL.JoinPath(path)

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// backoff returns the delay to wait before the given attempt is retried.
func (c *Client) backoff(attempt int) time.Duration {
	delay := c.retryBackoff
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay >= maxRetryBackoff {
			return maxRetryBackoff
		}
	}

	return delay
}

// retryDelay returns the delay to wait before the given attempt is retried,
// honoring the Retry-After header when the API sends one.
func (c *Client) retryDelay(resp *http.Response, attempt int) time.Duration {
	if delay := parseRetryAfter(resp.Header.Get("Retry-After")); delay > 0 {
		return delay
	}

	return c.backoff(attempt)
}

// parseRetryAfter parses the value of a Retry-After header. It returns 0 when
// the header is absent or cannot be understood.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}

	if deadline, err := http.ParseTime(value); err == nil {
		if remaining := time.Until(deadline); remaining > 0 {
			return remaining
		}
	}

	return 0
}

// encodePayload marshals a request payload. A nil payload produces no body.
func encodePayload(payload interface{}) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	return bodyBytes, nil
}

// sleep waits for the given duration, or until the context is canceled.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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

	parsed, parseErr := parseSubmissionResponse(resp)
	closeErr := resp.Body.Close()

	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close response body: %w", closeErr)
	}

	return parsed, nil
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

	parsed, parseErr := parseSubmissionResponse(resp)
	closeErr := resp.Body.Close()

	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close response body: %w", closeErr)
	}

	return parsed, nil
}

// SubmitDailySchedule submits daily work schedules for employees.
// It takes a slice of CompanyDailySchedule, each representing the schedules for a specific
// business branch.
func (c *Client) SubmitDailySchedule(ctx context.Context, companyDailySchedules []CompanyDailySchedule) ([]SubmissionResponse, error) {
	// The API expects the payload to be nested within "WTOS" and "WTO" keys.
	payload := map[string]map[string][]CompanyDailySchedule{
		"WTOS": {"WTO": companyDailySchedules},
	}
	resp, err := c.request(ctx, http.MethodPost, "/Documents/WTODaily", payload)
	if err != nil {
		return nil, err
	}

	parsed, parseErr := parseSubmissionResponse(resp)
	closeErr := resp.Body.Close()

	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close response body: %w", closeErr)
	}

	return parsed, nil
}

// SubmitWeeklySchedule submits weekly work schedules for employees.
// It takes a slice of CompanyWeeklySchedule, each representing the schedules for a specific
// business branch.
func (c *Client) SubmitWeeklySchedule(ctx context.Context, companyWeeklySchedules []CompanyWeeklySchedule) ([]SubmissionResponse, error) {
	// The API expects the payload to be nested within "WTOS" and "WTO" keys.
	payload := map[string]map[string][]CompanyWeeklySchedule{
		"WTOS": {"WTO": companyWeeklySchedules},
	}
	resp, err := c.request(ctx, http.MethodPost, "/Documents/WTOWeek", payload)
	if err != nil {
		return nil, err
	}

	parsed, parseErr := parseSubmissionResponse(resp)
	closeErr := resp.Body.Close()

	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close response body: %w", closeErr)
	}

	return parsed, nil
}
