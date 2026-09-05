package miniprogram

import (
	"context"
)

// ============================================================
// Complaints (投诉).
// ============================================================

// XPayComplaintOrderInfo links a complaint to its orders.
type XPayComplaintOrderInfo struct {
	// TransactionID of the WeChat Pay transaction.
	TransactionID string `json:"transaction_id"`
	// MchOrderNo of the merchant.
	MchOrderNo string `json:"mch_order_no"`
	// RefundID of the refund if any.
	RefundID string `json:"refund_id"`
}

// XPayComplaintMedia is an evidence media attached to a complaint.
type XPayComplaintMedia struct {
	// MediaType of the media.
	MediaType int `json:"media_type"`
	// MediaURL of the media.
	MediaURL string `json:"media_url"`
	// MediaTags of the media.
	MediaTags []string `json:"media_tags"`
}

// XPayComplaintItem is the detailed payload of one complaint.
type XPayComplaintItem struct {
	// ComplaintID of the complaint.
	ComplaintID string `json:"complaint_id"`
	// ComplaintTime of the complaint.
	ComplaintTime string `json:"complaint_time"`
	// ComplaintDetail text of the complaint.
	ComplaintDetail string `json:"complaint_detail"`
	// ComplaintState: PENDING / PROCESSING / PROCESSED.
	ComplaintState string `json:"complaint_state"`
	// PayerPhone of the complainant.
	PayerPhone string `json:"payer_phone"`
	// PayerOpenID of the complainant.
	PayerOpenID string `json:"payer_openid"`
	// ComplaintOrderInfo of the linked orders.
	ComplaintOrderInfo []*XPayComplaintOrderInfo `json:"complaint_order_info"`
	// ComplaintFullRefunded indicates whether the order was fully refunded.
	ComplaintFullRefunded bool `json:"complaint_full_refunded"`
	// IncomingUserResponse indicates a pending user message.
	IncomingUserResponse bool `json:"incoming_user_response"`
	// UserComplaintTimes of the user.
	UserComplaintTimes int `json:"user_complaint_times"`
	// ComplaintMediaList of the user evidence.
	ComplaintMediaList []*XPayComplaintMedia `json:"complaint_media_list"`
}

// XPayGetComplaintListRequest lists the complaints of a date window.
type XPayGetComplaintListRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// BeginDate in yyyy-mm-dd.
	BeginDate string `json:"begin_date"`
	// EndDate in yyyy-mm-dd.
	EndDate string `json:"end_date"`
	// Offset from 0.
	Offset int `json:"offset"`
	// Limit of the page.
	Limit int `json:"limit"`
}

// XPayGetComplaintListResponse is returned by XPayGetComplaintList.
type XPayGetComplaintListResponse struct {
	ErrResponse
	// Total number of matching complaints.
	Total int `json:"total"`
	// Complaints of the page.
	Complaints []*XPayComplaintItem `json:"complaints"`
}

// XPayGetComplaintList lists the user complaints of the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_get_complaint_list.html
func (w *MiniProgram) XPayGetComplaintList(ctx context.Context, req *XPayGetComplaintListRequest) (*XPayGetComplaintListResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/get_complaint_list", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayGetComplaintListResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayGetComplaintDetailRequest selects a complaint.
type XPayGetComplaintDetailRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// ComplaintID of the complaint.
	ComplaintID string `json:"complaint_id"`
}

// XPayGetComplaintDetailResponse is returned by XPayGetComplaintDetail.
type XPayGetComplaintDetailResponse struct {
	ErrResponse
	// Complaint of the query.
	Complaint *XPayComplaintItem `json:"complaint,omitempty"`
}

// XPayGetComplaintDetail returns the full detail of one complaint.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_get_complaint_detail.html
func (w *MiniProgram) XPayGetComplaintDetail(ctx context.Context, req *XPayGetComplaintDetailRequest) (*XPayGetComplaintDetailResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/get_complaint_detail", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayGetComplaintDetailResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayNegotiationRecord is one entry of a complaint negotiation history.
type XPayNegotiationRecord struct {
	// LogID of the operation.
	LogID string `json:"log_id"`
	// Operator of the operation.
	Operator string `json:"operator"`
	// OperateTime of the operation.
	OperateTime string `json:"operate_time"`
	// OperateType of the operation.
	OperateType string `json:"operate_type"`
	// OperateDetails of the operation.
	OperateDetails string `json:"operate_details"`
	// ComplaintMediaList of the uploaded evidence.
	ComplaintMediaList []*XPayComplaintMedia `json:"complaint_media_list"`
}

// XPayGetNegotiationHistoryRequest selects the negotiation history of a
// complaint.
type XPayGetNegotiationHistoryRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// ComplaintID of the complaint.
	ComplaintID string `json:"complaint_id"`
	// Offset from 0.
	Offset int `json:"offset"`
	// Limit of the page.
	Limit int `json:"limit"`
}

// XPayGetNegotiationHistoryResponse is returned by
// XPayGetNegotiationHistory.
type XPayGetNegotiationHistoryResponse struct {
	ErrResponse
	// Total of the history records.
	Total int `json:"total"`
	// History of the page.
	History []*XPayNegotiationRecord `json:"history"`
}

// XPayGetNegotiationHistory returns the merchant/user negotiation history of
// a complaint.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_get_negotiation_history.html
func (w *MiniProgram) XPayGetNegotiationHistory(ctx context.Context, req *XPayGetNegotiationHistoryRequest) (*XPayGetNegotiationHistoryResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/get_negotiation_history", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayGetNegotiationHistoryResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayResponseComplaintRequest replies to a user complaint.
type XPayResponseComplaintRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// ComplaintID of the complaint.
	ComplaintID string `json:"complaint_id"`
	// ResponseContent of the reply.
	ResponseContent string `json:"response_content"`
	// ResponseImages are the file ids returned by XPayUploadVPFile.
	ResponseImages []string `json:"response_images"`
}

// XPayResponseComplaint replies to the user inside a complaint.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_response_complaint.html
func (w *MiniProgram) XPayResponseComplaint(ctx context.Context, req *XPayResponseComplaintRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/response_complaint", req, "", false)
	return err
}

// XPayCompleteComplaintRequest marks a complaint as fully handled.
type XPayCompleteComplaintRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// ComplaintID of the complaint.
	ComplaintID string `json:"complaint_id"`
}

// XPayCompleteComplaint marks a complaint as processed and closes it.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_complete_complaint.html
func (w *MiniProgram) XPayCompleteComplaint(ctx context.Context, req *XPayCompleteComplaintRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/complete_complaint", req, "", false)
	return err
}

// XPayUploadVPFileRequest uploads a media file used as complaint evidence.
type XPayUploadVPFileRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// Base64Img is the base64 encoded image (up to 1 MB).
	Base64Img string `json:"base64_img,omitempty"`
	// ImgURL of the image (downloadable, up to 2 MB). Preferred over
	// Base64Img.
	ImgURL string `json:"img_url,omitempty"`
	// FileName of the uploaded image.
	FileName string `json:"file_name"`
}

// XPayUploadVPFileResponse is returned by XPayUploadVPFile.
type XPayUploadVPFileResponse struct {
	ErrResponse
	// FileID of the uploaded media used in XPayResponseComplaint.
	FileID string `json:"file_id"`
}

// XPayUploadVPFile uploads an image to be attached to a complaint reply.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_upload_vp_file.html
func (w *MiniProgram) XPayUploadVPFile(ctx context.Context, req *XPayUploadVPFileRequest) (*XPayUploadVPFileResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/upload_vp_file", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayUploadVPFileResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayGetUploadFileSignRequest fetches the signature header used to download a
// WeChat-Pay complaint image.
type XPayGetUploadFileSignRequest struct {
	// Env of the account.
	Env XPayEnv `json:"env"`
	// WxpayURL of the WeChat Pay image, format
	// "https://api.mch.weixin.qq.com/v3/merchant-service/images/{id}".
	WxpayURL string `json:"wxpay_url"`
	// ConvertCOS asks for a temporary COS download link (valid 30 min).
	ConvertCOS bool `json:"convert_cos"`
	// ComplaintID of the related complaint.
	ComplaintID string `json:"complaint_id"`
}

// XPayGetUploadFileSignResponse is returned by XPayGetUploadFileSign.
type XPayGetUploadFileSignResponse struct {
	ErrResponse
	// Sign value to place in the Authorization header when downloading.
	Sign string `json:"sign"`
	// CosURL returned when ConvertCOS is true (valid 30 min).
	CosURL string `json:"cos_url"`
}

// XPayGetUploadFileSign returns the Authorization header value needed to
// download WeChat-Pay complaint feedback images.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_get_upload_file_sign.html
func (w *MiniProgram) XPayGetUploadFileSign(ctx context.Context, req *XPayGetUploadFileSignRequest) (*XPayGetUploadFileSignResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/get_upload_file_sign", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayGetUploadFileSignResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// iOS Apple-payment subscriptions (订阅).
// ============================================================

// XPayQuerySubscribeContractRequest queries an Apple-payment subscription
// contract.
type XPayQuerySubscribeContractRequest struct {
	// Mode: 1 开通/自动续费查询, 2 其他（具体值按文档约定）.
	Mode int `json:"mode"`
	// ProductID of the subscribed product.
	ProductID string `json:"product_id"`
	// OutContractCode of the developer side contract code.
	OutContractCode string `json:"out_contract_code"`
	// Env of the account.
	Env XPayEnv `json:"env"`
}

// XPaySubscribeContract is the payload of a subscription contract.
type XPaySubscribeContract struct {
	// State of the contract.
	State int `json:"state"`
	// LastDeductTime of the last deduction.
	LastDeductTime int64 `json:"last_deduct_time,omitempty"`
	// NextDeductTime of the upcoming deduction.
	NextDeductTime int64 `json:"next_deduct_time,omitempty"`
}

// XPayQuerySubscribeContractResponse is returned by
// XPayQuerySubscribeContract.
type XPayQuerySubscribeContractResponse struct {
	ErrResponse
	// Contract of the query.
	Contract XPaySubscribeContract `json:"contract,omitempty"`
}

// XPayQuerySubscribeContract queries the state of an iOS Apple-payment
// subscription contract.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_subscribe_contract.html
func (w *MiniProgram) XPayQuerySubscribeContract(ctx context.Context, req *XPayQuerySubscribeContractRequest) (*XPayQuerySubscribeContractResponse, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_subscribe_contract", req, "", false)
	if err != nil {
		return nil, err
	}
	var result XPayQuerySubscribeContractResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// XPayCancelSubscribeContractRequest cancels an Apple-payment subscription
// contract.
type XPayCancelSubscribeContractRequest struct {
	// ProductID of the subscribed product.
	ProductID string `json:"product_id"`
	// OutContractCode of the developer side contract code.
	OutContractCode string `json:"out_contract_code"`
	// TerminationReason shown to the user.
	TerminationReason string `json:"termination_reason"`
	// Env of the account.
	Env XPayEnv `json:"env"`
}

// XPayCancelSubscribeContract terminates an Apple-payment subscription.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_cancel_subscribe_contract.html
func (w *MiniProgram) XPayCancelSubscribeContract(ctx context.Context, req *XPayCancelSubscribeContractRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/cancel_subscribe_contract", req, "", false)
	return err
}

// XPaySendSubscribePrepaymentRequest pushes an Apple-payment prepayment
// notification before the actual deduction.
type XPaySendSubscribePrepaymentRequest struct {
	// ProductID of the subscribed product.
	ProductID string `json:"product_id"`
	// DeductPrice to deduct in cents.
	DeductPrice int `json:"deduct_price"`
	// OutContractCode of the developer side contract code.
	OutContractCode string `json:"out_contract_code"`
	// Env of the account.
	Env XPayEnv `json:"env"`
}

// XPaySendSubscribePrepayment sends the subscription prepayment notice used by
// the Apple-payment flow.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_send_subscribe_pre_payment.html
func (w *MiniProgram) XPaySendSubscribePrepayment(ctx context.Context, req *XPaySendSubscribePrepaymentRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/send_subscribe_pre_payment", req, "", false)
	return err
}

// XPaySubmitSubscribePayOrderRequest submits an Apple-payment subscription
// order.
type XPaySubmitSubscribePayOrderRequest struct {
	// OfferID of the 米大师 application.
	OfferID string `json:"offer_id"`
	// ProductID of the subscribed product.
	ProductID string `json:"product_id"`
	// DeductPrice to deduct in cents.
	DeductPrice int `json:"deduct_price"`
	// BuyQuantity of the subscription.
	BuyQuantity int `json:"buy_quantity"`
	// CurrencyType, e.g. "CNY".
	CurrencyType string `json:"currency_type"`
	// OrderID of the merchant.
	OrderID string `json:"order_id"`
	// Attach transparent data.
	Attach string `json:"attach"`
	// Env of the account.
	Env XPayEnv `json:"env"`
}

// XPaySubmitSubscribePayOrder submits the actual Apple-payment subscription
// pay order after the prepayment notice.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_submit_subscribe_pay_order.html
func (w *MiniProgram) XPaySubmitSubscribePayOrder(ctx context.Context, req *XPaySubmitSubscribePayOrderRequest) error {
	_, err := w.xpayRequest(ctx, "/xpay/submit_subscribe_pay_order", req, "", false)
	return err
}
