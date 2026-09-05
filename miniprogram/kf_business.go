package miniprogram

import (
	"context"
	"encoding/json"
)

// KfBusinessResponse is the generic response of the customer-service business
// management APIs. WeChat returns the business profile with unstable field
// naming, so the decoded fields are retained raw.
type KfBusinessResponse struct {
	ErrResponse
	// Raw is the full upstream payload.
	Raw json.RawMessage `json:"-"`
}

// RegisterKfBusinessRequest opens a WeChat customer-service business (微信客服
// 商家版) entity for the Mini Program.
type RegisterKfBusinessRequest struct {
	// KfNick of the customer-service business.
	KfNick string `json:"kf_nick,omitempty"`
	// HeadImgURL of the business avatar.
	HeadImgURL string `json:"head_img_url,omitempty"`
	// Intro of the business.
	Intro string `json:"intro,omitempty"`
	// AvatarMediaID of the avatar uploaded via the media API.
	AvatarMediaID string `json:"avatar_media_id,omitempty"`
	// WechatChannelsQRCode of the bound channels QR code.
	WechatChannelsQRCode string `json:"wechat_channels_qr_code,omitempty"`
}

// RegisterKfBusiness registers a WeChat customer-service (微信客服) business for
// the Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-management/api_registerbusiness.html
func (w *MiniProgram) RegisterKfBusiness(ctx context.Context, req *RegisterKfBusinessRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/business/register", nil, defaultReqOptions(), req, nil)
}

// UpdateKfBusinessRequest updates the profile of the customer-service business.
type UpdateKfBusinessRequest struct {
	// KfNick of the business.
	KfNick string `json:"kf_nick,omitempty"`
	// Intro of the business.
	Intro string `json:"intro,omitempty"`
	// AvatarMediaID of the avatar media.
	AvatarMediaID string `json:"avatar_media_id,omitempty"`
}

// UpdateKfBusiness updates the WeChat customer-service business profile.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-management/api_updatebusiness.html
func (w *MiniProgram) UpdateKfBusiness(ctx context.Context, req *UpdateKfBusinessRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/business/update", nil, defaultReqOptions(), req, nil)
}

// GetKfBusinessResponse is returned by GetKfBusiness.
type GetKfBusinessResponse struct {
	ErrResponse
	// OpenKfID of the customer-service business.
	OpenKfID string `json:"open_kf_id,omitempty"`
	// Raw profile payload.
	Raw json.RawMessage `json:"-"`
}

// GetKfBusiness returns the WeChat customer-service business bound to the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-management/api_getbusiness.html
func (w *MiniProgram) GetKfBusiness(ctx context.Context) (*GetKfBusinessResponse, error) {
	var result GetKfBusinessResponse
	body := map[string]string{}
	raw, err := w.withAccessTokenRaw(ctx, "POST", "/cgi-bin/business/get", nil, defaultReqOptions(), body)
	if err != nil {
		return nil, err
	}
	if err := decodeJSON(raw, &result); err != nil {
		return nil, err
	}
	result.Raw = raw
	return &result, nil
}

// ListKfBusinessResponse is returned by ListKfBusiness.
type ListKfBusinessResponse struct {
	ErrResponse
	// KfList of the bound customer-service businesses.
	KfList []json.RawMessage `json:"kf_list,omitempty"`
}

// ListKfBusiness lists the customer-service businesses reachable from the Mini
// Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/kf-mgnt/kf-management/api_listbusiness.html
func (w *MiniProgram) ListKfBusiness(ctx context.Context) (*ListKfBusinessResponse, error) {
	var result ListKfBusinessResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/business/list", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetChatToolMessageRequest sets the "chat tool" entry message of an updatable
// (chat) message. The chat tool message is a special live activity message that
// can hold a join button.
type SetChatToolMessageRequest struct {
	// ActivityID of the created activity.
	ActivityID string `json:"activity_id"`
	// TargetState selects the message instance (0 original, 1/2 replicas).
	TargetState int `json:"target_state"`
	// TemplateInfo carries the template parameters (a "join" parameter with a
	// value "wxopen://..."), matching wx.updateMessage chat-tool usage.
	TemplateInfo struct {
		// ParameterList of the template values.
		ParameterList []UpdatableMessageParameter `json:"parameter_list"`
	} `json:"template_info"`
}

// SetChatToolMessage updates the chat tool (微信客服工具) message content shown in
// a chat activity card.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/updatable-message/api_setchattoolmsg.html
func (w *MiniProgram) SetChatToolMessage(ctx context.Context, req *SetChatToolMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/wxopen/chattoolmsg/send", nil, defaultReqOptions(), req, nil)
}
