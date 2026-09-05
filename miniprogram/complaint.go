package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// ============================================================
// 交易保障 complaints (投诉) — merchant responses to order complaints. The
// endpoints live under /wxaapi/minishop and /wxaapi/comment.
// ============================================================

// GetComplaintOrderDetailResponse is returned by GetComplaintOrderDetail.
type GetComplaintOrderDetailResponse struct {
	ErrResponse
	// ComplaintOrder is the complained order payload.
	ComplaintOrder *ComplaintOrderDetail `json:"complaintOrder,omitempty"`
	// ComplaintHistoryList of the complaint.
	ComplaintHistoryList []map[string]any `json:"complaintHistoryList,omitempty"`
	// ReturnBill of the related return.
	ReturnBill map[string]any `json:"returnBill,omitempty"`
}

// ComplaintOrderDetail is the semi-structured complained-order payload.
type ComplaintOrderDetail struct {
	// ComplaintOrderID of the complaint.
	ComplaintOrderID int64 `json:"complaintOrderId,omitempty"`
	// Status of the complaint.
	Status int `json:"status,omitempty"`
	// Amount in cents.
	Amount int64 `json:"amount,omitempty"`
	// RawData keeps the remaining upstream fields.
	RawData map[string]any `json:"-"`
}

// GetComplaintOrderDetail returns the detail of one order complaint.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/complaint/api_getorderdetail.html
func (w *MiniProgram) GetComplaintOrderDetail(ctx context.Context, complaintOrderID int64) (*GetComplaintOrderDetailResponse, error) {
	query := url.Values{}
	query.Set("complaintOrderId", fmtInt64(complaintOrderID))
	var result GetComplaintOrderDetailResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/minishop/complaintOrderDetail", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RespondOrderComplaintRequest replies to an order complaint.
type RespondOrderComplaintRequest struct {
	// ComplaintOrderID of the complaint.
	ComplaintOrderID int64 `json:"complaintOrderId"`
	// Content of the merchant response.
	Content string `json:"content"`
	// MediaIDList of the uploaded evidence media.
	MediaIDList []string `json:"mediaIdList,omitempty"`
	// BussiHandle: 1 accepted responsibility, 2 not accepted.
	BussiHandle int `json:"bussiHandle"`
}

// RespondOrderComplaint answers an order complaint with the merchant's
// position.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/complaint/api_respondcomplaint.html
func (w *MiniProgram) RespondOrderComplaint(ctx context.Context, req *RespondOrderComplaintRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/minishop/bussiRespondComplaint", nil, defaultReqOptions(), req, nil)
}

// SupplyOrderComplaintProofRequest submits evidence for a complaint response.
type SupplyOrderComplaintProofRequest struct {
	// ComplaintOrderID of the complaint.
	ComplaintOrderID int64 `json:"complaintOrderId"`
	// Content describing the submitted proof.
	Content string `json:"content,omitempty"`
	// MediaIDList of the proof media.
	MediaIDList []string `json:"mediaIdList,omitempty"`
}

// SupplyOrderComplaintProof uploads supplementary proof to an order complaint
// response.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/complaint/api_supplyproof.html
func (w *MiniProgram) SupplyOrderComplaintProof(ctx context.Context, req *SupplyOrderComplaintProofRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/minishop/bussiSupplyProof", nil, defaultReqOptions(), req, nil)
}

// SubmitOrderComplaintRefundRequest offers a refund to settle an order
// complaint.
type SubmitOrderComplaintRefundRequest struct {
	SupplyOrderComplaintProofRequest
}

// SubmitOrderComplaintRefund submits a refund proposal that closes the order
// complaint.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/complaint/api_submitrefund.html
func (w *MiniProgram) SubmitOrderComplaintRefund(ctx context.Context, req *SubmitOrderComplaintRefundRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/minishop/bussiSupplyRefund", nil, defaultReqOptions(), req, nil)
}

// AppealOrderComplaintRequest appeals an order complaint verdict.
type AppealOrderComplaintRequest struct {
	// ComplaintOrderID of the complaint.
	ComplaintOrderID int64 `json:"complaintOrderId"`
	// Content of the appeal.
	Content string `json:"content"`
	// MediaIDList of the appeal evidence.
	MediaIDList []string `json:"mediaIdList,omitempty"`
}

// AppealOrderComplaint appeals a complaint (申诉) when the merchant disputes
// the decision.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/complaint/api_busiappeal.html
func (w *MiniProgram) AppealOrderComplaint(ctx context.Context, req *AppealOrderComplaintRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/minishop/busiAppeal", nil, defaultReqOptions(), req, nil)
}
