package miniprogram

import (
	"context"
)

// ============================================================
// 微信快递 - provider side (快递服务方). Delivery companies push real-time
// status updates and quote fees for waybills ordered through WeChat express.
// ============================================================

// UpdateExpressWaybillPathRequest pushes a tracking event for a waybill.
type UpdateExpressWaybillPathRequest struct {
	// Token pushed by the order event.
	Token string `json:"token"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
	// ActionTime of the event (unix seconds).
	ActionTime int64 `json:"action_time"`
	// ActionType of the tracking event (see WeChat appendix).
	ActionType int `json:"action_type"`
	// ActionMsg detail shown on the tracking page (UTF-8).
	ActionMsg string `json:"action_msg"`
}

// UpdateExpressWaybillPath reports a tracking point to WeChat (轨迹更新).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_updatepath.html
func (w *MiniProgram) UpdateExpressWaybillPath(ctx context.Context, req *UpdateExpressWaybillPathRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/path/update", nil, defaultReqOptions(), req, nil)
}

// UpdateSingleWaybillStatusRequest updates a single-waybill status with the
// courier details.
type UpdateSingleWaybillStatusRequest struct {
	UpdateExpressWaybillPathRequest
	// PickupCourierName of the pickup courier.
	PickupCourierName string `json:"pickup_courier_name"`
	// PickupCourierPhone of the pickup courier.
	PickupCourierPhone string `json:"pickup_courier_phone"`
	// DeliveryCourierName of the delivery courier.
	DeliveryCourierName string `json:"delivery_courier_name"`
	// DeliveryCourierPhone of the delivery courier.
	DeliveryCourierPhone string `json:"delivery_courier_phone"`
}

// UpdateSingleWaybillStatus updates a 单号 (single waybill) order with courier
// info.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_scatterupdateorderstatus.html
func (w *MiniProgram) UpdateSingleWaybillStatus(ctx context.Context, req *UpdateSingleWaybillStatusRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/single_waybill/update", nil, defaultReqOptions(), req, nil)
}

// UpdateSingleWaybillFeeRequest updates the fee of a single waybill.
type UpdateSingleWaybillFeeRequest struct {
	// Token pushed by the order event.
	Token string `json:"token"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
	// NeedPay: 0 no online payment, 1 needs payment, 2 needs payment (支付分).
	NeedPay int `json:"need_pay"`
	// Fee to pay in cents.
	Fee int `json:"fee"`
	// OriginalFee (base + insured + other) in cents.
	OriginalFee int `json:"original_fee"`
	// BaseFee of the freight in cents.
	BaseFee int `json:"base_fee"`
	// InsuredFee of the insurance in cents.
	InsuredFee int `json:"insured_fee,omitempty"`
	// OtherFee in cents.
	OtherFee int `json:"other_fee,omitempty"`
	// Remark of the other fee.
	Remark string `json:"remark,omitempty"`
	// PayGoodsName shown in the WeChat Pay goods detail.
	PayGoodsName string `json:"pay_goods_name,omitempty"`
}

// UpdateSingleWaybillFee reports the payable fee of a single waybill.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_scatterupdateorderfee.html
func (w *MiniProgram) UpdateSingleWaybillFee(ctx context.Context, req *UpdateSingleWaybillFeeRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/single_waybill/fee", nil, defaultReqOptions(), req, nil)
}

// CancelSingleWaybillRequest cancels a single waybill order.
type CancelSingleWaybillRequest struct {
	// Token pushed by the order event.
	Token string `json:"token"`
	// Reason of the cancellation.
	Reason string `json:"reason"`
}

// CancelSingleWaybill cancels a single-waybill order from the provider side.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_scatterdeliverycancel.html
func (w *MiniProgram) CancelSingleWaybill(ctx context.Context, token, reason string) error {
	req := &CancelSingleWaybillRequest{Token: token, Reason: reason}
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/single_waybill/cancel_order", nil, defaultReqOptions(), req, nil)
}

// RefundSingleWaybillRequest refunds a single waybill order.
type RefundSingleWaybillRequest struct {
	// Token of the original order event.
	Token string `json:"token"`
	// Fee to refund in cents.
	Fee int `json:"fee"`
}

// RefundSingleWaybill refunds the fee of a single-waybill order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_scatterdeliveryrefund.html
func (w *MiniProgram) RefundSingleWaybill(ctx context.Context, token string, fee int) error {
	req := &RefundSingleWaybillRequest{Token: token, Fee: fee}
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/single_waybill/refund_order", nil, defaultReqOptions(), req, nil)
}

// GetSingleWaybillBillRequest downloads the single-waybill reconciliation
// bill.
type GetSingleWaybillBillRequest struct {
	// Date of the bill in YYYYMMDD.
	Date string `json:"date"`
	// Type of the bill: ALL, SUCCESS, REFUND.
	Type string `json:"type"`
}

// GetSingleWaybillBillResponse is returned by GetSingleWaybillBill.
type GetSingleWaybillBillResponse struct {
	ErrResponse
	// RawData of the bill payload.
	RawData map[string]any `json:"-"`
}

// GetSingleWaybillBill downloads the provider's reconciliation bill.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_scatter_get_bill.html
func (w *MiniProgram) GetSingleWaybillBill(ctx context.Context, req *GetSingleWaybillBillRequest) (*GetSingleWaybillBillResponse, error) {
	var result GetSingleWaybillBillResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/single_waybill/get_bill", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateComplaintResultRequest answers a complaint about a single waybill.
type UpdateComplaintResultRequest struct {
	// Token pushed by the order event.
	Token string `json:"token"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
	// Result of the complaint handling.
	Result string `json:"result"`
	// Desc of the handling result.
	Desc string `json:"desc"`
}

// UpdateComplaintResult reports the complaint handling result of a waybill.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_scatterupdatecomplainresult.html
func (w *MiniProgram) UpdateComplaintResult(ctx context.Context, req *UpdateComplaintResultRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/scatter/update_complaint_result", nil, defaultReqOptions(), req, nil)
}

// GetDeliveryContactRequest fetches the contact of a delivery order.
type GetDeliveryContactRequest struct {
	// Token pushed by the order event.
	Token string `json:"token"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
}

// GetDeliveryContactResponse is returned by GetDeliveryContact.
type GetDeliveryContactResponse struct {
	ErrResponse
	// RawData of the contact payload.
	RawData map[string]any `json:"-"`
}

// GetDeliveryContact returns the recipient/sender contact of a delivery order
// (provider-side).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_getcontact.html
func (w *MiniProgram) GetDeliveryContact(ctx context.Context, req *GetDeliveryContactRequest) (*GetDeliveryContactResponse, error) {
	var result GetDeliveryContactResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/contact/get", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PreviewWaybillTemplateRequest previews a waybill print template.
type PreviewWaybillTemplateRequest struct {
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
	// WaybillTemplate HTML content (base64 encoded).
	WaybillTemplate string `json:"waybill_template"`
	// WaybillData JSON data of the waybill.
	WaybillData string `json:"waybill_data"`
	// Custom is the original merchant addOrder request payload.
	Custom map[string]any `json:"custom"`
}

// PreviewWaybillTemplate previews the provider's waybill template rendering.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_previewtemplate.html
func (w *MiniProgram) PreviewWaybillTemplate(ctx context.Context, req *PreviewWaybillTemplateRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/template/preview", nil, defaultReqOptions(), req, nil)
}

// UpdateDeliveryBusinessRequest answers a merchant binding audit.
type UpdateDeliveryBusinessRequest struct {
	// ShopAppID of the merchant Mini Program.
	ShopAppID string `json:"shop_app_id"`
	// BizID of the merchant account.
	BizID string `json:"biz_id"`
	// ResultCode: 0 approved, other rejected.
	ResultCode int `json:"result_code"`
	// ResultMsg reason when rejected.
	ResultMsg string `json:"result_msg,omitempty"`
}

// UpdateDeliveryBusiness replies to the merchant account binding review.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-provider/api_updatebusiness.html
func (w *MiniProgram) UpdateDeliveryBusiness(ctx context.Context, req *UpdateDeliveryBusinessRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/service/business/update", nil, defaultReqOptions(), req, nil)
}

// NoWorryReturnRequest is the shared body of the 无忧退货 (no-worry return)
// APIs.
type NoWorryReturnRequest struct {
	// AppID of the merchant Mini Program.
	AppID string `json:"appid,omitempty"`
	// RawData extra fields.
	RawData map[string]any `json:"-"`
}

// BindNoWorryReturn binds the no-worry-return capability.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-return/api_addreturnid.html
func (w *MiniProgram) BindNoWorryReturn(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/no_worry_return/add", nil, defaultReqOptions(), body, nil)
}

// QueryNoWorryReturn queries the no-worry-return status.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-return/api_getreturnid.html
func (w *MiniProgram) QueryNoWorryReturn(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/no_worry_return/get", nil, defaultReqOptions(), body, nil)
}

// UnbindNoWorryReturn unbinds the no-worry-return capability.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/express-return/api_unbindreturnid.html
func (w *MiniProgram) UnbindNoWorryReturn(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/no_worry_return/unbind", nil, defaultReqOptions(), body, nil)
}
