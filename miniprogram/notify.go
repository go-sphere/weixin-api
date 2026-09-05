package miniprogram

import (
	"context"
)

// UserNotifyQuery selects the notify (服务通知 / subscription status) of a user.
// notify_type and notify_code come from the client-side interaction that
// created the notify state (service card capability).
type UserNotifyQuery struct {
	// OpenID of the user.
	OpenID string `json:"openid"`
	// NotifyType of the notify card.
	NotifyType int `json:"notify_type"`
	// NotifyCode of the notify.
	NotifyCode string `json:"notify_code"`
}

// UserNotifyInfo is the notify state of one user.
type UserNotifyInfo struct {
	// NotifyType of the notify card.
	NotifyType int `json:"notify_type,omitempty"`
	// CodeState of the notify: 0 = not bound, 1 = valid, 2 = invalid.
	CodeState int `json:"code_state,omitempty"`
	// CodeExpireTime is the Unix timestamp (seconds) when the notify code
	// expires.
	CodeExpireTime int64 `json:"code_expire_time,omitempty"`
	// ContentJSON is the card-state JSON data.
	ContentJSON string `json:"content_json,omitempty"`
}

// GetUserNotifyResponse is returned by GetUserNotify.
type GetUserNotifyResponse struct {
	ErrResponse
	// NotifyInfo of the queried user notify.
	NotifyInfo UserNotifyInfo `json:"notify_info,omitempty"`
}

// GetUserNotify queries the current state of a user's notify subscription
// (订阅通知状态管理).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_getusernotify.html
func (w *MiniProgram) GetUserNotify(ctx context.Context, req *UserNotifyQuery) (*GetUserNotifyResponse, error) {
	var result GetUserNotifyResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/get_user_notify", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetUserNotifyRequest sets the notify (服务卡片) state content of a user.
type SetUserNotifyRequest struct {
	UserNotifyQuery
	// ContentJSON carries the card-state data.
	ContentJSON string `json:"content_json,omitempty"`
	// CheckJSON carries the validity check data.
	CheckJSON string `json:"check_json,omitempty"`
}

// SetUserNotify updates the state of a user notify, typically after a client
// interaction that changes the card.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_setusernotify.html
func (w *MiniProgram) SetUserNotify(ctx context.Context, req *SetUserNotifyRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/set_user_notify", nil, defaultReqOptions(), req, nil)
}

// SetUserNotifyExtRequest attaches extra state to a user notify.
type SetUserNotifyExtRequest struct {
	UserNotifyQuery
	// ExtraJSON carries the additional state data.
	ExtraJSON string `json:"extra_json,omitempty"`
}

// SetUserNotifyExt updates the extra fields of a user notify card.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_setusernotifyext.html
func (w *MiniProgram) SetUserNotifyExt(ctx context.Context, req *SetUserNotifyExtRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/set_user_notifyext", nil, defaultReqOptions(), req, nil)
}
