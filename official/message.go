package official

import (
	"context"

	"github.com/go-sphere/weixin-api/core"
)

// ============================================================
// 客服消息 / 消息发送 (customer-service messages).
// A 公众号 can push a message to a user within 48h of their last interaction
// through the customer-service message API.
// ============================================================

// CustomerMessageType is the msgtype of a customer-service message.
type CustomerMessageType string

const (
	// MsgTypeText is a plain text message.
	MsgTypeText CustomerMessageType = "text"
	// MsgTypeImage is an image message.
	MsgTypeImage CustomerMessageType = "image"
	// MsgTypeVoice is a voice message.
	MsgTypeVoice CustomerMessageType = "voice"
	// MsgTypeVideo is a video message.
	MsgTypeVideo CustomerMessageType = "video"
	// MsgTypeMusic is a music message.
	MsgTypeMusic CustomerMessageType = "music"
	// MsgTypeNews is a news (article) message.
	MsgTypeNews CustomerMessageType = "news"
	// MsgTypeMpNews is a permanent-material article message.
	MsgTypeMpNews CustomerMessageType = "mpnews"
	// MsgTypeMpVideo is a Mini Program video message (服务号 only).
	MsgTypeMpVideo CustomerMessageType = "mpvideo"
	// MsgTypeMiniProgramPage is a Mini Program page message.
	MsgTypeMiniProgramPage CustomerMessageType = "miniprogrampage"
	// MsgTypeWxCard is a WeChat card message.
	MsgTypeWxCard CustomerMessageType = "wxcard"
)

// CustomerMessage is a customer-service message sent through the 48h window.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Service_Center_messages.html
type CustomerMessage struct {
	// ToUser is the openid of the receiver.
	ToUser string `json:"touser"`
	// MsgType of the message.
	MsgType CustomerMessageType `json:"msgtype"`
	// Text for text messages.
	Text *TextMessage `json:"text,omitempty"`
	// Image for image messages.
	Image *MediaMessage `json:"image,omitempty"`
	// Voice for voice messages.
	Voice *MediaMessage `json:"voice,omitempty"`
	// Video for video messages.
	Video *VideoMessage `json:"video,omitempty"`
	// Music for music messages.
	Music *MusicMessage `json:"music,omitempty"`
	// News for article messages (1-8 items).
	News *NewsMessage `json:"news,omitempty"`
	// MpNews for permanent-material articles.
	MpNews *MediaMessage `json:"mpnews,omitempty"`
	// MpVideo for Mini Program video (服务号).
	MpVideo *MpVideoMessage `json:"mpvideo,omitempty"`
	// MiniProgramPage for Mini Program page messages.
	MiniProgramPage *MiniProgramPageMessage `json:"miniprogrampage,omitempty"`
	// WxCard for WeChat card (wxcard) messages.
	WxCard *WxCardMessage `json:"wxcard,omitempty"`
}

// WxCardMessage sends a WeChat card (卡券).
type WxCardMessage struct {
	// CardID of the WeChat card to send.
	CardID string `json:"card_id"`
}

// TextMessage is the text payload of a message.
type TextMessage struct {
	// Content of the message.
	Content string `json:"content"`
}

// MediaMessage references an uploaded media or material.
type MediaMessage struct {
	// MediaID of the media/material.
	MediaID string `json:"media_id"`
}

// VideoMessage is a video payload (temp media).
type VideoMessage struct {
	// MediaID of the video.
	MediaID string `json:"media_id"`
	// ThumbMediaID of the video thumbnail.
	ThumbMediaID string `json:"thumb_media_id"`
	// Title of the video.
	Title string `json:"title,omitempty"`
	// Description of the video.
	Description string `json:"description,omitempty"`
}

// MpVideoMessage is a Mini Program video payload (服务号).
type MpVideoMessage struct {
	// MediaID of the Mini Program video.
	MediaID string `json:"media_id"`
	// Title of the video.
	Title string `json:"title,omitempty"`
	// Description of the video.
	Description string `json:"description,omitempty"`
	// ThumbMediaID of the thumbnail.
	ThumbMediaID string `json:"thumb_media_id"`
}

// MusicMessage is a music payload.
type MusicMessage struct {
	// Title of the music.
	Title string `json:"title,omitempty"`
	// Description of the music.
	Description string `json:"description,omitempty"`
	// MusicURL of the playable music.
	MusicURL string `json:"musicurl,omitempty"`
	// HQMusicURL of the high quality music.
	HQMusicURL string `json:"hqmusicurl,omitempty"`
	// ThumbMediaID of the thumbnail.
	ThumbMediaID string `json:"thumb_media_id"`
}

// ArticleItem is one article of a news message.
type ArticleItem struct {
	// Title of the article.
	Title string `json:"title"`
	// Description of the article.
	Description string `json:"description,omitempty"`
	// URL of the article.
	URL string `json:"url"`
	// PicURL of the article image.
	PicURL string `json:"picurl,omitempty"`
}

// NewsMessage is a news (articles) payload.
type NewsMessage struct {
	// Articles of the news (1-8 items).
	Articles []ArticleItem `json:"articles"`
}

// MiniProgramPageMessage opens a Mini Program page.
type MiniProgramPageMessage struct {
	// Title of the message.
	Title string `json:"title"`
	// AppID of the Mini Program.
	AppID string `json:"appid"`
	// PagePath of the Mini Program page.
	PagePath string `json:"pagepath"`
	// ThumbMediaID of the cover.
	ThumbMediaID string `json:"thumb_media_id"`
}

// SendCustomerMessage delivers a customer-service message to a user.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Service_Center_messages.html
func (oa *OfficialAccount) SendCustomerMessage(ctx context.Context, msg *CustomerMessage) error {
	return oa.withTokenPost(ctx, "/cgi-bin/message/custom/send", nil, core.DefaultRequestOptions(), msg, nil)
}

// SetCustomerTyping notifies the user that the account is typing. Send it
// before a real message and cancel it with command "CancelTyping".
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Service_Center_messages.html
func (oa *OfficialAccount) SetCustomerTyping(ctx context.Context, openid, command string) error {
	body := map[string]string{"touser": openid, "command": command}
	return oa.withTokenPost(ctx, "/cgi-bin/message/custom/typing", nil, core.DefaultRequestOptions(), body, nil)
}

// ============================================================
// 模板消息 (template messages).
// ============================================================

// TemplateMessageDataValue is one data value of a template message.
type TemplateMessageDataValue struct {
	// Value of the field.
	Value string `json:"value"`
	// Color of the field text (#RRGGBB).
	Color string `json:"color,omitempty"`
}

// SendTemplateMessageRequest is the payload of the template message API.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
type SendTemplateMessageRequest struct {
	// ToUser is the openid of the receiver.
	ToUser string `json:"touser"`
	// TemplateID of the approved template.
	TemplateID string `json:"template_id"`
	// URL to jump to when tapped.
	URL string `json:"url,omitempty"`
	// MiniProgram to open when tapped.
	MiniProgram *TemplateMessageMiniprogram `json:"miniprogram,omitempty"`
	// AppID (used with the client-side API) - optional.
	AppID string `json:"appid,omitempty"`
	// PagePath of the Mini Program page.
	PagePath string `json:"pagepath,omitempty"`
	// Data of the template fields.
	Data map[string]TemplateMessageDataValue `json:"data"`
}

// TemplateMessageMiniprogram is the Mini Program jump target of a template
// message.
type TemplateMessageMiniprogram struct {
	// AppID of the Mini Program.
	AppID string `json:"appid"`
	// PagePath of the page.
	PagePath string `json:"pagepath"`
}

// SendTemplateMessageResponse is returned by SendTemplateMessage.
type SendTemplateMessageResponse struct {
	ErrResponse
	// MsgID of the sent template message.
	MsgID int64 `json:"msgid"`
}

// SendTemplateMessage sends a template message (one-time; the user must have
// interacted recently or subscribed to the template).
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
func (oa *OfficialAccount) SendTemplateMessage(ctx context.Context, req *SendTemplateMessageRequest) (*SendTemplateMessageResponse, error) {
	var result SendTemplateMessageResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/message/template/send", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetTemplateIndustryRequest configures the two template industries of the
// account (服务号 认证 only).
type SetTemplateIndustryRequest struct {
	// IndustryID1 of the primary industry.
	IndustryID1 string `json:"industry_id1"`
	// IndustryID2 of the secondary industry.
	IndustryID2 string `json:"industry_id2"`
}

// SetTemplateIndustry sets the template message industries.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
func (oa *OfficialAccount) SetTemplateIndustry(ctx context.Context, req *SetTemplateIndustryRequest) error {
	return oa.withTokenPost(ctx, "/cgi-bin/template/api_set_industry", nil, core.DefaultRequestOptions(), req, nil)
}

// TemplateIndustry is one industry entry of the account.
type TemplateIndustry struct {
	// FirstClass of the industry.
	FirstClass string `json:"first_class"`
	// SecondClass of the industry.
	SecondClass string `json:"second_class"`
}

// GetTemplateIndustryResponse is returned by GetTemplateIndustry.
type GetTemplateIndustryResponse struct {
	ErrResponse
	// PrimaryIndustry of the account.
	PrimaryIndustry TemplateIndustry `json:"primary_industry"`
	// SecondaryIndustry of the account.
	SecondaryIndustry TemplateIndustry `json:"secondary_industry"`
}

// GetTemplateIndustry returns the configured template industries.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
func (oa *OfficialAccount) GetTemplateIndustry(ctx context.Context) (*GetTemplateIndustryResponse, error) {
	var result GetTemplateIndustryResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/template/get_industry", nil, core.DefaultRequestOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TemplateMessageListItem is one template added to the account.
type TemplateMessageListItem struct {
	// TemplateID of the template.
	TemplateID string `json:"template_id"`
	// Title of the template.
	Title string `json:"title"`
	// PrimaryIndustry of the template.
	PrimaryIndustry string `json:"primary_industry"`
	// DeputyIndustry of the template.
	DeputyIndustry string `json:"deputy_industry"`
	// Content of the template.
	Content string `json:"content"`
	// Example of the template.
	Example string `json:"example"`
}

// GetTemplateMessageListResponse is returned by GetTemplateMessageList.
type GetTemplateMessageListResponse struct {
	ErrResponse
	// TemplateList of the added templates.
	TemplateList []TemplateMessageListItem `json:"template_list"`
}

// GetTemplateMessageList lists the templates added to the account.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
func (oa *OfficialAccount) GetTemplateMessageList(ctx context.Context) (*GetTemplateMessageListResponse, error) {
	var result GetTemplateMessageListResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/template/get_all_private_template", nil, core.DefaultRequestOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteTemplateMessage removes an added template.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
func (oa *OfficialAccount) DeleteTemplateMessage(ctx context.Context, templateID string) error {
	body := map[string]string{"template_id": templateID}
	return oa.withTokenPost(ctx, "/cgi-bin/template/del_private_template", nil, core.DefaultRequestOptions(), body, nil)
}
