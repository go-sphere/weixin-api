package core

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
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

// JSONBytes is a pre-encoded JSON payload. It behaves like a raw []byte request
// body but is sent with the JSON content type, for callers that must sign the
// exact bytes that go on the wire (e.g. the Mini Program XPay pay_sig) while
// still honouring the endpoint's JSON contract.
type JSONBytes []byte

// Client is the shared HTTP + token foundation of every WeChat platform
// package. Platform clients embed *Client and add their own typed endpoint
// methods on top.
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
	// Cache for access tokens and tickets; nil creates an in-process memory
	// cache.
	Cache Cache
	// Proxy is an optional HTTP(S) proxy URL used for outbound requests,
	// e.g. "http://127.0.0.1:7890".
	Proxy string
	// BaseURL overrides the WeChat API host (used by tests).
	BaseURL string
	// HTTPClient injects a custom client (used by tests); when nil a default
	// client with a 30s timeout is created.
	HTTPClient *http.Client
}

// NewClient builds a core Client. creds may be nil when the platform provides
// its own token source override.
func NewClient(creds CredentialProvider, opts ClientOptions) *Client {
	cache := opts.Cache
	if cache == nil {
		cache = NewMemoryCache()
	}
	transport := http.DefaultTransport
	if opts.Proxy != "" {
		if base, ok := http.DefaultTransport.(*http.Transport); ok {
			transport = base.Clone()
			if proxyURL, err := url.Parse(opts.Proxy); err == nil {
				transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
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
// shares the transport and cache. Useful to manage several accounts with one
// HTTP setup.
func (c *Client) WithCredentials(creds Credentials) *Client {
	return NewClient(CredentialsOf(creds), ClientOptions{
		Cache:      c.cache,
		BaseURL:    c.baseURL,
		HTTPClient: c.httpClient,
	})
}

// tokenResponse is the shared shape of the /cgi-bin/token and
// /cgi-bin/ticket/getticket responses.
type tokenResponse struct {
	ErrResponse
	AccessToken string `json:"access_token,omitempty"`
	Ticket      string `json:"ticket,omitempty"`
	ExpiresIn   int    `json:"expires_in"`
}

// AccessTokenResponse is the JSON returned by the stable access token API
// (cgi-bin/stable_token).
type AccessTokenResponse struct {
	ErrResponse
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in"`
}

// fetchAccessToken calls the classic /cgi-bin/token endpoint and returns the
// raw token plus its server-reported lifetime in seconds.
func (c *Client) fetchAccessToken(ctx context.Context) (string, int, error) {
	var result tokenResponse
	query := url.Values{}
	query.Set("grant_type", "client_credential")
	query.Set("appid", c.AppID())
	query.Set("secret", c.Credentials().AppSecret)
	if err := c.GetJSON(ctx, "/cgi-bin/token", query, &result); err != nil {
		return "", 0, err
	}
	return result.AccessToken, result.ExpiresIn, nil
}

// GetAccessToken returns a usable access token, serving it from the cache when
// one is still valid. Concurrent callers share a single upstream token request
// (singleflight). A token whose cached entry is missing or expired triggers a
// refresh; pass reload to force the refresh.
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

// GetStableAccessToken fetches a stable access token via the
// cgi-bin/stable_token endpoint. Most callers should rely on GetAccessToken.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-access-token/api_getstableaccesstoken.html
func (c *Client) GetStableAccessToken(ctx context.Context, forceRefresh bool) (*AccessTokenResponse, error) {
	body := map[string]any{
		"grant_type":    "client_credential",
		"appid":         c.AppID(),
		"secret":        c.Credentials().AppSecret,
		"force_refresh": forceRefresh,
	}
	var result AccessTokenResponse
	if err := c.PostJSON(ctx, "/cgi-bin/stable_token", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetJsTicket returns a valid JS-SDK ticket (type=jsapi) used to sign wx.config
// for the JS-SDK. Tickets are cached like access tokens and shared across
// concurrent callers through singleflight. Pass reload to bypass the cache.
func (c *Client) GetJsTicket(ctx context.Context, reload bool) (string, error) {
	cacheKey := c.cacheKeyPrefix + "jsapi_ticket"
	if !reload {
		if ticket, ok, err := c.cache.Get(ctx, cacheKey); err == nil && ok {
			return ticket, nil
		}
	}
	val, err, _ := c.sf.Do(cacheKey, func() (any, error) {
		var result tokenResponse
		query := url.Values{}
		query.Set("type", "jsapi")
		err := c.WithToken(ctx, http.MethodGet, "/cgi-bin/ticket/getticket", query, RequestOptions{Retryable: false}, nil, &result)
		if err != nil {
			return "", err
		}
		if result.ExpiresIn > 0 {
			ttl := time.Duration(result.ExpiresIn)*time.Second - tokenExpirySafetyMargin
			if err := c.cache.SetWithTTL(ctx, cacheKey, result.Ticket, ttl); err != nil {
				return "", err
			}
		}
		return result.Ticket, nil
	})
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// ============================================================
// Request options and token-based request helpers.
// ============================================================

// RequestOptions controls per-call behaviour of token-based methods.
type RequestOptions struct {
	// Retryable retries once after an access-token-expiry error with a forced
	// token refresh.
	Retryable bool
	// ReloadAccessToken forces a fresh access token, bypassing the cache.
	ReloadAccessToken bool
}

// DefaultRequestOptions returns the default options: retries enabled and no
// forced token reload.
func DefaultRequestOptions() RequestOptions {
	return RequestOptions{Retryable: true}
}

// WithToken runs an access-token based HTTP call: it resolves the token,
// appends it as the access_token query parameter, issues the request and, when
// opts.Retryable is true and WeChat reports an access-token expiry style error,
// retries once after forcing a token refresh.
//
// body is JSON encoded when non-nil; dst, when non-nil, receives the decoded
// JSON response.
func (c *Client) WithToken(ctx context.Context, method, path string, query url.Values, opts RequestOptions, body, dst any) error {
	token, err := c.GetAccessToken(ctx, opts.ReloadAccessToken)
	if err != nil {
		return err
	}
	merged := mergeQuery(query)
	merged.Set("access_token", token)
	if err := c.CallJSON(ctx, method, path, merged, body, dst); err != nil {
		if opts.Retryable && isNeedRetryError(err) {
			opts.Retryable = false
			opts.ReloadAccessToken = true
			return c.WithToken(ctx, method, path, query, opts, body, dst)
		}
		return err
	}
	return nil
}

// WithTokenPost is WithToken restricted to POST, matching the endpoints that
// only accept POST bodies.
func (c *Client) WithTokenPost(ctx context.Context, path string, query url.Values, opts RequestOptions, body, dst any) error {
	return c.WithToken(ctx, http.MethodPost, path, query, opts, body, dst)
}

// WithTokenRaw performs an HTTP call behind an access token and returns the
// raw response bytes (binary payload such as an image or media file, or a JSON
// body when the caller prefers manual decoding). A JSON business-error body is
// detected and returned as an error.
func (c *Client) WithTokenRaw(ctx context.Context, method, path string, query url.Values, opts RequestOptions, body any) ([]byte, error) {
	token, err := c.GetAccessToken(ctx, opts.ReloadAccessToken)
	if err != nil {
		return nil, err
	}
	merged := mergeQuery(query)
	merged.Set("access_token", token)
	data, err := c.Do(ctx, method, path, merged, body)
	if err != nil {
		if opts.Retryable && isNeedRetryError(err) {
			opts.Retryable = false
			opts.ReloadAccessToken = true
			return c.WithTokenRaw(ctx, method, path, query, opts, body)
		}
		return nil, err
	}
	return data, nil
}

// WithTokenUpload uploads multipart/form-data content behind an access token
// and returns the raw response bytes (JSON for most upload endpoints).
func (c *Client) WithTokenUpload(ctx context.Context, path string, query url.Values, opts RequestOptions, form url.Values, fileField, fileName, fileContentType string, content io.Reader) ([]byte, error) {
	token, err := c.GetAccessToken(ctx, opts.ReloadAccessToken)
	if err != nil {
		return nil, err
	}
	merged := mergeQuery(query)
	merged.Set("access_token", token)
	data, err := c.multipartUpload(ctx, path, merged, form, fileField, fileName, fileContentType, content)
	if err != nil {
		if opts.Retryable && isNeedRetryError(err) {
			opts.Retryable = false
			opts.ReloadAccessToken = true
			return c.WithTokenUpload(ctx, path, query, opts, form, fileField, fileName, fileContentType, content)
		}
		return nil, err
	}
	return data, nil
}

// ============================================================
// Raw HTTP helpers (exported so platform packages can send any request).
// ============================================================

// Do performs a single HTTP request and returns the full response body.
// The caller controls content encoding via body:
//
//   - nil: no request body;
//   - []byte: raw payload with application/octet-stream content type;
//   - url.Values: form encoded payload;
//   - any other value: JSON encoded payload.
//
// The response is validated: HTTP status codes other than 200 are turned into
// errors, and a JSON response whose errcode is non-zero is turned into a
// classified business error. Raw bytes are returned to the caller for further
// decoding.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	var (
		reader      io.Reader
		contentType = "application/json; charset=utf-8"
	)
	switch b := body.(type) {
	case nil:
	case []byte:
		reader = bytes.NewReader(b)
		contentType = "application/octet-stream"
	case JSONBytes:
		reader = bytes.NewReader(b)
	case url.Values:
		reader = strings.NewReader(b.Encode())
		contentType = "application/x-www-form-urlencoded; charset=utf-8"
	case io.Reader:
		reader = b
		contentType = "application/octet-stream"
	default:
		data, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("wechat: encode request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("wechat: build request: %w", err)
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wechat: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("wechat: read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if bizErr := scanErrorBody(data); bizErr != nil {
			return nil, bizErr
		}
		return nil, fmt.Errorf("wechat: unexpected HTTP status %d for %s %s", resp.StatusCode, method, path)
	}
	if bizErr := scanErrorBody(data); bizErr != nil {
		return nil, bizErr
	}
	return data, nil
}

// PostJSON performs a JSON POST request and decodes the response into dst
// (dst may be nil when the caller only cares about the error). path must
// already include the full API path and query any required credentials.
func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, body, dst any) error {
	return c.CallJSON(ctx, http.MethodPost, path, query, body, dst)
}

// GetJSON performs a JSON GET request and decodes the response into dst.
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, dst any) error {
	return c.CallJSON(ctx, http.MethodGet, path, query, nil, dst)
}

// CallJSON performs a JSON HTTP request and decodes the response into dst.
func (c *Client) CallJSON(ctx context.Context, method, path string, query url.Values, body, dst any) error {
	data, err := c.Do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if dst == nil {
		return nil
	}
	return DecodeJSON(data, dst)
}

// multipartUpload performs a multipart/form-data upload request. form carries
// the regular fields, fileField/fileName the media field and content its bytes.
// It returns the raw response bytes (typically JSON).
func (c *Client) multipartUpload(ctx context.Context, path string, query url.Values, form url.Values, fileField, fileName string, contentType string, content io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for key, values := range form {
		for _, v := range values {
			if err := mw.WriteField(key, v); err != nil {
				return nil, err
			}
		}
	}
	if fileField != "" {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, fileField, fileName))
		if contentType != "" {
			h.Set("Content-Type", contentType)
		}
		fw, err := mw.CreatePart(h)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(fw, content); err != nil {
			return nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return nil, err
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wechat: upload %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("wechat: read upload response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if bizErr := scanErrorBody(data); bizErr != nil {
			return nil, bizErr
		}
		return nil, fmt.Errorf("wechat: unexpected HTTP status %d for upload %s", resp.StatusCode, path)
	}
	return data, nil
}

// mergeQuery shallow-copies src into a fresh url.Values.
func mergeQuery(src url.Values) url.Values {
	dst := make(url.Values, len(src))
	for k, vs := range src {
		dst[k] = append([]string(nil), vs...)
	}
	return dst
}

// hmacSHA256Hex computes the lowercase hex HMAC-SHA256 of msg using key.
func hmacSHA256Hex(key, msg []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(msg)
	return hex.EncodeToString(mac.Sum(nil))
}
