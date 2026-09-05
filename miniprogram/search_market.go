package miniprogram

import (
	"context"
)

// SubmitPageItem is one page to submit to the WeChat search index.
type SubmitPageItem struct {
	// Path of the Mini Program page, e.g. "pages/index/index?foo=bar".
	Path string `json:"path"`
	// Query of the page (merged into Path when Path has none).
	Query string `json:"query,omitempty"`
	// Title of the page shown in search results (up to 200 chars, no
	// emoji).
	Title string `json:"title"`
	// Content of the page used to build the search snippet (up to 2000
	// chars, plain text).
	Content string `json:"content,omitempty"`
}

// SubmitPagesRequest is the payload of SubmitPages.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/wxsearch/api_submitpages.html
type SubmitPagesRequest struct {
	// Pages to submit, up to 2000 in a single call.
	Pages []SubmitPageItem `json:"pages"`
}

// SubmitPages submits Mini Program pages into the WeChat search index so users
// can discover them through 搜一搜 (WeChat search). Submit regularly (nightly
// recommended).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/wxsearch/api_submitpages.html
func (w *MiniProgram) SubmitPages(ctx context.Context, req *SubmitPagesRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/search/wxaapi_submitpages", nil, defaultReqOptions(), req, nil)
}

// ServiceMarketInvokeRequest is the payload of InvokeServiceMarket. The
// service market (服务市场) lets Mini Programs call paid third-party APIs.
type ServiceMarketInvokeRequest struct {
	// Service of the market API, e.g. "wx79a8d8a1e2f3b4c5" style service id.
	Service string `json:"service"`
	// API of the invoked method within the service.
	API string `json:"api"`
	// Data is the invocation payload.
	Data map[string]any `json:"data,omitempty"`
	// ClientMsgID deduplicates requests.
	ClientMsgID string `json:"client_msg_id,omitempty"`
}

// ServiceMarketInvokeResponse is returned by InvokeServiceMarket. The upstream
// echoes errcode/errmsg only for transport problems; service-level responses
// arrive in Data.
type ServiceMarketInvokeResponse struct {
	ErrResponse
	// Data is the raw service response.
	Data map[string]any `json:"data,omitempty"`
}

// InvokeServiceMarket calls a third-party API purchased through the WeChat
// service market.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/wx-service-market/api_invokeservice.html
func (w *MiniProgram) InvokeServiceMarket(ctx context.Context, req *ServiceMarketInvokeRequest) (*ServiceMarketInvokeResponse, error) {
	var result ServiceMarketInvokeResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/servicemarket", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ServiceMarketRetrieveRequest fetches an async service-market result.
type ServiceMarketRetrieveRequest struct {
	// RequestID returned by InvokeServiceMarket for async calls.
	RequestID string `json:"request_id,omitempty"`
}

// ServiceMarketRetrieveResponse is returned by ServiceMarketRetrieve.
type ServiceMarketRetrieveResponse struct {
	ErrResponse
	// Data of the retrieved result.
	Data map[string]any `json:"data,omitempty"`
}

// ServiceMarketRetrieve retrieves the result of an asynchronous service-market
// invocation.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/wx-service-market/api_servicemarketretrieve.html
func (w *MiniProgram) ServiceMarketRetrieve(ctx context.Context, requestID string) (*ServiceMarketRetrieveResponse, error) {
	req := &ServiceMarketRetrieveRequest{RequestID: requestID}
	var result ServiceMarketRetrieveResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/servicemarketretrieve", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
