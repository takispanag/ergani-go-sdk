package ergani

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// TrialBaseURL is the base URL of the Ergani trial environment.
	//
	// The declarations submitted here are marked as void and have no legal
	// effect. It is the default, because a mistake against the production
	// environment has consequences.
	TrialBaseURL = "https://trialv2eservices.yeka.gr/WebServicesAPI/Api/"

	// ProductionBaseURL is the base URL of the production environment of the
	// Ergani platform.
	ProductionBaseURL = "https://eservices.yeka.gr/WebServicesAPI/Api/"

	// DefaultTimeout is the timeout of the HTTP client the SDK creates when
	// no other client is given.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the number of retries after the initial attempt
	// when WithMaxRetries is not given.
	DefaultMaxRetries = 2

	// DefaultRetryBackoff is the base delay between two retries when
	// WithRetryBackoff is not given.
	DefaultRetryBackoff = 500 * time.Millisecond
)

const (
	// maxRetryBackoff caps the exponential backoff and the delay the
	// Retry-After header may ask for.
	maxRetryBackoff = 30 * time.Second

	// maxResponseBytes bounds the size of a response body the SDK reads.
	maxResponseBytes = 32 << 20
)

// Client submits declarations to the Ergani API.
//
// A Client is safe for concurrent use by multiple goroutines. It is immutable
// after construction; only the cached access token is mutated, and that is
// guarded internally.
//
// The zero value of Client is not usable. Create one with New.
type Client struct {
	baseURL      *url.URL
	httpClient   *http.Client
	username     string
	password     string
	userType     UserType
	userAgent    string
	timeout      time.Duration
	maxRetries   int
	retryBackoff time.Duration
	logger       *slog.Logger

	// mu guards the cached credentials. It is held across the
	// authentication, so that concurrent callers trigger a single login
	// rather than one each.
	mu            sync.Mutex
	token         string
	tokenExpiry   time.Time
	refreshToken  string
	refreshExpiry time.Time

	// now and sleep are the seams the tests use to make the expiry and the
	// backoff deterministic.
	now   func() time.Time
	sleep func(ctx context.Context, d time.Duration) error
}

// Option configures a Client.
//
// Options are applied in order, so a later option overrides an earlier one
// that sets the same setting.
type Option func(*Client) error

// New returns a Client that authenticates with the given credentials.
//
// New does not contact the API; the first request performs the
// authentication.
//
// The credentials are the Ergani account of the employer. Use one Client per
// employer, because the declarations are checked against the connected
// employer.
//
// New returns an error when an option is invalid, for example when the base
// URL cannot be parsed.
func New(username, password string, opts ...Option) (*Client, error) {
	base, err := url.Parse(TrialBaseURL)
	if err != nil {
		return nil, fmt.Errorf("ergani: failed to parse the default base URL: %w", err)
	}

	c := &Client{
		baseURL:      base,
		username:     username,
		password:     password,
		userType:     UserTypeErgani,
		userAgent:    defaultUserAgent(),
		timeout:      DefaultTimeout,
		maxRetries:   DefaultMaxRetries,
		retryBackoff: DefaultRetryBackoff,
		logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		now:          time.Now,
		sleep:        sleepContext,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: c.timeout}
	}

	return c, nil
}

// WithBaseURL sets the base URL of the API.
//
// It defaults to [TrialBaseURL]. Use [ProductionBaseURL] to submit against
// the production environment.
//
// The URL must be absolute.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		parsed, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("ergani: invalid base URL %q: %w", baseURL, err)
		}
		if !parsed.IsAbs() {
			return fmt.Errorf("ergani: invalid base URL %q: it must be absolute", baseURL)
		}
		c.baseURL = parsed
		return nil
	}
}

// WithHTTPClient sets the HTTP client used for every request.
//
// When this option is not given, the SDK creates an *http.Client with the
// timeout of [WithTimeout]. The timeout of the given client is used as it
// is; it is not overridden.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("ergani: the HTTP client must not be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithTimeout sets the timeout of the HTTP client the SDK creates for itself.
//
// It is ignored when [WithHTTPClient] is given. It defaults to
// [DefaultTimeout].
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout < 0 {
			return fmt.Errorf("ergani: the timeout must not be negative")
		}
		c.timeout = timeout
		return nil
	}
}

// WithMaxRetries sets the maximum number of retries after the initial
// attempt.
//
// A request is retried when the API answers with 429 or with a 5xx status,
// or when the connection fails. A 401 triggers a fresh authentication and a
// new attempt.
//
// It defaults to [DefaultMaxRetries]. A value of 0 disables the retries.
//
// The submission endpoints are not idempotent: a submission that the API
// accepted but whose response was lost may be sent again, and the API may
// record it twice.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) error {
		if maxRetries < 0 {
			return fmt.Errorf("ergani: the number of retries must not be negative")
		}
		c.maxRetries = maxRetries
		return nil
	}
}

// WithRetryBackoff sets the base delay between two retries. The delay grows
// exponentially with every attempt and is capped at 30 seconds.
//
// It defaults to [DefaultRetryBackoff]. The Retry-After header of the API
// takes precedence, as long as it asks for less than the cap.
func WithRetryBackoff(backoff time.Duration) Option {
	return func(c *Client) error {
		if backoff < 0 {
			return fmt.Errorf("ergani: the backoff must not be negative")
		}
		c.retryBackoff = backoff
		return nil
	}
}

// WithUserAgent sets the User-Agent header of every request.
//
// It defaults to a value that identifies the SDK and its version.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		if userAgent == "" {
			return fmt.Errorf("ergani: the user agent must not be empty")
		}
		c.userAgent = userAgent
		return nil
	}
}

// WithUserType sets the kind of account the credentials belong to.
//
// It defaults to [UserTypeErgani].
func WithUserType(userType UserType) Option {
	return func(c *Client) error {
		if !userType.Valid() {
			return fmt.Errorf("ergani: invalid user type %q", userType)
		}
		c.userType = userType
		return nil
	}
}

// WithLogger sets the logger that reports the requests and the responses.
//
// The credentials, the access token and the identifiers of the employees are
// never written to the log.
//
// It defaults to a logger that discards everything.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) error {
		if logger == nil {
			return fmt.Errorf("ergani: the logger must not be nil")
		}
		c.logger = logger
		return nil
	}
}

// BaseURL returns the base URL the client talks to.
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}

// call performs one logical request, with the retry policy of the client.
//
// When withAuth is set, the request carries the access token. A 401 then
// invalidates the token and triggers a new attempt.
func (c *Client) call(
	ctx context.Context,
	method, path string,
	query url.Values,
	body []byte,
	withAuth bool,
) ([]byte, error) {
	for attempt := 0; ; attempt++ {
		var token string
		if withAuth {
			var err error
			token, err = c.accessToken(ctx)
			if err != nil {
				return nil, err
			}
		}

		req, err := c.newRequest(ctx, method, path, query, body, token)
		if err != nil {
			return nil, err
		}

		c.logger.DebugContext(ctx, "ergani: request", "method", method, "path", path)

		resp, err := c.send(req)
		if err != nil {
			if attempt < c.maxRetries {
				if waitErr := c.wait(ctx, attempt, nil); waitErr != nil {
					return nil, waitErr
				}
				continue
			}
			return nil, fmt.Errorf("ergani: request to %s failed: %w", path, err)
		}

		c.logger.DebugContext(ctx, "ergani: response", "method", method, "path", path, "status", resp.StatusCode)

		switch {
		case resp.StatusCode == http.StatusUnauthorized:
			if withAuth {
				// The access token was rejected. It is most likely expired,
				// and the refresh token, when there is one, gets a new one.
				c.invalidateAccessToken()
				if attempt < c.maxRetries {
					continue
				}
			}
			return nil, newAPIError(resp.StatusCode, resp.Body)

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			if attempt < c.maxRetries {
				if waitErr := c.wait(ctx, attempt, resp.Header); waitErr != nil {
					return nil, waitErr
				}
				continue
			}
			return nil, newAPIError(resp.StatusCode, resp.Body)

		case resp.StatusCode < 200 || resp.StatusCode >= 300:
			return nil, newAPIError(resp.StatusCode, resp.Body)
		}

		return resp.Body, nil
	}
}

// response is what a call to the API returned.
type response struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int

	// Header holds the headers of the response.
	Header http.Header

	// Body holds the body of the response.
	Body []byte
}

// send performs a request and reads its response.
func (c *Client) send(req *http.Request) (response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return response{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return response{}, fmt.Errorf("ergani: failed to read the response body: %w", err)
	}

	return response{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

// newRequest builds a request for the given API path.
func (c *Client) newRequest(
	ctx context.Context,
	method, path string,
	query url.Values,
	body []byte,
	token string,
) (*http.Request, error) {
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimSuffix(endpoint.Path, "/") + path
	if len(query) > 0 {
		endpoint.RawQuery = query.Encode()
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("ergani: failed to create the request for %s: %w", path, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// wait sleeps before the next attempt.
func (c *Client) wait(ctx context.Context, attempt int, header http.Header) error {
	delay := c.retryBackoff
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay >= maxRetryBackoff {
			delay = maxRetryBackoff
			break
		}
	}

	if header != nil {
		if after := parseRetryAfter(header.Get("Retry-After")); after > 0 && after < maxRetryBackoff {
			delay = after
		}
	}

	return c.sleep(ctx, delay)
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

// sleepContext waits for the given duration, or until the context is done.
func sleepContext(ctx context.Context, d time.Duration) error {
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

// defaultUserAgent identifies the SDK and its version.
func defaultUserAgent() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		version := info.Main.Version
		if version != "" && version != "(devel)" {
			return "ergani-go-sdk/" + version
		}
	}
	return "ergani-go-sdk/devel"
}

// encode marshals a request payload.
func encode(payload any) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ergani: failed to marshal the request payload: %w", err)
	}

	return data, nil
}
