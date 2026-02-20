// Package giant provides for reuse of common json api client patterns
// while doing its best to not get in the way :)
package giant

// Todo: resolve dep tangle with minibike monorepo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/clarktrimble/giant/basicrt"
	"github.com/clarktrimble/giant/logger"
	"github.com/clarktrimble/giant/logrt"
	"github.com/clarktrimble/giant/oauth2rt"
	"github.com/clarktrimble/giant/statusrt"
	"github.com/clarktrimble/launch"
	"github.com/pkg/errors"
)

// Todo: use BaseUrl: url.URL instead of string

// Config represents giant config
type Config struct {
	// BaseUri is the scheme, domain, optionally port and/or path
	// though which the client connects to a webserver
	// for example: http://192.168.64.3:4080/graphql
	BaseUri string `json:"base_uri" desc:"ex: http://llp.org:4080/graphql" required:"true"`
	// Timeout is the overall request timeout
	Timeout time.Duration `json:"timeout" desc:"request timeout" default:"1m"`
	// TimeoutShort is the dialer and response header timeout
	TimeoutShort time.Duration `json:"timeout_short" desc:"dialer and header timeout" default:"10s"`
	// Headers are set when making a request
	Headers []string `json:"headers,omitempty" desc:"header pairs to be sent with every request"`
	// SkipVerify skips verification of ssl certificates (dev only pls!)
	SkipVerify bool `json:"skip_verify" desc:"skip cert verification" default:"false"`
	// Ciphers overrides default tls ciphers
	Ciphers []uint16 `json:"ciphers" desc:"ciphers override"`
	// User is for basic auth in NewWithTrippers.
	User string `json:"user,omitempty" desc:"username for basic auth"`
	// Pass is for basic auth in NewWithTrippers.
	Pass launch.Redact `json:"pass,omitempty" desc:"password for basic auth"`
	// Todo: what's this??
	//KeyHeader string `json:"api_key_header,omitempty" desc:"Todo"`
	//ApiKey    Redact `json:"api_key_value,omitempty" desc:"Todo and sneak into redact headers"`
	// Todo: orrrrrrrr a giant client helper in bfc would work?
	// RedactHeaders are headers to be redacted from logging in NewWithTrippers.
	RedactHeaders []string `json:"redact_headers,omitempty" desc:"headers to redact from request logging"`
	// SkipBody when true request and response bodies are not logged in NewWithTrippers..
	SkipBody bool `json:"skip_body" desc:"skip logging of body for request and response" default:"false"`
	// UnixSocket
	UnixSocket string `json:"unix_socket,omitempty" desc:"unix socket"`
	// OAuth2 is for OAuth2 client credentials in NewWithTrippers.
	OAuth2 *OAuth2Config `json:"oauth2,omitempty" desc:"OAuth2 client credentials config"`
}

// OAuth2Config represents OAuth2 client credentials configuration.
type OAuth2Config struct {
	// BaseUri is the OAuth2 token endpoint base URI.
	// If empty, defaults to Config.BaseUri.
	BaseUri string `json:"base_uri,omitempty" desc:"OAuth2 base URI (defaults to client base_uri)"`
	// TokenPath is the path to the token endpoint.
	TokenPath string `json:"token_path" desc:"path to token endpoint" default:"/oauth/token"`
	// ClientID is the OAuth2 client ID.
	ClientID string `json:"client_id" desc:"OAuth2 client ID"`
	// ClientSecret is the OAuth2 client secret.
	ClientSecret launch.Redact `json:"client_secret" desc:"OAuth2 client secret or path to secret file"`
}

// Giant represents an http client
type Giant struct {
	// Client is a stdlib http client
	Client http.Client
	// BaseUri is as described in Config
	BaseUri string
	// Headers are set when making a request
	Headers map[string]string
}

// New constructs a new client from Config
func (cfg *Config) New() *Giant {

	var ciphers []uint16
	if len(cfg.Ciphers) > 0 {
		ciphers = cfg.Ciphers
	}

	transport := &http.Transport{
		Dial:                (&net.Dialer{Timeout: cfg.TimeoutShort}).Dial,
		TLSHandshakeTimeout: cfg.TimeoutShort,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipVerify, //nolint: gosec
			CipherSuites:       ciphers,
		},
	}

	// Todo: unit
	if cfg.UnixSocket != "" {
		transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", cfg.UnixSocket)
			},
		}
	}

	// copy header cfg pairs into map ignoring odd count

	hdrs := map[string]string{}
	hdrsLen := len(cfg.Headers)
	if hdrsLen%2 != 0 {
		hdrsLen--
	}
	for i := 0; i < hdrsLen; i += 2 {
		hdrs[cfg.Headers[i]] = cfg.Headers[i+1]
	}

	return &Giant{
		Client: http.Client{
			Transport:     transport,
			CheckRedirect: noRedirect,
			Timeout:       cfg.Timeout,
		},
		BaseUri: cfg.BaseUri,
		Headers: hdrs,
	}
}

// NewWithTrippers is a convenience method that adds StatusRt and Logrt after creating a client.
// If OAuth2 is defined in Config OAuth2Rt is added as well.
// If User and Pass are defined in Config BasicRt is added as well.
func (cfg *Config) NewWithTrippers(lgr logger.Logger) (giant *Giant) {

	giant = cfg.New()

	// OAuth2 goes first (innermost) so auth header is set before logging/status
	if cfg.OAuth2 != nil && cfg.OAuth2.ClientID != "" {
		baseUri := cfg.OAuth2.BaseUri
		if baseUri == "" {
			baseUri = cfg.BaseUri
		}
		giant.Use(&oauth2rt.OAuth2Rt{
			BaseUri:      baseUri,
			TokenPath:    cfg.OAuth2.TokenPath,
			ClientID:     cfg.OAuth2.ClientID,
			ClientSecret: string(cfg.OAuth2.ClientSecret),
			Logger:       lgr,
		})
	}

	giant.Use(&statusrt.StatusRt{})
	giant.Use(logrt.New(lgr, cfg.RedactHeaders, cfg.SkipBody))

	if cfg.User != "" && cfg.Pass != "" {
		basicRt := basicrt.New(cfg.User, string(cfg.Pass))
		giant.Use(basicRt)
	}

	return
}

// Use wraps the current transport with a round tripper
func (giant *Giant) Use(tripper tripper) {

	tripper.Wrap(giant.Client.Transport)
	giant.Client.Transport = tripper
}

// Request represents an http request
type Request struct {
	// Method is one of the http RFC methods (no net!)
	Method string
	// Path is appended to BaseUri when making a request
	// (leading and trailing slashes recommended here, convention for sanity!)
	Path string
	// Body is read from when making a request
	Body io.Reader
	// Headers are set when making a request
	Headers map[string]string
}

// Send sends a request
// leaving read/close of response body to caller
func (giant *Giant) Send(ctx context.Context, rq Request) (response *http.Response, err error) {

	if rq.Headers == nil {
		rq.Headers = map[string]string{}
	}
	maps.Copy(rq.Headers, giant.Headers)

	request, err := rq.httpRequest(ctx, giant.BaseUri)
	if err != nil {
		return
	}

	response, err = giant.Client.Do(request)
	err = errors.Wrapf(err, "http %s request to %s %s failed", rq.Method, giant.BaseUri, rq.Path)
	return
}

// SendJson constructs a request, sends and receives json closing the response body
func (giant *Giant) SendJson(ctx context.Context, method, path string, body io.Reader) (data []byte, err error) {

	rq := Request{
		Method: method,
		Path:   path,
		Body:   body,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
	}

	response, err := giant.Send(ctx, rq)
	if err != nil {
		return
	}
	defer response.Body.Close()

	data, err = io.ReadAll(response.Body)
	return
}

// SendObject marshalls the object to be sent, unmarshalls the response body, and calls SendJson
func (giant *Giant) SendObject(ctx context.Context, method, path string, sndObj, rcvObj any) (err error) {

	sndData, err := marshal(sndObj)
	if err != nil {
		return
	}

	rcvData, err := giant.SendJson(ctx, method, path, bytes.NewBuffer(sndData))
	if err != nil {
		return
	}

	if rcvObj != nil {
		err = json.Unmarshal(rcvData, &rcvObj)
		err = errors.Wrapf(err, "failed to decode response into %#v", rcvObj)
	}
	return
}

// Uri returns the base uri for use in links, etc.
func (giant *Giant) Uri() string {

	// Todo: privatize BaseUri and rename
	// Todo: unit, yeah, but unit!
	return giant.BaseUri
}

// unexported

type tripper interface {
	http.RoundTripper
	Wrap(next http.RoundTripper)
}

func marshal(obj any) (data []byte, err error) {

	data = []byte{}
	if obj == nil {
		return
	}

	data, err = json.Marshal(obj)
	err = errors.Wrapf(err, "failed to marshal object: %#v", obj)
	return
}

func (rq Request) httpRequest(ctx context.Context, baseUri string) (request *http.Request, err error) {

	uri, err := url.ParseRequestURI(fmt.Sprintf("%s%s", baseUri, rq.Path))
	if err != nil {
		err = errors.Wrapf(err, "unable to parse uri from %s %s", baseUri, rq.Path)
		return
	}

	request, err = http.NewRequestWithContext(ctx, rq.Method, uri.String(), rq.Body)
	if err != nil {
		err = errors.Wrapf(err, "unable to create %s request to %s %s", rq.Method, baseUri, rq.Path)
		return
	}

	for key, val := range rq.Headers {
		request.Header.Set(key, val)
	}

	return
}

func noRedirect(request *http.Request, via []*http.Request) error {
	// do not want posts redirected to a get
	// a-and generally expect to get it right, yeah

	if len(via) == 0 {
		return errors.Errorf("somehow redirected to %s %s from nowhere!?", request.Method, request.URL)
	}

	return errors.Errorf("cowardly refusing to accept redirect to %s %s from %#v", request.Method, request.URL, via)
}
