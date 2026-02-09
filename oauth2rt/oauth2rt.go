// Package oauth2rt implements the Tripper interface for OAuth2 client credentials.
package oauth2rt

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/pkg/errors"
)

// Todo: support form-encoded token requests (application/x-www-form-urlencoded)
// Todo: proactive refresh via expires_in and/or adopt golang.org/x/oauth2

// Logger is an optional logger for token operations.
// Todo: depend on logger from above
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Debug(ctx context.Context, msg string, kv ...any)
}

// OAuth2Rt implements the Tripper interface.
type OAuth2Rt struct {
	BaseUri      string
	TokenPath    string
	ClientID     string
	ClientSecret string
	Logger       Logger

	mu    sync.RWMutex
	token string
	next  http.RoundTripper
}

// Wrap sets the next round tripper, thereby wrapping it.
func (rt *OAuth2Rt) Wrap(next http.RoundTripper) {
	rt.next = next
}

// RoundTrip adds a Bearer token to requests, fetching lazily and retrying on 401.
func (rt *OAuth2Rt) RoundTrip(req *http.Request) (*http.Response, error) {

	ctx := req.Context()

	// buffer body for potential retry
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read request body")
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	token, err := rt.getToken(ctx)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := rt.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// retry once on 401
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		rt.clearToken()
		token, err = rt.getToken(ctx)
		if err != nil {
			return nil, err
		}

		// reset body for retry
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		req.Header.Set("Authorization", "Bearer "+token)
		resp, err = rt.next.RoundTrip(req)
	}

	return resp, err
}

// unexported

func (rt *OAuth2Rt) getToken(ctx context.Context) (token string, err error) {

	// "double-checked locking"
	// - check once with read lock (fast path)
	// - check again after write lock (safe path)

	rt.mu.RLock()
	token = rt.token
	rt.mu.RUnlock()
	if token != "" {
		return
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.token != "" {
		return rt.token, nil
	}

	token, err = rt.refreshToken(ctx)
	if err != nil {
		return
	}
	rt.token = token

	return
}

func (rt *OAuth2Rt) clearToken() {

	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.token = ""
}

func (rt *OAuth2Rt) refreshToken(ctx context.Context) (string, error) {

	tokenURL := rt.BaseUri + rt.TokenPath

	payload, err := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     rt.ClientID,
		"client_secret": rt.ClientSecret,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal token request")
	}

	// Todo: clean up logging in here
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewReader(payload))
	if err != nil {
		return "", errors.Wrap(err, "failed to create token request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rt.next.RoundTrip(req)
	if err != nil {
		return "", errors.Wrap(err, "token request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read token response")
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("token request returned %d: %s", resp.StatusCode, body)
	}

	rt.Logger.Debug(ctx, "refreshed token", "response", string(body))

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", errors.Wrap(err, "failed to decode token response")
	}

	if result.AccessToken == "" {
		return "", errors.New("token response missing access_token")
	}

	return result.AccessToken, nil
}
