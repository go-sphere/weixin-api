package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// UserFeedbackItem is one entry of the user feedback list (feedback submitted
// through the Mini Program "help & feedback" console).
type UserFeedbackItem struct {
	// RecordID of the feedback entry.
	RecordID int64 `json:"record_id"`
	// CreateTime is the creation time (Unix seconds).
	CreateTime int64 `json:"create_time"`
	// Content of the feedback.
	Content string `json:"content"`
	// Phone filled by the user when provided.
	Phone string `json:"phone"`
	// OpenID of the reporting user.
	OpenID string `json:"openid"`
	// Nickname of the reporting user.
	Nickname string `json:"nickname"`
	// HeadURL of the reporting user avatar.
	HeadURL string `json:"head_url"`
	// Type of the feedback.
	Type int `json:"type"`
	// MediaIDs of the attached images.
	MediaIDs []string `json:"mediaIds"`
	// SystemInfo as reported by the client.
	SystemInfo string `json:"systemInfo"`
}

// GetFeedbackResponse is returned by GetFeedback.
type GetFeedbackResponse struct {
	ErrResponse
	// TotalNum is the total number of matching feedback entries.
	TotalNum int64 `json:"total_num"`
	// List is the requested page of feedback entries.
	List []UserFeedbackItem `json:"list"`
}

// GetFeedback returns the user feedback list, paged.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getfeedback.html
func (w *MiniProgram) GetFeedback(ctx context.Context, page, count, feedbackType int) (*GetFeedbackResponse, error) {
	query := url.Values{}
	query.Set("page", fmtInt(page))
	query.Set("num", fmtInt(count))
	if feedbackType > 0 {
		query.Set("type", fmtInt(feedbackType))
	}
	var result GetFeedbackResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/feedback/list", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFeedbackMedia returns the media (images) attached to a user feedback
// entry. On success the endpoint responds with the raw image bytes; record ids
// and media ids come from GetFeedback.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getfeedbackmedia.html
func (w *MiniProgram) GetFeedbackMedia(ctx context.Context, recordID int64, mediaID string) ([]byte, error) {
	query := url.Values{}
	query.Set("record_id", fmtInt(int(recordID)))
	query.Set("media_id", mediaID)
	return w.withAccessTokenRaw(ctx, http.MethodGet, "/cgi-bin/media/getfeedbackmedia", query, defaultReqOptions(), nil)
}

// JsErrQueryItem is one aggregated JS error signature of the error list.
type JsErrQueryItem struct {
	// ErrorMsgMd5 of the error message.
	ErrorMsgMd5 string `json:"errorMsgMd5"`
	// ErrorMsg is the error message text.
	ErrorMsg string `json:"errorMsg"`
	// UV is the number of affected users.
	UV int `json:"uv"`
	// PV is the number of affected page visits.
	PV int `json:"pv"`
	// ErrorStackMd5 of the error stack.
	ErrorStackMd5 string `json:"errorStackMd5"`
	// ErrorStack is the error stack text.
	ErrorStack string `json:"errorStack"`
	// PVPercent is the ratio of this error among all errors.
	PVPercent string `json:"pvPercent"`
	// UVPercent is the user ratio of this error.
	UVPercent string `json:"uvPercent"`
}

// GetJsErrListResponse is returned by GetJsErrList.
type GetJsErrListResponse struct {
	ErrResponse
	// TotalCount of the matching error signatures.
	TotalCount int64 `json:"totalCount"`
	// OpenID is echoed back when an openid filter was used.
	OpenID string `json:"openid"`
	// Data is the page of aggregated error signatures.
	Data []JsErrQueryItem `json:"data"`
}

// GetJsErrListRequest filters the JS error list query.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getjserrlist.html
type GetJsErrListRequest struct {
	// AppVersion of the Mini Program release to filter by.
	AppVersion string `json:"appVersion,omitempty"`
	// ErrType of the error, e.g. "1" for JS errors.
	ErrType string `json:"errType,omitempty"`
	// StartTime of the window in "YYYY-MM-DD HH:MM:SS" (required).
	StartTime string `json:"startTime"`
	// EndTime of the window in "YYYY-MM-DD HH:MM:SS" (required).
	EndTime string `json:"endTime"`
	// Keyword to search in the error message.
	Keyword string `json:"keyword,omitempty"`
	// OpenID filter when the errors of one user are needed.
	OpenID string `json:"openid,omitempty"`
	// OrderBy selects the sort field (e.g. "uv").
	OrderBy string `json:"orderby,omitempty"`
	// Desc reverses the ordering when true.
	Desc string `json:"desc,omitempty"`
	// Offset of the first returned row.
	Offset int `json:"offset,omitempty"`
	// Limit of the page size.
	Limit int `json:"limit,omitempty"`
}

// GetJsErrList lists the JS errors aggregated for the Mini Program within the
// requested window.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getjserrlist.html
func (w *MiniProgram) GetJsErrList(ctx context.Context, req *GetJsErrListRequest) (*GetJsErrListResponse, error) {
	var result GetJsErrListResponse
	if err := w.withAccessTokenPost(ctx, "/wxaapi/log/jserr_list", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// JsErrOccurrence is one JS error occurrence.
type JsErrOccurrence struct {
	// Count of occurrences aggregated in this row.
	Count string `json:"Count"`
	// SdkVersion of the reporting client.
	SdkVersion string `json:"sdkVersion"`
	// ClientVersion of the reporting client.
	ClientVersion string `json:"ClientVersion"`
	// ErrorStackMd5 of the error.
	ErrorStackMd5 string `json:"errorStackMd5"`
	// TimeStamp of the occurrence.
	TimeStamp string `json:"TimeStamp"`
	// AppVersion of the Mini Program release.
	AppVersion string `json:"appVersion"`
	// ErrorMsgMd5 of the error.
	ErrorMsgMd5 string `json:"errorMsgMd5"`
	// ErrorMsg text of the error.
	ErrorMsg string `json:"errorMsg"`
	// ErrorStack text of the error.
	ErrorStack string `json:"errorStack"`
	// Ds is the deployment scene of the error.
	Ds string `json:"Ds"`
	// OsName of the reporting device.
	OsName string `json:"OsName"`
	// OpenID of the affected user.
	OpenID string `json:"openId"`
	// PluginVersion of the affected plugin.
	PluginVersion string `json:"pluginversion"`
	// AppID of the Mini Program.
	AppID string `json:"appId"`
	// DeviceModel of the reporting device.
	DeviceModel string `json:"DeviceModel"`
	// Source of the error.
	Source string `json:"source"`
	// Route where the error happened.
	Route string `json:"route"`
	// Uin of the affected user.
	Uin string `json:"Uin"`
	// Nickname of the affected user.
	Nickname string `json:"nickname"`
}

// GetJsErrDetailResponse is returned by GetJsErrDetail.
type GetJsErrDetailResponse struct {
	ErrResponse
	// TotalCount of the matching occurrences.
	TotalCount int64 `json:"totalCount"`
	// OpenID is echoed when a user filter was used.
	OpenID string `json:"openid"`
	// Data is the page of occurrences.
	Data []JsErrOccurrence `json:"data"`
}

// GetJsErrDetailRequest filters the JS error detail query.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getjserrdetail.html
type GetJsErrDetailRequest struct {
	// StartTime of the window in "YYYY-MM-DD HH:MM:SS" (required).
	StartTime string `json:"startTime"`
	// EndTime of the window in "YYYY-MM-DD HH:MM:SS" (required).
	EndTime string `json:"endTime"`
	// ErrorMsgMd5 of the error signature to inspect.
	ErrorMsgMd5 string `json:"errorMsgMd5"`
	// ErrorStackMd5 of the error signature to inspect.
	ErrorStackMd5 string `json:"errorStackMd5"`
	// AppVersion of the Mini Program release.
	AppVersion string `json:"appVersion,omitempty"`
	// SdkVersion of the client SDK.
	SdkVersion string `json:"sdkVersion,omitempty"`
	// OsName of the device OS.
	OsName string `json:"osName,omitempty"`
	// ClientVersion of the WeChat client.
	ClientVersion string `json:"clientVersion,omitempty"`
	// OpenID filter.
	OpenID string `json:"openid,omitempty"`
	// Offset of the first returned row.
	Offset int `json:"offset,omitempty"`
	// Limit of the page size.
	Limit int `json:"limit,omitempty"`
	// Desc reverses ordering when true.
	Desc string `json:"desc,omitempty"`
}

// GetJsErrDetail lists the individual occurrences of one JS error signature
// (identified by its message/stack md5).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getjserrdetail.html
func (w *MiniProgram) GetJsErrDetail(ctx context.Context, req *GetJsErrDetailRequest) (*GetJsErrDetailResponse, error) {
	var result GetJsErrDetailResponse
	if err := w.withAccessTokenPost(ctx, "/wxaapi/log/jserr_detail", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RealtimeLogEntry is one reported realtime log line.
type RealtimeLogEntry struct {
	// Level of the log (2 info, 4 warn, 8 error).
	Level int `json:"level"`
	// LibraryVersion of the base library.
	LibraryVersion string `json:"libraryVersion"`
	// ClientVersion of the WeChat client.
	ClientVersion string `json:"clientVersion"`
	// ID of the log report.
	ID string `json:"id"`
	// Timestamp of the report.
	Timestamp int64 `json:"timestamp"`
	// Platform of the client.
	Platform int `json:"platform"`
	// URL where the log was recorded.
	URL string `json:"url"`
	// TraceID of the log entry.
	TraceID string `json:"traceid"`
	// FilterMsg set through wx.getRealtimeLogManager().addFilterMsg.
	FilterMsg string `json:"filterMsg"`
	// Msg is the list of logged messages with timestamps.
	Msg []RealtimeLogMsg `json:"msg"`
}

// RealtimeLogMsg is one message inside a realtime log entry.
type RealtimeLogMsg struct {
	// Time of the message (Unix seconds).
	Time int64 `json:"time"`
	// Level of the message.
	Level int `json:"level"`
	// Msg text parts of the logged message.
	Msg []string `json:"msg"`
}

// RealtimeLogSearchData is the payload of RealtimeLogSearch.
type RealtimeLogSearchData struct {
	// List of the matching log entries.
	List []RealtimeLogEntry `json:"list"`
	// Total number of matching entries.
	Total int `json:"total"`
}

// RealtimeLogSearchResponse is returned by RealtimeLogSearch.
type RealtimeLogSearchResponse struct {
	ErrResponse
	// Data of the search result.
	Data RealtimeLogSearchData `json:"data"`
}

// RealtimeLogSearchRequest filters the realtime log query.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_realtimelogsearch.html
type RealtimeLogSearchRequest struct {
	// Date in "YYYYMMDD" format to search (required).
	Date string
	// BeginTime of the day window as Unix seconds.
	BeginTime int64
	// EndTime of the day window as Unix seconds.
	EndTime int64
	// Start is the first row index of the page.
	Start int64
	// Limit is the page size.
	Limit int64
	// Level of the logs to include.
	Level int64
	// TraceID of a specific log entry.
	TraceID string
	// URL where the log was recorded.
	URL string
	// ID of a specific log report.
	ID string
	// FilterMsg content filter.
	FilterMsg string
}

// RealtimeLogSearch queries the realtime logs uploaded by the Mini Program
// client for diagnostics.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_realtimelogsearch.html
func (w *MiniProgram) RealtimeLogSearch(ctx context.Context, req *RealtimeLogSearchRequest) (*RealtimeLogSearchResponse, error) {
	query := url.Values{}
	query.Set("date", req.Date)
	if req.BeginTime > 0 {
		query.Set("begintime", fmtInt(int(req.BeginTime)))
	}
	if req.EndTime > 0 {
		query.Set("endtime", fmtInt(int(req.EndTime)))
	}
	if req.Start > 0 {
		query.Set("start", fmtInt(int(req.Start)))
	}
	if req.Limit > 0 {
		query.Set("limit", fmtInt(int(req.Limit)))
	}
	if req.Level > 0 {
		query.Set("level", fmtInt(int(req.Level)))
	}
	if req.TraceID != "" {
		query.Set("traceId", req.TraceID)
	}
	if req.URL != "" {
		query.Set("url", req.URL)
	}
	if req.ID != "" {
		query.Set("id", req.ID)
	}
	if req.FilterMsg != "" {
		query.Set("filterMsg", req.FilterMsg)
	}
	var result RealtimeLogSearchResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/userlog/userlog_search", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VisitScene is one visit scene of the Mini Program.
type VisitScene struct {
	// Name of the scene.
	Name string `json:"name"`
	// Value of the scene id.
	Value string `json:"value"`
}

// GetSceneListResponse is returned by GetSceneList.
type GetSceneListResponse struct {
	ErrResponse
	// Scene is the list of visit scenes.
	Scene []VisitScene `json:"scene"`
}

// GetSceneList lists the visit scenes (referral channels) of the Mini Program
// for the previous day.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getscenelist.html
func (w *MiniProgram) GetSceneList(ctx context.Context) (*GetSceneListResponse, error) {
	var result GetSceneListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/log/get_scene", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ClientVersionGroup groups client versions by type.
type ClientVersionGroup struct {
	// Type of the client (e.g. 1 iOS, 2 Android).
	Type int `json:"type"`
	// ClientVersionList is the list of version strings.
	ClientVersionList []string `json:"client_version_list"`
}

// GetVersionListResponse is returned by GetVersionList.
type GetVersionListResponse struct {
	ErrResponse
	// CvList is the list of client version groups.
	CvList []ClientVersionGroup `json:"cvlist"`
}

// GetVersionList lists the client versions in use of the Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getversionlist.html
func (w *MiniProgram) GetVersionList(ctx context.Context) (*GetVersionListResponse, error) {
	var result GetVersionListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/log/get_client_version", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GrayReleasePlan describes an in-flight gray release.
type GrayReleasePlan struct {
	// Status of the gray release: 0 = none, 1 = running, 2 = paused,
	// 3 = finished, 4 = waiting for manual release.
	Status int `json:"status"`
	// CreateTimestamp of the plan (Unix seconds).
	CreateTimestamp int64 `json:"create_timestamp"`
	// GrayPercentage of traffic routed to the gray version (0-100).
	GrayPercentage int `json:"gray_percentage"`
	// SupportExperiencerFirst lets experiencers verify the gray version first.
	SupportExperiencerFirst bool `json:"support_experiencer_first"`
	// SupportDebugerFirst lets debuggers verify the gray version first.
	SupportDebugerFirst bool `json:"support_debuger_first"`
}

// GetGrayReleasePlanResponse is returned by GetGrayReleasePlan.
type GetGrayReleasePlanResponse struct {
	ErrResponse
	// GrayReleasePlan of the Mini Program, zero value when none is running.
	GrayReleasePlan GrayReleasePlan `json:"gray_release_plan"`
}

// GetGrayReleasePlan returns the current gray release plan of the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getgrayreleaseplan.html
func (w *MiniProgram) GetGrayReleasePlan(ctx context.Context) (*GetGrayReleasePlanResponse, error) {
	var result GetGrayReleasePlanResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxa/getgrayreleaseplan", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DomainInfo describes the configured network domains of the Mini Program.
type DomainInfo struct {
	// RequestDomain is the list of allowed request domains.
	RequestDomain []string `json:"requestdomain"`
	// WsRequestDomain is the list of allowed socket domains.
	WsRequestDomain []string `json:"wsrequestdomain"`
	// UploadDomain is the list of allowed upload domains.
	UploadDomain []string `json:"uploaddomain"`
	// DownloadDomain is the list of allowed download domains.
	DownloadDomain []string `json:"downloaddomain"`
	// UDPDomain is the list of allowed UDP domains.
	UDPDomain []string `json:"udpdomain"`
	// BizDomain is the list of allowed business domains.
	BizDomain []string `json:"bizdomain"`
}

// GetDomainInfoResponse is returned by GetDomainInfo.
type GetDomainInfoResponse struct {
	ErrResponse
	DomainInfo DomainInfo `json:"domain_info"`
}

// GetDomainInfo returns the currently configured network domains of the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getdomaininfo.html
func (w *MiniProgram) GetDomainInfo(ctx context.Context) (*GetDomainInfoResponse, error) {
	var result GetDomainInfoResponse
	body := map[string]string{"action": "get_domain_info"}
	if err := w.withAccessTokenPost(ctx, "/wxa/getwxadevinfo", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPerformanceRequest selects a performance report (page/network metrics).
type GetPerformanceRequest struct {
	// CostTimeType is the metric type (e.g. 1 for the selected set per docs).
	CostTimeType int `json:"cost_time_type,omitempty"`
	// DefaultStartTime of the default comparison window.
	DefaultStartTime int64 `json:"default_start_time,omitempty"`
	// DefaultEndTime of the default comparison window.
	DefaultEndTime int64 `json:"default_end_time,omitempty"`
	// Device filter.
	Device string `json:"device,omitempty"`
	// IsDownloadCode selects whether download-code triggers are included.
	IsDownloadCode string `json:"is_download_code,omitempty"`
	// Scene filter.
	Scene string `json:"scene,omitempty"`
	// NetworkType filter.
	NetworkType string `json:"networktype,omitempty"`
}

// PerformanceTimeData is one window of performance data.
type PerformanceTimeData struct {
	// List of metric rows.
	List []PerformanceDataRow `json:"list"`
}

// PerformanceDataRow is one performance metric row.
type PerformanceDataRow struct {
	// RefData is the reference date/key of the row.
	RefData string `json:"ref_data"`
	// CostTimeType of the metric.
	CostTimeType int `json:"cost_time_type"`
	// CostTime is the metric value.
	CostTime int `json:"cost_time"`
}

// GetPerformanceResponse is returned by GetPerformanceData.
type GetPerformanceResponse struct {
	ErrResponse
	// DefaultTimeData is the requested window payload.
	DefaultTimeData string `json:"default_time_data"`
	// CompareTimeData is the comparison window payload.
	CompareTimeData string `json:"compare_time_data"`
}

// GetPerformanceData returns Mini Program performance statistics for the
// configured metric and windows.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/operation/api_getperformance.html
func (w *MiniProgram) GetPerformanceData(ctx context.Context, req *GetPerformanceRequest) (*GetPerformanceResponse, error) {
	var result GetPerformanceResponse
	if err := w.withAccessTokenPost(ctx, "/wxaapi/log/get_performance", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
