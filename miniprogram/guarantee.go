package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// ============================================================
// 交易保障 (transaction guarantee) — order comments & complaints.
// Reference docs are grouped under
// https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/
// ============================================================

// GuaranteeCommentMedia is one media item attached to a comment (image/video).
type GuaranteeCommentMedia struct {
	// Img of an image media.
	Img string `json:"img,omitempty"`
	// ThumbImg of the image thumbnail.
	ThumbImg string `json:"thumbImg,omitempty"`
	// Video of a video media.
	Video string `json:"video,omitempty"`
	// VideoCover of the video.
	VideoCover string `json:"videoCover,omitempty"`
	// VideoDuration in seconds.
	VideoDuration int `json:"videoDuration,omitempty"`
}

// GuaranteeOrderInfo is the order referenced by a comment.
type GuaranteeOrderInfo struct {
	// BusiOrderId of the merchant.
	BusiOrderId string `json:"busiOrderId"`
}

// GuaranteeUser is the commenting user.
type GuaranteeUser struct {
	// OpenID of the user.
	OpenID string `json:"openid"`
	// Nickname of the user.
	Nickname string `json:"nickName"`
	// HeadImg of the user avatar.
	HeadImg string `json:"headImg"`
}

// GuaranteeBusiness is the Mini Program that received the comment.
type GuaranteeBusiness struct {
	// AppID of the Mini Program.
	AppID string `json:"appid"`
	// Nickname of the Mini Program.
	Nickname string `json:"nickName"`
	// HeadImg of the Mini Program avatar.
	HeadImg string `json:"headImg"`
}

// GuaranteeProduct is one product of the commented order.
type GuaranteeProduct struct {
	// ProductID of the product.
	ProductID string `json:"productId,omitempty"`
	// SkuID of the product SKU.
	SkuID string `json:"skuId,omitempty"`
	// Title of the product.
	Title string `json:"title,omitempty"`
	// PicURL of the product.
	PicURL string `json:"picUrl,omitempty"`
}

// GuaranteeComment is one order comment (评价).
type GuaranteeComment struct {
	// CommentID of the comment.
	CommentID string `json:"commentId"`
	// Amount of the order in cents.
	Amount int `json:"amount"`
	// OrderID of the merchant.
	OrderID string `json:"orderId"`
	// WxPayID of the WeChat Pay transaction.
	WxPayID string `json:"wxPayId"`
	// PayTime of the payment (unix seconds).
	PayTime int64 `json:"payTime"`
	// CreateTime of the comment (unix seconds).
	CreateTime int64 `json:"createTime"`
	// Score of the comment (1-5 stars).
	Score int `json:"score"`
	// Txt of the comment text.
	Txt string `json:"txt"`
	// Media of the comment.
	Media []GuaranteeCommentMedia `json:"media"`
	// IsAlreadySendTmpl indicates whether the comment template was pushed.
	IsAlreadySendTmpl bool `json:"isAlreadySendTmpl"`
	// ProductList of the commented products.
	ProductList []GuaranteeProduct `json:"productList"`
	// UserInfo of the commenting user.
	UserInfo GuaranteeUser `json:"userInfo"`
	// BusinessInfo of the receiving Mini Program.
	BusinessInfo GuaranteeBusiness `json:"businessInfo"`
	// OrderInfo of the referenced order.
	OrderInfo GuaranteeOrderInfo `json:"orderInfo"`
}

// GetOrderCommentListResponse is returned by GetOrderCommentList.
type GetOrderCommentListResponse struct {
	ErrResponse
	// CommentList of the page.
	CommentList []GuaranteeComment `json:"commentList"`
	// Total number of matching comments.
	Total int `json:"total"`
	// Offset of the returned page.
	Offset int `json:"offset"`
}

// GetOrderCommentListRequest selects the comments of an order window.
type GetOrderCommentListRequest struct {
	// StartTime of the window (unix seconds).
	StartTime int64
	// EndTime of the window (unix seconds).
	EndTime int64
	// FilterType of the comments: 0 all, 1 positive, 2 neutral, 3 negative.
	FilterType int
	// Offset of the page.
	Offset int
	// Limit of the page.
	Limit int
}

func (r *GetOrderCommentListRequest) query() url.Values {
	q := url.Values{}
	q.Set("startTimestamp", fmtInt64(r.StartTime))
	q.Set("endTimestamp", fmtInt64(r.EndTime))
	if r.FilterType != 0 {
		q.Set("filterType", fmtInt(r.FilterType))
	}
	if r.Offset != 0 {
		q.Set("offset", fmtInt(r.Offset))
	}
	if r.Limit != 0 {
		q.Set("limit", fmtInt(r.Limit))
	}
	return q
}

// GetOrderCommentList lists the order comments (评价) of the Mini Program
// within a time window.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_getccommentlist.html
func (w *MiniProgram) GetOrderCommentList(ctx context.Context, req *GetOrderCommentListRequest) (*GetOrderCommentListResponse, error) {
	var result GetOrderCommentListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/comment/mpcommentlist/get", req.query(), defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOrderCommentInfoResponse is returned by GetOrderCommentInfo.
type GetOrderCommentInfoResponse struct {
	ErrResponse
	// Comment of the query.
	Comment GuaranteeComment `json:"comment,omitempty"`
}

// GetOrderCommentInfo returns one order comment by id.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_getcommentinfo.html
func (w *MiniProgram) GetOrderCommentInfo(ctx context.Context, commentID string) (*GetOrderCommentInfoResponse, error) {
	query := url.Values{}
	query.Set("commentId", commentID)
	var result GetOrderCommentInfoResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/comment/commentinfo/get", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GuaranteeCommentReply is one reply of a comment.
type GuaranteeCommentReply struct {
	// ReplyID of the reply.
	ReplyID string `json:"replyId"`
	// Content of the reply.
	Content string `json:"content"`
	// ReplyTime of the reply (unix seconds).
	ReplyTime int64 `json:"replyTime"`
	// Type of the reply: 1 merchant, 2 buyer.
	Type int `json:"type"`
	// UserInfo of the replier.
	UserInfo GuaranteeUser `json:"userInfo,omitempty"`
}

// GetCommentReplyListResponse is returned by GetCommentReplyList.
type GetCommentReplyListResponse struct {
	ErrResponse
	// ReplyList of the comment.
	ReplyList []GuaranteeCommentReply `json:"replyList,omitempty"`
}

// GetCommentReplyList returns the merchant/user replies of one comment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_getcommentreplylist.html
func (w *MiniProgram) GetCommentReplyList(ctx context.Context, commentID string) (*GetCommentReplyListResponse, error) {
	query := url.Values{}
	query.Set("commentId", commentID)
	var result GetCommentReplyListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/comment/replyandcommentreplylist/get", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddCommentReplyRequest adds a merchant reply to a comment.
type AddCommentReplyRequest struct {
	// CommentID of the comment to reply to.
	CommentID string `json:"commentId"`
	// Content of the reply (up to 300 chars).
	Content string `json:"content"`
}

// AddCommentReply replies to an order comment on behalf of the merchant.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_addcommentreply.html
func (w *MiniProgram) AddCommentReply(ctx context.Context, req *AddCommentReplyRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/comment/commentreply/add", nil, defaultReqOptions(), req, nil)
}

// DeleteCommentReplyRequest removes a merchant reply.
type DeleteCommentReplyRequest struct {
	// CommentID of the comment.
	CommentID string `json:"commentId"`
}

// DeleteCommentReply removes a merchant reply from an order comment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_deletecommentreply.html
func (w *MiniProgram) DeleteCommentReply(ctx context.Context, commentID string) error {
	req := &DeleteCommentReplyRequest{CommentID: commentID}
	return w.withAccessTokenPost(ctx, "/wxaapi/comment/commentreply/delete", nil, defaultReqOptions(), req, nil)
}

// AddReplyRequest creates a comment (评价) on behalf of the developer. The
// first reply created by the developer becomes the order comment itself.
type AddReplyRequest struct {
	// CommentID of the order to comment on.
	CommentID string `json:"commentId"`
	// Content of the comment (up to 300 chars).
	Content string `json:"content"`
}

// AddReply creates a comment for an order; the first developer reply turns
// into the order comment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_addreply.html
func (w *MiniProgram) AddReply(ctx context.Context, req *AddReplyRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/comment/reply/add", nil, defaultReqOptions(), req, nil)
}

// DeleteReplyRequest removes a comment/reply.
type DeleteReplyRequest struct {
	// CommentID of the comment.
	CommentID string `json:"commentId"`
	// ReplyID of the reply to delete.
	ReplyID string `json:"replyId"`
}

// DeleteReply deletes a reply (or comment) by its reply id.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_deletereply.html
func (w *MiniProgram) DeleteReply(ctx context.Context, commentID, replyID string) error {
	req := &DeleteReplyRequest{CommentID: commentID, ReplyID: replyID}
	return w.withAccessTokenPost(ctx, "/wxaapi/comment/reply/delete", nil, defaultReqOptions(), req, nil)
}

// ConfirmCommentCompromiseRequest settles (和解) an order comment dispute.
type ConfirmCommentCompromiseRequest struct {
	// CommentID of the comment to compromise.
	CommentID string `json:"commentId"`
	// PicList of the negotiation evidence images.
	PicList []string `json:"picList,omitempty"`
	// Content describing the compromise.
	Content string `json:"content,omitempty"`
}

// ConfirmCommentCompromise marks an order comment as settled through merchant
// negotiation (和解).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_confirmcompromise.html
func (w *MiniProgram) ConfirmCommentCompromise(ctx context.Context, req *ConfirmCommentCompromiseRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/comment/confirmcompromise", nil, defaultReqOptions(), req, nil)
}

// ResetCommentKfQuota resets the reply quota consumed by the customer-service
// auto-reply of a comment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/comment/api_resetapikfquota.html
func (w *MiniProgram) ResetCommentKfQuota(ctx context.Context, commentID string) error {
	body := map[string]string{"commentId": commentID}
	return w.withAccessTokenPost(ctx, "/wxaapi/comment/apikfquota/reset", nil, defaultReqOptions(), body, nil)
}

// ============================================================
// Guarantee status & penalties (交易保障状态/处罚).
// ============================================================

// GetGuaranteeStatusResponse reports whether the trade-guarantee capability is
// active for the account.
type GetGuaranteeStatusResponse struct {
	ErrResponse
	// Raw payload (the guarantee state object) is semi-structured.
	Raw map[string]any `json:"-"`
}

// GetGuaranteeStatus returns the transaction-guarantee (交易保障) status of the
// Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/basic/api_getguaranteestatus.html
func (w *MiniProgram) GetGuaranteeStatus(ctx context.Context) (*GetGuaranteeStatusResponse, error) {
	var result GetGuaranteeStatusResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/wxamptrade/get_guarantee_status", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GuaranteePenalty is one penalty (处罚) record of the Mini Program.
type GuaranteePenalty struct {
	// RawData of the penalty record.
	RawData map[string]any `json:"-"`
}

// GetPenaltyListRequest pages the penalty list.
type GetPenaltyListRequest struct {
	// Offset of the page.
	Offset int
	// Limit of the page.
	Limit int
}

// GetPenaltyListResponse is returned by GetPenaltyList.
type GetPenaltyListResponse struct {
	ErrResponse
	// CurrentScore of the trade-guarantee scoring.
	CurrentScore int `json:"current_score,omitempty"`
	// Total of the penalties.
	Total int `json:"total,omitempty"`
	// AppealList of the penalty records.
	AppealList []map[string]any `json:"appeal_list,omitempty"`
}

// GetPenaltyList lists the trade-guarantee penalties (处罚) of the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/transaction-guarantee/basic/api_getpenaltylist.html
func (w *MiniProgram) GetPenaltyList(ctx context.Context, req *GetPenaltyListRequest) (*GetPenaltyListResponse, error) {
	query := url.Values{}
	query.Set("offset", fmtInt(req.Offset))
	query.Set("limit", fmtInt(req.Limit))
	var result GetPenaltyListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/wxamptrade/get_penalty_list", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// fmtInt64 renders an int64 as a decimal string.
func fmtInt64(v int64) string { return fmtInt(int(v)) }
