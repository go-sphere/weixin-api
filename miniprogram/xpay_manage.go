package miniprogram

import (
	"context"
)

// ============================================================
// Withdraw (提现).
// ============================================================

// XPayCreateWithdrawOrderRequest creates a merchant withdrawal order.
type XPayCreateWithdrawOrderRequest struct {
	// WithdrawNO of the withdrawal (length 8-32, letters/digits/_/-).
	WithdrawNO string `json:"withdraw_no"`
	// WithdrawAmount in yuan, e.g. "0.01" for 1 cent.
	WithdrawAmount string `json:"withdraw_amount"`
	// Env of the withdrawal.
	Env XPayEnv `json:"env"`
}

// XPayCreateWithdrawOrderResponse is returned by XPayCreateWithdrawOrder.
type XPayCreateWithdrawOrderResponse struct {
	ErrResponse
	// WithdrawNO echoed back.
	WithdrawNO string `json:"withdraw_no"`
	// WxWithdrawNO of the WeChat side.
	WxWithdrawNO string `json:"wx_withdraw_no"`
}

// XPayCreateWithdrawOrder creates a withdrawal order moving the merchant
// balance to the bound bank account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_create_withdraw_order.html
func (w *MiniProgram) XPayCreateWithdrawOrder(ctx context.Context, req *XPayCreateWithdrawOrderRequest) (*XPayCreateWithdrawOrderResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/create_withdraw_order", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayCreateWithdrawOrderResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayQueryWithdrawOrderRequest selects a withdrawal order to query.
type XPayQueryWithdrawOrderRequest struct {
	// WithdrawNO of the merchant (alternative to WxWithdrawNO).
	WithdrawNO string `json:"withdraw_no,omitempty"`
	// WxWithdrawNO of the WeChat side.
	WxWithdrawNO string `json:"wx_withdraw_no,omitempty"`
	// Env of the withdrawal.
	Env XPayEnv `json:"env"`
}

// XPayQueryWithdrawOrderResponse is returned by XPayQueryWithdrawOrder.
type XPayQueryWithdrawOrderResponse struct {
	ErrResponse
	// WithdrawNO of the merchant.
	WithdrawNO string `json:"withdraw_no,omitempty"`
	// Status: 1 withdrawing, 2 succeeded, 3 failed.
	Status int `json:"status"`
	// WithdrawAmount in yuan.
	WithdrawAmount string `json:"withdraw_amount,omitempty"`
	// WxWithdrawNo of the WeChat side.
	WxWithdrawNo string `json:"wx_withdraw_no,omitempty"`
	// WithdrawSuccessTimestamp (unix seconds).
	WithdrawSuccessTimestamp int64 `json:"withdraw_success_timestamp,omitempty"`
	// CreateTime of the withdrawal.
	CreateTime string `json:"create_time,omitempty"`
	// FailReason of a failed withdrawal.
	FailReason string `json:"failReason,omitempty"`
}

// XPayQueryWithdrawOrder queries the state of a withdrawal order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_withdraw_order.html
func (w *MiniProgram) XPayQueryWithdrawOrder(ctx context.Context, req *XPayQueryWithdrawOrderRequest) (*XPayQueryWithdrawOrderResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_withdraw_order", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryWithdrawOrderResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// Bill download.
// ============================================================

// XPayDownloadBillRequest selects a bill period to download.
type XPayDownloadBillRequest struct {
	// BeginDs of the bill period in yyyymmdd.
	BeginDs string `json:"begin_ds"`
	// EndDs of the bill period in yyyymmdd.
	EndDs string `json:"end_ds"`
}

// XPayDownloadBillResponse is returned by XPayDownloadBill. The first call
// triggers bill generation; poll again until the URL is present.
type XPayDownloadBillResponse struct {
	ErrResponse
	// URL of the generated bill file.
	URL string `json:"url,omitempty"`
}

// XPayDownloadBill downloads the Mini Program virtual-payment bill. Amounts in
// the bill are in cents.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_download_bill.html
func (w *MiniProgram) XPayDownloadBill(ctx context.Context, req *XPayDownloadBillRequest) (*XPayDownloadBillResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/download_bill", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayDownloadBillResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayStartDownloadOrderRequest launches an order-detail export task.
type XPayStartDownloadOrderRequest struct {
	// Env of the orders.
	Env XPayEnv `json:"env"`
	// BeginDs in yyyymmdd.
	BeginDs string `json:"begin_ds"`
	// EndDs in yyyymmdd.
	EndDs string `json:"end_ds"`
}

// XPayStartDownloadOrder starts the async export of Mini Program order detail.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_start_download_order.html
func (w *MiniProgram) XPayStartDownloadOrder(ctx context.Context, req *XPayStartDownloadOrderRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/start_download_order", req, "", false)
	return err
}

// XPayQueryDownloadOrderRequest selects the export task to poll.
type XPayQueryDownloadOrderRequest struct {
	XPayStartDownloadOrderRequest
}

// XPayQueryDownloadOrderResponse is returned by XPayQueryDownloadOrder.
type XPayQueryDownloadOrderResponse struct {
	ErrResponse
	// URL of the exported order file once ready.
	URL string `json:"url,omitempty"`
}

// XPayQueryDownloadOrder polls the result of an order-detail export task.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_download_order.html
func (w *MiniProgram) XPayQueryDownloadOrder(ctx context.Context, req *XPayQueryDownloadOrderRequest) (*XPayQueryDownloadOrderResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_download_order", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryDownloadOrderResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayDownloadIOSSettlementBillRequest selects an iOS settlement bill period.
type XPayDownloadIOSSettlementBillRequest struct {
	// StartMonth in yyyymm.
	StartMonth string `json:"start_month"`
	// EndMonth in yyyymm.
	EndMonth string `json:"end_month"`
}

// XPayDownloadIOSSettlementBillResponse is returned by
// XPayDownloadIOSSettlementBill.
type XPayDownloadIOSSettlementBillResponse struct {
	ErrResponse
	// URL of the iOS settlement bill.
	URL string `json:"url,omitempty"`
}

// XPayDownloadIOSSettlementBill downloads the iOS Apple-payment settlement
// bill for a month range.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_download_ios_settlement_bill.html
func (w *MiniProgram) XPayDownloadIOSSettlementBill(ctx context.Context, req *XPayDownloadIOSSettlementBillRequest) (*XPayDownloadIOSSettlementBillResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/download_ios_settlement_bill", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayDownloadIOSSettlementBillResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// Merchant balance.
// ============================================================

// XPayQueryBizBalanceRequest queries the withdrawable merchant balance.
type XPayQueryBizBalanceRequest struct {
	// Env of the account (only used for signing).
	Env XPayEnv `json:"env"`
}

// XPayBizBalanceAvailable is the withdrawable balance payload.
type XPayBizBalanceAvailable struct {
	// Amount withdrawable in yuan.
	Amount string `json:"amount"`
	// CurrencyCode, usually "CNY".
	CurrencyCode string `json:"currency_code"`
}

// XPayQueryBizBalanceResponse is returned by XPayQueryBizBalance.
type XPayQueryBizBalanceResponse struct {
	ErrResponse
	// BalanceAvailable of the merchant.
	BalanceAvailable *XPayBizBalanceAvailable `json:"balance_available,omitempty"`
}

// XPayQueryBizBalance returns the withdrawable balance of the merchant
// virtual-payment account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_biz_balance.html
func (w *MiniProgram) XPayQueryBizBalance(ctx context.Context, req *XPayQueryBizBalanceRequest) (*XPayQueryBizBalanceResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_biz_balance", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryBizBalanceResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// Goods (道具) batch management.
// ============================================================

// XPayGoodsItem is one virtual good (道具).
type XPayGoodsItem struct {
	// ID of the good (letters/digits/_/-, up to 64).
	ID string `json:"id"`
	// Name of the good.
	Name string `json:"name,omitempty"`
	// Price of the good in cents.
	Price int `json:"price,omitempty"`
	// Remark of the good.
	Remark string `json:"remark,omitempty"`
	// ItemURL of the good image (jpg/png).
	ItemURL string `json:"item_url,omitempty"`
	// Status of the batch task for this good: 0 uploading, 1 id exists,
	// 2 success, 3 failed.
	Status int `json:"status,omitempty"`
	// ErrMsg of a failed good.
	ErrMsg string `json:"errmsg,omitempty"`
}

// XPayStartUploadGoodsRequest starts a batch goods upload task.
type XPayStartUploadGoodsRequest struct {
	// UploadItem list of goods to upload.
	UploadItem []*XPayGoodsItem `json:"upload_item"`
	// Env of the goods.
	Env XPayEnv `json:"env"`
}

// XPayStartUploadGoods starts the batch upload of virtual goods into the
// sandbox/development environment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_start_upload_goods.html
func (w *MiniProgram) XPayStartUploadGoods(ctx context.Context, req *XPayStartUploadGoodsRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/start_upload_goods", req, "", false)
	return err
}

// XPayQueryUploadGoodsRequest polls the upload task.
type XPayQueryUploadGoodsRequest struct {
	// Env of the goods.
	Env XPayEnv `json:"env"`
}

// XPayQueryUploadGoodsResponse is returned by XPayQueryUploadGoods.
type XPayQueryUploadGoodsResponse struct {
	ErrResponse
	// UploadItem is the per-good upload result list.
	UploadItem []*XPayGoodsItem `json:"upload_item"`
	// Status of the task: 0 none, 1 running, 2 failed/partial, 3 success.
	Status int `json:"status"`
}

// XPayQueryUploadGoods polls the batch goods upload task status.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_upload_goods.html
func (w *MiniProgram) XPayQueryUploadGoods(ctx context.Context, req *XPayQueryUploadGoodsRequest) (*XPayQueryUploadGoodsResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_upload_goods", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryUploadGoodsResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayPublishGoodsItem is one good of a publish task.
type XPayPublishGoodsItem struct {
	// ID of the good to publish (must exist in the dev environment).
	ID string `json:"id"`
	// PublishStatus: 0 publishing, 1 id exists, 2 success, 3 failed.
	PublishStatus int `json:"publish_status,omitempty"`
	// ErrMsg of a failed publish.
	ErrMsg string `json:"errmsg,omitempty"`
}

// XPayStartPublishGoodsRequest starts a batch publish task.
type XPayStartPublishGoodsRequest struct {
	// Env of the goods.
	Env XPayEnv `json:"env"`
	// PublishItem of the goods to publish.
	PublishItem []*XPayPublishGoodsItem `json:"publish_item"`
}

// XPayStartPublishGoods publishes previously uploaded sandbox goods to the
// production store.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_start_publish_goods.html
func (w *MiniProgram) XPayStartPublishGoods(ctx context.Context, req *XPayStartPublishGoodsRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/start_publish_goods", req, "", false)
	return err
}

// XPayQueryPublishGoodsRequest polls the publish task.
type XPayQueryPublishGoodsRequest struct {
	// Env of the goods.
	Env XPayEnv `json:"env"`
}

// XPayQueryPublishGoodsResponse is returned by XPayQueryPublishGoods.
type XPayQueryPublishGoodsResponse struct {
	ErrResponse
	// PublishItem is the per-good publish result list.
	PublishItem []*XPayPublishGoodsItem `json:"publish_item"`
	// Status of the task: 0 none, 1 running, 2 failed/partial, 3 success.
	Status int `json:"status"`
}

// XPayQueryPublishGoods polls the batch goods publish task status.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_publish_goods.html
func (w *MiniProgram) XPayQueryPublishGoods(ctx context.Context, req *XPayQueryPublishGoodsRequest) (*XPayQueryPublishGoodsResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_publish_goods", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryPublishGoodsResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// Ad funds (广告金) management.
// ============================================================

// XPayTransferAccount is an ad-fund transfer (充值) account.
type XPayTransferAccount struct {
	// Name of the transfer account.
	TransferAccountName string `json:"transfer_account_name"`
	// UID of the transfer account.
	TransferAccountUID int64 `json:"transfer_account_uid"`
	// AgencyID of the agency.
	TransferAccountAgencyID int64 `json:"transfer_account_agency_id"`
	// AgencyName of the agency.
	TransferAccountAgencyName string `json:"transfer_account_agency_name"`
	// State of the review: 0 pending, 1 passed, 2 rejected.
	State int `json:"state"`
	// BindResult: 1 success, 2 failed.
	BindResult int `json:"bind_result"`
	// ErrorMsg of a failed binding.
	ErrorMsg string `json:"error_msg"`
}

// XPayQueryTransferAccountRequest lists the ad-fund accounts.
type XPayQueryTransferAccountRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
}

// XPayQueryTransferAccountResponse is returned by XPayQueryTransferAccount.
type XPayQueryTransferAccountResponse struct {
	ErrResponse
	// AcctList of the bound transfer accounts.
	AcctList []*XPayTransferAccount `json:"acct_list"`
}

// XPayQueryTransferAccount lists the ad-fund recharge accounts bound to the
// Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_transfer_account.html
func (w *MiniProgram) XPayQueryTransferAccount(ctx context.Context, req *XPayQueryTransferAccountRequest) (*XPayQueryTransferAccountResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_transfer_account", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryTransferAccountResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayBindTransferAccountRequest binds an ad-fund recharge account. Provide
// either the UID or the organisation name.
type XPayBindTransferAccountRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// UID of the transfer account.
	TransferAccountUID int64 `json:"transfer_account_uid,omitempty"`
	// OrgName of the transfer account subject.
	TransferAccountOrgName string `json:"transfer_account_org_name,omitempty"`
}

// XPayBindTransferAccount binds an ad-fund recharge account to the Mini
// Program for a review.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_bind_transfer_accout.html
func (w *MiniProgram) XPayBindTransferAccount(ctx context.Context, req *XPayBindTransferAccountRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/bind_transfer_accout", req, "", false)
	return err
}

// XPayAdverFundsFilter narrows an ad-fund record query.
type XPayAdverFundsFilter struct {
	// SettleBegin of the settlement window (unix seconds).
	SettleBegin int64 `json:"settle_begin,omitempty"`
	// SettleEnd of the settlement window (unix seconds).
	SettleEnd int64 `json:"settle_end,omitempty"`
	// FundType: 0 normal, 1 ad incentive, 2 targeted incentive.
	FundType int `json:"fund_type,omitempty"`
}

// XPayAdverFundsRecord is one ad-fund grant record.
type XPayAdverFundsRecord struct {
	// SettleBegin of the settlement window.
	SettleBegin int64 `json:"settle_begin"`
	// SettleEnd of the settlement window.
	SettleEnd int64 `json:"settle_end"`
	// TotalAmount granted, in cents.
	TotalAmount int64 `json:"total_amount"`
	// RemainAmount still available, in cents.
	RemainAmount int64 `json:"remain_amount"`
	// ExpireTime of the fund (unix seconds).
	ExpireTime int64 `json:"expire_time"`
	// FundType of the grant.
	FundType int `json:"fund_type"`
	// FundID of the grant.
	FundID string `json:"fund_id"`
}

// XPayQueryAdverFundsRequest queries the ad-fund grant records.
type XPayQueryAdverFundsRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// Page number (>=1).
	Page int `json:"page,omitempty"`
	// PageSize of each page.
	PageSize int `json:"page_size,omitempty"`
	// Filter of the query.
	Filter *XPayAdverFundsFilter `json:"filter,omitempty"`
}

// XPayQueryAdverFundsResponse is returned by XPayQueryAdverFunds.
type XPayQueryAdverFundsResponse struct {
	ErrResponse
	// AdverFundsList of the granted records.
	AdverFundsList []*XPayAdverFundsRecord `json:"adver_funds_list"`
	// TotalPage count.
	TotalPage int `json:"total_page"`
}

// XPayQueryAdverFunds lists the ad-fund grant records of the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_adver_funds.html
func (w *MiniProgram) XPayQueryAdverFunds(ctx context.Context, req *XPayQueryAdverFundsRequest) (*XPayQueryAdverFundsResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_adver_funds", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryAdverFundsResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayCreateFundsBillRequest recharges ad funds into a transfer account.
type XPayCreateFundsBillRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// TransferAmount in cents.
	TransferAmount int64 `json:"transfer_amount"`
	// TransferAccountUID of the receiving account.
	TransferAccountUID int64 `json:"transfer_account_uid"`
	// TransferAccountName of the receiving account.
	TransferAccountName string `json:"transfer_account_name"`
	// TransferAccountAgencyID of the agency.
	TransferAccountAgencyID int64 `json:"transfer_account_agency_id"`
	// RequestID idempotency key (up to 1024 chars).
	RequestID string `json:"request_id"`
	// SettleBegin of the settlement window.
	SettleBegin int64 `json:"settle_begin"`
	// SettleEnd of the settlement window.
	SettleEnd int64 `json:"settle_end"`
	// AuthorizeAdvertise: 0 no, 1 yes.
	AuthorizeAdvertise int `json:"authorize_advertise"`
	// FundType of the recharge.
	FundType int `json:"fund_type"`
}

// XPayCreateFundsBillResponse is returned by XPayCreateFundsBill.
type XPayCreateFundsBillResponse struct {
	ErrResponse
	// BillID of the recharge order.
	BillID string `json:"bill_id"`
}

// XPayCreateFundsBill recharges ad funds into a bound transfer account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_create_funds_bill.html
func (w *MiniProgram) XPayCreateFundsBill(ctx context.Context, req *XPayCreateFundsBillRequest) (*XPayCreateFundsBillResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/create_funds_bill", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayCreateFundsBillResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayFundsBillFilter narrows an ad-fund recharge record query.
type XPayFundsBillFilter struct {
	// OperTimeBegin of the recharge window (unix seconds).
	OperTimeBegin int64 `json:"oper_time_begin,omitempty"`
	// OperTimeEnd of the recharge window (unix seconds).
	OperTimeEnd int64 `json:"oper_time_end,omitempty"`
	// BillID of a specific recharge.
	BillID string `json:"bill_id,omitempty"`
	// RequestID of a specific recharge.
	RequestID string `json:"request_id,omitempty"`
}

// XPayFundsBillRecord is one ad-fund recharge record.
type XPayFundsBillRecord struct {
	// BillID of the recharge.
	BillID string `json:"bill_id"`
	// OperTime of the recharge (unix seconds).
	OperTime int64 `json:"oper_time"`
	// SettleBegin of the linked fund settlement.
	SettleBegin int64 `json:"settle_begin"`
	// SettleEnd of the linked fund settlement.
	SettleEnd int64 `json:"settle_end"`
	// FundID of the linked fund grant.
	FundID string `json:"fund_id"`
	// TransferAccountName of the recharge target.
	TransferAccountName string `json:"transfer_account_name"`
	// TransferAccountUID of the recharge target.
	TransferAccountUID int64 `json:"transfer_account_uid"`
	// TransferAmount in cents.
	TransferAmount int64 `json:"transfer_amount"`
	// Status: 0 recharging, 1 success, 2 failed.
	Status int `json:"status"`
	// RequestID of the recharge.
	RequestID string `json:"request_id"`
}

// XPayQueryFundsBillRequest queries the ad-fund recharge records.
type XPayQueryFundsBillRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// Page number (>=1).
	Page int `json:"page"`
	// PageSize of each page.
	PageSize int `json:"page_size"`
	// Filter of the query.
	Filter XPayFundsBillFilter `json:"filter"`
}

// XPayQueryFundsBillResponse is returned by XPayQueryFundsBill.
type XPayQueryFundsBillResponse struct {
	ErrResponse
	// BillList of the recharge records.
	BillList []*XPayFundsBillRecord `json:"bill_list"`
	// TotalPage count.
	TotalPage int `json:"total_page"`
}

// XPayQueryFundsBill lists the ad-fund recharge records.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_funds_bill.html
func (w *MiniProgram) XPayQueryFundsBill(ctx context.Context, req *XPayQueryFundsBillRequest) (*XPayQueryFundsBillResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_funds_bill", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryFundsBillResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayRecoverBillFilter narrows an ad-fund recovery record query.
type XPayRecoverBillFilter struct {
	// RecoverTimeBegin of the recovery window (unix seconds).
	RecoverTimeBegin int64 `json:"recover_time_begin,omitempty"`
	// RecoverTimeEnd of the recovery window (unix seconds).
	RecoverTimeEnd int64 `json:"recover_time_end,omitempty"`
	// BillID of a specific recovery.
	BillID string `json:"bill_id,omitempty"`
}

// XPayRecoverBillRecord is one ad-fund recovery record.
type XPayRecoverBillRecord struct {
	// BillID of the recovery.
	BillID string `json:"bill_id"`
	// RecoverTime of the recovery (unix seconds).
	RecoverTime int64 `json:"recover_time"`
	// SettleBegin of the linked fund settlement.
	SettleBegin int64 `json:"settle_begin"`
	// SettleEnd of the linked fund settlement.
	SettleEnd int64 `json:"settle_end"`
	// FundID of the linked fund grant.
	FundID string `json:"fund_id"`
	// RecoverAccountName of the recovering account.
	RecoverAccountName string `json:"recover_account_name"`
	// RecoverAmount in cents.
	RecoverAmount int64 `json:"recover_amount"`
	// RefundOrderList of the refund orders behind the recovery.
	RefundOrderList []string `json:"refund_order_list"`
}

// XPayQueryRecoverBillRequest queries the ad-fund recovery records.
type XPayQueryRecoverBillRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// Page number (>=1).
	Page int `json:"page"`
	// PageSize of each page.
	PageSize int `json:"page_size"`
	// Filter of the query.
	Filter XPayRecoverBillFilter `json:"filter"`
}

// XPayQueryRecoverBillResponse is returned by XPayQueryRecoverBill.
type XPayQueryRecoverBillResponse struct {
	ErrResponse
	// BillList of the recovery records.
	BillList []*XPayRecoverBillRecord `json:"bill_list"`
	// TotalPage count.
	TotalPage int `json:"total_page"`
}

// XPayQueryRecoverBill lists the ad-fund recovery records.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_recover_bill.html
func (w *MiniProgram) XPayQueryRecoverBill(ctx context.Context, req *XPayQueryRecoverBillRequest) (*XPayQueryRecoverBillResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_recover_bill", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQueryRecoverBillResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayDownloadAdverFundsOrderRequest downloads the merchant orders behind an
// ad-fund grant.
type XPayDownloadAdverFundsOrderRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// FundID of the ad-fund grant.
	FundID string `json:"fund_id"`
}

// XPayDownloadAdverFundsOrderResponse is returned by
// XPayDownloadAdverFundsOrder.
type XPayDownloadAdverFundsOrderResponse struct {
	ErrResponse
	// URL of the merchant order download.
	URL string `json:"url,omitempty"`
}

// XPayDownloadAdverFundsOrder downloads the merchant orders funded by an
// ad-fund grant.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_download_adverfunds_order.html
func (w *MiniProgram) XPayDownloadAdverFundsOrder(ctx context.Context, req *XPayDownloadAdverFundsOrderRequest) (*XPayDownloadAdverFundsOrderResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/download_adverfunds_order", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayDownloadAdverFundsOrderResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
