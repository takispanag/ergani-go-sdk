package ergani

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

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
