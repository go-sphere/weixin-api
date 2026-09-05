package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// ============================================================
// CloudBase "others": SMS marketing, statistics, VoIP sign and open data.
// ============================================================

// TCBSendSMSRequest sends SMS messages to users (cloud SMS).
type TCBSendSMSRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// PhoneNumberList (up to 1000, each prefixed with +86).
	PhoneNumberList []string `json:"phone_number_list"`
	// SMSType: "Marketing" or "Notification".
	SMSType string `json:"sms_type"`
	// TemplateID required when SMSType is Notification.
	TemplateID string `json:"template_id,omitempty"`
	// Content for marketing SMS (up to 70 chars, customisable 30).
	Content string `json:"content,omitempty"`
	// Path of the cloud static site page for marketing SMS.
	Path string `json:"path,omitempty"`
	// TemplateParamList of the notification template variables.
	TemplateParamList []string `json:"template_param_list,omitempty"`
	// UseShortName uses the Mini Program short name.
	UseShortName bool `json:"use_short_name"`
	// ResourceAppID of the resource owner (for third-party development).
	ResourceAppID string `json:"resource_appid,omitempty"`
}

// TCBSendSMS delivers marketing or notification SMS messages.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_sendcloudbasesms.html
func (w *MiniProgram) TCBSendSMS(ctx context.Context, req *TCBSendSMSRequest) error {
	return w.withAccessTokenPost(ctx, "/tcb/sendsms", nil, defaultReqOptions(), req, nil)
}

// TCBSendSMSV2Request sends cloud SMS with an URL link.
type TCBSendSMSV2Request struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// URLLink attached to the SMS.
	URLLink string `json:"url_link"`
	// TemplateID of the SMS template (e.g. 844110 marketing).
	TemplateID string `json:"template_id"`
	// TemplateParamList of the template variables.
	TemplateParamList []string `json:"template_param_list"`
	// PhoneNumberList (up to 1000, prefixed with +86).
	PhoneNumberList []string `json:"phone_number_list"`
	// UseShortName uses the Mini Program short name.
	UseShortName bool `json:"use_short_name"`
	// ResourceAppID of the resource owner.
	ResourceAppID string `json:"resource_appid,omitempty"`
}

// TCBSendSMSV2 sends cloud SMS through the URL-link based endpoint.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_newsendcloudbasesms.html
func (w *MiniProgram) TCBSendSMSV2(ctx context.Context, req *TCBSendSMSV2Request) error {
	return w.withAccessTokenPost(ctx, "/tcb/sendsmsv2", nil, defaultReqOptions(), req, nil)
}

// TCBCreateSendSMSTaskRequest starts a bulk SMS task from a CSV file.
type TCBCreateSendSMSTaskRequest struct {
	// Env of the cloud environment.
	Env string `json:"env"`
	// FileURL of the SMS CSV file.
	FileURL string `json:"file_url"`
	// TemplateID of the SMS template.
	TemplateID string `json:"template_id"`
}

// TCBCreateSendSMSTask launches a batch SMS-send task.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_createsendsmstask.html
func (w *MiniProgram) TCBCreateSendSMSTask(ctx context.Context, req *TCBCreateSendSMSTaskRequest) error {
	return w.withAccessTokenPost(ctx, "/tcb/createsendsmstask", nil, defaultReqOptions(), req, nil)
}

// TCBDescribeSMSRecordsRequest filters the SMS send records.
type TCBDescribeSMSRecordsRequest struct {
	// EnvID of the cloud environment.
	EnvID string `json:"EnvId"`
	// StartDate of the window ("2021-01-01").
	StartDate string `json:"StartDate"`
	// EndDate of the window.
	EndDate string `json:"EndDate"`
	// Mobile phone number filter.
	Mobile string `json:"Mobile"`
	// QueryID filter.
	QueryID string `json:"QueryId"`
	// PageNumber (1-based).
	PageNumber int `json:"PageNumber"`
	// PageSize of each page.
	PageSize int `json:"PageSize"`
}

// TCBDescribeSMSRecordsResponse is returned by TCBDescribeSMSRecords.
type TCBDescribeSMSRecordsResponse struct {
	ErrResponse
	// RawData of the records payload.
	RawData map[string]any `json:"-"`
}

// TCBDescribeSMSRecords lists the cloud SMS send records.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_describesmsrecords.html
func (w *MiniProgram) TCBDescribeSMSRecords(ctx context.Context, req *TCBDescribeSMSRecordsRequest) (*TCBDescribeSMSRecordsResponse, error) {
	var result TCBDescribeSMSRecordsResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/describesmsrecords", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBCloudBaseReportRequest reports a cloud marketing activity event.
type TCBCloudBaseReportRequest struct {
	// ReportAction: "sendSmsTask" or "openH5".
	ReportAction string `json:"report_action"`
	// EnvID of the cloud environment.
	EnvID string `json:"env_id"`
	// ActivityID of the marketing activity.
	ActivityID string `json:"activity_id"`
	// TaskID required when ReportAction is sendSmsTask.
	TaskID string `json:"task_id,omitempty"`
	// PhoneCount required when ReportAction is sendSmsTask.
	PhoneCount string `json:"phone_count,omitempty"`
	// ChannelID required when ReportAction is openH5.
	ChannelID string `json:"channel_id,omitempty"`
	// SessionID required when ReportAction is openH5.
	SessionID string `json:"session_id,omitempty"`
}

// TCBCloudBaseReport reports cloud SMS/H5 activity events for analytics.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_cloudbasereportapi.html
func (w *MiniProgram) TCBCloudBaseReport(ctx context.Context, req *TCBCloudBaseReportRequest) error {
	return w.withAccessTokenPost(ctx, "/tcb/cloudbasereport", nil, defaultReqOptions(), req, nil)
}

// TCBGetStatisticsRequest selects the cloud statistics report.
type TCBGetStatisticsRequest struct {
	// Action of the report: smsMarketingOverviewData,
	// smsMarketingConversionData or smsMarketingRealTimeData.
	Action string `json:"action"`
	// BeginDate unix timestamp.
	BeginDate int64 `json:"begin_date"`
	// EndDate unix timestamp.
	EndDate int64 `json:"end_date"`
	// PageLimit for the overview/conversion actions.
	PageLimit int `json:"page_limit,omitempty"`
	// PageOffset for the overview/conversion actions.
	PageOffset int `json:"page_offset,omitempty"`
}

// TCBGetStatisticsResponse is returned by TCBGetStatistics.
type TCBGetStatisticsResponse struct {
	ErrResponse
	// RawData of the statistics payload.
	RawData map[string]any `json:"-"`
}

// TCBGetStatistics returns the cloud SMS marketing statistics.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_getcloudbasestatistics.html
func (w *MiniProgram) TCBGetStatistics(ctx context.Context, req *TCBGetStatisticsRequest) (*TCBGetStatisticsResponse, error) {
	var result TCBGetStatisticsResponse
	if err := w.withAccessTokenPost(ctx, "/tcb/getstatistics", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBGetQCloudTokenRequest requests a QCloud temporary token.
type TCBGetQCloudTokenRequest struct {
	// Lifespan of the token in seconds (max 7200).
	Lifespan int `json:"lifespan"`
}

// TCBGetQCloudTokenResponse is returned by TCBGetQCloudToken.
type TCBGetQCloudTokenResponse struct {
	ErrResponse
	// RawData of the token payload.
	RawData map[string]any `json:"-"`
}

// TCBGetQCloudToken returns a temporary QCloud access token for the cloud
// environment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_getcloudtoken.html
func (w *MiniProgram) TCBGetQCloudToken(ctx context.Context, lifespan int) (*TCBGetQCloudTokenResponse, error) {
	var result TCBGetQCloudTokenResponse
	req := &TCBGetQCloudTokenRequest{Lifespan: lifespan}
	if err := w.withAccessTokenPost(ctx, "/tcb/getqcloudtoken", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBGetVoIPSignRequest asks for a real-time-voice (语音) room signature.
type TCBGetVoIPSignRequest struct {
	// GroupID of the game/voice room.
	GroupID string `json:"group_id"`
	// Timestamp of the signature (unix seconds).
	Timestamp int64 `json:"timestamp"`
	// Nonce random string (up to 128 chars).
	Nonce string `json:"nonce"`
}

// TCBGetVoIPSignResponse is returned by TCBGetVoIPSign.
type TCBGetVoIPSignResponse struct {
	ErrResponse
	// RawData of the sign payload.
	RawData map[string]any `json:"-"`
}

// TCBGetVoIPSign returns the signature used to join a cloud voice room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_getcloudbasevoipsign.html
func (w *MiniProgram) TCBGetVoIPSign(ctx context.Context, req *TCBGetVoIPSignRequest) (*TCBGetVoIPSignResponse, error) {
	var result TCBGetVoIPSignResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/getvoipsign", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBGetOpenDataRequest resolves cloudids (open data references) to plain
// data for an openid.
type TCBGetOpenDataRequest struct {
	// CloudIDList of the data references to resolve.
	CloudIDList []string `json:"cloudid_list"`
}

// TCBGetOpenDataResponse is returned by TCBGetOpenData.
type TCBGetOpenDataResponse struct {
	ErrResponse
	// RawData of the resolved open data.
	RawData map[string]any `json:"-"`
}

// TCBGetOpenData resolves cloud-open-data references for a user (the openid is
// passed as a query parameter).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_getopendata.html
func (w *MiniProgram) TCBGetOpenData(ctx context.Context, openid string, cloudIDs []string) (*TCBGetOpenDataResponse, error) {
	query := url.Values{}
	query.Set("openid", openid)
	body := &TCBGetOpenDataRequest{CloudIDList: cloudIDs}
	var result TCBGetOpenDataResponse
	if err := w.withAccessToken(ctx, http.MethodPost, "/wxa/getopendata", query, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
