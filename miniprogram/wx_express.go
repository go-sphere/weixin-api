package miniprogram

import (
	"context"
)

// OpenMsgDeliveryCompany is one delivery company supported by the logistics
// open-msg (物流服务消息) capability.
type OpenMsgDeliveryCompany struct {
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// DeliveryName of the company.
	DeliveryName string `json:"delivery_name"`
}

// GetOpenMsgDeliveryListResponse is returned by GetOpenMsgDeliveryList.
type GetOpenMsgDeliveryListResponse struct {
	ErrResponse
	// DeliveryList of the supported companies.
	DeliveryList []OpenMsgDeliveryCompany `json:"delivery_list,omitempty"`
}

// GetOpenMsgDeliveryList returns the delivery companies supported by the
// logistics service message capability.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-msg/api_get_delivery_list.html
func (w *MiniProgram) GetOpenMsgDeliveryList(ctx context.Context) (*GetOpenMsgDeliveryListResponse, error) {
	var result GetOpenMsgDeliveryListResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/open_msg/get_delivery_list", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FollowWaybillRequest subscribes a user to the tracking updates of a waybill.
type FollowWaybillRequest struct {
	// OpenID of the user to notify.
	OpenID string `json:"openid"`
	// DeliveryID of the delivery company.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the waybill to follow.
	WaybillID string `json:"waybill_id"`
}

// FollowWaybill subscribes a user to a waybill's tracking messages.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-msg/api_follow_waybill.html
func (w *MiniProgram) FollowWaybill(ctx context.Context, req *FollowWaybillRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/open_msg/follow_waybill", nil, defaultReqOptions(), req, nil)
}

// ExpressTrackPoint is one tracking point of a waybill trace.
type ExpressTrackPoint struct {
	// Time of the tracking point (Unix seconds).
	Time int64 `json:"time,omitempty"`
	// Status of the tracking point.
	Status int `json:"status,omitempty"`
	// Desc of the tracking point.
	Desc string `json:"desc,omitempty"`
	// City of the tracking point.
	City string `json:"city,omitempty"`
}

// QueryFollowTraceResponse is returned by QueryFollowTrace.
type QueryFollowTraceResponse struct {
	ErrResponse
	// Trace of the waybill.
	Trace []ExpressTrackPoint `json:"trace,omitempty"`
}

// QueryFollowTrace returns the tracking timeline of a followed waybill.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-msg/api_query_follow_trace.html
func (w *MiniProgram) QueryFollowTrace(ctx context.Context, req *FollowWaybillRequest) (*QueryFollowTraceResponse, error) {
	var result QueryFollowTraceResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/open_msg/query_follow_trace", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateFollowWaybillGoodsRequest attaches the purchased goods name to a
// followed waybill.
type UpdateFollowWaybillGoodsRequest struct {
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
	// GoodsName of the shipped goods.
	GoodsName string `json:"goods_name"`
}

// UpdateFollowWaybillGoods updates the goods name shown with a followed
// waybill.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-msg/api_update_follow_waybill_goods.html
func (w *MiniProgram) UpdateFollowWaybillGoods(ctx context.Context, req *UpdateFollowWaybillGoodsRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/open_msg/update_follow_waybill_goods", nil, defaultReqOptions(), req, nil)
}

// TraceWaybillRequest queries the live trace of a waybill via the open msg
// trace capability.
type TraceWaybillRequest struct {
	// OpenID of the user.
	OpenID string `json:"openid,omitempty"`
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
}

// TraceWaybillResponse is returned by TraceWaybill.
type TraceWaybillResponse struct {
	ErrResponse
	// Trace of the waybill.
	Trace []ExpressTrackPoint `json:"trace,omitempty"`
}

// TraceWaybill returns the tracking timeline of a waybill through the
// 运单查询 (trace) open-msg capability.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-search/api_trace_waybill.html
func (w *MiniProgram) TraceWaybill(ctx context.Context, req *TraceWaybillRequest) (*TraceWaybillResponse, error) {
	var result TraceWaybillResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/open_msg/trace_waybill", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryTraceResponse is returned by QueryTrace.
type QueryTraceResponse struct {
	ErrResponse
	// Trace of the queried waybill.
	Trace []ExpressTrackPoint `json:"trace,omitempty"`
}

// QueryTrace queries the tracking timeline for a waybill by delivery id and
// waybill number.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-search/api_query_trace.html
func (w *MiniProgram) QueryTrace(ctx context.Context, req *TraceWaybillRequest) (*QueryTraceResponse, error) {
	var result QueryTraceResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/open_msg/query_trace", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWaybillGoodsRequest updates the goods name of a queried waybill.
type UpdateWaybillGoodsRequest struct {
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
	// GoodsName of the shipped goods.
	GoodsName string `json:"goods_name,omitempty"`
}

// UpdateWaybillGoods updates the goods display name of a waybill within the
// express search box.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-search/api_update_waybill_goods.html
func (w *MiniProgram) UpdateWaybillGoods(ctx context.Context, req *UpdateWaybillGoodsRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/open_msg/update_waybill_goods", nil, defaultReqOptions(), req, nil)
}
