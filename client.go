package ergani

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
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
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}

// call performs one logical request, with the retry policy of the client.
//
// When withAuth is set, the request carries the access token. A 401 then
// invalidates the token and triggers a new attempt.

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
