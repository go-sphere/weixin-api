package miniprogram

import (
	"context"
	"fmt"
)

// fmtInt renders an int as a decimal string.
func fmtInt(v int) string { return fmt.Sprintf("%d", v) }

// dailyAnalysisBody builds the standard {begin_date,end_date} JSON body shared
// by the /datacube analysis endpoints. beginDate/endDate use the "YYYY-MM-DD"
// layout required by WeChat.
func dailyAnalysisBody(beginDate, endDate string) map[string]string {
	return map[string]string{
		"begin_date": beginDate,
		"end_date":   endDate,
	}
}

// DataCubeAnalysisValue counts one analysis dimension bucket.
type DataCubeAnalysisValue struct {
	// Key is the dimension value (e.g. an access-source code or a stay-time
	// bucket like "0-1s").
	Key string `json:"key"`
	// Value is the number of users in that bucket.
	Value int `json:"value"`
}

// VisitPageDailySummary is one row of the daily visit summary report.
type VisitPageDailySummary struct {
	// RefDate is the day the row covers, "YYYY-MM-DD".
	RefDate string `json:"ref_date"`
	// SessionCount is the number of sessions.
	SessionCount int `json:"session_cnt"`
	// VisitPV is the number of visits (page views).
	VisitPV int `json:"visit_pv"`
	// VisitUV is the number of distinct visiting users.
	VisitUV int `json:"visit_uv"`
	// VisitUVNew is the number of first time visitors.
	VisitUVNew int `json:"visit_uv_new"`
	// StayTimeUV is the total stay time (seconds) of distinct users.
	StayTimeUV float64 `json:"stay_time_uv"`
	// StayTimeSession is the average stay time per session (seconds).
	StayTimeSession float64 `json:"stay_time_session"`
	// VisitDepth is the average visit depth.
	VisitDepth float64 `json:"visit_depth"`
	// VisitDepthNew is the average visit depth of new users.
	VisitDepthNew float64 `json:"visit_depth_new"`
}

// GetDailyVisitSummaryResponse is returned by GetDailyVisitSummary.
type GetDailyVisitSummaryResponse struct {
	ErrResponse
	// List holds one row per requested day.
	List []VisitPageDailySummary `json:"list"`
}

// GetDailyVisitSummary returns the daily visit statistics (page views, unique
// users, stay durations) for the requested range. Up to 30 days between begin
// and end, queried in 1-day granularity.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/others/api_getdailysummary.html
func (w *MiniProgram) GetDailyVisitSummary(ctx context.Context, beginDate, endDate string) (*GetDailyVisitSummaryResponse, error) {
	var result GetDailyVisitSummaryResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappiddailysummarytrend", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VisitTrendDaily is one row of the daily visit trend report.
type VisitTrendDaily struct {
	// RefDate of the row, "YYYY-MM-DD".
	RefDate string `json:"ref_date"`
	// SessionCount is the number of sessions.
	SessionCount int `json:"session_cnt"`
	// VisitPV is the number of visits.
	VisitPV int `json:"visit_pv"`
	// VisitUV is the number of distinct visiting users.
	VisitUV int `json:"visit_uv"`
	// VisitUVNew is the number of first time visitors.
	VisitUVNew int `json:"visit_uv_new"`
	// StayTimeUV is the total stay time (seconds) of distinct users.
	StayTimeUV float64 `json:"stay_time_uv"`
	// StayTimeSession is the average stay time per session (seconds).
	StayTimeSession float64 `json:"stay_time_session"`
	// VisitDepth is the average visit depth.
	VisitDepth float64 `json:"visit_depth"`
	// VisitDepthNew is the average visit depth of new users.
	VisitDepthNew float64 `json:"visit_depth_new"`
}

// GetDailyVisitTrendResponse is returned by GetDailyVisitTrend.
type GetDailyVisitTrendResponse struct {
	ErrResponse
	// List holds one row per requested day.
	List []VisitTrendDaily `json:"list"`
}

// GetDailyVisitTrend returns the daily visit trend. Query a single day or a
// whole-day range of up to 30 days.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/visit-trend/api_getdailyvisittrend.html
func (w *MiniProgram) GetDailyVisitTrend(ctx context.Context, beginDate, endDate string) (*GetDailyVisitTrendResponse, error) {
	var result GetDailyVisitTrendResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappiddailyvisittrend", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VisitTrendWeekly mirrors VisitTrendDaily for weekly aggregates.
type VisitTrendWeekly struct {
	// RefDate is the start day (Monday) of the week, "YYYY-MM-DD".
	RefDate string `json:"ref_date"`
	// SessionCount is the number of sessions.
	SessionCount int `json:"session_cnt"`
	// VisitPV is the number of visits.
	VisitPV int `json:"visit_pv"`
	// VisitUV is the number of distinct visiting users.
	VisitUV int `json:"visit_uv"`
	// VisitUVNew is the number of first time visitors.
	VisitUVNew int `json:"visit_uv_new"`
	// StayTimeUV is the total stay time (seconds) of distinct users.
	StayTimeUV float64 `json:"stay_time_uv"`
	// StayTimeSession is the average stay time per session (seconds).
	StayTimeSession float64 `json:"stay_time_session"`
	// VisitDepth is the average visit depth.
	VisitDepth float64 `json:"visit_depth"`
	// VisitDepthNew is the average visit depth of new users.
	VisitDepthNew float64 `json:"visit_depth_new"`
}

// GetWeeklyVisitTrendResponse is returned by GetWeeklyVisitTrend.
type GetWeeklyVisitTrendResponse struct {
	ErrResponse
	// List holds one row per requested week.
	List []VisitTrendWeekly `json:"list"`
}

// GetWeeklyVisitTrend returns the weekly visit trend. The range must span whole
// weeks (Monday to Sunday) and cover up to 10 weeks.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/visit-trend/api_getweeklyvisittrend.html
func (w *MiniProgram) GetWeeklyVisitTrend(ctx context.Context, beginDate, endDate string) (*GetWeeklyVisitTrendResponse, error) {
	var result GetWeeklyVisitTrendResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappidweeklyvisittrend", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VisitTrendMonthly mirrors VisitTrendDaily for monthly aggregates.
type VisitTrendMonthly struct {
	// RefDate is the start day of the month, "YYYY-MM-DD".
	RefDate string `json:"ref_date"`
	// SessionCount is the number of sessions.
	SessionCount int `json:"session_cnt"`
	// VisitPV is the number of visits.
	VisitPV int `json:"visit_pv"`
	// VisitUV is the number of distinct visiting users.
	VisitUV int `json:"visit_uv"`
	// VisitUVNew is the number of first time visitors.
	VisitUVNew int `json:"visit_uv_new"`
	// StayTimeUV is the total stay time (seconds) of distinct users.
	StayTimeUV float64 `json:"stay_time_uv"`
	// StayTimeSession is the average stay time per session (seconds).
	StayTimeSession float64 `json:"stay_time_session"`
	// VisitDepth is the average visit depth.
	VisitDepth float64 `json:"visit_depth"`
	// VisitDepthNew is the average visit depth of new users.
	VisitDepthNew float64 `json:"visit_depth_new"`
}

// GetMonthlyVisitTrendResponse is returned by GetMonthlyVisitTrend.
type GetMonthlyVisitTrendResponse struct {
	ErrResponse
	// List holds one row per requested month.
	List []VisitTrendMonthly `json:"list"`
}

// GetMonthlyVisitTrend returns the monthly visit trend. The range must span
// whole natural months and cover up to 30 months.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/visit-trend/api_getmonthlyvisittrend.html
func (w *MiniProgram) GetMonthlyVisitTrend(ctx context.Context, beginDate, endDate string) (*GetMonthlyVisitTrendResponse, error) {
	var result GetMonthlyVisitTrendResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappidmonthlyvisittrend", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VisitDistributionDaily is one row of the visit distribution report.
type VisitDistributionDaily struct {
	// RefDate of the row.
	RefDate string `json:"ref_date"`
	// List holds the per-dimension distribution buckets: index 0 is the
	// access-source distribution, index 1 the visit-depth buckets and index 2
	// the stay-time buckets.
	List []VisitDistributionDimension `json:"list"`
}

// VisitDistributionDimension is one distribution dimension of a day.
type VisitDistributionDimension struct {
	// Index selects the dimension (0 access source, 1 visit depth, 2 stay
	// time); the meaning of Items.Key depends on it.
	Index int `json:"index"`
	// Items are the per-bucket user counts.
	Items []DataCubeAnalysisValue `json:"item_list"`
}

// GetVisitDistributionResponse is returned by GetVisitDistribution.
type GetVisitDistributionResponse struct {
	ErrResponse
	// List holds one row per requested day.
	List []VisitDistributionDaily `json:"list"`
}

// GetVisitDistribution reports how visits distribute across access sources,
// visit depth and stay time per day.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/others/api_getvisitdistribution.html
func (w *MiniProgram) GetVisitDistribution(ctx context.Context, beginDate, endDate string) (*GetVisitDistributionResponse, error) {
	var result GetVisitDistributionResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappidvisitdistribution", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VisitPageRow is one row of the per-page visit report.
type VisitPageRow struct {
	// PagePath is the page path the row covers.
	PagePath string `json:"page_path"`
	// PageVisitPV is the number of visits of the page.
	PageVisitPV int `json:"page_visit_pv"`
	// PageVisitUV is the number of distinct visiting users.
	PageVisitUV int `json:"page_visit_uv"`
	// PageStayTime is the average stay time (seconds) on the page.
	PageStayTime float64 `json:"page_staytime"`
	// EntryPageVisitPV is the number of entries landing on this page.
	EntryPageVisitPV int `json:"entrypage_pv"`
	// ExitPageVisitPV is the number of exits happening on this page.
	ExitPageVisitPV int `json:"exitpage_pv"`
	// PageSharePV is the number of shares triggered from this page.
	PageSharePV int `json:"page_share_pv"`
	// PageShareUV is the number of users who shared from this page.
	PageShareUV int `json:"page_share_uv"`
}

// GetVisitPageResponse is returned by GetVisitPage.
type GetVisitPageResponse struct {
	ErrResponse
	// List holds one row per page over the whole range.
	List []VisitPageRow `json:"list"`
}

// GetVisitPage reports per-page visit statistics over the requested range
// (up to 30 days).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/others/api_getvisitpage.html
func (w *MiniProgram) GetVisitPage(ctx context.Context, beginDate, endDate string) (*GetVisitPageResponse, error) {
	var result GetVisitPageResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappidvisitpage", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UserPortraitValue describes one user portrait bucket.
type UserPortraitValue struct {
	// ID of the bucket (age code, gender 1/2, province code...).
	ID int `json:"id"`
	// Name of the bucket.
	Name string `json:"name"`
	// Value is the number of users in the bucket.
	Value int `json:"value"`
}

// UserPortraitData carries the buckets of one portrait dimension.
type UserPortraitData struct {
	// Province list of the user distribution.
	Province []UserPortraitValue `json:"province"`
	// City list of the user distribution.
	City []UserPortraitValue `json:"city"`
	// Gender list (1 male, 2 female).
	Gender []UserPortraitValue `json:"gender"`
	// Platform list (Android/iOS/devtools and so on).
	Platform []UserPortraitValue `json:"platform"`
	// Device list of device models.
	Device []UserPortraitValue `json:"device"`
	// Age list of age buckets.
	Age []UserPortraitValue `json:"age"`
}

// GetUserPortraitResponse is returned by GetUserPortrait.
type GetUserPortraitResponse struct {
	ErrResponse
	// RefDate of the report.
	RefDate string `json:"ref_date"`
	// VisitUVNew is the number of new users the report covers.
	VisitUVNew int `json:"visit_uv_new"`
	// VisitUV is the number of active users the report covers.
	VisitUV int `json:"visit_uv"`
	// VisitUVNewData is the portrait of new users.
	VisitUVNewData UserPortraitData `json:"visit_uv_new_data"`
	// VisitUVData is the portrait of active users.
	VisitUVData UserPortraitData `json:"visit_uv_data"`
}

// GetUserPortrait returns demographic, geographic and platform portraits of
// the Mini Program users for a single day (beginDate == endDate).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/others/api_getuserportrait.html
func (w *MiniProgram) GetUserPortrait(ctx context.Context, beginDate, endDate string) (*GetUserPortraitResponse, error) {
	var result GetUserPortraitResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappiduserportrait", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RetainItem is one retention data point: Key is the day/week/month offset and
// Value the number of retained users.
type RetainItem struct {
	// Key is the retention offset (days since the reference date).
	Key int `json:"key"`
	// Value is the number of users who returned at that offset.
	Value int `json:"value"`
}

// GetDailyRetainResponse is returned by GetDailyRetain.
type GetDailyRetainResponse struct {
	ErrResponse
	// RefDate of the report, "YYYY-MM-DD".
	RefDate string `json:"ref_date"`
	// VisitUVNew is the retention curve of new users (offset in days).
	VisitUVNew []RetainItem `json:"visit_uv_new"`
	// VisitUV is the retention curve of active users (offset in days).
	VisitUV []RetainItem `json:"visit_uv"`
}

// GetDailyRetain returns the daily retention report: how many of the users
// active on the reference day come back over the following days. Query a single
// day.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/visit-retain/api_getdailyretain.html
func (w *MiniProgram) GetDailyRetain(ctx context.Context, beginDate, endDate string) (*GetDailyRetainResponse, error) {
	var result GetDailyRetainResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappiddailyretaininfo", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWeeklyRetainResponse is returned by GetWeeklyRetain.
type GetWeeklyRetainResponse struct {
	ErrResponse
	// RefDate is the Monday of the reference week, "YYYY-MM-DD".
	RefDate string `json:"ref_date"`
	// VisitUVNew is the retention curve of new users (offset in weeks).
	VisitUVNew []RetainItem `json:"visit_uv_new"`
	// VisitUV is the retention curve of active users (offset in weeks).
	VisitUV []RetainItem `json:"visit_uv"`
	// VisitUVNewData breaks new-user retention down per week day.
	VisitUVNewData []RetainItem `json:"visit_uv_new_data"`
	// VisitUVData breaks active-user retention down per week day.
	VisitUVData []RetainItem `json:"visit_uv_data"`
}

// GetWeeklyRetain returns the weekly retention report. Query a single week
// (Monday to Sunday).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/visit-retain/api_getweeklyretain.html
func (w *MiniProgram) GetWeeklyRetain(ctx context.Context, beginDate, endDate string) (*GetWeeklyRetainResponse, error) {
	var result GetWeeklyRetainResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappidweeklyretaininfo", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMonthlyRetainResponse is returned by GetMonthlyRetain.
type GetMonthlyRetainResponse struct {
	ErrResponse
	// RefDate is the first day of the reference month, "YYYY-MM-DD".
	RefDate string `json:"ref_date"`
	// VisitUVNew is the retention curve of new users (offset in months).
	VisitUVNew []RetainItem `json:"visit_uv_new"`
	// VisitUV is the retention curve of active users (offset in months).
	VisitUV []RetainItem `json:"visit_uv"`
}

// GetMonthlyRetain returns the monthly retention report. Query a single month.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/visit-retain/api_getmonthlyretain.html
func (w *MiniProgram) GetMonthlyRetain(ctx context.Context, beginDate, endDate string) (*GetMonthlyRetainResponse, error) {
	var result GetMonthlyRetainResponse
	err := w.withAccessTokenPost(ctx, "/datacube/getweanalysisappidmonthlyretaininfo", nil, defaultReqOptions(), dailyAnalysisBody(beginDate, endDate), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
