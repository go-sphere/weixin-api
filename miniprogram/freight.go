package miniprogram

import (
	"context"
)

// ============================================================
// 运费险 (freight insurance) — 微信运费险 group. Routes live under
// /wxa/business/insurance_freight/*. The Mini Program opens the capability,
// reports paid orders (createorder) and lets buyers claim refunds of their
// freight through the insurance.
// ============================================================

// FreightInsuranceOrderAddress is a delivery address of an insured order.
type FreightInsuranceOrderAddress struct {
	// Province of the address.
	Province string `json:"province"`
	// City of the address.
	City string `json:"city"`
	// District of the address.
	District string `json:"county"`
	// DetailAddress of the address.
	DetailAddress string `json:"address"`
}

// CreateFreightInsuranceOrderRequest reports a paid freight-insurance order.
type CreateFreightInsuranceOrderRequest struct {
	// OpenID of the paying user.
	OpenID string `json:"openid"`
	// OrderNo of the merchant order.
	OrderNo string `json:"order_no"`
	// PayTime of the order (unix seconds).
	PayTime int64 `json:"pay_time"`
	// PayAmount of the order in cents.
	PayAmount int `json:"pay_amount"`
	// DeliveryNo of the shipping waybill.
	DeliveryNo string `json:"delivery_no"`
	// DeliveryAddress of the shipment.
	DeliveryAddress *FreightInsuranceOrderAddress `json:"delivery_place"`
	// ReceiptAddress of the recipient.
	ReceiptAddress *FreightInsuranceOrderAddress `json:"receipt_place"`
}

// CreateFreightInsuranceOrder reports an order eligible for freight insurance.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_createorder.html
func (w *MiniProgram) CreateFreightInsuranceOrder(ctx context.Context, req *CreateFreightInsuranceOrderRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/createorder", nil, defaultReqOptions(), req, nil)
}

// CreateFreightChargeIDRequest buys freight-insurance quota (投保).
type CreateFreightChargeIDRequest struct {
	// Quota amount to purchase in cents.
	Quota int `json:"quota"`
}

// CreateFreightChargeIDResponse is returned by CreateFreightChargeID.
type CreateFreightChargeIDResponse struct {
	ErrResponse
	// ChargeID of the purchased quota order.
	ChargeID string `json:"charge_id,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// CreateFreightChargeID purchases freight-insurance quota (充值额度) for the
// merchant account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_createchargeid.html
func (w *MiniProgram) CreateFreightChargeID(ctx context.Context, quota int) (*CreateFreightChargeIDResponse, error) {
	var result CreateFreightChargeIDResponse
	req := &CreateFreightChargeIDRequest{Quota: quota}
	if err := w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/createchargeid", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ApplyFreightPayRequest pays for the purchased freight-insurance quota.
type ApplyFreightPayRequest struct {
	// OrderID of the charge order returned by CreateFreightChargeID.
	OrderID string `json:"order_id"`
}

// ApplyFreightPay pays the freight-insurance quota order created by
// CreateFreightChargeID.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_applypay.html
func (w *MiniProgram) ApplyFreightPay(ctx context.Context, orderID string) error {
	req := &ApplyFreightPayRequest{OrderID: orderID}
	return w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/applypay", nil, defaultReqOptions(), req, nil)
}

// FreightOrderQuery selects insured freight orders.
type FreightOrderQuery struct {
	// OpenID of the user.
	OpenID string `json:"openid,omitempty"`
	// OrderNo of the merchant.
	OrderNo string `json:"order_no,omitempty"`
	// PolicyNo of the insurance policy.
	PolicyNo string `json:"policy_no,omitempty"`
	// ReportNo of a claim report.
	ReportNo string `json:"report_no,omitempty"`
	// DeliveryNo of the waybill.
	DeliveryNo string `json:"delivery_no,omitempty"`
	// RefundDeliveryNo of a return waybill.
	RefundDeliveryNo string `json:"refund_delivery_no,omitempty"`
	// BeginTime of the window (unix seconds).
	BeginTime int64 `json:"begin_time,omitempty"`
	// EndTime of the window (unix seconds).
	EndTime int64 `json:"end_time,omitempty"`
	// StatusList filter of the order states.
	StatusList []int `json:"status_list,omitempty"`
	// Offset of the page.
	Offset int `json:"offset,omitempty"`
	// Limit of the page.
	Limit int `json:"limit,omitempty"`
	// SortDirect: 0 desc, 1 asc.
	SortDirect int `json:"sort_direct,omitempty"`
}

// GetFreightOrderListResponse is returned by GetFreightOrderList.
type GetFreightOrderListResponse struct {
	ErrResponse
	// OrderList of the query.
	OrderList []map[string]any `json:"order_list,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// GetFreightOrderList lists the insured freight orders of the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_getorderlist.html
func (w *MiniProgram) GetFreightOrderList(ctx context.Context, req *FreightOrderQuery) (*GetFreightOrderListResponse, error) {
	var result GetFreightOrderListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/getorderlist", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFreightPayOrderListRequest pages the freight-insurance payment orders.
type GetFreightPayOrderListRequest struct {
	// StatusList filter of the payment states.
	StatusList []int `json:"status_list,omitempty"`
	// Offset of the page.
	Offset int `json:"offset,omitempty"`
	// Limit of the page.
	Limit int `json:"limit,omitempty"`
}

// GetFreightPayOrderListResponse is returned by GetFreightPayOrderList.
type GetFreightPayOrderListResponse struct {
	ErrResponse
	// OrderList of the quota payment orders.
	OrderList []map[string]any `json:"order_list,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// GetFreightPayOrderList lists the freight-insurance quota payment orders.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_getpayorderlist.html
func (w *MiniProgram) GetFreightPayOrderList(ctx context.Context, req *GetFreightPayOrderListRequest) (*GetFreightPayOrderListResponse, error) {
	var result GetFreightPayOrderListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/getpayorderlist", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFreightSummaryRequest selects the window of the freight-insurance summary.
type GetFreightSummaryRequest struct {
	// BeginTime of the window (unix seconds).
	BeginTime int64 `json:"begin_time"`
	// EndTime of the window (unix seconds).
	EndTime int64 `json:"end_time"`
}

// GetFreightSummaryResponse is returned by GetFreightSummary.
type GetFreightSummaryResponse struct {
	ErrResponse
	// Summary of the freight insurance usage in the window.
	Summary map[string]any `json:"summary,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// GetFreightSummary returns the freight-insurance usage summary of a window.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_getsummary.html
func (w *MiniProgram) GetFreightSummary(ctx context.Context, req *GetFreightSummaryRequest) (*GetFreightSummaryResponse, error) {
	var result GetFreightSummaryResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/getsummary", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ClaimFreightInsuranceRequest starts a freight-insurance claim.
type ClaimFreightInsuranceRequest struct {
	// OpenID of the claimant user.
	OpenID string `json:"openid"`
	// OrderNo of the original insured order.
	OrderNo string `json:"order_no"`
	// RefundDeliveryNo of the return waybill.
	RefundDeliveryNo string `json:"refund_delivery_no"`
	// RefundCompany of the return delivery company.
	RefundCompany string `json:"refund_company"`
}

// ClaimFreightInsurance submits a buyer's freight-insurance claim (理赔).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_claim.html
func (w *MiniProgram) ClaimFreightInsurance(ctx context.Context, req *ClaimFreightInsuranceRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/claim", nil, defaultReqOptions(), req, nil)
}

// FreightInsuranceRefund refunds unused freight-insurance quota.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_refund.html
func (w *MiniProgram) FreightInsuranceRefund(ctx context.Context) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/refund", nil, defaultReqOptions(), map[string]string{}, nil)
}

// OpenFreightInsurance enables the freight-insurance capability for the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_open.html
func (w *MiniProgram) OpenFreightInsurance(ctx context.Context) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/open", nil, defaultReqOptions(), map[string]string{}, nil)
}

// QueryFreightInsuranceOpenResponse is returned by QueryFreightInsuranceOpen.
type QueryFreightInsuranceOpenResponse struct {
	ErrResponse
	// RawData of the status payload.
	RawData map[string]any `json:"-"`
}

// QueryFreightInsuranceOpen queries whether the freight-insurance capability
// is enabled.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_query_open.html
func (w *MiniProgram) QueryFreightInsuranceOpen(ctx context.Context) (*QueryFreightInsuranceOpenResponse, error) {
	var result QueryFreightInsuranceOpenResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/query_open", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateFreightNotifyFundsRequest notifies WeChat about the freight-insurance
// fund settlement.
type UpdateFreightNotifyFundsRequest struct {
	// NotifyFunds payload of the settlement.
	NotifyFunds string `json:"notify_funds"`
}

// UpdateFreightNotifyFunds notifies the freight-insurance fund settlement
// result to WeChat.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/freight/api_insurance_freight_update_notify_funds.html
func (w *MiniProgram) UpdateFreightNotifyFunds(ctx context.Context, notifyFunds string) error {
	req := &UpdateFreightNotifyFundsRequest{NotifyFunds: notifyFunds}
	return w.withAccessTokenPost(ctx, "/wxa/business/insurance_freight/update_notify_funds", nil, defaultReqOptions(), req, nil)
}
