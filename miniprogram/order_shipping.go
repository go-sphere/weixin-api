package miniprogram

import (
	"context"
)

// OrderShippingItem is one shipped/returned order entry reported via the
// shipping information APIs (also known as "发货信息管理" / transaction
// guarantee order info).
type OrderShippingItem struct {
	// TransactionID is the WeChat Pay transaction id (required unless
	// OutTradeNo+MerchantID given).
	TransactionID string `json:"transaction_id,omitempty"`
	// OutTradeNo is the merchant order number (required unless
	// TransactionID given).
	OutTradeNo string `json:"out_trade_no,omitempty"`
	// MerchantID (mchid) of the WeChat Pay merchant (required together with
	// OutTradeNo).
	MerchantID string `json:"merchant_id,omitempty"`
	// OrderType distinguishes shipment of a paid order: 1 = 虚拟物品发货,
	// 2 = 线下交易, 3 = 线下配送, 4 = 上门自提.
	OrderType int `json:"order_type,omitempty"`
}

// UploadOrderShippingInfoRequest is the payload of
// UploadOrderShippingInfo.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_uploadshippinginfo.html
type UploadOrderShippingInfoRequest struct {
	// OrderKey identifies the order being shipped.
	OrderKey struct {
		// OrderNumberType selects which field identifies the order: 1 = the
		// merchant order number, 2 = the WeChat transaction id.
		OrderNumberType int `json:"order_number_type"`
		// TransactionID used when OrderNumberType is 2.
		TransactionID string `json:"transaction_id,omitempty"`
		// OutTradeNo used when OrderNumberType is 1.
		OutTradeNo string `json:"out_trade_no,omitempty"`
		// MchID of the WeChat Pay merchant.
		MchID string `json:"mchid,omitempty"`
	} `json:"order_key"`
	// LogisticsType of the delivery: 1 = real logistics, 2 = virtual goods
	// fulfillment, 3 = offline self-operated delivery.
	LogisticsType int `json:"logistics_type"`
	// DeliveryMode of the delivery: 1 = dispatched via a delivery company,
	// 2 = same city / self operated, 3 = self pick-up.
	DeliveryMode int `json:"delivery_mode"`
	// ShippingList is the list of shipped packages (one per courier waybill).
	ShippingList []ShippingPackage `json:"shipping_list"`
	// UploadTime is the Unix timestamp (seconds) of the shipment.
	UploadTime int64 `json:"upload_time,omitempty"`
	// Payer contains the paying user openid or unionid.
	Payer struct {
		// OpenID of the paying user.
		OpenID string `json:"openid,omitempty"`
	} `json:"payer,omitempty"`
}

// ShippingPackage is one shipped package of an order.
type ShippingPackage struct {
	// TrackingNo is the tracking/waybill number of the package.
	TrackingNo string `json:"tracking_no,omitempty"`
	// ExpressCompany is the delivery company code (see the logistics
	// company list API).
	ExpressCompany string `json:"express_company,omitempty"`
	// ItemDesc is the description of the shipped items.
	ItemDesc string `json:"item_desc,omitempty"`
	// Contact carries the recipient contact for offline deliveries.
	Contact struct {
		// ConsignorContact is the sender's contact (for same-city mode).
		ConsignorContact string `json:"consignor_contact,omitempty"`
		// ReceiverContact is the recipient's contact.
		ReceiverContact string `json:"receiver_contact,omitempty"`
	} `json:"contact,omitempty"`
}

// UploadOrderShippingInfo reports to WeChat that the goods of an order were
// shipped. This is required for Mini Programs using the transaction-guarantee
// ("交易保障") feature and drives the user-visible order logistics status.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_uploadshippinginfo.html
func (w *MiniProgram) UploadOrderShippingInfo(ctx context.Context, req *UploadOrderShippingInfoRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/order/upload_shipping_info", nil, defaultReqOptions(), req, nil)
}

// UploadCombinedShippingInfoRequest combines several packages into one
// shipment notice.
type UploadCombinedShippingInfoRequest struct {
	// OrderKey identifies the order (same shape as the single upload).
	OrderKey struct {
		OrderNumberType int    `json:"order_number_type"`
		TransactionID   string `json:"transaction_id,omitempty"`
		OutTradeNo      string `json:"out_trade_no,omitempty"`
		MchID           string `json:"mchid,omitempty"`
	} `json:"order_key"`
	// UploadTime is the Unix timestamp (seconds) of the shipment.
	UploadTime int64 `json:"upload_time,omitempty"`
	// Payer of the order.
	Payer struct {
		OpenID string `json:"openid,omitempty"`
	} `json:"payer,omitempty"`
	// ShippingList of the combined packages.
	ShippingList []ShippingPackage `json:"shipping_list,omitempty"`
}

// UploadCombinedShippingInfo reports multiple packages for one order in a
// single call.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_uploadcombinedshippinginfo.html
func (w *MiniProgram) UploadCombinedShippingInfo(ctx context.Context, req *UploadCombinedShippingInfoRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/order/upload_combined_shipping_info", nil, defaultReqOptions(), req, nil)
}

// NotifyConfirmReceiveRequest notifies WeChat that the user confirmed receipt
// of the order goods.
type NotifyConfirmReceiveRequest struct {
	// OrderKey identifies the order (same shape as the single upload).
	OrderKey struct {
		OrderNumberType int    `json:"order_number_type"`
		TransactionID   string `json:"transaction_id,omitempty"`
		OutTradeNo      string `json:"out_trade_no,omitempty"`
		MchID           string `json:"mchid,omitempty"`
	} `json:"order_key"`
	// Payer of the order.
	Payer struct {
		OpenID string `json:"openid,omitempty"`
	} `json:"payer,omitempty"`
}

// NotifyConfirmReceive tells WeChat the user confirmed receipt, moving the
// order into the confirmed state.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_notifyconfirmreceive.html
func (w *MiniProgram) NotifyConfirmReceive(ctx context.Context, req *NotifyConfirmReceiveRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/order/notify_confirm_receive", nil, defaultReqOptions(), req, nil)
}

// OrderKey mirrors the nested order key shape shared by order queries.
type OrderKey struct {
	// OrderNumberType selects the identifier (1 merchant order no, 2
	// transaction id).
	OrderNumberType int `json:"order_number_type"`
	// TransactionID used when OrderNumberType is 2.
	TransactionID string `json:"transaction_id,omitempty"`
	// OutTradeNo used when OrderNumberType is 1.
	OutTradeNo string `json:"out_trade_no,omitempty"`
	// MchID of the WeChat Pay merchant.
	MchID string `json:"mchid,omitempty"`
}

// GetOrderShippingInfoResponse is returned by GetOrderShippingInfo. The
// PayInfo payload carries the logistics state reported previously.
type GetOrderShippingInfoResponse struct {
	ErrResponse
	// ShippingList is the current shipping info of the order.
	ShippingList []ShippingPackage `json:"shipping_list,omitempty"`
}

// GetOrderShippingInfo returns the shipping information previously reported
// for an order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_getorder.html
func (w *MiniProgram) GetOrderShippingInfo(ctx context.Context, orderKey *OrderKey) (*GetOrderShippingInfoResponse, error) {
	body := map[string]any{"order_key": orderKey}
	var result GetOrderShippingInfoResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/order/get_order", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IsTradeManagedResponse is returned by IsTradeManaged.
type IsTradeManagedResponse struct {
	ErrResponse
	// IsTradeManaged is 1 when the Mini Program uses the trade-managed
	// (transaction guarantee) capability.
	IsTradeManaged int `json:"is_trade_managed"`
}

// IsTradeManaged reports whether the Mini Program is managed by WeChat's
// trade-guarantee ("交易保障") capability, i.e. whether shipping info must be
// uploaded.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_istrademanaged.html
func (w *MiniProgram) IsTradeManaged(ctx context.Context) (*IsTradeManagedResponse, error) {
	var result IsTradeManagedResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/order/is_trade_managed", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetMessageJumpPathRequest configures the message card jump path shown to
// users after an order event.
type SetMessageJumpPathRequest struct {
	// JumpPath is the Mini Program page path for the message card to jump to.
	JumpPath string `json:"jump_path,omitempty"`
	// AppID of a related Mini Program to jump to (optional, must be bound).
	AppID string `json:"appid,omitempty"`
	// Path is used together with AppID when jumping to another Mini Program.
	Path string `json:"path,omitempty"`
}

// SetMessageJumpPath sets the jump configuration of the WeChat Pay message
// cards (order status push).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_setmsgjumppath.html
func (w *MiniProgram) SetMessageJumpPath(ctx context.Context, req *SetMessageJumpPathRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/order/set_msg_jump_path", nil, defaultReqOptions(), req, nil)
}

// IsTradeManagementConfirmationCompleted reports whether the developer has
// confirmed the trade-management (交易保障) capability onboarding; appid is the
// Mini Program to check.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_istrademanagementconfirmationcompleted.html
func (w *MiniProgram) IsTradeManagementConfirmationCompleted(ctx context.Context, appid string) (bool, error) {
	var result struct {
		ErrResponse
		Completed bool `json:"completed"`
	}
	body := map[string]string{"appid": appid}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/order/is_trade_management_confirmation_completed", nil, defaultReqOptions(), body, &result); err != nil {
		return false, err
	}
	return result.Completed, nil
}

// ListManagedOrderRequest pages the trade-managed orders.
type ListManagedOrderRequest struct {
	// PayTimeRange restricts the orders by pay time (unix seconds).
	PayTimeRange *struct {
		BeginTime int64 `json:"begin_time"`
		EndTime   int64 `json:"end_time"`
	} `json:"pay_time_range,omitempty"`
	// OrderState filter of the orders.
	OrderState int `json:"order_state,omitempty"`
	// OpenID of a specific user.
	OpenID string `json:"openid,omitempty"`
	// LastIndex cursor of the previous page (from ListManagedOrderResponse).
	LastIndex string `json:"last_index,omitempty"`
	// PageSize of the page.
	PageSize int `json:"page_size"`
}

// ManagedOrderEntry is one order of the list.
type ManagedOrderEntry struct {
	// RawData of the order entry.
	RawData map[string]any `json:"-"`
}

// ListManagedOrderResponse is returned by ListManagedOrder.
type ListManagedOrderResponse struct {
	ErrResponse
	// OrderList of the page.
	OrderList []map[string]any `json:"order_list,omitempty"`
	// HasMore is true when more pages follow.
	HasMore bool `json:"has_more"`
	// LastIndex cursor for the next page.
	LastIndex string `json:"last_index"`
}

// ListManagedOrder lists the orders under trade-management, pageable through
// the last_index cursor.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_getorderlist.html
func (w *MiniProgram) ListManagedOrder(ctx context.Context, req *ListManagedOrderRequest) (*ListManagedOrderResponse, error) {
	var result ListManagedOrderResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/order/get_order_list", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetWxaTradeTypeRequest is the payload of SetWxaTradeType.
type SetWxaTradeTypeRequest struct {
	// TradeType to switch to (target trade type enum per WeChat docs).
	TradeType int `json:"trade_type"`
	// MaterialList of the supporting materials (up to 10; images/videos via
	// media_id).
	MaterialList []SetWxaTradeTypeMaterial `json:"material_list"`
	// Reason of the trade-type change.
	Reason string `json:"reason"`
}

// SetWxaTradeTypeMaterial is one material of a trade-type change application.
type SetWxaTradeTypeMaterial struct {
	// Type of the material: 1 image, 2 video.
	Type int `json:"type"`
	// MediaID uploaded through the temporary-media API.
	MediaID string `json:"media_id"`
}

// SetWxaTradeType applies to change the trade type (发货模式) of the
// trade-managed Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/order_shipping/api_setwxatradetypecgi.html
func (w *MiniProgram) SetWxaTradeType(ctx context.Context, req *SetWxaTradeTypeRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/order/setwxatradetypecgi", nil, defaultReqOptions(), req, nil)
}
