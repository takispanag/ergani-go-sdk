package ergani

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

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
