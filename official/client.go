package official

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-sphere/weixin-api/core"
)

// Type aliases re-exported so official callers use one import.
type (
	// ErrResponse is the universal WeChat error envelope.
	ErrResponse = core.ErrResponse
	// APIError is a WeChat business error carrying the rid trace id.
	APIError = core.APIError
	// Cache stores access tokens and tickets.
	Cache = core.Cache
)

// Sentinel errors re-exported for convenience.
var (
	ErrorInvalidCredential  = core.ErrorInvalidCredential
	ErrorAccessTokenExpired = core.ErrorAccessTokenExpired
	ErrorInvalidAccessToken = core.ErrorInvalidAccessToken
)

// Config holds everything needed to authenticate an Official Account client.
type Config struct {
	// AppID is the official account appid.
	AppID string `json:"app_id" yaml:"app_id"`
	// AppSecret of the official account.
	AppSecret string `json:"app_secret" yaml:"app_secret"`
	// Proxy is an optional HTTP(S) proxy URL.
	Proxy string `json:"proxy" yaml:"proxy"`
}

// OfficialAccount is the WeChat Official Account (公众号/服务号) API client. It
// embeds *core.Client (transport + token management) and exposes typed
// endpoint methods.
type OfficialAccount struct {
	*core.Client
	config Config
}

// NewOfficialAccount builds an Official Account client. cache may be nil.
func NewOfficialAccount(config Config, cache core.Cache) *OfficialAccount {
	client := core.NewClient(core.CredentialsOf(core.Credentials{
		AppID:     config.AppID,
		AppSecret: config.AppSecret,
	}), core.ClientOptions{Cache: cache, Proxy: config.Proxy})
	return &OfficialAccount{Client: client, config: config}
}

// newOfficialWithHTTPClient is the test seam used by unit tests.
func newOfficialWithHTTPClient(config Config, cache core.Cache, baseURL string, httpClient *http.Client) *OfficialAccount {
	oa := &OfficialAccount{config: config}
	oa.Client = core.NewClient(core.CredentialsOf(core.Credentials{
		AppID:     config.AppID,
		AppSecret: config.AppSecret,
	}), core.ClientOptions{
		Cache:      cache,
		Proxy:      config.Proxy,
		BaseURL:    baseURL,
		HTTPClient: httpClient,
	})
	return oa
}

// AppID returns the configured official account appid.
func (oa *OfficialAccount) AppID() string { return oa.config.AppID }

// withToken is an internal alias of core.Client.WithToken.
func (oa *OfficialAccount) withToken(ctx context.Context, method, path string, query url.Values, opts core.RequestOptions, body, dst any) error {
	return oa.WithToken(ctx, method, path, query, opts, body, dst)
}

// withTokenPost is withToken restricted to POST.
func (oa *OfficialAccount) withTokenPost(ctx context.Context, path string, query url.Values, opts core.RequestOptions, body, dst any) error {
	return oa.WithTokenPost(ctx, path, query, opts, body, dst)
}

// withTokenRaw performs an access-token request returning raw bytes.
func (oa *OfficialAccount) withTokenRaw(ctx context.Context, method, path string, query url.Values, opts core.RequestOptions, body any) ([]byte, error) {
	return oa.WithTokenRaw(ctx, method, path, query, opts, body)
}
