package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// UpdatableMessageParameter is one "parameter" entry of an updatable message.
type UpdatableMessageParameter struct {
	// Name is the template parameter key.
	Name string `json:"name"`
	// Value is the value to substitute.
	Value string `json:"value"`
}

// CreateActivityIDResponse is returned by CreateActivityID.
type CreateActivityIDResponse struct {
	ErrResponse
	// ActivityID references the message for later updates.
	ActivityID string `json:"activity_id"`
	// ExpirationTime is the Unix timestamp (seconds) when the activity id
	// expires.
	ExpirationTime int64 `json:"expiration_time"`
	// TargetState carries an intermediate state for the up to 6 message
	// instances a dynamic activity id can drive.
	TargetState int `json:"target_state,omitempty"`
}

// CreateActivityID obtains an activity id used to build a "updatable message"
// (a live activity entry like a game/match score card). Pass an empty unionid
// unless the message should be updatable across apps sharing the Open Platform
// account. openid is optional and needed only for cross-account activity ids.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/updatable-message/api_createactivityid.html
func (w *MiniProgram) CreateActivityID(ctx context.Context, unionid, openid string) (*CreateActivityIDResponse, error) {
	query := url.Values{}
	query.Set("unionid", unionid)
	if openid != "" {
		query.Set("openid", openid)
	}
	var result CreateActivityIDResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/cgi-bin/message/wxopen/activityid/create", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetUpdatableMessageRequest updates an in-flight updatable message.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/updatable-message/api_setupdatablemsg.html
type SetUpdatableMessageRequest struct {
	// ActivityID returned by CreateActivityID (required).
	ActivityID string `json:"activity_id"`
	// TargetState selects the message instance to update: 0 = the original
	// message, 1 = the first replica, 2 = the second replica.
	TargetState int `json:"target_state"`
	// TemplateInfo carries the new parameter values.
	TemplateInfo struct {
		// ParameterList replaces the template placeholders.
		ParameterList []UpdatableMessageParameter `json:"parameter_list"`
	} `json:"template_info"`
}

// SetUpdatableMessage updates the content of a previously created updatable
// message (the wx.updateMessage usage on the client).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/updatable-message/api_setupdatablemsg.html
func (w *MiniProgram) SetUpdatableMessage(ctx context.Context, req *SetUpdatableMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/wxopen/updatablemsg/send", nil, defaultReqOptions(), req, nil)
}

// SendUniformMessageRequest is the payload of SendUniformMessage. A uniform
// message can send an official-account template message and/or a Mini Program
// subscribe message in one call. The Mini Program and official account must be
// bound to the same WeChat Open Platform account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/uniform-message/api_senduniformmessage.html
type SendUniformMessageRequest struct {
	// ToUser is the openid of the user under the official account (or under
	// the Mini Program when MpTemplateMsg only is used).
	ToUser string `json:"touser"`
	// WeappTemplateMsg is the Mini Program subscribe message payload. Follow
	// the subscribe message template rules (user must have granted the
	// subscription).
	WeappTemplateMsg *MiniProgramTemplateMessage `json:"weapp_template_msg,omitempty"`
	// MpTemplateMsg is the official account template message payload. The user
	// must have followed the official account.
	MpTemplateMsg *OfficialAccountTemplateMessage `json:"mp_template_msg,omitempty"`
}

// MiniProgramTemplateMessage is the Mini Program part of a uniform message.
type MiniProgramTemplateMessage struct {
	// TemplateID is the subscribe message template id (priTmplId).
	TemplateID string `json:"template_id,omitempty"`
	// Page to open when the message is tapped.
	Page string `json:"page,omitempty"`
	// Data mirrors the subscribe message keyword values.
	Data map[string]string `json:"data,omitempty"`
	// MiniprogramState (developer/trial/formal) of the target page.
	MiniprogramState string `json:"miniprogram_state,omitempty"`
	// Lang of the content.
	Lang string `json:"lang,omitempty"`
}

// OfficialAccountTemplateMessage is the official account part of a uniform
// message.
type OfficialAccountTemplateMessage struct {
	// AppID of the bound official account.
	AppID string `json:"appid,omitempty"`
	// TemplateID of the official account template message.
	TemplateID string `json:"template_id,omitempty"`
	// URL to open when the message is tapped (optional; when both URL and
	// Miniprogram are absent the official account home is used).
	URL string `json:"url,omitempty"`
	// Miniprogram to open instead of the URL when tapped.
	Miniprogram *MpTemplateMiniprogram `json:"miniprogram,omitempty"`
	// Data of the official account template, keyed by template field names.
	Data map[string]MpTemplateDataValue `json:"data,omitempty"`
}

// MpTemplateMiniprogram points a uniform message tap to a Mini Program page.
type MpTemplateMiniprogram struct {
	// AppID of the Mini Program (must be bound to the same Open Platform
	// account as the official account).
	AppID string `json:"appid,omitempty"`
	// PagePath to open inside the Mini Program.
	PagePath string `json:"pagepath,omitempty"`
}

// MpTemplateDataValue is one template data value of an official account
// message.
type MpTemplateDataValue struct {
	// Value of the field.
	Value string `json:"value,omitempty"`
	// Color of the field text (optional).
	Color string `json:"color,omitempty"`
}

// SendUniformMessage sends a uniform message mixing an official account
// template message and/or a Mini Program subscribe message.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/uniform-message/api_senduniformmessage.html
func (w *MiniProgram) SendUniformMessage(ctx context.Context, req *SendUniformMessageRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/wxopen/template/uniform_send", nil, defaultReqOptions(), req, nil)
}
