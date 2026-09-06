package official

import (
	"context"
	"net/http"

	"github.com/go-sphere/weixin-api/core"
)

// ============================================================
// 基础支持 (basic support) — callback IP, API domain IP.
// ============================================================

// GetCallbackIPResponse is returned by GetCallbackIP.
type GetCallbackIPResponse struct {
	ErrResponse
	// IPList of the WeChat callback egress addresses.
	IPList []string `json:"ip_list"`
}

// GetCallbackIP lists the IP ranges WeChat uses when pushing message callbacks
// so the server can verify inbound requests.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/IP_List_interface.html
func (oa *OfficialAccount) GetCallbackIP(ctx context.Context) (*GetCallbackIPResponse, error) {
	var result GetCallbackIPResponse
	if err := oa.withToken(ctx, http.MethodGet, "/cgi-bin/getcallbackip", nil, core.DefaultRequestOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// APIDomainIPItem is one outbound IP of WeChat's API egress.
type APIDomainIPItem struct {
	// IP of WeChat's API egress.
	IP string `json:"ip"`
	// IPType of the address.
	IPType int `json:"ip_type,omitempty"`
	// IsSubDomain, when 1, marks a shared egress subset.
	IsSubDomain int `json:"is_sub_domain,omitempty"`
	// Domain of the egress.
	Domain string `json:"domain,omitempty"`
}

// GetAPIDomainIPResponse is returned by GetAPIDomainIP.
type GetAPIDomainIPResponse struct {
	ErrResponse
	// IPList of the API egress addresses.
	IPList []APIDomainIPItem `json:"ip_list"`
}

// GetAPIDomainIP lists the IP ranges WeChat API servers use outbound so a
// firewall can allow-list them.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/IP_List_interface.html
func (oa *OfficialAccount) GetAPIDomainIP(ctx context.Context) (*GetAPIDomainIPResponse, error) {
	var result GetAPIDomainIPResponse
	if err := oa.withToken(ctx, http.MethodGet, "/cgi-bin/get_api_domain_ip", nil, core.DefaultRequestOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ClearQuota clears the whole account's daily API call quota via the legacy
// /cgi-bin/clear_quota route (the request body carries the account appid).
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/API_Call_Limits.html
func (oa *OfficialAccount) ClearQuota(ctx context.Context) error {
	body := map[string]string{"appid": oa.config.AppID}
	return oa.withTokenPost(ctx, "/cgi-bin/clear_quota", nil, core.DefaultRequestOptions(), body, nil)
}

// QuotaDetailItem is one row of the API call quota detail.
type QuotaDetailItem struct {
	// Quota of the cgi path for the day.
	Quota int `json:"quota"`
	// Used of the cgi path for the day.
	Used int `json:"used"`
	// Surplus of the cgi path for the day.
	Surplus int `json:"surplus"`
	// ComponentQuota of the cgi path for the day (component accounts).
	ComponentQuota int `json:"component_quota,omitempty"`
}

// GetAPICallQuotaResponse is returned by GetAPICallQuota.
type GetAPICallQuotaResponse struct {
	ErrResponse
	// Quota of the queried cgi path for the day.
	Quota int `json:"quota"`
	// Used of the queried cgi path for the day.
	Used int `json:"used"`
	// Surplus of the queried cgi path for the day.
	Surplus int `json:"surplus"`
	// ComponentQuota of the queried cgi path (component accounts).
	ComponentQuota int `json:"component_quota"`
	// DetailList carries per-interface rows for the whole-account quota query.
	DetailList []QuotaDetailItem `json:"detail_list,omitempty"`
}

// GetAPICallQuota returns the quota usage of a cgi path.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/API_Call_Limits.html
func (oa *OfficialAccount) GetAPICallQuota(ctx context.Context, cgiPath string) (*GetAPICallQuotaResponse, error) {
	body := map[string]string{"cgi_path": cgiPath}
	var result GetAPICallQuotaResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/openapi/quota/get", nil, core.DefaultRequestOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// 消息加密 (message push AES crypto) — re-exported from core.
// ============================================================

// MessageCrypto handles the message-callback signature verification and
// AES-CBC body (de)encryption.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Message_encryption_and_signature_protocol.html
type MessageCrypto = core.MessageCrypto

// NewMessageCrypto builds a MessageCrypto for the official account callback.
func NewMessageCrypto(token, encodingAESKey, appID string) (*MessageCrypto, error) {
	return core.NewMessageCrypto(token, encodingAESKey, appID)
}
