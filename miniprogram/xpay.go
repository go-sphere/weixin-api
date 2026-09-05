package miniprogram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ============================================================
// XPay (虚拟支付) — signing and transport helpers.
//
// All /xpay endpoints accept POST JSON bodies and are signed:
//   - pay_sig = HMAC-SHA256(AppKey, path + "&" + rawBody) is required for
//     every request;
//   - signature = HMAC-SHA256(user session_key, rawBody) is additionally
//     required for coin/user-level requests (query_user_balance, currency_pay,
//     cancel_currency_pay, present_currency).
//
// The signatures are appended as query parameters beside access_token.
// ============================================================

// xpayRequest runs one signed XPay request. sessionKey is optional; provide it
// when the upstream endpoint requires the user-level signature.
func (w *MiniProgram) xpayRequest(ctx context.Context, path string, body any, sessionKey string, requireUserSign bool) ([]byte, error) {
	if w.config.AppKey == "" {
		return nil, fmt.Errorf("wechat: xpay %s requires a non-empty Config.AppKey", path)
	}
	rawBody, err := jsonMarshal(body)
	if err != nil {
		return nil, err
	}
	paySign := hmacSHA256Hex([]byte(w.config.AppKey), []byte(path+"&"+string(rawBody)))
	query := url.Values{}
	query.Set("pay_sig", paySign)
	if requireUserSign {
		if sessionKey == "" {
			return nil, fmt.Errorf("wechat: xpay %s requires a user session key", path)
		}
		query.Set("signature", hmacSHA256Hex([]byte(sessionKey), rawBody))
	}
	data, err := w.withAccessTokenRaw(ctx, http.MethodPost, path, query, defaultReqOptions(), rawBody)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ============================================================
// User coin balance & coin transactions.
// ============================================================

// XPayQueryUserBalanceRequest queries the coin balance of a user.
type XPayQueryUserBalanceRequest struct {
	XPayCommon
}

// XPayQueryUserBalanceResponse is returned by XPayQueryUserBalance.
type XPayQueryUserBalanceResponse struct {
	ErrResponse
	// Balance is the total coin balance (paid + present).
	Balance int `json:"balance"`
	// PresentBalance is the balance of the granted (无价) coins.
	PresentBalance int `json:"present_balance"`
	// SumSave is the cumulative amount of paid coins recharged.
	SumSave int `json:"sum_save"`
	// SumPresent is the cumulative amount of present coins granted.
	SumPresent int `json:"sum_present"`
	// SumBalance is the historical total of coins added.
	SumBalance int `json:"sum_balance"`
	// SumCost is the historical total of coins spent.
	SumCost int `json:"sum_cost"`
	// FirstSaveFlag is 1 when the user satisfies the first recharge reward.
	FirstSaveFlag int `json:"first_save_flag"`
}

// XPayQueryUserBalance returns the coin (代币) balance of a user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_user_balance.html
func (w *MiniProgram) XPayQueryUserBalance(ctx context.Context, req *XPayQueryUserBalanceRequest, sessionKey string) (*XPayQueryUserBalanceResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_user_balance", req, sessionKey, true)
	if err != nil {
		return nil, err
	}
	var result XPayQueryUserBalanceResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayCurrencyPayRequest deducts coins from a user (usually to pay for virtual
// goods priced in coins).
type XPayCurrencyPayRequest struct {
	XPayCommon
	// Amount of coins to deduct.
	Amount int `json:"amount"`
	// OrderID is the merchant order id (must be unique).
	OrderID string `json:"order_id"`
	// PayItem is a JSON-encoded array of the purchased items, recorded in the
	// user's flow, e.g. [{"productid":"id","unit_price":10,"quantity":1}].
	PayItem string `json:"payitem"`
	// Remark displayed in the bill.
	Remark string `json:"remark,omitempty"`
}

// XPayCurrencyPayResponse is returned by XPayCurrencyPay.
type XPayCurrencyPayResponse struct {
	ErrResponse
	// OrderID echoed back.
	OrderID string `json:"order_id"`
	// Balance after the deduction.
	Balance int `json:"balance"`
	// UsedPresentAmount of present coins consumed by the payment.
	UsedPresentAmount int `json:"used_present_amount"`
}

// XPayCurrencyPay deducts coins (代币扣减) for a user purchase.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_currency_pay.html
func (w *MiniProgram) XPayCurrencyPay(ctx context.Context, req *XPayCurrencyPayRequest, sessionKey string) (*XPayCurrencyPayResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/currency_pay", req, sessionKey, true)
	if err != nil {
		return nil, err
	}
	var result XPayCurrencyPayResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayCancelCurrencyPayRequest refunds a coin payment (the reverse operation
// of XPayCurrencyPay).
type XPayCancelCurrencyPayRequest struct {
	XPayCommon
	// PayOrderID is the order_id used by the original currency_pay call.
	PayOrderID string `json:"pay_order_id"`
	// OrderID of this refund order.
	OrderID string `json:"order_id"`
	// Amount to refund.
	Amount int `json:"amount"`
}

// XPayCancelCurrencyPayResponse is returned by XPayCancelCurrencyPay.
type XPayCancelCurrencyPayResponse struct {
	ErrResponse
	// OrderID of the refund order.
	OrderID string `json:"order_id"`
}

// XPayCancelCurrencyPay refunds a previously deducted coin payment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_cancel_currency_pay.html
func (w *MiniProgram) XPayCancelCurrencyPay(ctx context.Context, req *XPayCancelCurrencyPayRequest, sessionKey string) (*XPayCancelCurrencyPayResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/cancel_currency_pay", req, sessionKey, true)
	if err != nil {
		return nil, err
	}
	var result XPayCancelCurrencyPayResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayPresentCurrencyRequest grants (赠送) coins to a user.
type XPayPresentCurrencyRequest struct {
	XPayCommon
	// OrderID of the present order (must be unique).
	OrderID string `json:"order_id"`
	// Amount of coins to grant.
	Amount int `json:"amount"`
}

// XPayPresentCurrencyResponse is returned by XPayPresentCurrency.
type XPayPresentCurrencyResponse struct {
	ErrResponse
	// Balance after granting.
	Balance int `json:"balance"`
	// OrderID of the present order.
	OrderID string `json:"order_id"`
	// PresentBalance received by the user in total.
	PresentBalance int `json:"present_balance"`
}

// XPayPresentCurrency grants present (赠送) coins to a user. Because presents
// cannot be looked up by order id, retry until errcode 0 or 268490004
// (repeat operation) is returned.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_present_currency.html
func (w *MiniProgram) XPayPresentCurrency(ctx context.Context, req *XPayPresentCurrencyRequest, sessionKey string) (*XPayPresentCurrencyResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/present_currency", req, sessionKey, true)
	if err != nil {
		return nil, err
	}
	var result XPayPresentCurrencyResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// Cash (现金) orders.
// ============================================================

// XPayQueryOrderRequest queries an XPay cash order by either the merchant or
// the WeChat order id.
type XPayQueryOrderRequest struct {
	XPayCommon
	// OrderID of the merchant (alternative to WxOrderID).
	OrderID string `json:"order_id,omitempty"`
	// WxOrderID of WeChat (alternative to OrderID).
	WxOrderID string `json:"wx_order_id,omitempty"`
}

// XPayOrder is the detailed payload of a cash order.
type XPayOrder struct {
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// CreateTime of the order (unix seconds).
	CreateTime int64 `json:"create_time"`
	// UpdateTime of the order (unix seconds).
	UpdateTime int64 `json:"update_time"`
	// Status of the order (see XPayOrderStatus).
	Status XPayOrderStatus `json:"status"`
	// BizType of the order (0 = 短剧 drama).
	BizType int `json:"biz_type"`
	// OrderFee is the order amount in cents.
	OrderFee int `json:"order_fee"`
	// CouponFee is the discounted amount in cents.
	CouponFee int `json:"coupon_fee"`
	// PaidFee is the amount the user paid in cents.
	PaidFee int `json:"paid_fee"`
	// OrderType: 0 payment order, 1 refund order.
	OrderType int `json:"order_type"`
	// RefundFee for refund orders, in cents.
	RefundFee int `json:"refund_fee"`
	// PaidTime of the payment/refund (unix seconds).
	PaidTime int64 `json:"paid_time"`
	// ProvideTime of the delivery (unix seconds).
	ProvideTime int64 `json:"provide_time"`
	// BizMeta custom data passed at order creation.
	BizMeta string `json:"biz_meta"`
	// EnvType: 1 production, 2 sandbox.
	EnvType int `json:"env_type"`
	// Token returned by 米大师 when the order was placed.
	Token string `json:"token"`
	// LeftFee is the still-refundable amount of a payment order.
	LeftFee int `json:"left_fee"`
	// WxOrderID of WeChat.
	WxOrderID string `json:"wx_order_id"`
	// ChannelOrderID is the merchant order id shown in the WeChat Pay detail.
	ChannelOrderID string `json:"channel_order_id"`
	// WxPayOrderID is the WeChat Pay transaction id.
	WxPayOrderID string `json:"wxpay_order_id"`
}

// XPayQueryOrderResponse is returned by XPayQueryOrder.
type XPayQueryOrderResponse struct {
	ErrResponse
	// Order of the query.
	Order *XPayOrder `json:"order,omitempty"`
}

// XPayQueryOrder queries the state of a cash (现金) order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_order.html
func (w *MiniProgram) XPayQueryOrder(ctx context.Context, req *XPayQueryOrderRequest) (*XPayQueryOrderResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_order", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryOrderResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayRefundOrderRequest refunds a cash order placed through the JSAPI flow.
type XPayRefundOrderRequest struct {
	XPayCommon
	// OrderID of the original merchant order (alternative to WxOrderID).
	OrderID string `json:"order_id,omitempty"`
	// WxOrderID of the original order (alternative to OrderID).
	WxOrderID string `json:"wx_order_id,omitempty"`
	// RefundOrderID of this refund (length 8-32, letters/digits/_/-).
	RefundOrderID string `json:"refund_order_id"`
	// LeftFee is the remaining refundable amount of the order in cents (see
	// XPayQueryOrder).
	LeftFee int `json:"left_fee"`
	// RefundFee to refund, in (0, LeftFee] cents.
	RefundFee int `json:"refund_fee"`
	// BizMeta custom data echoed back by query_order (up to 1024 chars).
	BizMeta string `json:"biz_meta,omitempty"`
	// RefundReason code: 0 none, 1 product problem, 2 after-sales, 3 user
	// willingness, 4 price problem, 5 other.
	RefundReason int `json:"refund_reason,omitempty"`
	// ReqFrom: 1 manual by customer service, 2 user self service, 3 other.
	ReqFrom int `json:"req_from,omitempty"`
}

// XPayRefundOrderResponse is returned by XPayRefundOrder.
type XPayRefundOrderResponse struct {
	ErrResponse
	// RefundOrderID of this refund.
	RefundOrderID string `json:"refund_order_id"`
	// RefundWxOrderID of WeChat.
	RefundWxOrderID string `json:"refund_wx_order_id"`
	// PayOrderID of the original payment order.
	PayOrderID string `json:"pay_order_id"`
	// PayWxOrderID of the original WeChat payment order.
	PayWxOrderID string `json:"pay_wx_order_id"`
}

// XPayRefundOrder refunds a cash order placed through the JSAPI payment flow.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_refund_order.html
func (w *MiniProgram) XPayRefundOrder(ctx context.Context, req *XPayRefundOrderRequest) (*XPayRefundOrderResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/refund_order", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayRefundOrderResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayNotifyProvideGoodsRequest marks a cash order as delivered. This is only
// needed when the xpay_goods_deliver_notify push was not received; normally the
// delivery is acknowledged through that callback.
type XPayNotifyProvideGoodsRequest struct {
	// OrderID of the merchant (alternative to WxOrderID).
	OrderID string `json:"order_id,omitempty"`
	// WxOrderID of WeChat (alternative to OrderID).
	WxOrderID string `json:"wx_order_id,omitempty"`
	// Env of the order.
	Env XPayEnv `json:"env"`
}

// XPayNotifyProvideGoods notifies WeChat that the virtual goods of a paid cash
// order were delivered.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_notify_provide_goods.html
func (w *MiniProgram) XPayNotifyProvideGoods(ctx context.Context, req *XPayNotifyProvideGoodsRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/notify_provide_goods", req, "", false)
	return err
}
