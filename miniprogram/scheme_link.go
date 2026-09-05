package miniprogram

import (
	"context"
)

// URLSchemeLinkPage is the target of a URL Scheme / URL Link when the app
// decides where the link opens inside the Mini Program.
type URLSchemeLinkPage struct {
	// Path inside the Mini Program, e.g. "pages/index/index".
	Path string `json:"path,omitempty"`
	// Query string for the target page, e.g. "foo=bar".
	Query string `json:"query,omitempty"`
	// EnvVersion of the Mini Program to open: develop, trial or release
	// (default).
	EnvVersion string `json:"env_version,omitempty"`
}

// LinkExpiry describes the validity window of a scheme or link.
type LinkExpiry struct {
	// IsExpire controls whether the link expires. false means permanent
	// (default). Permanence is subject to the quota rules of each endpoint.
	IsExpire bool `json:"is_expire,omitempty"`
	// ExpireType selects the expiry model: 1 = relative interval
	// (ExpireInterval), 2 = absolute timestamp (ExpireTime).
	ExpireType int `json:"expire_type,omitempty"`
	// ExpireTime is the Unix timestamp (seconds) of the absolute expiry.
	ExpireTime int64 `json:"expire_time,omitempty"`
	// ExpireInterval is the number of seconds from creation until expiry.
	ExpireInterval int64 `json:"expire_interval,omitempty"`
}

// GenerateSchemeRequest are the options of GenerateScheme. Exactly one of
// (jump target / expiry) groupings should be filled.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-scheme/api_generatescheme.html
type GenerateSchemeRequest struct {
	// JumpWxa is the Mini Program page to open when the scheme is used.
	JumpWxa *URLSchemeLinkPage `json:"jump_wxa,omitempty"`
	LinkExpiry
}

// GenerateSchemeResponse is returned by GenerateScheme.
type GenerateSchemeResponse struct {
	ErrResponse
	// OpenLink is the generated weixin://dl/business/?t=... scheme, to be
	// embedded in web pages or sent in messages.
	OpenLink string `json:"openlink"`
}

// GenerateScheme creates an encrypted URL Scheme (weixin://dl/business/?t=...)
// that opens a Mini Program page. Typically used when the Mini Program needs
// the link to work from outside the WeChat app ecosystem (SMS, email).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-scheme/api_generatescheme.html
func (w *MiniProgram) GenerateScheme(ctx context.Context, req *GenerateSchemeRequest) (*GenerateSchemeResponse, error) {
	var result GenerateSchemeResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/generatescheme", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateNFCSchemeRequest carries the NFC-specific parameters for
// GenerateNFCScheme.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-scheme/api_generatenfcscheme.html
type GenerateNFCSchemeRequest struct {
	// ModelID of the NFC tag model obtained from the Mini Program NFC module.
	ModelID string `json:"model_id,omitempty"`
	// SN is the serial number of the NFC tag.
	SN string `json:"sn,omitempty"`
	// JumpWxa is the Mini Program page to open (required).
	JumpWxa *URLSchemeLinkPage `json:"jump_wxa,omitempty"`
	LinkExpiry
}

// GenerateNFCSchemeResponse is returned by GenerateNFCScheme.
type GenerateNFCSchemeResponse struct {
	ErrResponse
	OpenLink string `json:"openlink"`
}

// GenerateNFCScheme creates an encrypted URL Scheme destined for an NFC tag.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-scheme/api_generatenfcscheme.html
func (w *MiniProgram) GenerateNFCScheme(ctx context.Context, req *GenerateNFCSchemeRequest) (*GenerateNFCSchemeResponse, error) {
	var result GenerateNFCSchemeResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/generatenfcscheme", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SchemeInfo describes a queried URL Scheme.
type SchemeInfo struct {
	// AppID that created the scheme.
	AppID string `json:"appid"`
	// Path the scheme opens.
	Path string `json:"path"`
	// Query the scheme carries.
	Query string `json:"query"`
	// CreateTime is the Unix timestamp (seconds) of creation.
	CreateTime int64 `json:"create_time"`
	// ExpireInterval is the remaining validity in seconds.
	ExpireInterval int64 `json:"expire_interval"`
	// ExpireTime is the absolute Unix expiry timestamp when applicable.
	ExpireTime int64 `json:"expire_time"`
	// EnvVersion opened by the scheme.
	EnvVersion string `json:"env_version"`
}

// QuerySchemeResponse is returned by QueryScheme.
type QuerySchemeResponse struct {
	ErrResponse
	// SchemeInfo of the queried scheme.
	SchemeInfo SchemeInfo `json:"scheme_info"`
}

// QueryScheme resolves the target page of an existing URL Scheme, useful to
// recover where a scheme leads or to manage redirects server side.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-scheme/api_queryscheme.html
func (w *MiniProgram) QueryScheme(ctx context.Context, scheme string) (*QuerySchemeResponse, error) {
	var result QuerySchemeResponse
	body := map[string]string{"scheme": scheme}
	if err := w.withAccessTokenPost(ctx, "/wxa/queryscheme", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateURLShortenLinkRequest are the options of GenerateShortLink (the
// "短链接" short link API).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/short-link/api_generateshortlink.html
type GenerateURLShortenLinkRequest struct {
	// PageURL is the page path of the Mini Program, optionally with query,
	// e.g. "pages/index/index?foo=bar".
	PageURL string `json:"page_url"`
	// PageTitle shown when the link is shared (up to 256 characters).
	PageTitle string `json:"page_title,omitempty"`
	// IsPermanent controls whether the link is permanent. Only one permanent
	// short link per page path is allowed and it cannot be regenerated, so
	// plan accordingly.
	IsPermanent bool `json:"is_permanent,omitempty"`
}

// GenerateShortLinkResponse is returned by GenerateShortLink.
type GenerateShortLinkResponse struct {
	ErrResponse
	// Link is the shortened link.
	Link string `json:"link"`
}

// GenerateShortLink shortens a page path into a link. One permanent link per
// path, or temporary links with a 30 day validity.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/short-link/api_generateshortlink.html
func (w *MiniProgram) GenerateShortLink(ctx context.Context, req *GenerateURLShortenLinkRequest) (*GenerateShortLinkResponse, error) {
	var result GenerateShortLinkResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/genwxashortlink", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateURLLinkRequest are the options of GenerateURLLink. URL Links
// (https://wxaurl.cn/...) are usable directly in a mobile browser and can be
// launched with wx-open-launch-weapp on web pages.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-link/api_generateurllink.html
type GenerateURLLinkRequest struct {
	// Path of the page inside the Mini Program.
	Path string `json:"path,omitempty"`
	// Query string for the target page.
	Query string `json:"query,omitempty"`
	// EnvVersion of the Mini Program to open.
	EnvVersion string `json:"env_version,omitempty"`
	LinkExpiry
}

// GenerateURLLinkResponse is returned by GenerateURLLink.
type GenerateURLLinkResponse struct {
	ErrResponse
	// URLLink is the generated https://wxaurl.cn/... link.
	URLLink string `json:"url_link"`
}

// GenerateURLLink creates an https URL Link that opens a Mini Program page.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-link/api_generateurllink.html
func (w *MiniProgram) GenerateURLLink(ctx context.Context, req *GenerateURLLinkRequest) (*GenerateURLLinkResponse, error) {
	var result GenerateURLLinkResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/generate_urllink", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// URLLinkInfo describes a queried URL Link.
type URLLinkInfo struct {
	// AppID that created the link.
	AppID string `json:"appid"`
	// Path the link opens.
	Path string `json:"path"`
	// Query the link carries.
	Query string `json:"query"`
	// CreateTime is the Unix timestamp (seconds) of creation.
	CreateTime int64 `json:"create_time"`
	// ExpireTime is the absolute Unix expiry timestamp when applicable.
	ExpireTime int64 `json:"expire_time"`
	// EnvVersion opened by the link.
	EnvVersion string `json:"env_version"`
	// CloudBase schema info when the link targets a cloud page.
	CloudBase string `json:"cloud_base,omitempty"`
}

// QueryURLLinkResponse is returned by QueryURLLink.
type QueryURLLinkResponse struct {
	ErrResponse
	URLLinkInfo URLLinkInfo `json:"url_link_info"`
}

// QueryURLLink resolves the target configuration of an existing URL Link.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/url-link/api_queryurllink.html
func (w *MiniProgram) QueryURLLink(ctx context.Context, urlLink string) (*QueryURLLinkResponse, error) {
	var result QueryURLLinkResponse
	body := map[string]string{"url_link": urlLink}
	if err := w.withAccessTokenPost(ctx, "/wxa/query_urllink", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
