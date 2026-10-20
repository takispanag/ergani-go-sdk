package ergani

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UserType is the kind of account the credentials belong to.
type UserType string

// The kinds of account the API accepts.
const (
	// UserTypeExternal is an external user, an account of the API itself.
	UserTypeExternal UserType = "01"

	// UserTypeErgani is an account that logs in with the codes of the Ergani
	// platform. It is the usual one for an employer.
	UserTypeErgani UserType = "02"

	// UserTypeConstruction is an account that logs in with the codes of EFKA
	// for construction works.
	UserTypeConstruction UserType = "03"
)

// Valid reports whether the user type is one the API documents.
func (t UserType) Valid() bool {
	switch t {
	case UserTypeExternal, UserTypeErgani, UserTypeConstruction:
		return true
	default:
		return false
	}
}

// Token is the set of credentials the API hands out.
//
// The access token is a JSON Web Token that goes into the Authorization
// header of every call. It is short lived: the API reports its lifetime in
// seconds, three hours at the time of writing.
//
// The refresh token is a long lived secret that renews the access token. It
// is revoked by [Client.Logout].
type Token struct {
	// AccessToken is the JSON Web Token to send in the Authorization header.
	AccessToken string

	// AccessTokenExpires is the moment the access token stops being valid.
	// It is the zero time when the API did not report a lifetime.
	AccessTokenExpires time.Time

	// RefreshToken renews the access token.
	RefreshToken string

	// RefreshTokenExpires is the moment the refresh token stops being valid.
	// It is the zero time when the API did not report an expiry.
	RefreshTokenExpires time.Time
}

// Expired reports whether the access token is known to be expired.
func (t *Token) Expired(now time.Time) bool {
	if t == nil || t.AccessToken == "" {
		return true
	}
	if t.AccessTokenExpires.IsZero() {
		return false
	}

	return !now.Before(t.AccessTokenExpires)
}

// authenticationRequest is the body of the authentication call.
//
// The API binds the fields case insensitively, but the official
// documentation spells the last one "Usertype".
type authenticationRequest struct {
	Username string   `json:"Username"`
	Password string   `json:"Password"`
	UserType UserType `json:"Usertype"`
}

// authenticationResponse is the body the API answers the authentication
// calls with.
type authenticationResponse struct {
	AccessToken         string `json:"accessToken"`
	AccessTokenExpired  int64  `json:"accessTokenExpired"`
	RefreshToken        string `json:"refreshToken"`
	RefreshTokenExpired string `json:"refreshTokenExpired"`
}

// refreshRequest is the body of the refresh call. The field names are in
// PascalCase, unlike the ones of the authentication call.
type refreshRequest struct {
	AccessToken  string `json:"AccessToken"`
	RefreshToken string `json:"RefreshToken"`
}

// Authenticate authenticates with the API and returns the credentials.
//
// The client authenticates on its own before the first call, so calling this
// method is only needed to obtain the token itself, or to authenticate
// eagerly.
//
// It returns an [*APIError] when the API rejects the credentials.
func (c *Client) Authenticate(ctx context.Context) (*Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.authenticateLocked(ctx)
}

// Refresh renews the access token with the refresh token.
//
// It returns an [*APIError] when the refresh token has expired, in which case
// the client authenticates again with the credentials.
func (c *Client) Refresh(ctx context.Context) (*Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.refreshLocked(ctx)
}

// Logout revokes the refresh token.
//
// The API accepts the refresh token as a bare JSON string, not as an object.
func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	refresh := c.refreshToken
	c.mu.Unlock()

	if refresh == "" {
		return nil
	}

	body, err := encode(refresh)
	if err != nil {
		return err
	}

	if _, err := c.call(ctx, http.MethodPost, "/Authentication/Logout", nil, body, false); err != nil {
		return err
	}

	c.mu.Lock()
	c.refreshToken = ""
	c.refreshExpiry = time.Time{}
	c.mu.Unlock()

	return nil
}

// authenticateLocked performs the authentication. The caller must hold mu.
func (c *Client) authenticateLocked(ctx context.Context) (*Token, error) {
	body, err := encode(authenticationRequest{
		Username: c.username,
		Password: c.password,
		UserType: c.userType,
	})
	if err != nil {
		return nil, err
	}

	data, err := c.call(ctx, http.MethodPost, "/Authentication", nil, body, false)
	if err != nil {
		return nil, err
	}

	return c.storeTokenLocked(data)
}

// refreshLocked renews the access token. The caller must hold mu.
func (c *Client) refreshLocked(ctx context.Context) (*Token, error) {
	if c.refreshToken == "" {
		return c.authenticateLocked(ctx)
	}

	body, err := encode(refreshRequest{
		AccessToken:  c.token,
		RefreshToken: c.refreshToken,
	})
	if err != nil {
		return nil, err
	}

	data, err := c.call(ctx, http.MethodPost, "/Authentication/Refresh", nil, body, false)
	if err != nil {
		return nil, err
	}

	return c.storeTokenLocked(data)
}

// storeTokenLocked decodes an authentication response and caches it. The
// caller must hold mu.
func (c *Client) storeTokenLocked(data []byte) (*Token, error) {
	var response authenticationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the authentication response: %w", err)
	}

	if response.AccessToken == "" {
		return nil, &AuthenticationError{Reason: "the API did not return an access token"}
	}

	now := c.now()

	token := &Token{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	}
	if response.AccessTokenExpired > 0 {
		token.AccessTokenExpires = now.Add(time.Duration(response.AccessTokenExpired) * time.Second)
	}
	if response.RefreshTokenExpired != "" {
		if expiry, err := time.Parse(time.RFC3339Nano, response.RefreshTokenExpired); err == nil {
			token.RefreshTokenExpires = expiry
		}
	}

	c.token = token.AccessToken
	c.tokenExpiry = token.AccessTokenExpires
	c.refreshToken = token.RefreshToken
	c.refreshExpiry = token.RefreshTokenExpires

	return token, nil
}

// accessToken returns a valid access token, authenticating or refreshing
// when needed.
func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()

	if c.token != "" && (c.tokenExpiry.IsZero() || now.Before(c.tokenExpiry)) {
		return c.token, nil
	}

	if c.refreshToken != "" && (c.refreshExpiry.IsZero() || now.Before(c.refreshExpiry)) {
		if token, err := c.refreshLocked(ctx); err == nil {
			return token.AccessToken, nil
		}
		// The refresh token did not work. Fall back to the credentials.
	}

	token, err := c.authenticateLocked(ctx)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// invalidateAccessToken drops the cached access token, keeping the refresh
// token so that the next call can renew it.
func (c *Client) invalidateAccessToken() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.token = ""
	c.tokenExpiry = time.Time{}
}
