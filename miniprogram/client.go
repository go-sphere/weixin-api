// Package miniprogram is a typed Go client for the WeChat Mini Program server
// APIs (https://developers.weixin.qq.com/miniprogram/dev/server/API/).
//
// It is built on the shared github.com/go-sphere/weixin-api/core package,
// which supplies the HTTP transport, the universal {"errcode":..,"errmsg":..}
// envelope handling, and access-token management. The MiniProgram struct embeds
// *core.Client, so every core helper (GetAccessToken, Do, CallJSON, WithToken,
// ...) is promoted onto it.
//
// # Conventions
//
//   - Methods that require an access token acquire (and cache) it automatically
//     and transparently retry once after a token-expiry style error, forcing a
//     token refresh before the retry.
//   - All methods return typed results. API responses are checked for the
//     universal {"errcode":..,"errmsg":..} envelope and surface a non-zero code
//     as a [*core.APIError] (or a core sentinel error for access-token
//     failures) instead of silently returning partial data.
//   - Methods returning binary payloads (QR codes, media files) return the raw
//     bytes and sniff for an embedded JSON error envelope.
//   - Every method documents the exact upstream reference page so both humans
//     and AI agents can jump from Go doc straight to the authoritative
//     request/response contract.
package miniprogram

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-sphere/weixin-api/core"
)

// Core aliases re-exported for callers that used the single package before the
// split. New code should import core directly.
type (
	// ErrResponse is the universal WeChat error envelope.
	ErrResponse = core.ErrResponse
	// APIError is a WeChat business error carrying the rid trace id.
	APIError = core.APIError
	// Cache stores access tokens and tickets.
	Cache = core.Cache
	// AccessTokenResponse is returned by the stable token API.
	AccessTokenResponse = core.AccessTokenResponse
)

// Sentinel errors re-exported for convenience.
var (
	ErrorInvalidCredential  = core.ErrorInvalidCredential
	ErrorAccessTokenExpired = core.ErrorAccessTokenExpired
	ErrorInvalidAccessToken = core.ErrorInvalidAccessToken
)

// Well-known error codes re-exported for convenience.
const (
	ErrCodeInvalidCredential  = core.ErrCodeInvalidCredential
	ErrCodeInvalidAccessToken = core.ErrCodeInvalidAccessToken
	ErrCodeAccessTokenExpired = core.ErrCodeAccessTokenExpired
	ErrCodeAPIFreqOutOfLimit  = core.ErrCodeAPIFreqOutOfLimit
)

// DefaultBaseURL is the host of the Mini Program server APIs.
const DefaultBaseURL = core.DefaultBaseURL

// MiniAppEnv is the deployment environment selector of a Mini Program.
// Several APIs need it to decide which version of the Mini Program to target.
type MiniAppEnv string

const (
	// MiniAppEnvRelease is the released Mini Program version.
	MiniAppEnvRelease MiniAppEnv = "release" // 正式版
	// MiniAppEnvTrial is the trial (体验版) Mini Program version.
	MiniAppEnvTrial MiniAppEnv = "trial" // 体验版
	// MiniAppEnvDevelop is the in-development (开发版) Mini Program version.
	MiniAppEnvDevelop MiniAppEnv = "develop" // 开发版
)

// String returns the raw environment string.
func (e MiniAppEnv) String() string { return string(e) }

// subscribeState maps a MiniAppEnv to the value expected by the subscribe
// message API (miniprogram_state field): "formal", "trial" and "developer".
func (e MiniAppEnv) subscribeState() string {
	switch e {
	case MiniAppEnvTrial:
		return "trial"
	case MiniAppEnvDevelop:
		return "developer"
	default:
		return "formal"
	}
}

// Config holds everything needed to authenticate a Mini Program client.
type Config struct {
	// AppID is the Mini Program application id (appid).
	AppID string `json:"app_id" yaml:"app_id"`
	// AppSecret is the Mini Program application secret (secret).
	AppSecret string `json:"app_secret" yaml:"app_secret"`
	// AppKey is the 米大师 / 虚拟支付 pay signing key configured in the Mini
	// Program console. It is only used by the XPay (virtual payment) methods
	// to compute the pay_sig request signature.
	AppKey string `json:"app_key" yaml:"app_key"`
	// Proxy is an optional HTTP(S) proxy URL used for outbound requests,
	// e.g. "http://127.0.0.1:7890".
	Proxy string `json:"proxy" yaml:"proxy"`
	// Env is the Mini Program deployment environment used to derive defaults
	// for APIs that depend on it. Empty means MiniAppEnvRelease.
	Env MiniAppEnv `json:"env" yaml:"env"`
}

// MiniProgram is the WeChat Mini Program API client. It embeds *core.Client
// (HTTP transport + token management) and exposes one method per upstream
// endpoint.
type MiniProgram struct {
	*core.Client
	config Config
}

// NewMiniProgram builds a Mini Program client from a Config. cache may be nil,
// in which case a fresh in-process memory cache is created; pass a shared
// core.Cache when the client runs in a multi-instance deployment. An empty Env
// defaults to the release environment.
func NewMiniProgram(config Config, cache core.Cache) *MiniProgram {
	if config.Env == "" {
		config.Env = MiniAppEnvRelease
	}
	client := core.NewClient(core.CredentialsOf(core.Credentials{
		AppID:     config.AppID,
		AppSecret: config.AppSecret,
	}), core.ClientOptions{Cache: cache, Proxy: config.Proxy})
	return &MiniProgram{Client: client, config: config}
}

// NewWechat is kept as an alias of NewMiniProgram for callers migrating from
// the pre-split package name.
//
// Deprecated: use NewMiniProgram.
func NewWechat(config Config, cache core.Cache) *MiniProgram {
	return NewMiniProgram(config, cache)
}

// newMiniProgramWithHTTPClient is the test seam used by unit tests to point the
// client at an httptest server and inject a fake HTTP client.
func newMiniProgramWithHTTPClient(config Config, cache core.Cache, baseURL string, httpClient *http.Client) *MiniProgram {
	if config.Env == "" {
		config.Env = MiniAppEnvRelease
	}
	mp := &MiniProgram{config: config}
	mp.Client = core.NewClient(core.CredentialsOf(core.Credentials{
		AppID:     config.AppID,
		AppSecret: config.AppSecret,
	}), core.ClientOptions{
		Cache:      cache,
		Proxy:      config.Proxy,
		BaseURL:    baseURL,
		HTTPClient: httpClient,
	})
	return mp
}

// AppID returns the configured Mini Program application id.
func (w *MiniProgram) AppID() string { return w.config.AppID }

// Config returns the effective configuration after defaults were applied.
func (w *MiniProgram) Config() Config { return w.config }

// withAccessToken is an internal alias of core.Client.WithToken.
func (w *MiniProgram) withAccessToken(ctx context.Context, method, path string, query url.Values, opts requestOptions, body, dst any) error {
	return w.WithToken(ctx, method, path, query, opts.toCore(), body, dst)
}

// getJSON performs a JSON GET and decodes dst.
func (w *MiniProgram) getJSON(ctx context.Context, path string, query url.Values, dst any) error {
	return w.GetJSON(ctx, path, query, dst)
}

// withAccessTokenPost is withAccessToken restricted to POST.
func (w *MiniProgram) withAccessTokenPost(ctx context.Context, path string, query url.Values, opts requestOptions, body, dst any) error {
	return w.WithTokenPost(ctx, path, query, opts.toCore(), body, dst)
}

// withAccessTokenRaw performs an access-token request returning raw bytes.
func (w *MiniProgram) withAccessTokenRaw(ctx context.Context, method, path string, query url.Values, opts requestOptions, body any) ([]byte, error) {
	return w.WithTokenRaw(ctx, method, path, query, opts.toCore(), body)
}

// withAccessTokenUpload performs an access-token multipart upload returning the
// raw response bytes.
func (w *MiniProgram) withAccessTokenUpload(ctx context.Context, path string, query url.Values, opts requestOptions, form url.Values, fileField, fileName, fileContentType string, content interface{ Read([]byte) (int, error) }) ([]byte, error) {
	return w.WithTokenUpload(ctx, path, query, opts.toCore(), form, fileField, fileName, fileContentType, readerOf(content))
}
