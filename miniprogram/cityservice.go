package miniprogram

import (
	"context"
)

// ============================================================
// 城市服务 (city service) — basic + medical flows. These capabilities are
// enabled per business application and are mostly used by 城市服务/政务 Mini
// Programs.
// ============================================================

// CheckRealNameInfoRequest validates a user's real-name identity.
type CheckRealNameInfoRequest struct {
	// OpenID of the user under the business Mini Program.
	OpenID string `json:"openid"`
	// RealName to check.
	RealName string `json:"real_name"`
	// CredID of the identity document (currently only ID cards).
	CredID string `json:"cred_id"`
	// CredType, default 1 (ID card).
	CredType string `json:"cred_type"`
	// Code obtained from the Mini Program redirect.
	Code string `json:"code"`
}

// CheckRealNameInfoResponse is returned by CheckRealNameInfo.
type CheckRealNameInfoResponse struct {
	ErrResponse
	// RawData of the verification result.
	RawData map[string]any `json:"-"`
}

// CheckRealNameInfo checks whether a user's name and credential match
// (城市服务实名校验).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/basic/api_checkrealnameinfo.html
func (w *MiniProgram) CheckRealNameInfo(ctx context.Context, req *CheckRealNameInfoRequest) (*CheckRealNameInfoResponse, error) {
	var result CheckRealNameInfoResponse
	if err := w.withAccessTokenPost(ctx, "/intp/realname/checkrealnameinfo", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CityServiceMessageData is the template data payload of a city-service
// message.
type CityServiceMessageData struct {
	// Fields of the message template.
	Fields map[string]any `json:"-"`
}

// SendCityServiceMessageRequest delivers a city-service (办事) message to a
// user.
type SendCityServiceMessageRequest struct {
	// OpenID of the user.
	OpenID string `json:"openid"`
	// BizTemplateID assigned to the business by city service.
	BizTemplateID string `json:"biz_template_id"`
	// ResultPageStyleID when the message contains a result page.
	ResultPageStyleID string `json:"result_page_style_id,omitempty"`
	// DealMsgStyleID when the message contains a handling record.
	DealMsgStyleID string `json:"deal_msg_style_id,omitempty"`
	// CardStyleID when the message contains a page card.
	CardStyleID string `json:"card_style_id,omitempty"`
	// OrderNo merges the handling records of the same order.
	OrderNo string `json:"order_no"`
	// URL to jump to.
	URL string `json:"url,omitempty"`
	// Data JSON of the template values.
	Data map[string]any `json:"data"`
}

// SendCityServiceMessage sends a 城市服务 message (service notification /
// handling record).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/basic/api_cityservice_sendmsgdata.html
func (w *MiniProgram) SendCityServiceMessage(ctx context.Context, req *SendCityServiceMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cityservice/sendmsgdata", nil, defaultReqOptions(), req, nil)
}

// GetCityServicePathRequest selects the city-service page path to obtain.
type GetCityServicePathRequest struct {
	// PageType of the page: 0 service detail, 1 city home, 3 topic, 5 search.
	PageType int `json:"page_type"`
	// SrcChannel of the jump source.
	SrcChannel int `json:"src_channel"`
	// NeedPathType set to 1 when an H5 URL is needed.
	NeedPathType int `json:"need_path_type,omitempty"`
	// DeviceType set to 2 when an H5 URL is needed.
	DeviceType int `json:"device_type,omitempty"`
	// CityName for page types 1/3/5.
	CityName string `json:"city_name,omitempty"`
	// ContentName for page type 3.
	ContentName string `json:"content_name,omitempty"`
	// ExtParams for page type 5.
	ExtParams []map[string]any `json:"ext_params,omitempty"`
	// ServiceID for page type 0.
	ServiceID int `json:"service_id,omitempty"`
	// Params JSON-encoded pass-through params for page type 0.
	Params string `json:"params,omitempty"`
	// CityID of the user's city for page type 0.
	CityID string `json:"city_id,omitempty"`
}

// GetCityServicePathResponse is returned by GetCityServicePath.
type GetCityServicePathResponse struct {
	ErrResponse
	// RawData of the path payload.
	RawData map[string]any `json:"-"`
}

// GetCityServicePath returns the service-home page path / H5 URL of a city
// service.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/basic/api_cityserviceservicehomepath.html
func (w *MiniProgram) GetCityServicePath(ctx context.Context, req *GetCityServicePathRequest) (*GetCityServicePathResponse, error) {
	var result GetCityServicePathResponse
	if err := w.withAccessTokenPost(ctx, "/cityservice/getservicepath", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TransportCodeBusinessViewRequest requests the business page of the transport
// code.
type TransportCodeBusinessViewRequest struct {
	// PathType of the target page.
	PathType int `json:"path_type"`
}

// TransportCodeBusinessViewResponse is returned by
// GetTransportCodeBusinessView.
type TransportCodeBusinessViewResponse struct {
	ErrResponse
	// RawData of the view payload.
	RawData map[string]any `json:"-"`
}

// GetTransportCodeBusinessView returns the business-view page of the transport
// code (乘车码).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/basic/api_transportcode_getbusinessview.html
func (w *MiniProgram) GetTransportCodeBusinessView(ctx context.Context, pathType int) (*TransportCodeBusinessViewResponse, error) {
	req := &TransportCodeBusinessViewRequest{PathType: pathType}
	var result TransportCodeBusinessViewResponse
	if err := w.withAccessTokenPost(ctx, "/intp/transportcode/getbusinessview", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMedicalRealNameRequest fetches the medical real-name of a user via the
// wxmed_authcode.
type GetMedicalRealNameRequest struct {
	// AppID of the business official account.
	AppID string `json:"app_id"`
	// OpenID of the user.
	OpenID string `json:"open_id"`
	// WxmedAuthcode from the URL (valid 10 minutes).
	WxmedAuthcode string `json:"wxmed_authcode"`
}

// GetMedicalRealNameResponse is returned by GetMedicalRealName.
type GetMedicalRealNameResponse struct {
	ErrResponse
	// RawData of the real-name payload.
	RawData map[string]any `json:"-"`
}

// GetMedicalRealName returns the 医疗实名 (medical real-name) information of a
// user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/elderMedical/api_cityservice_getmedrealname.html
func (w *MiniProgram) GetMedicalRealName(ctx context.Context, req *GetMedicalRealNameRequest) (*GetMedicalRealNameResponse, error) {
	var result GetMedicalRealNameResponse
	if err := w.withAccessTokenPost(ctx, "/cityservice/getmedrealname", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCityMessageRelationRequest queries a user's message relation for a
// business.
type GetCityMessageRelationRequest struct {
	// BusinessID of the business (e.g. 130).
	BusinessID string `json:"business_id"`
	// OpenID of the user.
	OpenID string `json:"open_id"`
}

// GetCityMessageRelationResponse is returned by GetCityMessageRelation.
type GetCityMessageRelationResponse struct {
	ErrResponse
	// RawData of the relation payload.
	RawData map[string]any `json:"-"`
}

// GetCityMessageRelation returns the message relation of a user under a city
// service business.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/elderMedical/api_cityservice_getmsgrelation.html
func (w *MiniProgram) GetCityMessageRelation(ctx context.Context, req *GetCityMessageRelationRequest) (*GetCityMessageRelationResponse, error) {
	var result GetCityMessageRelationResponse
	if err := w.withAccessTokenPost(ctx, "/cityservice/getmsgrelation", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MedicalChannelMessageRequest sends a channel (渠道) message for the medical
// assistant flow.
type MedicalChannelMessageRequest struct {
	// Status of the message (e.g. 1501001).
	Status int `json:"status"`
	// OpenID of the user.
	OpenID string `json:"open_id"`
	// OrderID of the medical order.
	OrderID string `json:"order_id"`
	// MsgID of the message.
	MsgID string `json:"msg_id"`
	// AppID of the business.
	AppID string `json:"app_id"`
	// BusinessID of the business (e.g. 150).
	BusinessID int `json:"business_id"`
	// BusinessInfo extra payload.
	BusinessInfo map[string]any `json:"business_info,omitempty"`
}

// SendMedicalChannelMessage delivers a medical-assistant channel message.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/medicalassistant/api_cityservice_sendchannelmsg.html
func (w *MiniProgram) SendMedicalChannelMessage(ctx context.Context, req *MedicalChannelMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cityservice/sendchannelmsg", nil, defaultReqOptions(), req, nil)
}

// MedElderNoticeRequest is the shared payload of the elder-medical notice
// APIs.
type MedElderNoticeRequest struct {
	// AppID of the business official account.
	AppID string `json:"app_id"`
	// NoticeType of the notice (公告类型).
	NoticeType int `json:"notice_type"`
}

// GetHospitalNoticeListRequest selects the hospital notices to list.
type GetHospitalNoticeListRequest struct {
	MedElderNoticeRequest
}

// GetHospitalNoticeListResponse is returned by GetHospitalNoticeList.
type GetHospitalNoticeListResponse struct {
	ErrResponse
	// RawData of the notice list.
	RawData map[string]any `json:"-"`
}

// GetHospitalNoticeList returns the hospital notices of a notice type.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/elderMedical/api_intp_eldermed_gethospnoticelist.html
func (w *MiniProgram) GetHospitalNoticeList(ctx context.Context, req *GetHospitalNoticeListRequest) (*GetHospitalNoticeListResponse, error) {
	var result GetHospitalNoticeListResponse
	if err := w.withAccessTokenPost(ctx, "/intp/eldermedical/gethospnoticelist", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PublishHospitalNoticeRequest publishes a hospital notice draft.
type PublishHospitalNoticeRequest struct {
	MedElderNoticeRequest
	// NoticeID of the draft to publish.
	NoticeID int `json:"notice_id"`
}

// PublishHospitalNotice publishes a hospital notice (发布公告).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/elderMedical/api_intp_eldermed_publichopsnotice.html
func (w *MiniProgram) PublishHospitalNotice(ctx context.Context, req *PublishHospitalNoticeRequest) error {
	return w.withAccessTokenPost(ctx, "/intp/eldermedical/publichopsnotice", nil, defaultReqOptions(), req, nil)
}

// PreviewHospitalNoticeRequest previews a hospital notice to a WeChat user.
type PreviewHospitalNoticeRequest struct {
	MedElderNoticeRequest
	// NoticeID of the notice to preview.
	NoticeID string `json:"notice_id"`
	// PreviewUsername of the WeChat id allowed to preview.
	PreviewUsername string `json:"preview_username"`
	// Operation of the preview.
	Operation int `json:"operation"`
}

// PreviewHospitalNotice sends a notice preview to a WeChat user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/elderMedical/api_previewhopsnotice.html
func (w *MiniProgram) PreviewHospitalNotice(ctx context.Context, req *PreviewHospitalNoticeRequest) error {
	return w.withAccessTokenPost(ctx, "/intp/eldermedical/previewhopsnotice", nil, defaultReqOptions(), req, nil)
}

// SetHospitalNoticeRequest creates or updates a hospital notice draft.
type SetHospitalNoticeRequest struct {
	MedElderNoticeRequest
	// NoticeContent of the notice (up to 3000 chars, rich text supported).
	NoticeContent string `json:"notice_content"`
	// NoticeID to update an existing draft (omitted creates a new draft).
	NoticeID int `json:"notice_id,omitempty"`
}

// SetHospitalNotice creates or overwrites a hospital notice draft.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cityservice/elderMedical/api_sethopsnotice.html
func (w *MiniProgram) SetHospitalNotice(ctx context.Context, req *SetHospitalNoticeRequest) error {
	return w.withAccessTokenPost(ctx, "/intp/eldermedical/sethopsnotice", nil, defaultReqOptions(), req, nil)
}
