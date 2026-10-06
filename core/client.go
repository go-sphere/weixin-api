package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/singleflight"
)

// DefaultBaseURL is the host of the WeChat server APIs shared by all
// platforms.
const DefaultBaseURL = "https://api.weixin.qq.com"

const (
	// maxResponseSize caps how many bytes are read from a response body as a
	// defensive measure. 64 MiB is far beyond the largest media payload.
	maxResponseSize = 64 << 20
	// defaultHTTPTimeout bounds every HTTP round trip made by a client.
	defaultHTTPTimeout = 30 * time.Second
	// tokenExpirySafetyMargin is subtracted from the server reported
	// expires_in when caching access tokens and tickets so that cached
	// credentials are never used after the server has actually revoked them.
	tokenExpirySafetyMargin = 2 * time.Second
)

// Credentials is the authentication pair shared by all WeChat platform APIs.
type Credentials struct {
	// AppID of the WeChat application (Mini Program, official account, ...).
	AppID string `json:"app_id" yaml:"app_id"`
	// AppSecret of the application.
	AppSecret string `json:"app_secret" yaml:"app_secret"`
}

// CredentialProvider returns the credentials of the platform. It is a hook so
// platform clients can supply per-call credentials (e.g. different official
// accounts). Most callers pass Credentials directly via CredentialsProvider.
type CredentialProvider func() Credentials

// CredentialsOf adapts a static Credentials value into a CredentialProvider.
func CredentialsOf(creds Credentials) CredentialProvider {
	return func() Credentials { return creds }
}

// Client is the shared HTTP + token foundation of every WeChat platform
// package. Platform clients embed *Client (through [EndpointClient]) and add
// their own typed endpoint methods on top. Build it with [NewClient]; the zero
// value is not usable.
//
// A Client is safe for concurrent use as long as its [Cache] is. Its
// configuration is fixed at construction.
type Client struct {
	sf         singleflight.Group
	cache      Cache
	baseURL    string
	httpClient *http.Client
	creds      CredentialProvider
	// custom token cache keys: the key prefix is derived from the appid so
	// different accounts sharing one client do not collide.
	cacheKeyPrefix string
}

// ClientOptions customises a core client.
type ClientOptions struct {
	// Cache stores access tokens; nil creates an in-process memory cache
	// (see NewMemoryCache).
	Cache Cache
	// Proxy is an optional HTTP(S) proxy URL used for outbound requests,
	// e.g. "http://127.0.0.1:7890". It only applies to the default HTTP
	// client: it is ignored when HTTPClient is set, and an unparsable URL is
	// silently ignored.
	Proxy string
	// BaseURL overrides the WeChat API host, without a trailing slash (tests,
	// gateways). Empty means DefaultBaseURL.
	BaseURL string
	// HTTPClient injects a custom client, used as is; when nil a client with a
	// 30s timeout (and Proxy, if set) is created.
	HTTPClient *http.Client
}

// NewClient builds a core Client. creds may be nil when the platform provides
// its own token source override; GetAccessToken then requests a token with
// empty credentials. creds is called once here to derive the token cache key
// prefix from the AppID, and again whenever credentials are needed. NewClient never fails;
// see ClientOptions for defaults.
func NewClient(creds CredentialProvider, opts ClientOptions) *Client {
	cache := opts.Cache
	if cache == nil {
		cache = NewMemoryCache()
	}
	transport := http.DefaultTransport
	if opts.Proxy != "" {
		if proxyURL, err := url.Parse(opts.Proxy); err == nil {
			if base, ok := http.DefaultTransport.(*http.Transport); ok {
				cloned := base.Clone()
				cloned.Proxy = http.ProxyURL(proxyURL)
				transport = cloned
			}
		}
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout:   defaultHTTPTimeout,
			Transport: transport,
		}
	}
	c := &Client{
		cache:      cache,
		baseURL:    baseURL,
		httpClient: httpClient,
		creds:      creds,
	}
	if creds != nil {
		c.cacheKeyPrefix = creds().AppID + ":"
	}
	return c
}

// Credentials returns the current credentials.
func (c *Client) Credentials() Credentials {
	if c.creds == nil {
		return Credentials{}
	}
	return c.creds()
}

// AppID returns the configured application id.
func (c *Client) AppID() string { return c.Credentials().AppID }

// WithCredentials builds a derived client that uses different credentials but
// shares the HTTP client, base URL and cache. Tokens are cached under a key
// prefixed with the AppID, so accounts sharing the cache do not collide. c is
// not modified.
func (c *Client) WithCredentials(creds Credentials) *Client {
	return NewClient(CredentialsOf(creds), ClientOptions{
		Cache:      c.cache,
		BaseURL:    c.baseURL,
		HTTPClient: c.httpClient,
	})
}

// fetchAccessToken calls the classic /cgi-bin/token endpoint and returns the raw
// token plus its server-reported lifetime in seconds.
func (c *Client) fetchAccessToken(ctx context.Context) (string, int, error) {
	query := url.Values{}
	query.Set("grant_type", "client_credential")
	query.Set("appid", c.AppID())
	query.Set("secret", c.Credentials().AppSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/cgi-bin/token?"+query.Encode(), nil)
	if err != nil {
		return "", 0, fmt.Errorf("wechat: build token request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("wechat: fetch access token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return "", 0, fmt.Errorf("wechat: read token response: %w", err)
	}
	if bizErr := scanErrorBody(data); bizErr != nil {
		return "", 0, bizErr
	}
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", 0, fmt.Errorf("wechat: decode access token: %w", err)
	}
	if result.AccessToken == "" {
		return "", 0, errors.New("wechat: empty access_token in response")
	}
	return result.AccessToken, result.ExpiresIn, nil
}

// GetAccessToken returns a usable access token, serving it from the cache when
// one is still valid. A token whose cached entry is missing or expired triggers
// a refresh from /cgi-bin/token; pass reload to skip the cache and force the
// refresh. A cache read error is treated as a miss.
//
// Concurrent callers for the same account share a single upstream token
// request (singleflight); they all receive its result, including an error
// caused by cancellation of the context of the caller that started it. The
// fetched token is cached for the server's expires_in minus a 2s safety
// margin; a response without a positive expires_in is not cached.
//
// Errors: errcode 40001, 40014 and 42001 return ErrorInvalidCredential,
// ErrorInvalidAccessToken and ErrorAccessTokenExpired; any other non-zero
// errcode returns an *APIError. Transport, decode and cache write failures are
// returned wrapped, and a response with an empty access_token is an error.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-access-token/api_getaccesstoken.html
func (c *Client) GetAccessToken(ctx context.Context, reload bool) (string, error) {
	cacheKey := c.cacheKeyPrefix + "access_token"
	if !reload {
		if token, ok, err := c.cache.Get(ctx, cacheKey); err == nil && ok {
			return token, nil
		}
	}
	val, err, _ := c.sf.Do(cacheKey, func() (any, error) {
		token, expiresIn, err := c.fetchAccessToken(ctx)
		if err != nil {
			return "", err
		}
		if expiresIn > 0 {
			ttl := time.Duration(expiresIn)*time.Second - tokenExpirySafetyMargin
			if err := c.cache.SetWithTTL(ctx, cacheKey, token, ttl); err != nil {
				return "", fmt.Errorf("wechat: cache access token: %w", err)
			}
		}
		return token, nil
	})
	if err != nil {
		return "", err
	}
	return val.(string), nil
}
