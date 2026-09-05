package miniprogram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// OpenAPIMgmtDoc is the canonical reference for the open API management
// endpoints.
const OpenAPIMgmtDoc = "https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/"

// QuotaItem describes one bucket of an API quota.
type QuotaItem struct {
	// QuotaID of the bucket, "" for the daily quota bucket.
	QuotaID string `json:"quota_id,omitempty"`
	// QuotaLimit of the bucket (requests per day, -1 means unlimited).
	QuotaLimit int `json:"quota_limit,omitempty"`
	// QuotaUsed of the bucket today.
	QuotaUsed int `json:"quota_used,omitempty"`
	// QuotaRemain of the bucket today.
	QuotaRemain int `json:"quota_remain,omitempty"`
	// Period of the bucket quota (e.g. "2592000" for monthly buckets).
	Period int `json:"period,omitempty"`
}

// GetAPICallQuotaResponse is returned by GetAPICallQuota.
type GetAPICallQuotaResponse struct {
	ErrResponse
	// Quota of the queried cgi path.
	Quota []QuotaItem `json:"quota,omitempty"`
	// RateLimit describes the per-minute/per-hour throttling of the path.
	RateLimit struct {
		// Calls of the current rate window.
		Calls int `json:"calls,omitempty"`
		// MaxCalls allowed per window.
		MaxCalls int `json:"max_calls,omitempty"`
		// MaxCallsPerSecond allowed.
		MaxCallsPerSecond int `json:"max_calls_per_sec,omitempty"`
		// Period of the rate window (seconds).
		Period int `json:"period,omitempty"`
	} `json:"rate_limit,omitempty"`
	// ComponentRateLimit is only populated for third-party platform tokens.
	ComponentRateLimit struct {
		// Calls of the current window.
		Calls int `json:"calls,omitempty"`
		// MaxCalls of the window.
		MaxCalls int `json:"max_calls,omitempty"`
		// Period of the window (seconds).
		Period int `json:"period,omitempty"`
	} `json:"component_rate_limit,omitempty"`
}

// GetAPICallQuota returns the call quota consumption of one API path
// (e.g. "wxa/getwxacodeunlimit").
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_getapiquota.html
func (w *MiniProgram) GetAPICallQuota(ctx context.Context, cgiPath string) (*GetAPICallQuotaResponse, error) {
	query := url.Values{}
	query.Set("cgi_path", cgiPath)
	var result GetAPICallQuotaResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/openapi/quota/get", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ClearAPICallQuota clears the daily quota consumption of one API path. Only a
// limited number of clears per month is allowed.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_clearquota.html
func (w *MiniProgram) ClearAPICallQuota(ctx context.Context, cgiPath string) error {
	query := url.Values{}
	query.Set("cgi_path", cgiPath)
	return w.withAccessToken(ctx, http.MethodPost, "/cgi-bin/openapi/quota/clear", query, defaultReqOptions(), nil, nil)
}

// GetAPIQuotaResponse is returned by GetAPIQuota (the alias of
// GetAPICallQuota).
type GetAPIQuotaResponse = GetAPICallQuotaResponse

// GetAPIQuota is an alias of GetAPICallQuota kept for readability.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_getapiquota.html
func (w *MiniProgram) GetAPIQuota(ctx context.Context, cgiPath string) (*GetAPIQuotaResponse, error) {
	return w.GetAPICallQuota(ctx, cgiPath)
}

// ClearAPIQuota is an alias of ClearAPICallQuota.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_clearquota.html
func (w *MiniProgram) ClearAPIQuota(ctx context.Context, cgiPath string) error {
	return w.ClearAPICallQuota(ctx, cgiPath)
}

// CallbackCheckRequest asks WeChat to perform a callback reachability test.
type CallbackCheckRequest struct {
	// Action must be "check".
	Action string `json:"action,omitempty"`
	// CheckOperator selects the probe: "all" (both domain and IP reachability
	// via DNS+connect) or "dns".
	CheckOperator string `json:"check_operator,omitempty"`
	// CheckType is 1 for domain, 2 for IP reachability (default all).
	CheckType int `json:"check_type,omitempty"`
}

// CallbackCheckResponseItem is one probe result.
type CallbackCheckResponseItem struct {
	// CallbackURL checked.
	CallbackURL string `json:"callback_url"`
	// DNS of the callback domain resolved.
	DNS string `json:"dns,omitempty"`
	// IP of the resolved callback host.
	IP string `json:"ip,omitempty"`
	// Port probed.
	Port int `json:"port,omitempty"`
	// IsOk is 1 when the probe succeeded.
	IsOk int `json:"is_ok,omitempty"`
	// Error message when the probe failed.
	Error string `json:"error,omitempty"`
}

// CallbackCheckResponse is returned by CallbackCheck.
type CallbackCheckResponse struct {
	ErrResponse
	// ToVerify are the check results.
	ToVerify []CallbackCheckResponseItem `json:"to_verify,omitempty"`
}

// CallbackCheck probes the configured callback URL from WeChat's network so
// that the domain/IP reachability problem is diagnosed.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_callbackcheck.html
func (w *MiniProgram) CallbackCheck(ctx context.Context, req *CallbackCheckRequest) (*CallbackCheckResponse, error) {
	var result CallbackCheckResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/callback/check", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// APIDomainIPItem is one outbound IP of WeChat's API servers.
type APIDomainIPItem struct {
	// IP of WeChat's API egress.
	IP string `json:"ip,omitempty"`
	// IPType of the address.
	IPType int `json:"ip_type,omitempty"`
	// IsSubDomain, when 1, marks a shared egress subset.
	IsSubDomain int `json:"is_sub_domain,omitempty"`
	// Domain is the outbound domain, e.g. "api.weixin.qq.com".
	Domain string `json:"domain,omitempty"`
	// Hints for firewall allow-listing.
	Hints string `json:"hints,omitempty"`
}

// GetAPIDomainIPResponse is returned by GetAPIDomainIP.
type GetAPIDomainIPResponse struct {
	ErrResponse
	// IPList of the API egress addresses.
	IPList []APIDomainIPItem `json:"ip_list,omitempty"`
}

// GetAPIDomainIP lists the IP ranges WeChat API servers use outbound, so the
// merchant firewall can allow-list them. The addresses rotate; refresh the list
// periodically.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_getapidomainip.html
func (w *MiniProgram) GetAPIDomainIP(ctx context.Context) (*GetAPIDomainIPResponse, error) {
	var result GetAPIDomainIPResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/get_api_domain_ip", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCallbackIPResponse is returned by GetCallbackIP.
type GetCallbackIPResponse struct {
	ErrResponse
	// IPList of the WeChat callback egress addresses.
	IPList []string `json:"ip_list,omitempty"`
}

// GetCallbackIP lists the IP ranges that WeChat uses when pushing message
// callbacks, so the merchant can verify that incoming requests really come from
// WeChat.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_getcallbackip.html
func (w *MiniProgram) GetCallbackIP(ctx context.Context) (*GetCallbackIPResponse, error) {
	var result GetCallbackIPResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/getcallbackip", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRidInfoResponse is returned by GetRidInfo. It explains an API error whose
// errmsg carried a rid.
type GetRidInfoResponse struct {
	ErrResponse
	// RequestTime is when the failing request reached WeChat.
	RequestTime int64 `json:"request_time,omitempty"`
	// ErrInfo carries the diagnostic chain of the request.
	ErrInfo GetRidInfoErrInfo `json:"err_info,omitempty"`
}

// GetRidInfoErrInfo is the diagnostic payload of a rid lookup.
type GetRidInfoErrInfo struct {
	// ErrCode returned by the failing request.
	ErrCode int `json:"errcode,omitempty"`
	// ErrMsg returned by the failing request.
	ErrMsg string `json:"errmsg,omitempty"`
	// OutTraceNo is the internal trace id for WeChat support.
	OutTraceNo string `json:"out_trace_no,omitempty"`
	// ErrTime of the failing request.
	ErrTime int64 `json:"err_time,omitempty"`
}

// GetRidInfo resolves the details behind an API error carrying a rid value.
// Use it to understand why a request failed (e.g. quota or permission
// problems) when debugging.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_getridinfo.html
func (w *MiniProgram) GetRidInfo(ctx context.Context, rid string) (*GetRidInfoResponse, error) {
	query := url.Values{}
	query.Set("rid", rid)
	var result GetRidInfoResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/openapi/rid/get", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// APILogDetail is the detail payload (json field) attached to some error
// responses.
type APILogDetail struct {
	// Keys of the log payload.
	Keys []string `json:"keys,omitempty"`
	// Values of the log payload.
	Values []any `json:"values,omitempty"`
	// Raw preserves the whole json payload.
	Raw json.RawMessage `json:"-"`
}
