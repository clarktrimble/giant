// Package oauth2rt implements the Tripper interface for OAuth2 client credentials.
package oauth2rt

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Todo: depend on logger in m-monorepo, when we get there someday
// Todo: swith to New rather than direct construction.
// Todo: no tripper logging etc here, prolly correct but doc
// Todo: look at wrapping golang.org/x/oauth2 instead of hand-vibed soln
// Todo: someday support form-encoded token requests (application/x-www-form-urlencoded)
// Todo: someday proactive refresh via expires_in and/or adopt golang.org/x/oauth2

// Logger specifies a contextual structured logger.
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

	// retry once on 401/403
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		rt.Logger.Info(ctx, "received 401/403, refreshing token")
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
		rt.Logger.Debug(ctx, "using cached oauth token")
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
	rt.Logger.Debug(ctx, "requesting oauth token", "url", tokenURL)

	payload, err := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     rt.ClientID,
		"client_secret": rt.ClientSecret,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal token request")
	}

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

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", errors.Wrap(err, "failed to decode token response")
	}

	accessToken, _ := data["access_token"].(string)
	if accessToken == "" {
		return "", errors.New("token response missing access_token")
	}

	data["access_token"] = strings.Repeat("x", len(accessToken))
	rt.Logger.Info(ctx, "oauth token refreshed", "response", data)

	return accessToken, nil
}
