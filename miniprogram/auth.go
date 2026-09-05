package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// LoginDoc is the canonical reference for the user login APIs.
const LoginDoc = "https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/"

// JsCode2SessionResponse is returned by JsCode2Session. The session_key must be
// kept secret server-side and is used to decrypt user data sent from the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/api_code2session.html
type JsCode2SessionResponse struct {
	ErrResponse
	// OpenID is the unique identifier of the user within this Mini Program.
	OpenID string `json:"openid"`
	// SessionKey is the login session key, valid until the user clears it.
	SessionKey string `json:"session_key"`
	// UnionID is present only when the user belongs to a WeChat Open Platform
	// account that links multiple apps.
	UnionID string `json:"unionid"`
}

// JsCode2Session exchanges a wx.login temporary code for the user openid and
// session_key. This is the entry point of every Mini Program user session.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/api_code2session.html
func (w *MiniProgram) JsCode2Session(ctx context.Context, code string) (*JsCode2SessionResponse, error) {
	query := url.Values{}
	query.Set("appid", w.config.AppID)
	query.Set("secret", w.config.AppSecret)
	query.Set("js_code", code)
	query.Set("grant_type", "authorization_code")
	var result JsCode2SessionResponse
	if err := w.getJSON(ctx, "/sns/jscode2session", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CheckSessionKey tells the server whether a user's login session is still
// valid. The signature is the HMAC-SHA256 of the empty string keyed by the
// session_key (lowercase hex). A nil error means the session is valid.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/api_checksessionkey.html
func (w *MiniProgram) CheckSessionKey(ctx context.Context, openid, sessionKey string) error {
	query := url.Values{}
	query.Set("openid", openid)
	query.Set("signature", hmacSHA256Hex([]byte(sessionKey), nil))
	query.Set("sig_method", "hmac_sha256")
	return w.withAccessToken(ctx, http.MethodGet, "/wxa/checksession", query, defaultReqOptions(), nil, nil)
}

// ResetUserSessionKey invalidates a user's current session key. When a new key
// is needed (for example after a security incident), the client must re-login
// via wx.login to obtain a fresh code. The signature is the HMAC-SHA256 of the
// empty string keyed by the current session_key.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/api_resetusersessionkey.html
func (w *MiniProgram) ResetUserSessionKey(ctx context.Context, openid, sessionKey string) error {
	query := url.Values{}
	query.Set("openid", openid)
	query.Set("signature", hmacSHA256Hex([]byte(sessionKey), nil))
	query.Set("sig_method", "hmac_sha256")
	return w.withAccessToken(ctx, http.MethodGet, "/wxa/resetusersessionkey", query, defaultReqOptions(), nil, nil)
}

// SnsOauth2Response mirrors the OAuth2 access_token payload of the WeChat
// website/Open Platform web authorization flow.
type SnsOauth2Response struct {
	ErrResponse
	// AccessToken is the OAuth2 web access token.
	AccessToken string `json:"access_token"`
	// ExpiresIn is the token lifetime in seconds.
	ExpiresIn int `json:"expires_in"`
	// RefreshToken can exchange a new access token after expiry.
	RefreshToken string `json:"refresh_token"`
	// OpenID identifies the user under this official account.
	OpenID string `json:"openid"`
	// Scope lists the granted permissions, space separated.
	Scope string `json:"scope"`
	// UnionID is present when the app is bound to a WeChat Open Platform
	// account.
	UnionID string `json:"unionid"`
}

// SnsOauth2 performs the OAuth2 web code-to-token exchange used by web pages
// served under an official account (scope=snsapi_userinfo). It is not part of
// the Mini Program login flow but is kept for services that also host web
// pages.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
func (w *MiniProgram) SnsOauth2(ctx context.Context, code string) (*SnsOauth2Response, error) {
	query := url.Values{}
	query.Set("appid", w.config.AppID)
	query.Set("secret", w.config.AppSecret)
	query.Set("code", code)
	query.Set("grant_type", "authorization_code")
	var result SnsOauth2Response
	if err := w.getJSON(ctx, "/sns/oauth2/access_token", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SnsRefreshOAuth2Token refreshes an expired OAuth2 web access token using its
// refresh_token.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
func (w *MiniProgram) SnsRefreshOAuth2Token(ctx context.Context, refreshToken string) (*SnsOauth2Response, error) {
	query := url.Values{}
	query.Set("appid", w.config.AppID)
	query.Set("grant_type", "refresh_token")
	query.Set("refresh_token", refreshToken)
	var result SnsOauth2Response
	if err := w.getJSON(ctx, "/sns/oauth2/refresh_token", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SnsOAuth2UserInfoResponse is the profile payload returned for an
// snsapi_userinfo scope access token.
type SnsOAuth2UserInfoResponse struct {
	ErrResponse
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

// GetOAuth2UserInfo fetches the user profile with an snsapi_userinfo access
// token.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
func (w *MiniProgram) GetOAuth2UserInfo(ctx context.Context, accessToken, openid string) (*SnsOAuth2UserInfoResponse, error) {
	query := url.Values{}
	query.Set("access_token", accessToken)
	query.Set("openid", openid)
	var result SnsOAuth2UserInfoResponse
	if err := w.getJSON(ctx, "/sns/userinfo", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPaidUnionIDRequest carries the parameters of GetPaidUnionID. All fields
// are passed as query parameters; exactly one of the payment identifier fields
// is required in addition to OpenID.
type GetPaidUnionIDRequest struct {
	// OpenID of the user whose unionid is requested (required).
	OpenID string
	// TransactionID of a WeChat Pay order of the user (alternative to
	// MchID+OutTradeNo).
	TransactionID string
	// MchID is the merchant id of the WeChat Pay order (alternative to
	// TransactionID).
	MchID string
	// OutTradeNo is the merchant order number of the WeChat Pay order.
	OutTradeNo string
}

// GetPaidUnionIDResponse is returned by GetPaidUnionID.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/basic-info/api_getpaidunionid.html
type GetPaidUnionIDResponse struct {
	ErrResponse
	// UnionID of the user, or an empty string when the account is not bound to
	// an Open Platform account.
	UnionID string `json:"unionid"`
}

// GetPaidUnionID resolves a user's unionid through a WeChat Pay order the user
// paid, which is useful when the paid order allows linking to the unionid.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/basic-info/api_getpaidunionid.html
func (w *MiniProgram) GetPaidUnionID(ctx context.Context, req *GetPaidUnionIDRequest) (*GetPaidUnionIDResponse, error) {
	query := url.Values{}
	query.Set("openid", req.OpenID)
	if req.TransactionID != "" {
		query.Set("transaction_id", req.TransactionID)
	}
	if req.MchID != "" {
		query.Set("mch_id", req.MchID)
	}
	if req.OutTradeNo != "" {
		query.Set("out_trade_no", req.OutTradeNo)
	}
	var result GetPaidUnionIDResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxa/getpaidunionid", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPluginOpenPidRequest identifies a user of a plugin that shares data with
// the host Mini Program.
type GetPluginOpenPidRequest struct {
	// OpenID of the user within the plugin that is sharing the data.
	OpenID string `json:"openid"`
}

// GetPluginOpenPidResponse carries the mapped openids.
type GetPluginOpenPidResponse struct {
	ErrResponse
	// OpenPID is the same openid that the plugin sees for the user.
	OpenPID string `json:"openpid"`
}

// GetPluginOpenPid returns the openpid a plugin observes for a user, so the
// host app can correlate data coming from the plugin.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/basic-info/api_getpluginopenpid.html
func (w *MiniProgram) GetPluginOpenPid(ctx context.Context, req *GetPluginOpenPidRequest) (*GetPluginOpenPidResponse, error) {
	var result GetPluginOpenPidResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/wxa/getpluginopenpid", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserEncryptKeyRequest selects which run-time encryption key to fetch.
type GetUserEncryptKeyRequest struct {
	// OpenID of the user.
	OpenID string `json:"openid"`
	// SessionKey of the user, from JsCode2Session.
	SessionKey string `json:"session_key"`
	// Signature of the "session_key" string produced with the WeChat Pay v3
	// private key (optional when the app does not use WeChat Pay).
	Signature string `json:"signature"`
	// SignatureMethod is the signing algorithm, "HMAC-SHA256" is the only
	// supported value (optional).
	SignatureMethod string `json:"sig_method,omitempty"`
}

// EncryptKeyInfo describes one encryption key issued to the user.
type EncryptKeyInfo struct {
	// EncryptKey is the base64 encryption key.
	EncryptKey string `json:"encrypt_key"`
	// Version of the encryption key.
	Version int `json:"version"`
	// ExpireAt is the Unix timestamp (seconds) when the key expires.
	ExpireAt int `json:"expire_at"`
	// CreateAt is the Unix timestamp (seconds) when the key was created.
	CreateAt int `json:"create_time"`
	// Iv is the 16-byte base64 initialisation vector paired with the key.
	Iv string `json:"iv"`
}

// GetUserEncryptKeyResponse is returned by GetUserEncryptKey.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/internet/api_getuserencryptkey.html
type GetUserEncryptKeyResponse struct {
	ErrResponse
	// EncryptKeyInfos carries the (current and previous) encryption keys.
	EncryptKeyInfos []EncryptKeyInfo `json:"key_info_list"`
}

// GetUserEncryptKey fetches the run-time encryption keys for a user, needed to
// decrypt the encrypted data received through some APIs (e.g. the encrypted
// phone number under the new privacy scheme).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/internet/api_getuserencryptkey.html
func (w *MiniProgram) GetUserEncryptKey(ctx context.Context, req *GetUserEncryptKeyRequest) (*GetUserEncryptKeyResponse, error) {
	var result GetUserEncryptKeyResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/wxa/business/getuserencryptkey", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
