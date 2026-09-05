package miniprogram

import (
	"context"
)

// ============================================================
// 同城配送 (WeChat Express same-city distribution).
// Routes live under /cgi-bin/express/intracity/*. The merchant first applies,
// creates a store, then quotes (pre_add) and places (add) delivery orders.
// ============================================================

// IntracityDeliveryProvider is a supported same-city delivery service.
type IntracityDeliveryProvider struct {
	// DeliveryID of the provider.
	DeliveryID string `json:"delivery_id"`
	// DeliveryName of the provider.
	DeliveryName string `json:"delivery_name"`
}

// IntracityGetCityResponse returns the cities supported for same-city
// delivery by a provider.
type IntracityGetCityResponse struct {
	ErrResponse
	// CityList of the supported cities.
	CityList []map[string]any `json:"city_list,omitempty"`
}

// IntracityGetCity lists the cities reachable by the same-city distribution
// providers.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_getcity.html
func (w *MiniProgram) IntracityGetCity(ctx context.Context) (*IntracityGetCityResponse, error) {
	var result IntracityGetCityResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/getcity", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracityApply opens the same-city distribution capability for the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_apply.html
func (w *MiniProgram) IntracityApply(ctx context.Context) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/apply", nil, defaultReqOptions(), map[string]string{}, nil)
}

// IntracityStoreAddress is the address payload of an intracity store.
type IntracityStoreAddress struct {
	// Province of the store.
	Province string `json:"province"`
	// City of the store.
	City string `json:"city"`
	// District of the store.
	District string `json:"area"`
	// Street of the store.
	Street string `json:"street"`
	// Address (house number) of the store.
	Address string `json:"house"`
	// Lng of the store location.
	Lng float64 `json:"lng"`
	// Lat of the store location.
	Lat float64 `json:"lat"`
	// Phone of the store.
	Phone string `json:"phone"`
}

// CreateIntracityStoreRequest creates a same-city distribution store.
type CreateIntracityStoreRequest struct {
	// OutStoreID of the merchant side.
	OutStoreID string `json:"out_store_id"`
	// StoreName of the store.
	StoreName string `json:"store_name"`
	// OrderPatternType: 1 预约单, 2 即时单, 3 预约+即时 (optional).
	OrderPatternType int `json:"order_pattern,omitempty"`
	// ServiceTransIDPrefer preferred service provider.
	ServiceTransIDPrefer string `json:"service_trans_prefer,omitempty"`
	// AddressInfo of the store.
	AddressInfo *IntracityStoreAddress `json:"address_info,omitempty"`
}

// CreateIntracityStore registers a store for the same-city distribution.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_createstore.html
func (w *MiniProgram) CreateIntracityStore(ctx context.Context, req *CreateIntracityStoreRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/createstore", nil, defaultReqOptions(), req, nil)
}

// QueryIntracityStoreRequest selects a store to query.
type QueryIntracityStoreRequest struct {
	// OutStoreID of the merchant side.
	OutStoreID string `json:"out_store_id,omitempty"`
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id,omitempty"`
}

// QueryIntracityStoreResponse is returned by QueryIntracityStore.
type QueryIntracityStoreResponse struct {
	ErrResponse
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id,omitempty"`
	// StoreName of the store.
	StoreName string `json:"store_name,omitempty"`
	// ServiceTransID of the bound provider.
	ServiceTransID string `json:"service_trans_id,omitempty"`
	// StoreStatus of the store.
	StoreStatus int `json:"store_status,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// QueryIntracityStore returns the state of one same-city store.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_querystore.html
func (w *MiniProgram) QueryIntracityStore(ctx context.Context, req *QueryIntracityStoreRequest) (*QueryIntracityStoreResponse, error) {
	var result QueryIntracityStoreResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/querystore", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateIntracityStoreRequest updates the store info.
type UpdateIntracityStoreRequest struct {
	// StoreID of the WeChat side (required).
	StoreID string `json:"wx_store_id"`
	// StoreName of the store.
	StoreName string `json:"store_name,omitempty"`
	// AddressInfo of the store.
	AddressInfo *IntracityStoreAddress `json:"address_info,omitempty"`
	// OrderPatternType (1 预约, 2 即时, 3 both).
	OrderPatternType int `json:"order_pattern,omitempty"`
	// ServiceTransIDPrefer preferred provider.
	ServiceTransIDPrefer string `json:"service_trans_prefer,omitempty"`
}

// UpdateIntracityStore updates an existing same-city store.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_updatestore.html
func (w *MiniProgram) UpdateIntracityStore(ctx context.Context, req *UpdateIntracityStoreRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/updatestore", nil, defaultReqOptions(), req, nil)
}

// StoreBalanceRequest queries the balance of a store.
type StoreBalanceRequest struct {
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id"`
	// ServiceTransID of the provider.
	ServiceTransID string `json:"service_trans_id"`
}

// StoreBalanceResponse is returned by IntracityBalanceQuery.
type StoreBalanceResponse struct {
	ErrResponse
	// Balance in cents.
	Balance int64 `json:"balance,omitempty"`
	// AvailableBalance in cents.
	AvailableBalance int64 `json:"available_balance,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// IntracityBalanceQuery returns the balance of a store account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_balancequery.html
func (w *MiniProgram) IntracityBalanceQuery(ctx context.Context, req *StoreBalanceRequest) (*StoreBalanceResponse, error) {
	var result StoreBalanceResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/balancequery", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StoreChargeRequest recharges a store account.
type StoreChargeRequest struct {
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id"`
	// Amount to charge in cents.
	Amount int64 `json:"amount"`
	// OutBillNo of the merchant recharge order.
	OutBillNo string `json:"out_bill_no,omitempty"`
}

// StoreChargeResponse is returned by IntracityStoreCharge.
type StoreChargeResponse struct {
	ErrResponse
	// RawData of the result.
	RawData map[string]any `json:"-"`
}

// IntracityStoreCharge recharges (充值) the store distribution account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_storecharge.html
func (w *MiniProgram) IntracityStoreCharge(ctx context.Context, req *StoreChargeRequest) (*StoreChargeResponse, error) {
	var result StoreChargeResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/storecharge", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StoreRefundRequest refunds a store recharge.
type StoreRefundRequest struct {
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id"`
	// RefundAmount in cents.
	RefundAmount int64 `json:"refund_amount"`
	// OutRefundNo of the merchant refund order.
	OutRefundNo string `json:"out_refund_no,omitempty"`
}

// StoreRefundResponse is returned by IntracityStoreRefund.
type StoreRefundResponse struct {
	ErrResponse
	// RawData of the result.
	RawData map[string]any `json:"-"`
}

// IntracityStoreRefund refunds money from the store distribution account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_store_refund.html
func (w *MiniProgram) IntracityStoreRefund(ctx context.Context, req *StoreRefundRequest) (*StoreRefundResponse, error) {
	var result StoreRefundResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/storerefund", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracityCargo describes the goods of an order.
type IntracityCargo struct {
	// CargoName of the goods.
	CargoName string `json:"cargo_name"`
	// CargoType of the goods (per provider dictionary).
	CargoType int `json:"cargo_type"`
	// CargoWeight in grams.
	CargoWeight int `json:"cargo_weight"`
	// CargoPrice declared in cents.
	CargoPrice int `json:"cargo_price"`
	// CargoCount of packages.
	CargoCount int `json:"cargo_num"`
}

// IntracityPreAddOrderRequest quotes a same-city delivery order.
type IntracityPreAddOrderRequest struct {
	// StoreID of the WeChat side (from CreateIntracityStore).
	StoreID string `json:"wx_store_id"`
	// UserName of the recipient.
	UserName string `json:"user_name"`
	// UserPhone of the recipient.
	UserPhone string `json:"user_phone"`
	// UserLng of the recipient location.
	UserLng float64 `json:"user_lng"`
	// UserLat of the recipient location.
	UserLat float64 `json:"user_lat"`
	// UserAddress of the recipient.
	UserAddress string `json:"user_address"`
	// Cargo of the order.
	Cargo *IntracityCargo `json:"cargo"`
	// UseSandbox runs the quote in the sandbox environment.
	UseSandbox bool `json:"use_sandbox,omitempty"`
}

// IntracityQuote is the fee payload of a pre-created order.
type IntracityQuote struct {
	// Fee in cents.
	Fee int64 `json:"fee"`
	// Distance in meters.
	Distance int64 `json:"distance"`
	// DeliveryFee in cents.
	DeliveryFee int64 `json:"delivery_fee,omitempty"`
	// CourierFee in cents.
	CourierFee int64 `json:"courier_fee,omitempty"`
}

// IntracityPreAddOrderResponse is returned by IntracityPreAddOrder.
type IntracityPreAddOrderResponse struct {
	ErrResponse
	// Fee of the quoted order.
	Fee IntracityQuote `json:"fee"`
	// StoreOrderID of the merchant order to confirm later.
	StoreOrderID string `json:"store_order_id,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// IntracityPreAddOrder quotes a same-city delivery before placing it.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_preaddorder.html
func (w *MiniProgram) IntracityPreAddOrder(ctx context.Context, req *IntracityPreAddOrderRequest) (*IntracityPreAddOrderResponse, error) {
	var result IntracityPreAddOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/preaddorder", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracityAddOrderRequest places a same-city delivery order after the quote.
type IntracityAddOrderRequest struct {
	IntracityPreAddOrderRequest
	// StoreOrderID returned by IntracityPreAddOrder (required).
	StoreOrderID string `json:"store_order_id"`
}

// IntracityAddOrderResponse is returned by IntracityAddOrder.
type IntracityAddOrderResponse struct {
	ErrResponse
	// StoreOrderID of the merchant.
	StoreOrderID string `json:"store_order_id,omitempty"`
	// OrderID of the WeChat side.
	OrderID string `json:"wx_order_id,omitempty"`
	// Status of the order.
	Status int `json:"status,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// IntracityAddOrder places the same-city delivery order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_addorder.html
func (w *MiniProgram) IntracityAddOrder(ctx context.Context, req *IntracityAddOrderRequest) (*IntracityAddOrderResponse, error) {
	var result IntracityAddOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/addorder", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracityOrderSelector selects a same-city order to query/cancel.
type IntracityOrderSelector struct {
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id,omitempty"`
	// StoreOrderID of the merchant order.
	StoreOrderID string `json:"store_order_id,omitempty"`
	// OrderID of the WeChat side.
	OrderID string `json:"wx_order_id,omitempty"`
}

// QueryIntracityOrderResponse is returned by QueryIntracityOrder.
type QueryIntracityOrderResponse struct {
	ErrResponse
	// OrderState of the order.
	OrderState int `json:"order_state,omitempty"`
	// DeliveryName of the provider.
	DeliveryName string `json:"delivery_name,omitempty"`
	// OrderID of the WeChat side.
	OrderID string `json:"wx_order_id,omitempty"`
	// CourierInfo of the assigned courier.
	CourierInfo map[string]any `json:"courier_info,omitempty"`
	// Fee of the order.
	Fee IntracityQuote `json:"fee"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// QueryIntracityOrder returns the state of a same-city delivery order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_queryorder.html
func (w *MiniProgram) QueryIntracityOrder(ctx context.Context, req *IntracityOrderSelector) (*QueryIntracityOrderResponse, error) {
	var result QueryIntracityOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/queryorder", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracityCancelOrderRequest cancels a same-city delivery order.
type IntracityCancelOrderRequest struct {
	IntracityOrderSelector
	// CancelReason of the cancellation.
	CancelReason string `json:"cancel_reason,omitempty"`
}

// IntracityCancelOrderResponse is returned by IntracityCancelOrder.
type IntracityCancelOrderResponse struct {
	ErrResponse
	// DeductFee charged for the cancellation in cents.
	DeductFee int64 `json:"deduct_fee,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// IntracityCancelOrder cancels a same-city delivery order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_cancelorder.html
func (w *MiniProgram) IntracityCancelOrder(ctx context.Context, req *IntracityCancelOrderRequest) (*IntracityCancelOrderResponse, error) {
	var result IntracityCancelOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/cancelorder", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracityQueryFlowRequest queries the courier flow of a store window.
type IntracityQueryFlowRequest struct {
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id,omitempty"`
	// BeginTime of the window (unix seconds).
	BeginTime int64 `json:"begin_time,omitempty"`
	// EndTime of the window (unix seconds).
	EndTime int64 `json:"end_time,omitempty"`
}

// IntracityQueryFlowResponse is returned by IntracityQueryFlow.
type IntracityQueryFlowResponse struct {
	ErrResponse
	// FlowList of the account movements.
	FlowList []map[string]any `json:"flow_list,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// IntracityQueryFlow returns the balance flow of a store window.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_queryflow.html
func (w *MiniProgram) IntracityQueryFlow(ctx context.Context, req *IntracityQueryFlowRequest) (*IntracityQueryFlowResponse, error) {
	var result IntracityQueryFlowResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/queryflow", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracityMockNotifyRequest simulates a provider callback (development).
type IntracityMockNotifyRequest struct {
	// StoreID of the WeChat side.
	StoreID string `json:"wx_store_id"`
	// StoreOrderID of the merchant order.
	StoreOrderID string `json:"store_order_id,omitempty"`
	// OrderID of the WeChat side.
	OrderID string `json:"wx_order_id,omitempty"`
	// MockType of the simulated callback.
	MockType int `json:"mock_type,omitempty"`
}

// IntracityMockNotify triggers a mock provider notification for test orders.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_mocknotify.html
func (w *MiniProgram) IntracityMockNotify(ctx context.Context, req *IntracityMockNotifyRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/mocknotify", nil, defaultReqOptions(), req, nil)
}

// IntracityGetPayModeResponse is returned by IntracityGetPayMode.
type IntracityGetPayModeResponse struct {
	ErrResponse
	// PayModeList of the supported payment modes.
	PayModeList []map[string]any `json:"pay_mode_list,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// IntracityGetPayMode returns the available payment modes (支付方式) of the
// same-city distribution capability.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_getpaymode.html
func (w *MiniProgram) IntracityGetPayMode(ctx context.Context) (*IntracityGetPayModeResponse, error) {
	var result IntracityGetPayModeResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/getpaymode", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IntracitySetPayModeRequest sets the payment mode of the capability.
type IntracitySetPayModeRequest struct {
	// PayMode to enable.
	PayMode int `json:"pay_mode"`
}

// IntracitySetPayMode updates the payment mode (支付方式) used for orders.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/same_city_distribution/api_intracity_setpaymode.html
func (w *MiniProgram) IntracitySetPayMode(ctx context.Context, payMode int) error {
	req := &IntracitySetPayModeRequest{PayMode: payMode}
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/intracity/setpaymode", nil, defaultReqOptions(), req, nil)
}
