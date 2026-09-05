package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// NearbyPoiListItem is one row of GetNearbyPoiList. RawData preserves the full
// upstream JSON object because WeChat returns a semi-structured payload that is
// not stable enough to model field-by-field.
type NearbyPoiListItem struct {
	// POIID of the nearby location entry.
	POIID int64 `json:"poi_id,omitempty"`
	// Status of the entry audit.
	Status int `json:"status,omitempty"`
	// AuditID of the last audit submission.
	AuditID int64 `json:"audit_id,omitempty"`
	// AuditStatus of the last audit.
	AuditStatus int `json:"audit_status,omitempty"`
	// AuditResult message of the last audit.
	AuditResult string `json:"audit_result,omitempty"`
	// IsShow is 1 when the entry is currently shown to users.
	IsShow int `json:"is_show,omitempty"`
	// RawData holds the complete entry JSON as returned by WeChat.
	RawData map[string]any `json:"-"`
}

// GetNearbyPoiListResponse is returned by GetNearbyPoiList.
type GetNearbyPoiListResponse struct {
	ErrResponse
	// LeftCount is the remaining quota of "nearby Mini Program" entries.
	LeftCount int `json:"left_count"`
	// TotalCount is the total quota of entries the account may configure.
	TotalCount int `json:"total_count"`
	// List is the requested page of location entries.
	List []NearbyPoiListItem `json:"list,omitempty"`
}

// GetNearbyPoiList lists the "nearby Mini Program" (附近的小程序) location
// entries of the account, paged.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/nearby-poi/api_getnearbypoilist.html
func (w *MiniProgram) GetNearbyPoiList(ctx context.Context, page, pageRows int) (*GetNearbyPoiListResponse, error) {
	query := url.Values{}
	query.Set("page", fmtInt(page))
	query.Set("page_rows", fmtInt(pageRows))
	var result GetNearbyPoiListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxa/getnearbypoilist", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddNearbyPoi submits a new "nearby Mini Program" location entry. The request
// schema is verbose and versioned (store info, credentials, business hours,
// service info...); pass the exact JSON object described by the reference doc.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/nearby-poi/api_addnearbypoi.html
func (w *MiniProgram) AddNearbyPoi(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/addnearbypoi", nil, defaultReqOptions(), body, nil)
}

// DeleteNearbyPoi removes a nearby location entry by its poi id.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/nearby-poi/api_deletenearbypoi.html
func (w *MiniProgram) DeleteNearbyPoi(ctx context.Context, poiID string) error {
	body := map[string]string{"poi_id": poiID}
	return w.withAccessTokenPost(ctx, "/wxa/delnearbypoi", nil, defaultReqOptions(), body, nil)
}

// SetNearbyPoiShowStatusRequest toggles the visibility of a nearby entry.
type SetNearbyPoiShowStatusRequest struct {
	// PoiID of the entry.
	PoiID string `json:"poi_id"`
	// Status: 0 hides the entry, 1 shows it.
	Status int `json:"status"`
}

// SetNearbyPoiShowStatus shows or hides a "nearby Mini Program" location entry.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/nearby-poi/api_setshowstatus.html
func (w *MiniProgram) SetNearbyPoiShowStatus(ctx context.Context, req *SetNearbyPoiShowStatusRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/setnearbypoishowstatus", nil, defaultReqOptions(), req, nil)
}
