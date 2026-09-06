package official

import (
	"context"

	"github.com/go-sphere/weixin-api/core"
)

// ============================================================
// 群发消息 (mass message sending). Available to verified accounts. Tag-based
// or openid-list based delivery.
// ============================================================

// MassMessage is the payload of a mass send by tag or openids.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Batch_Sends_and_Group_Sends.html
type MassMessage struct {
	// Filter sends to a tag (or "is_to_all": true for everyone).
	Filter *MassFilter `json:"filter,omitempty"`
	// ToUser sends to a specific openid list.
	ToUser []string `json:"touser,omitempty"`
	// MpNews sends a permanent news material.
	MpNews *MediaMessage `json:"mpnews,omitempty"`
	// Text sends text.
	Text *TextMessage `json:"text,omitempty"`
	// Voice sends a voice material.
	Voice *MediaMessage `json:"voice,omitempty"`
	// Image sends an image material.
	Image *MediaMessage `json:"image,omitempty"`
	// MsgType of the message.
	MsgType string `json:"msgtype"`
	// SendIgnoreReprint prevents sending when the article was reposted.
	SendIgnoreReprint int `json:"send_ignore_reprint,omitempty"`
	// ClientMsgID deduplicates the request.
	ClientMsgID string `json:"clientmsgid,omitempty"`
}

// MassFilter selects the mass-send audience.
type MassFilter struct {
	// IsToAll sends to every follower.
	IsToAll bool `json:"is_to_all"`
	// TagID sends to the followers of a tag.
	TagID int `json:"tag_id"`
}

// MassSendResponse is returned by the mass send endpoints.
type MassSendResponse struct {
	ErrResponse
	// MsgID of the mass message.
	MsgID int64 `json:"msg_id"`
	// MsgDataID of the message data.
	MsgDataID int64 `json:"msg_data_id,omitempty"`
}

// MassSendByTag sends a mass message to the followers of a tag.
//
// Mass sends are non-idempotent and the send helpers do not accept a
// clientmsgid, so token-expiry auto-retry is disabled: the SDK must not
// re-submit a crowd-targeting message on its own.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Batch_Sends_and_Group_Sends.html
func (oa *OfficialAccount) MassSendByTag(ctx context.Context, tagID int, msgType string, mediaIDOrText string) (*MassSendResponse, error) {
	msg := &MassMessage{
		Filter:  &MassFilter{TagID: tagID},
		MsgType: msgType,
	}
	attachMassContent(msg, mediaIDOrText)
	var result MassSendResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/message/mass/sendall", nil, core.RequestOptions{Retryable: false}, msg, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MassSendToOpenIDs sends a mass message to an explicit openid list. Like
// MassSendByTag it is non-idempotent, so token-expiry auto-retry is disabled.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Batch_Sends_and_Group_Sends.html
func (oa *OfficialAccount) MassSendToOpenIDs(ctx context.Context, openids []string, msgType, mediaIDOrText string) (*MassSendResponse, error) {
	msg := &MassMessage{
		ToUser:  openids,
		MsgType: msgType,
	}
	attachMassContent(msg, mediaIDOrText)
	var result MassSendResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/message/mass/send", nil, core.RequestOptions{Retryable: false}, msg, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func attachMassContent(msg *MassMessage, mediaIDOrText string) {
	media := &MediaMessage{MediaID: mediaIDOrText}
	switch msg.MsgType {
	case "text":
		msg.Text = &TextMessage{Content: mediaIDOrText}
	case "mpnews":
		msg.MpNews = media
	case "voice":
		msg.Voice = media
	case "image":
		msg.Image = media
	}
}

// MassPreviewRequest previews a mass message to a WeChat user.
type MassPreviewRequest struct {
	// ToUser of the preview (openid).
	ToUser string `json:"touser,omitempty"`
	// ToWXName of the preview (微信号).
	ToWXName string `json:"towxname,omitempty"`
	// MsgType of the message.
	MsgType string `json:"msgtype"`
	// Content of a text preview.
	Content string `json:"text,omitempty"`
	// MediaID of a media preview.
	MediaID string `json:"media_id,omitempty"`
}

// MassPreview sends a preview of a mass message.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Batch_Sends_and_Group_Sends.html
func (oa *OfficialAccount) MassPreview(ctx context.Context, req *MassPreviewRequest) error {
	payload := map[string]any{
		"msgtype": req.MsgType,
	}
	if req.ToUser != "" {
		payload["touser"] = req.ToUser
	}
	if req.ToWXName != "" {
		payload["towxname"] = req.ToWXName
	}
	if req.MsgType == "text" {
		payload["text"] = map[string]string{"content": req.Content}
	} else {
		payload[req.MsgType] = map[string]string{"media_id": req.MediaID}
	}
	return oa.withTokenPost(ctx, "/cgi-bin/message/mass/preview", nil, core.DefaultRequestOptions(), payload, nil)
}

// MassStatusResponse is returned by GetMassSendStatus.
type MassStatusResponse struct {
	ErrResponse
	// MsgID of the mass message.
	MsgID int64 `json:"msg_id"`
	// Status of the delivery: SEND_SUCCESS / SENDING / SEND_FAIL / DELETE.
	Status string `json:"msg_status"`
}

// GetMassSendStatus polls the status of a mass message.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Batch_Sends_and_Group_Sends.html
func (oa *OfficialAccount) GetMassSendStatus(ctx context.Context, msgID int64) (*MassStatusResponse, error) {
	var result MassStatusResponse
	body := map[string]int64{"msg_id": msgID}
	if err := oa.withTokenPost(ctx, "/cgi-bin/message/mass/get", nil, core.DefaultRequestOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMassMessage removes a sent mass message.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Batch_Sends_and_Group_Sends.html
func (oa *OfficialAccount) DeleteMassMessage(ctx context.Context, msgID int64) error {
	body := map[string]any{"msg_id": msgID}
	return oa.withTokenPost(ctx, "/cgi-bin/message/mass/delete", nil, core.DefaultRequestOptions(), body, nil)
}

// ============================================================
// 数据分析 (official account datacube).
// ============================================================

// UserAnalysisSummary is one row of the user analysis.
type UserAnalysisSummary struct {
	// RefDate of the row ("yyyy-mm-dd").
	RefDate string `json:"ref_date"`
	// UserSource of the row (channel code).
	UserSource int `json:"user_source"`
	// NewUser of the day.
	NewUser int `json:"new_user"`
	// CancelUser of the day.
	CancelUser int `json:"cancel_user"`
}

// GetUserAnalysisResponse is returned by GetUserAnalysis.
type GetUserAnalysisResponse struct {
	ErrResponse
	// List of the summary rows.
	List []UserAnalysisSummary `json:"list"`
}

// dateRangeBody builds the shared {begin_date,end_date} body.
func dateRangeBody(beginDate, endDate string) map[string]string {
	return map[string]string{"begin_date": beginDate, "end_date": endDate}
}

// GetUserAnalysis returns the daily user (follower) change data.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Analytics/User_Analysis_Data_Interface.html
func (oa *OfficialAccount) GetUserAnalysis(ctx context.Context, beginDate, endDate string) (*GetUserAnalysisResponse, error) {
	var result GetUserAnalysisResponse
	if err := oa.withTokenPost(ctx, "/datacube/getusersummary", nil, core.DefaultRequestOptions(), dateRangeBody(beginDate, endDate), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ArticleSummary is one row of the article analysis.
type ArticleSummary struct {
	// RefDate of the row.
	RefDate string `json:"ref_date"`
	// MsgID of the article message.
	MsgID int `json:"msgid"`
	// Title of the article.
	Title string `json:"title"`
	// IntPageReadUser of the internal readers.
	IntPageReadUser int `json:"int_page_read_user"`
	// IntPageReadCount of the internal reads.
	IntPageReadCount int `json:"int_page_read_count"`
	// OriPageReadUser of the direct readers.
	OriPageReadUser int `json:"ori_page_read_user"`
	// OriPageReadCount of the direct reads.
	OriPageReadCount int `json:"ori_page_read_count"`
	// ShareUser of the sharers.
	ShareUser int `json:"share_user"`
	// ShareCount of the shares.
	ShareCount int `json:"share_count"`
	// AddToFavUser of the favouriters.
	AddToFavUser int `json:"add_to_fav_user"`
	// AddToFavCount of the favourites.
	AddToFavCount int `json:"add_to_fav_count"`
}

// GetArticleSummaryResponse is returned by GetArticleSummary.
type GetArticleSummaryResponse struct {
	ErrResponse
	// List of the article rows.
	List []ArticleSummary `json:"list"`
}

// GetArticleSummary returns the article-read data of the latest 3 days.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Analytics/Article_Analysis_Data_Interface.html
func (oa *OfficialAccount) GetArticleSummary(ctx context.Context, beginDate, endDate string) (*GetArticleSummaryResponse, error) {
	var result GetArticleSummaryResponse
	if err := oa.withTokenPost(ctx, "/datacube/getarticlesummary", nil, core.DefaultRequestOptions(), dateRangeBody(beginDate, endDate), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// InterfaceAnalysisRow is one row of the interface call analysis.
type InterfaceAnalysisRow struct {
	// RefDate of the row.
	RefDate string `json:"ref_date"`
	// CallbackCount of the callback calls.
	CallbackCount int `json:"callback_count"`
	// FailCount of the failed calls.
	FailCount int `json:"fail_count"`
	// TotalTimeCost of the calls (ms).
	TotalTimeCost int `json:"total_time_cost"`
	// MaxTimeCost of the slowest call (ms).
	MaxTimeCost int `json:"max_time_cost"`
}

// GetInterfaceAnalysisResponse is returned by GetInterfaceAnalysis.
type GetInterfaceAnalysisResponse struct {
	ErrResponse
	// List of the rows.
	List []InterfaceAnalysisRow `json:"list"`
}

// GetInterfaceAnalysis returns the interface-call analysis data.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Analytics/Analytics_API.html
func (oa *OfficialAccount) GetInterfaceAnalysis(ctx context.Context, beginDate, endDate string) (*GetInterfaceAnalysisResponse, error) {
	var result GetInterfaceAnalysisResponse
	if err := oa.withTokenPost(ctx, "/datacube/getinterfacedata", nil, core.DefaultRequestOptions(), dateRangeBody(beginDate, endDate), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
