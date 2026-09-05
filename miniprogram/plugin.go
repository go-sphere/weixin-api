package miniprogram

import (
	"context"
)

// PluginStatus is the state of a plugin application.
type PluginStatus int

const (
	// PluginStatusApplied means a plugin has applied to be used.
	PluginStatusApplied PluginStatus = 1
	// PluginStatusApproved means the plugin application was approved.
	PluginStatusApproved PluginStatus = 2
	// PluginStatusRejected means the plugin application was rejected.
	PluginStatusRejected PluginStatus = 3
	// PluginStatusUnderReview means the plugin application is being reviewed.
	PluginStatusUnderReview PluginStatus = 4
)

// PluginItem describes one plugin attached to the Mini Program.
type PluginItem struct {
	// AppID of the plugin.
	AppID string `json:"appid"`
	// Status of the plugin (see PluginStatus).
	Status int `json:"status"`
	// Nickname of the plugin.
	Nickname string `json:"nickname"`
	// HeadImgURL of the plugin icon.
	HeadImgURL string `json:"headimgurl"`
}

// GetPluginListResponse is returned by GetPluginList.
type GetPluginListResponse struct {
	ErrResponse
	// PluginList is the list of plugins bound to the Mini Program.
	PluginList []PluginItem `json:"plugin_list"`
}

// GetPluginList returns the plugins currently usable by the Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/plugin-management/api_manageplugin.html
func (w *MiniProgram) GetPluginList(ctx context.Context) (*GetPluginListResponse, error) {
	var result GetPluginListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/plugin", nil, defaultReqOptions(), map[string]string{"action": "list"}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ApplyPluginRequest requests to use a plugin.
type ApplyPluginRequest struct {
	// AppID of the plugin to apply for.
	AppID string `json:"plugin_appid"`
	// Reason why the plugin is needed (max 200 chars, required when the
	// plugin requires an application).
	Reason string `json:"reason,omitempty"`
}

// ApplyPlugin submits an application to use a plugin. Whether the application
// goes through immediately depends on the plugin's policy.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/plugin-management/api_managepluginapplication.html
func (w *MiniProgram) ApplyPlugin(ctx context.Context, req *ApplyPluginRequest) error {
	body := map[string]any{"action": "apply", "plugin_appid": req.AppID}
	if req.Reason != "" {
		body["reason"] = req.Reason
	}
	return w.withAccessTokenPost(ctx, "/wxa/plugin", nil, defaultReqOptions(), body, nil)
}

// UnbindPlugin removes a plugin from the Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/plugin-management/api_manageplugin.html
func (w *MiniProgram) UnbindPlugin(ctx context.Context, pluginAppID string) error {
	body := map[string]string{
		"action":       "unbind",
		"plugin_appid": pluginAppID,
	}
	return w.withAccessTokenPost(ctx, "/wxa/plugin", nil, defaultReqOptions(), body, nil)
}

// PluginDevApplyListResponse lists plugin developers who applied to be used by
// the Mini Program.
type PluginDevApplyListResponse struct {
	ErrResponse
	// PluginList of the applications (plugin_appid in each entry).
	PluginList []PluginDevApplyItem `json:"plugin_list,omitempty"`
}

// PluginDevApplyItem is one plugin developer application row.
type PluginDevApplyItem struct {
	// AppID of the plugin.
	AppID string `json:"plugin_appid"`
	// Status of the application.
	Status int `json:"status"`
	// Nickname of the plugin.
	Nickname string `json:"nickname,omitempty"`
	// HeadImgURL of the plugin.
	HeadImgURL string `json:"headimgurl,omitempty"`
}

// GetPluginDevApplyList returns the plugin applications addressed to this Mini
// Program when the app is the plugin user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/plugin-management/api_managepluginapplication.html
func (w *MiniProgram) GetPluginDevApplyList(ctx context.Context) (*PluginDevApplyListResponse, error) {
	var result PluginDevApplyListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/devplugin", nil, defaultReqOptions(), map[string]string{"action": "dev_apply_list"}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetPluginDevApplyStatusRequest approves or rejects a plugin developer's
// application to be used by this Mini Program.
type SetPluginDevApplyStatusRequest struct {
	// AppID of the plugin whose application is decided.
	AppID string `json:"plugin_appid"`
	// Status to set: 2 = agree, 3 = reject.
	Status int `json:"status"`
	// Reason is required when rejecting.
	Reason string `json:"reason,omitempty"`
}

// SetPluginDevApplyStatus decides a pending plugin application.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/plugin-management/api_managepluginapplication.html
func (w *MiniProgram) SetPluginDevApplyStatus(ctx context.Context, req *SetPluginDevApplyStatusRequest) error {
	body := map[string]any{
		"action":       "dev_agree",
		"plugin_appid": req.AppID,
		"status":       req.Status,
	}
	if req.Reason != "" {
		body["reason"] = req.Reason
	}
	return w.withAccessTokenPost(ctx, "/wxa/devplugin", nil, defaultReqOptions(), body, nil)
}
