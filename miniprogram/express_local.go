package miniprogram

import (
	"context"
)

// ImmeDeliveryCompany is one local delivery provider supported by WeChat.
type ImmeDeliveryCompany struct {
	// DeliveryID of the provider (e.g. "BDB", "SFTC").
	DeliveryID string `json:"delivery_id"`
	// DeliveryName of the provider.
	DeliveryName string `json:"delivery_name"`
}

// GetAllImmeDeliveryResponse is returned by GetAllImmeDelivery.
type GetAllImmeDeliveryResponse struct {
	ErrResponse
	// List of the supported providers.
	List []ImmeDeliveryCompany `json:"list"`
}

// GetAllImmeDelivery lists the same-city delivery providers WeChat integrates
// with.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_getallimmedelivery.html
func (w *MiniProgram) GetAllImmeDelivery(ctx context.Context) (*GetAllImmeDeliveryResponse, error) {
	var result GetAllImmeDeliveryResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/delivery/getall", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OpenDeliveryRequest applies to open the same-city delivery capability.
type OpenDeliveryRequest struct {
	// DeliveryID of the chosen provider.
	DeliveryID string `json:"delivery_id"`
	// CallbackURL receiving the delivery events (public https URL).
	CallbackURL string `json:"callback_url"`
	// AppKey obtained from the provider after review.
	AppKey string `json:"appkey,omitempty"`
	// AppSecret of the provider account.
	AppSecret string `json:"appsecret,omitempty"`
}

// OpenDelivery enables the same-city delivery service with a provider. The
// Mini Program must have completed the provider-side review first.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_opendelivery.html
func (w *MiniProgram) OpenDelivery(ctx context.Context, req *OpenDeliveryRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/open", nil, defaultReqOptions(), req, nil)
}

// BindImmeDeliveryAccountRequest binds the merchant account of a provider.
type BindImmeDeliveryAccountRequest struct {
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// AppKey obtained from the provider.
	AppKey string `json:"appkey"`
	// AppSecret of the provider account.
	AppSecret string `json:"appsecret"`
}

// BindImmeDeliveryAccount binds the merchant's account with a same-city
// delivery provider.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_bindlocalaccount.html
func (w *MiniProgram) BindImmeDeliveryAccount(ctx context.Context, req *BindImmeDeliveryAccountRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/shop/add", nil, defaultReqOptions(), req, nil)
}

// GetImmeDeliveryBindAccountResponse is returned by
// GetImmeDeliveryBindAccount.
type GetImmeDeliveryBindAccountResponse struct {
	ErrResponse
	// Shop of the current provider binding.
	Shop struct {
		// DeliveryID of the provider.
		DeliveryID string `json:"delivery_id,omitempty"`
		// DeliveryName of the provider.
		DeliveryName string `json:"delivery_name,omitempty"`
		// AppKey bound.
		AppKey string `json:"appkey,omitempty"`
		// AppSecret bound (masked).
		AppSecret string `json:"appsecret,omitempty"`
	} `json:"shop"`
}

// GetImmeDeliveryBindAccount returns the currently bound same-city delivery
// provider account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_getbindaccount.html
func (w *MiniProgram) GetImmeDeliveryBindAccount(ctx context.Context) (*GetImmeDeliveryBindAccountResponse, error) {
	var result GetImmeDeliveryBindAccountResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/shop/get", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImmeDeliveryReceiver is the receiver of a same-city delivery order.
type ImmeDeliveryReceiver struct {
	// Name of the receiver.
	Name string `json:"name"`
	// Address of the receiver.
	Address string `json:"address"`
	// AddressDetail of the receiver.
	AddressDetail string `json:"address_detail"`
	// Lng of the receiver location (longitude * 10^6).
	Lng int64 `json:"lng,omitempty"`
	// Lat of the receiver location (latitude * 10^6).
	Lat int64 `json:"lat,omitempty"`
	// City of the receiver (must match the sender city).
	City string `json:"city"`
	// Phone of the receiver.
	Phone string `json:"phone"`
}

// ImmeDeliverySender is the sender of a same-city delivery order.
type ImmeDeliverySender struct {
	// Name of the sender.
	Name string `json:"name"`
	// Address of the sender.
	Address string `json:"address"`
	// AddressDetail of the sender.
	AddressDetail string `json:"address_detail"`
	// Lng of the sender location (longitude * 10^6).
	Lng int64 `json:"lng"`
	// Lat of the sender location (latitude * 10^6).
	Lat int64 `json:"lat"`
	// City of the sender.
	City string `json:"city"`
	// Phone of the sender.
	Phone string `json:"phone"`
}

// ImmeDeliveryCargo describes the goods and fees of the delivery.
type ImmeDeliveryCargo struct {
	// GoodsValue is the declared goods value in cents.
	GoodsValue int64 `json:"goods_value,omitempty"`
	// GoodsWeight is the goods weight in grams.
	GoodsWeight int64 `json:"goods_weight,omitempty"`
	// GoodsPrice is the price declared to the rider in cents.
	GoodsPrice int64 `json:"goods_price,omitempty"`
	// GoodsDetail textual description.
	GoodsDetail string `json:"goods_detail,omitempty"`
	// GoodsPickupInfo hint for the rider.
	GoodsPickupInfo string `json:"goods_pickup_info,omitempty"`
	// GoodsDeliveryInfo hint for the rider.
	GoodsDeliveryInfo string `json:"goods_delivery_info,omitempty"`
	// CargoFirstClass describes the cargo class, e.g. "生鲜".
	CargoFirstClass string `json:"cargo_first_class,omitempty"`
	// CargoSecondClass describes the cargo sub class.
	CargoSecondClass string `json:"cargo_second_class,omitempty"`
}

// ImmeDeliveryOrder is the shared base of the same-city delivery order
// payloads.
type ImmeDeliveryOrder struct {
	// ShopOrderID is the merchant order id.
	ShopOrderID string `json:"shop_order_id,omitempty"`
	// ShopNO is the merchant shop number.
	ShopNO string `json:"shop_no,omitempty"`
	// DeliverySign is the delivery signature (per provider contract).
	DeliverySign string `json:"delivery_sign,omitempty"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id,omitempty"`
	// OpenID of the ordering user.
	OpenID string `json:"openid,omitempty"`
	// ExpectedDeliveryTime of the order in "YYYY-MM-DD HH:mm" (unix seconds
	// for some providers, see docs).
	ExpectedDeliveryTime int64 `json:"expected_delivery_time,omitempty"`
	// ExpectedPickupTime of the order.
	ExpectedPickupTime int64 `json:"expected_pickup_time,omitempty"`
	// OrderPrice the user paid in cents.
	OrderPrice int64 `json:"order_price,omitempty"`
	// OrderTime of the order (Unix seconds).
	OrderTime int64 `json:"order_time,omitempty"`
	// IsInsured enables the insurance when 1.
	IsInsured int `json:"is_insured,omitempty"`
	// IsDirectDelivery is 1 when the order is direct without shop pickup.
	IsDirectDelivery int `json:"is_direct_delivery,omitempty"`
	// Cargo of the order.
	Cargo *ImmeDeliveryCargo `json:"cargo,omitempty"`
	// Receiver of the order.
	Receiver *ImmeDeliveryReceiver `json:"receiver,omitempty"`
	// Sender of the order (required for same-city handover orders).
	Sender *ImmeDeliverySender `json:"sender,omitempty"`
	// Tips for the rider in cents (added via AddImmeDeliveryTip).
	Tips int64 `json:"tips,omitempty"`
}

// ImmeDeliveryFee is the fee breakdown of a pre-created order.
type ImmeDeliveryFee struct {
	// Price of the delivery in cents.
	Price int64 `json:"price,omitempty"`
	// DeliveryFee for the rider in cents.
	DeliveryFee int64 `json:"delivery_fee,omitempty"`
	// InsuranceFee in cents.
	InsuranceFee int64 `json:"insurance_fee,omitempty"`
	// Distance of the delivery in meters.
	Distance int64 `json:"distance,omitempty"`
	// WaitingFee in cents.
	WaitingFee int64 `json:"waiting_fee,omitempty"`
	// Tip in cents.
	Tip int64 `json:"tip,omitempty"`
	// CouponDiscount in cents.
	CouponDiscount int64 `json:"coupon_discount,omitempty"`
}

// PreAddImmeDeliveryOrderRequest is the payload of
// PreAddImmeDeliveryOrder. It quotes the delivery and returns the delivery
// token needed to confirm the order.
type PreAddImmeDeliveryOrderRequest struct {
	ImmeDeliveryOrder
}

// PreAddImmeDeliveryOrderResponse is returned by PreAddImmeDeliveryOrder.
type PreAddImmeDeliveryOrderResponse struct {
	ErrResponse
	// Fee of the quoted delivery.
	Fee ImmeDeliveryFee `json:"fee"`
	// DeliveryToken must be passed when confirming the order.
	DeliveryToken string `json:"delivery_token"`
}

// PreAddImmeDeliveryOrder quotes a same-city delivery before actually placing
// the order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_preaddorder.html
func (w *MiniProgram) PreAddImmeDeliveryOrder(ctx context.Context, req *PreAddImmeDeliveryOrderRequest) (*PreAddImmeDeliveryOrderResponse, error) {
	var result PreAddImmeDeliveryOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/pre_add", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddImmeDeliveryOrderRequest is the payload of AddImmeDeliveryOrder.
type AddImmeDeliveryOrderRequest struct {
	ImmeDeliveryOrder
	// DeliveryToken returned by the pre-add quote.
	DeliveryToken string `json:"delivery_token,omitempty"`
}

// AddImmeDeliveryOrderResponse is returned by AddImmeDeliveryOrder.
type AddImmeDeliveryOrderResponse struct {
	ErrResponse
	// Fee of the created order.
	Fee ImmeDeliveryFee `json:"fee"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id,omitempty"`
	// WaybillID of the created order.
	WaybillID string `json:"waybill_id,omitempty"`
	// OrderStatus of the created order.
	OrderStatus int `json:"order_status,omitempty"`
	// WaybillData printable content.
	WaybillData []ExpressWaybillData `json:"waybill_data,omitempty"`
}

// AddImmeDeliveryOrder places a same-city delivery order confirmed after a
// pre-add quote.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_addlocalorder.html
func (w *MiniProgram) AddImmeDeliveryOrder(ctx context.Context, req *AddImmeDeliveryOrderRequest) (*AddImmeDeliveryOrderResponse, error) {
	var result AddImmeDeliveryOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/add", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetImmeDeliveryOrderRequest selects a same-city order to query.
type GetImmeDeliveryOrderRequest struct {
	// OrderID is the merchant order id.
	OrderID string `json:"order_id"`
	// OpenID of the ordering user.
	OpenID string `json:"openid"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id,omitempty"`
	// ShopOrderID of the merchant.
	ShopOrderID string `json:"shop_order_id,omitempty"`
	// ShopNO of the merchant shop.
	ShopNO string `json:"shop_no,omitempty"`
}

// GetImmeDeliveryOrderResponse is returned by GetImmeDeliveryOrder.
type GetImmeDeliveryOrderResponse struct {
	ErrResponse
	// OrderStatus of the order.
	OrderStatus int `json:"order_status,omitempty"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id,omitempty"`
	// RiderInfo of the assigned rider.
	RiderInfo struct {
		// Name of the rider.
		Name string `json:"name,omitempty"`
		// Phone of the rider.
		Phone string `json:"phone,omitempty"`
	} `json:"rider_info,omitempty"`
	// Fee of the order.
	Fee ImmeDeliveryFee `json:"fee,omitempty"`
	// ExpectedDeliveryTime of the order.
	ExpectedDeliveryTime int64 `json:"expected_delivery_time,omitempty"`
	// DeliveryToken for follow-up operations.
	DeliveryToken string `json:"delivery_token,omitempty"`
}

// GetImmeDeliveryOrder queries the state of a same-city delivery order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_getlocalorder.html
func (w *MiniProgram) GetImmeDeliveryOrder(ctx context.Context, req *GetImmeDeliveryOrderRequest) (*GetImmeDeliveryOrderResponse, error) {
	var result GetImmeDeliveryOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/get", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelImmeDeliveryOrderRequest cancels a same-city delivery order.
type CancelImmeDeliveryOrderRequest struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
	// CancelReasonId of the cancellation (from the provider).
	CancelReasonID int `json:"cancel_reason_id,omitempty"`
	// CancelReason text.
	CancelReason string `json:"cancel_reason,omitempty"`
}

// CancelImmeDeliveryOrderResponse is returned by CancelImmeDeliveryOrder.
type CancelImmeDeliveryOrderResponse struct {
	ErrResponse
	// DeductFee charged for the cancellation in cents.
	DeductFee int64 `json:"deduct_fee,omitempty"`
	// Desc of the deduction.
	Desc string `json:"desc,omitempty"`
}

// CancelImmeDeliveryOrder cancels a same-city order (a fee may be deducted
// depending on the provider policy).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_cancellocalorder.html
func (w *MiniProgram) CancelImmeDeliveryOrder(ctx context.Context, req *CancelImmeDeliveryOrderRequest) (*CancelImmeDeliveryOrderResponse, error) {
	var result CancelImmeDeliveryOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/cancel", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddImmeDeliveryTipRequest adds a tip to a same-city order.
type AddImmeDeliveryTipRequest struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
	// Tips amount in cents.
	Tips int64 `json:"tips"`
	// Remark of the tip (max 30 chars).
	Remark string `json:"remark,omitempty"`
}

// AddImmeDeliveryTipResponse is returned by AddImmeDeliveryTip.
type AddImmeDeliveryTipResponse struct {
	ErrResponse
	// Fee of the order after the tip.
	Fee ImmeDeliveryFee `json:"fee"`
}

// AddImmeDeliveryTip adds an extra tip for the rider of a same-city order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_addtips.html
func (w *MiniProgram) AddImmeDeliveryTip(ctx context.Context, req *AddImmeDeliveryTipRequest) (*AddImmeDeliveryTipResponse, error) {
	var result AddImmeDeliveryTipResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/addtips", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PreCancelImmeDeliveryOrderRequest previews the cancellation cost of a
// same-city order.
type PreCancelImmeDeliveryOrderRequest struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
	// CancelReasonId of the cancellation.
	CancelReasonID int `json:"cancel_reason_id,omitempty"`
	// CancelReason text.
	CancelReason string `json:"cancel_reason,omitempty"`
}

// PreCancelImmeDeliveryOrderResponse is returned by
// PreCancelImmeDeliveryOrder.
type PreCancelImmeDeliveryOrderResponse struct {
	ErrResponse
	// DeductFee that would be charged.
	DeductFee int64 `json:"deduct_fee"`
	// Desc of the deduction.
	Desc string `json:"desc"`
}

// PreCancelImmeDeliveryOrder previews the cancellation cost before actually
// cancelling.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_precancelorder.html
func (w *MiniProgram) PreCancelImmeDeliveryOrder(ctx context.Context, req *PreCancelImmeDeliveryOrderRequest) (*PreCancelImmeDeliveryOrderResponse, error) {
	var result PreCancelImmeDeliveryOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/precancel", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AbnormalImmeDeliveryOrderRequest reports an abnormal delivery (goods
// damaged / rider in trouble) to the provider.
type AbnormalImmeDeliveryOrderRequest struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
	// AbnormalReason of the report (1 = no rider assigned in time, 2 = rider
	// delay, 3 = other).
	AbnormalReason int `json:"abnormal_reason"`
	// AbnormalDesc textual detail (max 100 chars).
	AbnormalDesc string `json:"abnormal_desc,omitempty"`
	// AbnormalTime when the problem happened (Unix seconds, or the provider
	// value).
	AbnormalTime int64 `json:"abnormal_time,omitempty"`
}

// AbnormalImmeDeliveryOrderResponse is returned by
// AbnormalImmeDeliveryOrder.
type AbnormalImmeDeliveryOrderResponse struct {
	ErrResponse
	// DeductFee charged when the merchant is at fault.
	DeductFee int64 `json:"deduct_fee,omitempty"`
	// Desc of the handling.
	Desc string `json:"desc,omitempty"`
}

// AbnormalImmeDeliveryOrder reports an abnormal same-city delivery order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_abnormalconfirm.html
func (w *MiniProgram) AbnormalImmeDeliveryOrder(ctx context.Context, req *AbnormalImmeDeliveryOrderRequest) (*AbnormalImmeDeliveryOrderResponse, error) {
	var result AbnormalImmeDeliveryOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/confirm_return", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReAddImmeDeliveryOrderRequest re-dispatches a cancelled order.
type ReAddImmeDeliveryOrderRequest struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the original order.
	WaybillID string `json:"waybill_id"`
	// ShopOrderID of the merchant.
	ShopOrderID string `json:"shop_order_id"`
	// ShopNO of the shop.
	ShopNO string `json:"shop_no"`
	// ExpectedDeliveryTime of the re-dispatched order.
	ExpectedDeliveryTime int64 `json:"expected_delivery_time,omitempty"`
}

// ReAddImmeDeliveryOrderResponse is returned by ReAddImmeDeliveryOrder.
type ReAddImmeDeliveryOrderResponse struct {
	ErrResponse
	// Result of the re-dispatch: 0 success, 1 already dispatched.
	Result int `json:"result,omitempty"`
	// Fee of the order.
	Fee ImmeDeliveryFee `json:"fee,omitempty"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id,omitempty"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id,omitempty"`
	// OrderStatus of the order.
	OrderStatus int `json:"order_status,omitempty"`
}

// ReAddImmeDeliveryOrder re-dispatches a cancelled same-city delivery order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_reorder.html
func (w *MiniProgram) ReAddImmeDeliveryOrder(ctx context.Context, req *ReAddImmeDeliveryOrderRequest) (*ReAddImmeDeliveryOrderResponse, error) {
	var result ReAddImmeDeliveryOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/order/readd", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MockImmeDeliveryOrderRequest simulates an order event during development.
type MockImmeDeliveryOrderRequest struct {
	// OrderID of the merchant (must be a mock order).
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// MockType of the simulated action: 1 = order accepted, 2 = rider en
	// route to shop, 3 = rider arrived, 4 = picked up, 5 = delivered.
	MockType int `json:"mock_type"`
	// ActionTime of the simulated event.
	ActionTime int64 `json:"action_time,omitempty"`
}

// MockImmeDeliveryOrder pushes a simulated delivery event to a test order
// (development only).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_mockupdateorder.html
func (w *MiniProgram) MockImmeDeliveryOrder(ctx context.Context, req *MockImmeDeliveryOrderRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/test_update_order", nil, defaultReqOptions(), req, nil)
}
