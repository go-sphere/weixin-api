package miniprogram

import (
	"context"
	"net/http"
)

// ExpressDeliveryCompany is one delivery company supported by WeChat express.
type ExpressDeliveryCompany struct {
	// DeliveryID of the company, e.g. "SF", "YTO".
	DeliveryID string `json:"delivery_id"`
	// DeliveryName of the company.
	DeliveryName string `json:"delivery_name"`
	// CanUseCash indicates whether the merchant bound the company account.
	CanUseCash int `json:"can_use_cash"`
	// CanUseCash2 indicates whether a second cash-account binding exists.
	CanUseCash2 int `json:"can_use_cash2,omitempty"`
	// SupportType lists the supported service types (0 = no cash on delivery,
	// 1 = cash on delivery supported, 2 = cash and quality check...).
	SupportType int `json:"support_type"`
	// DefaultQuota of monthly bill-counts the merchant has.
	DefaultQuota int `json:"default_quota,omitempty"`
	// QuotaUsed of the current month.
	QuotaUsed int `json:"quota_used,omitempty"`
	// QuotaRemain of the current month.
	QuotaRemain int `json:"quota_remain,omitempty"`
}

// GetAllDeliveryCompanyResponse is returned by GetAllDeliveryCompany.
type GetAllDeliveryCompanyResponse struct {
	ErrResponse
	// Count of the returned companies.
	Count int `json:"count"`
	// Data is the list of supported delivery companies.
	Data []ExpressDeliveryCompany `json:"data"`
}

// GetAllDeliveryCompany lists every delivery company WeChat express integrates
// with, together with the merchant's per-company quota usage.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_getalldelivery.html
func (w *MiniProgram) GetAllDeliveryCompany(ctx context.Context) (*GetAllDeliveryCompanyResponse, error) {
	var result GetAllDeliveryCompanyResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/express/business/delivery/getall", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ExpressAccount is one delivery-company account bound to the merchant.
type ExpressAccount struct {
	// DeliveryID of the company the account belongs to.
	DeliveryID string `json:"delivery_id"`
	// BizID is the business id of the merchant with that company.
	BizID string `json:"biz_id"`
	// Password is the bound password (masked).
	Password string `json:"password"`
	// CreateTime of the binding (Unix seconds).
	CreateTime int64 `json:"create_time"`
	// UpdateTime of the binding (Unix seconds).
	UpdateTime int64 `json:"update_time"`
	// StatusCode of the binding.
	StatusCode int `json:"status_code"`
	// StatusMessage of the binding.
	StatusMessage string `json:"status_msg"`
}

// GetAllExpressAccountResponse is returned by GetAllExpressAccount.
type GetAllExpressAccountResponse struct {
	ErrResponse
	// Count of the bound accounts.
	Count int `json:"count"`
	// Data is the list of bound accounts.
	Data []ExpressAccount `json:"data"`
}

// GetAllExpressAccount lists the merchant accounts bound to the delivery
// companies.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_getallaccount.html
func (w *MiniProgram) GetAllExpressAccount(ctx context.Context) (*GetAllExpressAccountResponse, error) {
	var result GetAllExpressAccountResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/express/business/account/getall", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BindExpressAccountRequest binds an account of a delivery company.
type BindExpressAccountRequest struct {
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// BizID of the merchant with the company.
	BizID string `json:"biz_id"`
	// Password of the account.
	Password string `json:"password"`
}

// BindExpressAccount binds the merchant's account of a delivery company so
// that waybills can be created with it.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_bindaccount.html
func (w *MiniProgram) BindExpressAccount(ctx context.Context, req *BindExpressAccountRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/business/account/bind", nil, defaultReqOptions(), req, nil)
}

// ExpressReceiver is the receiver of an express order.
type ExpressReceiver struct {
	// Name of the receiver.
	Name string `json:"name"`
	// Tel of the receiver (must match the wechat-openid pattern rules; either
	// Tel or Mobile required).
	Tel string `json:"tel,omitempty"`
	// Mobile of the receiver.
	Mobile string `json:"mobile,omitempty"`
	// Address of the receiver.
	Address string `json:"address"`
}

// ExpressSender is the sender of an express order.
type ExpressSender struct {
	// Name of the sender.
	Name string `json:"name"`
	// Tel of the sender.
	Tel string `json:"tel,omitempty"`
	// Mobile of the sender.
	Mobile string `json:"mobile,omitempty"`
	// Address of the sender.
	Address string `json:"address"`
}

// ExpressOrderCargo describes the goods of an express order.
type ExpressOrderCargo struct {
	// Count of the packages.
	Count int `json:"count"`
	// Weight of the order in kg.
	Weight float64 `json:"weight,omitempty"`
	// SpaceX/Y/Z is the space size (cm).
	SpaceX int `json:"space_x,omitempty"`
	SpaceY int `json:"space_y,omitempty"`
	SpaceZ int `json:"space_z,omitempty"`
}

// ExpressShop is the shop info attached to an express order.
type ExpressShop struct {
	// WxaPath of the Mini Program page the user can visit to follow the order.
	WxaPath string `json:"wxa_path,omitempty"`
	// WxaAppID of the Mini Program (defaults to the current app when empty).
	WxaAppID string `json:"wxa_appid,omitempty"`
	// ShopName displayed in the order card.
	ShopName string `json:"shop_name,omitempty"`
	// ShopOrderID of the merchant.
	ShopOrderID string `json:"shop_order_id,omitempty"`
}

// ExpressAddOrderRequest creates an express waybill order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_addorder.html
type ExpressAddOrderRequest struct {
	// OrderID is the merchant order id (unique within the Mini Program).
	OrderID string `json:"order_id"`
	// OpenID of the user who placed the order (used for message pushes).
	OpenID string `json:"openid"`
	// DeliveryID of the chosen delivery company.
	DeliveryID string `json:"delivery_id"`
	// BizID of the merchant account of that company.
	BizID string `json:"biz_id"`
	// CustomRemark shown on the waybill.
	CustomRemark string `json:"custom_remark,omitempty"`
	// ExpectType is 0 (cheapest), 1 (fastest), 2 (balance). Default 0.
	ExpectType int `json:"expect_type,omitempty"`
	// Cargo of the order.
	Cargo ExpressOrderCargo `json:"cargo"`
	// Sender of the order.
	Sender ExpressSender `json:"sender"`
	// Receiver of the order.
	Receiver ExpressReceiver `json:"receiver"`
	// Shop info.
	Shop ExpressShop `json:"shop"`
}

// ExpressAddOrderResult is the successful creation result of an express order.
type ExpressAddOrderResult struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// WaybillID generated by WeChat.
	WaybillID string `json:"waybill_id"`
	// WaybillData is the printable waybill content (template rendering data).
	WaybillData []ExpressWaybillData `json:"waybill_data"`
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// OrderStatus of the created order.
	OrderStatus int `json:"order_status,omitempty"`
}

// ExpressWaybillData is one line of waybill printing data.
type ExpressWaybillData struct {
	// Key of the waybill field.
	Key string `json:"key"`
	// Value of the waybill field.
	Value string `json:"value"`
}

// ExpressAddOrderResponse is returned by ExpressAddOrder. A nil error means the
// order was created and WaybillData is ready to print.
type ExpressAddOrderResponse struct {
	ErrResponse
	ExpressAddOrderResult
}

// ExpressAddOrder creates a waybill order with the chosen delivery company and
// returns the printable waybill data.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_addorder.html
func (w *MiniProgram) ExpressAddOrder(ctx context.Context, req *ExpressAddOrderRequest) (*ExpressAddOrderResponse, error) {
	var result ExpressAddOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/business/order/add", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ExpressCancelOrderRequest cancels a created express order.
type ExpressCancelOrderRequest struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// WaybillID returned at creation.
	WaybillID string `json:"waybill_id"`
}

// ExpressCancelOrderResponse is returned by ExpressCancelOrder.
type ExpressCancelOrderResponse struct {
	ErrResponse
	// OrderStatus of the cancelled order.
	OrderStatus int `json:"order_status,omitempty"`
}

// ExpressCancelOrder cancels a pending express order before the delivery
// company picks it up.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_cancelorder.html
func (w *MiniProgram) ExpressCancelOrder(ctx context.Context, req *ExpressCancelOrderRequest) (*ExpressCancelOrderResponse, error) {
	var result ExpressCancelOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/business/order/cancel", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ExpressOrderDetailRequest selects an express order to query.
type ExpressOrderDetailRequest struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
}

// ExpressOrderDetail is the detailed state of one express order.
type ExpressOrderDetail struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
	// OrderStatus: 0 pending pickup, 1 in transit, 2 signed, 3 abnormal.
	OrderStatus int `json:"order_status,omitempty"`
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id,omitempty"`
	// ShipTime of the order.
	ShipTime string `json:"ship_time,omitempty"`
	// ReceiverName of the order.
	ReceiverName string `json:"receiver_name,omitempty"`
	// ReceiverAddress of the order.
	ReceiverAddress string `json:"receiver_address,omitempty"`
	// WaybillData is the printable waybill content.
	WaybillData []ExpressWaybillData `json:"waybill_data,omitempty"`
}

// GetExpressOrderResponse is returned by GetExpressOrder.
type GetExpressOrderResponse struct {
	ErrResponse
	// Order is the queried order state.
	Order *ExpressOrderDetail `json:"order,omitempty"`
}

// GetExpressOrder queries the current state of one express order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_getorder.html
func (w *MiniProgram) GetExpressOrder(ctx context.Context, req *ExpressOrderDetailRequest) (*GetExpressOrderResponse, error) {
	var result GetExpressOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/business/order/get", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchExpressOrderRequest filters the orders to query in batch.
type BatchExpressOrderRequest struct {
	// OrderList of the orders to query (up to 50).
	OrderList []ExpressOrderDetailRequest `json:"order_list"`
}

// BatchExpressOrderItem is one queried order of the batch result.
type BatchExpressOrderItem struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
	// OrderStatus of the order.
	OrderStatus int `json:"order_status,omitempty"`
	// WaybillData printable content.
	WaybillData []ExpressWaybillData `json:"waybill_data,omitempty"`
}

// BatchGetExpressOrderResponse is returned by BatchGetExpressOrder.
type BatchGetExpressOrderResponse struct {
	ErrResponse
	// OrderList is the returned orders.
	OrderList []BatchExpressOrderItem `json:"order_list,omitempty"`
}

// BatchGetExpressOrder queries up to 50 express orders at once.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_batchgetorder.html
func (w *MiniProgram) BatchGetExpressOrder(ctx context.Context, req *BatchExpressOrderRequest) (*BatchGetExpressOrderResponse, error) {
	var result BatchGetExpressOrderResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/business/order/batchget", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ExpressTrackPath is one leg of an express tracking path.
type ExpressTrackPath struct {
	// ActionTime of the leg (Unix seconds).
	ActionTime int `json:"action_time"`
	// ActionType of the leg (courier actions per WeChat docs).
	ActionType int `json:"action_type"`
	// ActionMsg describing the leg.
	ActionMsg string `json:"action_msg"`
}

// GetExpressPathResponse is returned by GetExpressPath.
type GetExpressPathResponse struct {
	ErrResponse
	// OpenID is echoed back.
	OpenID string `json:"openid,omitempty"`
	// DeliveryID of the company.
	DeliveryID string `json:"delivery_id,omitempty"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id,omitempty"`
	// PathList is the tracking timeline.
	PathList []ExpressTrackPath `json:"path_list,omitempty"`
}

// GetExpressPath returns the tracking timeline of an express waybill.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_getpath.html
func (w *MiniProgram) GetExpressPath(ctx context.Context, req *ExpressOrderDetailRequest) (*GetExpressPathResponse, error) {
	var result GetExpressPathResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/business/path/get", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ExpressPrinter describes one bound bluetooth printer.
type ExpressPrinter struct {
	// OpenID of the printer binding user.
	OpenID string `json:"openid"`
	// UpdateTime of the binding.
	UpdateTime int64 `json:"update_time,omitempty"`
	// TagName of the printer (自定义标签).
	TagName string `json:"tag_name,omitempty"`
}

// GetExpressPrinterResponse is returned by GetExpressPrinterList.
type GetExpressPrinterResponse struct {
	ErrResponse
	// Count of the bound printers.
	Count int `json:"count"`
	// OpenID of the current query user.
	OpenID string `json:"openid,omitempty"`
	// PrinterList of the bound printers.
	PrinterList []ExpressPrinter `json:"printer_list,omitempty"`
}

// GetExpressPrinterList lists the bluetooth printers bound by the user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_getprinter.html
func (w *MiniProgram) GetExpressPrinterList(ctx context.Context, openid string) (*GetExpressPrinterResponse, error) {
	var result GetExpressPrinterResponse
	body := map[string]string{"openid": openid}
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/business/printer/getall", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateExpressPrinterRequest binds a bluetooth printer for a user.
type UpdateExpressPrinterRequest struct {
	// OpenID of the user binding the printer.
	OpenID string `json:"openid"`
	// TagName is the printer alias.
	TagName string `json:"tag_name,omitempty"`
}

// UpdateExpressPrinter adds (action=bind) or removes (action=unbind) a printer.
// The bind flow requires the wx.getBluetoothDevices step on the client first.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_updateprinter.html
func (w *MiniProgram) UpdateExpressPrinter(ctx context.Context, action, openid string) error {
	body := map[string]string{
		"openid": openid,
		"action": action,
	}
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/business/printer/update", nil, defaultReqOptions(), body, nil)
}

// GetExpressQuotaResponse is returned by GetExpressQuota.
type GetExpressQuotaResponse struct {
	ErrResponse
	// QuotaNum is the remaining bill quota of the month.
	QuotaNum int `json:"quota_num"`
}

// GetExpressQuota returns the merchant's remaining express bill quota for the
// month.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_getquota.html
func (w *MiniProgram) GetExpressQuota(ctx context.Context) (*GetExpressQuotaResponse, error) {
	var result GetExpressQuotaResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/business/quota/get", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MockExpressUpdateOrderRequest simulates a delivery update for test orders.
type MockExpressUpdateOrderRequest struct {
	// OrderID of the merchant (order must carry test_express as delivery_id).
	OrderID string `json:"order_id"`
	// OpenID of the user.
	OpenID string `json:"openid"`
	// DeliveryID must be "TEST" (test_express).
	DeliveryID string `json:"delivery_id"`
	// WaybillID of the order.
	WaybillID string `json:"waybill_id"`
	// BizID of the test account.
	BizID string `json:"biz_id"`
	// ActionType of the simulated event: 10001 pickup, 10002 in transit,
	// 10003 signed, 10004 abnormal.
	ActionType int `json:"action_type"`
}

// MockExpressUpdateOrder pushes a fake tracking event to a test express order
// (used during development before real delivery companies are bound).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/express/express-by-business/api_testupdateorder.html
func (w *MiniProgram) MockExpressUpdateOrder(ctx context.Context, req *MockExpressUpdateOrderRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/business/test_update_order", nil, defaultReqOptions(), req, nil)
}
