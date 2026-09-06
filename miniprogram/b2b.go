package miniprogram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ============================================================
// B2b 零售 / 微信支付分账 (retail merchant profit sharing). The money-moving
// calls (profit sharing, refunds, withdrawals, auto-withdraw configuration)
// require the merchant-signed pay_sig query parameter which the caller must
// compute with their WeChat-Pay key and pass via WithPaySig; the pure query
// endpoints do not.
// ============================================================

// b2bOptions carries the WeChat-Pay signature for a retail/B2b call.
type b2bOptions struct {
	paySig string
}

// b2bCallOption customises a retail/B2b request.
type b2bCallOption func(*b2bOptions)

// WithPaySig supplies the WeChat-Pay pay_sig required by the retail B2b money
// APIs.
func WithPaySig(paySig string) b2bCallOption {
	return func(o *b2bOptions) { o.paySig = paySig }
}

func newB2bOptions(opts []b2bCallOption) *b2bOptions {
	o := &b2bOptions{}
	for _, apply := range opts {
		apply(o)
	}
	return o
}

// b2bPost performs a retail/B2b POST and decodes the JSON response into dst
// (may be nil). It does not require a pay_sig; use b2bPostRequireSig for the
// money-moving endpoints that WeChat mandates a signature for.
func (w *MiniProgram) b2bPost(ctx context.Context, path string, body any, dst any, opts ...b2bCallOption) error {
	o := newB2bOptions(opts)
	query := url.Values{}
	if o.paySig != "" {
		query.Set("pay_sig", o.paySig)
	}
	data, err := w.withAccessTokenRaw(ctx, http.MethodPost, path, query, defaultReqOptions(), body)
	if err != nil {
		return err
	}
	if dst == nil {
		return nil
	}
	return decodeJSON(data, dst)
}

// b2bPostRequireSig is b2bPost for endpoints that WeChat rejects without a
// pay_sig: the caller must supply one via WithPaySig or the call fails fast
// instead of sending an unsigned request.
func (w *MiniProgram) b2bPostRequireSig(ctx context.Context, path string, body any, dst any, opts ...b2bCallOption) error {
	o := newB2bOptions(opts)
	if o.paySig == "" {
		return fmt.Errorf("wechat: %s requires a pay_sig (pass WithPaySig)", path)
	}
	return w.b2bPost(ctx, path, body, dst, WithPaySig(o.paySig))
}

// CreateProfitSharingOrderRequest splits a payment among the receivers.
type CreateProfitSharingOrderRequest struct {
	// MchID of the sub merchant (from B2b onboarding).
	MchID string `json:"mchid"`
	// OutTradeNo of the payment order in the B2b Mini Program.
	OutTradeNo string `json:"out_trade_no"`
	// ProfitFee to share, in cents (not exceeding the order amount).
	ProfitFee int64 `json:"profit_fee"`
	// ReceiverType of the receiving account (as registered).
	ReceiverType string `json:"receiver_type"`
	// ReceiverAccount of the receiver (as registered).
	ReceiverAccount string `json:"receiver_account"`
}

// CreateProfitSharingOrder initiates a profit-sharing split.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_createprofitsharingorder.html
func (w *MiniProgram) CreateProfitSharingOrder(ctx context.Context, req *CreateProfitSharingOrderRequest, opts ...b2bCallOption) error {
	return w.b2bPostRequireSig(ctx, "/retail/B2b/createprofitsharingorder", req, nil, opts...)
}

// QueryProfitSharingOrderRequest selects a profit-sharing order.
type QueryProfitSharingOrderRequest struct {
	// OutTradeNo of the payment order.
	OutTradeNo string `json:"out_trade_no"`
	// ReceiverType of the receiver.
	ReceiverType string `json:"receiver_type"`
	// ReceiverAccount of the receiver.
	ReceiverAccount string `json:"receiver_account"`
	// MchID of the merchant.
	MchID string `json:"mchid"`
}

// QueryProfitSharingOrderResponse is returned by QueryProfitSharingOrder.
type QueryProfitSharingOrderResponse struct {
	ErrResponse
	// RawData of the sharing order payload.
	RawData map[string]any `json:"-"`
}

// QueryProfitSharingOrder queries a profit-sharing order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_queryprofitsharingorder.html
func (w *MiniProgram) QueryProfitSharingOrder(ctx context.Context, req *QueryProfitSharingOrderRequest, opts ...b2bCallOption) (*QueryProfitSharingOrderResponse, error) {
	var result QueryProfitSharingOrderResponse
	if err := w.b2bPost(ctx, "/retail/B2b/queryprofitsharingorder", req, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddProfitSharingAccountRequest registers a profit-sharing receiver account.
type AddProfitSharingAccountRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// ReceiverType of the account.
	ReceiverType string `json:"receiver_type"`
	// ReceiverAccount of the account.
	ReceiverAccount string `json:"receiver_account"`
}

// AddProfitSharingAccount adds a profit-sharing receiver.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_addprofitsharingaccount.html
func (w *MiniProgram) AddProfitSharingAccount(ctx context.Context, req *AddProfitSharingAccountRequest, opts ...b2bCallOption) error {
	return w.b2bPost(ctx, "/retail/B2b/addprofitsharingaccount", req, nil, opts...)
}

// DeleteProfitSharingAccountRequest removes a registered receiver.
type DeleteProfitSharingAccountRequest struct {
	AddProfitSharingAccountRequest
}

// DeleteProfitSharingAccount removes a profit-sharing receiver.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_delprofitsharingaccount.html
func (w *MiniProgram) DeleteProfitSharingAccount(ctx context.Context, req *DeleteProfitSharingAccountRequest, opts ...b2bCallOption) error {
	return w.b2bPost(ctx, "/retail/B2b/delprofitsharingaccount", req, nil, opts...)
}

// QueryProfitSharingAccountRequest selects a registered receiver.
type QueryProfitSharingAccountRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// ReceiverType of the account.
	ReceiverType string `json:"receiver_type"`
	// ReceiverAccount of the account.
	ReceiverAccount string `json:"receiver_account"`
}

// QueryProfitSharingAccountResponse is returned by QueryProfitSharingAccount.
type QueryProfitSharingAccountResponse struct {
	ErrResponse
	// RawData of the account payload.
	RawData map[string]any `json:"-"`
}

// QueryProfitSharingAccount queries a registered profit-sharing receiver.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_queryprofitsharingaccount.html
func (w *MiniProgram) QueryProfitSharingAccount(ctx context.Context, req *QueryProfitSharingAccountRequest, opts ...b2bCallOption) (*QueryProfitSharingAccountResponse, error) {
	var result QueryProfitSharingAccountResponse
	if err := w.b2bPost(ctx, "/retail/B2b/queryprofitsharingaccount", req, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryProfitSharingRemainAmountResponse is returned by
// QueryProfitSharingRemainAmount.
type QueryProfitSharingRemainAmountResponse struct {
	ErrResponse
	// RawData of the remaining amount payload.
	RawData map[string]any `json:"-"`
}

// QueryProfitSharingRemainAmount queries the un-split amount of an order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_queryprofitsharingremainamt.html
func (w *MiniProgram) QueryProfitSharingRemainAmount(ctx context.Context, mchID, outTradeNo string, opts ...b2bCallOption) (*QueryProfitSharingRemainAmountResponse, error) {
	body := map[string]string{"mchid": mchID, "out_trade_no": outTradeNo}
	var result QueryProfitSharingRemainAmountResponse
	if err := w.b2bPost(ctx, "/retail/B2b/queryprofitsharingremainamt", body, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// FinishProfitSharingOrderRequest closes the profit-sharing of an order.
type FinishProfitSharingOrderRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutTradeNo of the order.
	OutTradeNo string `json:"out_trade_no"`
	// Description of the finishing.
	Description string `json:"description"`
}

// FinishProfitSharingOrder marks the profit-sharing of an order as finished.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_finishprofitsharingorder.html
func (w *MiniProgram) FinishProfitSharingOrder(ctx context.Context, req *FinishProfitSharingOrderRequest, opts ...b2bCallOption) error {
	return w.b2bPostRequireSig(ctx, "/retail/B2b/finishprofitsharingorder", req, nil, opts...)
}

// RefundProfitSharingRequest refunds a previously shared amount.
type RefundProfitSharingRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutTradeNo of the original order.
	OutTradeNo string `json:"out_trade_no"`
	// OutRefundNo of the refund.
	OutRefundNo string `json:"out_refund_no"`
	// RefundFee in cents.
	RefundFee int64 `json:"refund_fee"`
	// RawData extra payload.
	RawData map[string]any `json:"-"`
}

// RefundProfitSharing refunds an amount already split to a receiver.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_refundprofitsharing.html
func (w *MiniProgram) RefundProfitSharing(ctx context.Context, req *RefundProfitSharingRequest, opts ...b2bCallOption) error {
	return w.b2bPostRequireSig(ctx, "/retail/B2b/refundprofitsharing", req, nil, opts...)
}

// QueryRefundProfitSharingRequest selects a profit-sharing refund.
type QueryRefundProfitSharingRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutRefundNo of the refund order.
	OutRefundNo string `json:"out_refund_no"`
}

// QueryRefundProfitSharingResponse is returned by QueryRefundProfitSharing.
type QueryRefundProfitSharingResponse struct {
	ErrResponse
	// RawData of the refund payload.
	RawData map[string]any `json:"-"`
}

// QueryRefundProfitSharing queries a profit-sharing refund order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_queryrefundprofitsharingorder.html
func (w *MiniProgram) QueryRefundProfitSharing(ctx context.Context, req *QueryRefundProfitSharingRequest, opts ...b2bCallOption) (*QueryRefundProfitSharingResponse, error) {
	var result QueryRefundProfitSharingResponse
	if err := w.b2bPost(ctx, "/retail/B2b/queryrefundprofitsharingorder", req, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMchBalanceRequest selects the merchant balance to query.
type GetMchBalanceRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
}

// GetMchBalanceResponse is returned by GetMchBalance.
type GetMchBalanceResponse struct {
	ErrResponse
	// RawData of the balance payload.
	RawData map[string]any `json:"-"`
}

// GetMchBalance returns the merchant account balance.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_getmchbalance.html
func (w *MiniProgram) GetMchBalance(ctx context.Context, mchID string, opts ...b2bCallOption) (*GetMchBalanceResponse, error) {
	var result GetMchBalanceResponse
	body := map[string]string{"mchid": mchID}
	if err := w.b2bPost(ctx, "/retail/B2b/getmchbalance", body, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// ManualWithdrawRequest withdraws money from the merchant balance.
type ManualWithdrawRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// WithdrawAmount in cents.
	WithdrawAmount int64 `json:"withdraw_amount"`
	// OutWithdrawNo of the withdrawal.
	OutWithdrawNo string `json:"out_withdraw_no"`
}

// ManualWithdraw performs a manual withdrawal.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_manualwithdraw.html
func (w *MiniProgram) ManualWithdraw(ctx context.Context, req *ManualWithdrawRequest, opts ...b2bCallOption) error {
	return w.b2bPostRequireSig(ctx, "/retail/B2b/withdraw", req, nil, opts...)
}

// QueryWithdrawRequest selects a withdrawal to query.
type QueryWithdrawRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutWithdrawNo of the withdrawal.
	OutWithdrawNo string `json:"out_withdraw_no,omitempty"`
	// WithdrawNo of the WeChat side.
	WithdrawNo string `json:"withdraw_no,omitempty"`
}

// QueryWithdrawResponse is returned by QueryWithdraw.
type QueryWithdrawResponse struct {
	ErrResponse
	// RawData of the withdrawal payload.
	RawData map[string]any `json:"-"`
}

// QueryWithdraw queries a manual withdrawal order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_querywithdraw.html
func (w *MiniProgram) QueryWithdraw(ctx context.Context, req *QueryWithdrawRequest, opts ...b2bCallOption) (*QueryWithdrawResponse, error) {
	var result QueryWithdrawResponse
	if err := w.b2bPost(ctx, "/retail/B2b/querywithdraw", req, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetAutoWithdrawRequest configures automatic withdrawals.
type SetAutoWithdrawRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// RawData of the auto-withdraw settings.
	RawData map[string]any `json:"-"`
}

// SetAutoWithdraw enables/updates the automatic withdrawal rule.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_setautowithdraw.html
func (w *MiniProgram) SetAutoWithdraw(ctx context.Context, body map[string]any, opts ...b2bCallOption) error {
	return w.b2bPostRequireSig(ctx, "/retail/B2b/setautowithdraw", body, nil, opts...)
}

// GetAppKeyRequest requests the merchant app key.
type GetAppKeyRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
}

// GetAppKeyResponse is returned by GetAppKey.
type GetAppKeyResponse struct {
	ErrResponse
	// RawData of the app key payload.
	RawData map[string]any `json:"-"`
}

// GetAppKey returns the app key of a B2b merchant.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_getappkey.html
func (w *MiniProgram) GetAppKey(ctx context.Context, mchID string) (*GetAppKeyResponse, error) {
	var result GetAppKeyResponse
	body := map[string]string{"mchid": mchID}
	if err := w.b2bPost(ctx, "/retail/B2b/getappkey", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RefundOrderRequest refunds a B2b payment order.
type RefundOrderRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutTradeNo of the order.
	OutTradeNo string `json:"out_trade_no"`
	// OutRefundNo of the refund.
	OutRefundNo string `json:"out_refund_no"`
	// RefundFee in cents.
	RefundFee int64 `json:"refund_fee"`
	// TotalFee of the original order in cents.
	TotalFee int64 `json:"total_fee"`
}

// RefundOrder refunds a B2b payment order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_refundorder.html
func (w *MiniProgram) RefundOrder(ctx context.Context, req *RefundOrderRequest, opts ...b2bCallOption) error {
	return w.b2bPost(ctx, "/retail/B2b/refund", req, nil, opts...)
}

// GetB2bOrderRequest selects a B2b payment order.
type GetB2bOrderRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutTradeNo of the order (alternative to TransactionID).
	OutTradeNo string `json:"out_trade_no,omitempty"`
	// TransactionID of the WeChat Pay order.
	TransactionID string `json:"transaction_id,omitempty"`
}

// GetB2bOrderResponse is returned by GetB2bOrder.
type GetB2bOrderResponse struct {
	ErrResponse
	// RawData of the order payload.
	RawData map[string]any `json:"-"`
}

// GetB2bOrder queries a B2b payment order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_getorder.html
func (w *MiniProgram) GetB2bOrder(ctx context.Context, req *GetB2bOrderRequest, opts ...b2bCallOption) (*GetB2bOrderResponse, error) {
	var result GetB2bOrderResponse
	if err := w.b2bPost(ctx, "/retail/B2b/getorder", req, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetB2bRefundRequest selects a B2b refund order.
type GetB2bRefundRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutRefundNo of the refund.
	OutRefundNo string `json:"out_refund_no,omitempty"`
	// RefundID of the WeChat side.
	RefundID string `json:"refund_id,omitempty"`
}

// GetB2bRefundResponse is returned by GetB2bRefund.
type GetB2bRefundResponse struct {
	ErrResponse
	// RawData of the refund payload.
	RawData map[string]any `json:"-"`
}

// GetB2bRefund queries a B2b refund order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_getrefund.html
func (w *MiniProgram) GetB2bRefund(ctx context.Context, req *GetB2bRefundRequest, opts ...b2bCallOption) (*GetB2bRefundResponse, error) {
	var result GetB2bRefundResponse
	if err := w.b2bPost(ctx, "/retail/B2b/getrefund", req, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// DownloadB2bBillRequest selects a B2b bill to download.
type DownloadB2bBillRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// BillDate in yyyyMMdd.
	BillDate string `json:"bill_date,omitempty"`
	// RawData of the remaining selectors.
	RawData map[string]any `json:"-"`
}

// DownloadB2bBillResponse is returned by DownloadB2bBill.
type DownloadB2bBillResponse struct {
	ErrResponse
	// URL of the bill file.
	URL string `json:"url,omitempty"`
	// RawData of the response.
	RawData map[string]any `json:"-"`
}

// DownloadB2bBill downloads the merchant bill.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_downloadbill.html
func (w *MiniProgram) DownloadB2bBill(ctx context.Context, body map[string]any, opts ...b2bCallOption) (*DownloadB2bBillResponse, error) {
	var result DownloadB2bBillResponse
	if err := w.b2bPost(ctx, "/retail/B2b/downloadbill", body, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMchInfoRequest selects a B2b merchant profile.
type GetMchInfoRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
}

// GetMchInfoResponse is returned by GetMchInfo.
type GetMchInfoResponse struct {
	ErrResponse
	// RawData of the merchant payload.
	RawData map[string]any `json:"-"`
}

// GetMchInfo returns the B2b merchant profile.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_getmchinfo.html
func (w *MiniProgram) GetMchInfo(ctx context.Context, mchID string) (*GetMchInfoResponse, error) {
	var result GetMchInfoResponse
	body := map[string]string{"mchid": mchID}
	if err := w.b2bPost(ctx, "/retail/B2b/getmchinfo", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateWqfLinkRequest opens a 微信支付分 (WeChat Pay Score) link.
type CreateWqfLinkRequest struct {
	// RequestNo of the bank-transfer activation order.
	RequestNo string `json:"request_no"`
}

// CreateWqfLink requests a WeChat-Pay-Fen (支付分) activation link.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_createwqflink.html
func (w *MiniProgram) CreateWqfLink(ctx context.Context, requestNo string) error {
	body := map[string]string{"request_no": requestNo}
	return w.b2bPost(ctx, "/retail/B2b/createwqflink", body, nil)
}

// GetWqfChargeFeeResponse is returned by GetWqfChargeFee.
type GetWqfChargeFeeResponse struct {
	ErrResponse
	// RawData of the fee payload.
	RawData map[string]any `json:"-"`
}

// GetWqfChargeFee returns the 微信支付分 service fee configuration.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_getwqfchargefee.html
func (w *MiniProgram) GetWqfChargeFee(ctx context.Context, body map[string]any) (*GetWqfChargeFeeResponse, error) {
	var result GetWqfChargeFeeResponse
	if err := w.b2bPost(ctx, "/retail/B2b/getwqfchargefee", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWqfChargeFee updates the 微信支付分 service fee configuration.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_updatewqfchargefee.html
func (w *MiniProgram) UpdateWqfChargeFee(ctx context.Context, body map[string]any) error {
	return w.b2bPost(ctx, "/retail/B2b/updatewqfchargefee", body, nil)
}

// SetMchProfitRateRequest configures the merchant profit-sharing rate.
type SetMchProfitRateRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// ProfitRate percentage.
	ProfitRate int `json:"profit_rate,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// SetMchProfitRate updates the profit rate of a B2b merchant.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_setmchprofitrate.html
func (w *MiniProgram) SetMchProfitRate(ctx context.Context, body map[string]any) error {
	return w.b2bPost(ctx, "/retail/B2b/setmchprofitrate", body, nil)
}

// CloseB2bOrderRequest closes a B2b payment order.
type CloseB2bOrderRequest struct {
	// MchID of the merchant.
	MchID string `json:"mchid"`
	// OutTradeNo of the order.
	OutTradeNo string `json:"out_trade_no"`
}

// CloseB2bOrder closes an unpaid B2b payment order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_closeb2border.html
func (w *MiniProgram) CloseB2bOrder(ctx context.Context, req *CloseB2bOrderRequest, opts ...b2bCallOption) error {
	return w.b2bPost(ctx, "/retail/B2b/closeb2border", req, nil, opts...)
}
