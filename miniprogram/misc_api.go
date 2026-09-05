package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// ============================================================
// 品牌 (famous brand) application for the trade-guarantee order shipping.
// ============================================================

// FamousBrandApplyAuditInfo carries the brand materials of an application.
type FamousBrandApplyAuditInfo struct {
	// BrandName of the brand.
	BrandName string `json:"brand_name"`
	// BrandType of the brand.
	BrandType int `json:"brand_type"`
	// FlagshipInWhichECPlatform names the e-commerce platform flagship store.
	FlagshipInWhichECPlatform string `json:"flagship_in_which_ec_platform,omitempty"`
	// ECPlatformProofMediaIDs of the flagship-store proof media.
	ECPlatformProofMediaIDs []string `json:"ec_platform_proof_list,omitempty"`
	// OtherMaterialMediaIDs of supplementary materials.
	OtherMaterialMediaIDs []string `json:"other_material_list,omitempty"`
	// AuthorityCertifiedProofMediaIDs of the authority-certified famous-trademark
	// materials.
	AuthorityCertifiedProofMediaIDs []string `json:"authority_certified_proof_list,omitempty"`
}

// FamousBrandApplication is the body of a famous-brand application.
type FamousBrandApplication struct {
	// ApplyForType of the application.
	ApplyForType int `json:"apply_for"`
	// AuditInfo of the brand.
	AuditInfo *FamousBrandApplyAuditInfo `json:"audit_info,omitempty"`
}

// ApplyFamousBrandRequest is the payload of ApplyFamousBrand.
type ApplyFamousBrandRequest struct {
	// Application of the famous-brand claim.
	Application FamousBrandApplication `json:"Application"`
}

// ApplyFamousBrand submits a famous-brand application (品牌申请) required by
// certain order-shipping trade types.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_famousbrandapply.html
func (w *MiniProgram) ApplyFamousBrand(ctx context.Context, req *ApplyFamousBrandRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/famousbrand/apply", nil, defaultReqOptions(), req, nil)
}

// FamousBrandAuditStatus is the audit state of a famous-brand application.
type FamousBrandAuditStatus struct {
	// Progress of the audit: 0 none, 1 under review, 2 passed, 3 rejected.
	Progress int `json:"progress,omitempty"`
	// Application is the application payload echoed back.
	Application map[string]any `json:"application,omitempty"`
}

// GetFamousBrandApplyStatus returns the audit status of the famous-brand
// application.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_getfamousbrandapplystatus.html
func (w *MiniProgram) GetFamousBrandApplyStatus(ctx context.Context) (*FamousBrandAuditStatus, error) {
	var result FamousBrandAuditStatus
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/famousbrand/get_status", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpecialOrderReport is the payload of ReportSpecialOrder.
type SpecialOrderReport struct {
	// OrderID of the order needing special reporting; a WeChat Pay transaction
	// id or a merchant order number.
	OrderID string `json:"order_id"`
	// Type of the special report: 1 = pre-sale (预售) order, 2 = test order.
	Type int `json:"type"`
	// DelayTo is the expected shipment time (unix seconds); required when
	// Type is 1.
	DelayTo int64 `json:"delay_to,omitempty"`
}

// ReportSpecialOrder reports an order that ships later than the requirement
// (特殊发货报备), so the trade-guarantee shipping deadline is adjusted.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_opspecialorder.html
func (w *MiniProgram) ReportSpecialOrder(ctx context.Context, req *SpecialOrderReport) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/order/opspecialorder", nil, defaultReqOptions(), req, nil)
}

// ============================================================
// 手机号快速验证 charging usage (云调用额度用量查询), from the "charge" group.
// ============================================================

// UsageRecord is one usage record row.
type UsageRecord struct {
	// Date of the usage in "yyyyMMdd".
	Date string `json:"date,omitempty"`
	// Count of the calls.
	Count int64 `json:"count,omitempty"`
}

// GetChargeUsageDetailResponse is returned by GetChargeUsageDetail.
type GetChargeUsageDetailResponse struct {
	ErrResponse
	// EffectiveUse cumulative usage for resource-pack style products.
	EffectiveUse string `json:"effectiveUse,omitempty"`
	// RawData of the usage detail payload.
	RawData map[string]any `json:"-"`
}

// GetChargeUsageDetailRequest pages the usage detail of a paid cloud-call
// product.
type GetChargeUsageDetailRequest struct {
	// SPUId of the purchased product.
	SPUId string `json:"spuId"`
	// Offset of the page (from 0).
	Offset int `json:"offset"`
	// Limit of the page (max 20).
	Limit int `json:"limit"`
}

// GetChargeUsageDetail returns the detailed usage of the paid API quota
// product.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/charge/api_getusagedetail.html
func (w *MiniProgram) GetChargeUsageDetail(ctx context.Context, req *GetChargeUsageDetailRequest) (*GetChargeUsageDetailResponse, error) {
	query := url.Values{}
	query.Set("spuId", req.SPUId)
	query.Set("offset", fmtInt(req.Offset))
	query.Set("limit", fmtInt(req.Limit))
	var result GetChargeUsageDetailResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxa/charge/usage/get", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRecentAverageUsageRequest selects the product for the recent-average
// query.
type GetRecentAverageUsageRequest struct {
	// SPUId of the purchased product.
	SPUId string `json:"spuId"`
}

// GetRecentAverageUsageResponse is returned by GetRecentAverageUsage.
type GetRecentAverageUsageResponse struct {
	ErrResponse
	// RawData of the recent average usage payload.
	RawData map[string]any `json:"-"`
}

// GetRecentAverageUsage returns the recent average usage of the paid API quota
// product.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/charge/api_getrecentaverageusage.html
func (w *MiniProgram) GetRecentAverageUsage(ctx context.Context, req *GetRecentAverageUsageRequest) (*GetRecentAverageUsageResponse, error) {
	query := url.Values{}
	query.Set("spuId", req.SPUId)
	var result GetRecentAverageUsageResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxa/charge/usage/get_recent_average", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// 微信红包封面 (red packet cover).
// ============================================================

// GetRedPacketCoverURLResponse is returned by GetRedPacketCoverURL.
type GetRedPacketCoverURLResponse struct {
	ErrResponse
	// URL of the red packet cover.
	URL string `json:"url,omitempty"`
	// RawData of the response.
	RawData map[string]any `json:"-"`
}

// GetRedPacketCoverURLRequest selects the red packet cover for a user.
type GetRedPacketCoverURLRequest struct {
	// OpenID of the user who may claim the cover.
	OpenID string `json:"openid"`
	// CToken obtained from the red packet cover platform (must allow the
	// current appid).
	CToken string `json:"ctoken"`
}

// GetRedPacketCoverURL returns the red packet cover (红包封面) URL for a user
// so the Mini Program can launch the red packet flow.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/red-packet-cover/api_getredpacketcoverurl.html
func (w *MiniProgram) GetRedPacketCoverURL(ctx context.Context, req *GetRedPacketCoverURLRequest) (*GetRedPacketCoverURLResponse, error) {
	var result GetRedPacketCoverURLResponse
	if err := w.withAccessTokenPost(ctx, "/redpacketcover/wxapp/cover_url/get_by_token", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// 用工 (labor use / employer relation) messages.
// ============================================================

// SendEmployeeRelationMessageRequest sends an employer-employee relation
// message (用工企业向务工人员发消息) based on a message template.
type SendEmployeeRelationMessageRequest struct {
	// TemplateID of the approved message template.
	TemplateID string `json:"template_id"`
	// Page to open when the message is tapped.
	Page string `json:"page"`
	// ToUser is the openid of the receiving worker.
	ToUser string `json:"touser"`
	// Data is a JSON-encoded string of the template values, e.g.
	// {"data":{"character_string1":{"value":"aaa"}}}.
	Data string `json:"data"`
}

// SendEmployeeRelationMessage sends a labor-relation message from an employer
// Mini Program to its registered workers.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/laboruse/api_sendemployeerelationmsg.html
func (w *MiniProgram) SendEmployeeRelationMessage(ctx context.Context, req *SendEmployeeRelationMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/wxopen/employeerelationmsg/send", nil, defaultReqOptions(), req, nil)
}

// UnbindUserB2CAuthInfoRequest unbinds the user authentication between a B2C
// (business-to-consumer) Mini Program pair.
type UnbindUserB2CAuthInfoRequest struct {
	// OpenID of the user under the source Mini Program.
	OpenID string `json:"openid"`
	// AuthAppID of the target Mini Program to unbind from.
	AuthAppID string `json:"auth_appid,omitempty"`
}

// UnbindUserB2CAuthInfo unbinds the cross-Mini-Program user authorisation
// (B2C 用户授权).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/laboruse/api_unbinduserb2cauthinfo.html
func (w *MiniProgram) UnbindUserB2CAuthInfo(ctx context.Context, req *UnbindUserB2CAuthInfoRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/unbinduserb2cauthinfo", nil, defaultReqOptions(), req, nil)
}
