package miniprogram

// This file groups the shared types of the XPay (虚拟支付) API family.

// XPayEnv selects the XPay environment: 0 production, 1 sandbox.
type XPayEnv int

const (
	// XPayEnvProduction is the production XPay environment.
	XPayEnvProduction XPayEnv = 0
	// XPayEnvSandbox is the sandbox XPay environment used for testing.
	XPayEnvSandbox XPayEnv = 1
)

// XPayErrCode documents the reserved XPay business error codes.
type XPayErrCode int

// Well-known XPay error codes; they surface through APIError with errcode in
// the 2684xxxx range.
const (
	XPayErrOpenIDError         XPayErrCode = 268490001 // invalid openid
	XPayErrRequestParamError   XPayErrCode = 268490002 // request parameter error
	XPayErrSignError           XPayErrCode = 268490003 // signature error
	XPayErrRepeatOperation     XPayErrCode = 268490004 // repeat operation (already succeeded)
	XPayErrOrderRefunded       XPayErrCode = 268490005 // order already refunded
	XPayErrInsufficientBalance XPayErrCode = 268490006 // insufficient coin balance
	XPayErrSensitiveContent    XPayErrCode = 268490007 // sensitive content detected
	XPayErrSessionKeyExpired   XPayErrCode = 268490009 // session_key missing or expired
	XPayErrBillGenerating      XPayErrCode = 268490011 // bill still generating
	XPayErrRefundInProgress    XPayErrCode = 268490014 // refund in progress, retry later
	XPayErrAccountNotVerified  XPayErrCode = 268490021 // account not verified (进件)
)

// XPayOrderStatus is the state machine of an XPay cash order.
type XPayOrderStatus int

const (
	XPayOrderStatusInit         XPayOrderStatus = 0 // order initialised, not usable for payment
	XPayOrderStatusCreated      XPayOrderStatus = 1 // order created
	XPayOrderStatusPaid         XPayOrderStatus = 2 // paid, awaiting delivery
	XPayOrderStatusDelivering   XPayOrderStatus = 3 // delivering
	XPayOrderStatusDelivered    XPayOrderStatus = 4 // delivered
	XPayOrderStatusRefunded     XPayOrderStatus = 5 // refunded
	XPayOrderStatusClosed       XPayOrderStatus = 6 // closed, not usable anymore
	XPayOrderStatusRefundFailed XPayOrderStatus = 7 // refund failed
)

// XPayCommon is the base of the requests that are bound to a user and an
// environment.
type XPayCommon struct {
	// OpenID of the user (absent for merchant-level calls).
	OpenID string `json:"openid,omitempty"`
	// Env selects production (0) or sandbox (1).
	Env XPayEnv `json:"env"`
	// UserIP of the caller, e.g. "1.1.1.1".
	UserIP string `json:"user_ip,omitempty"`
	// DeviceType: 1 Android, 2 iOS.
	DeviceType int `json:"device_type,omitempty"`
}

// XPayCommonResponse is the shared error envelope of every XPay response.
type XPayCommonResponse struct {
	ErrResponse
}
