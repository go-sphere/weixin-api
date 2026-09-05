package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// CategoryItem describes one Mini Program category usable by subscribe message
// templates.
type CategoryItem struct {
	// ID of the category.
	ID int `json:"id"`
	// Name of the category.
	Name string `json:"name"`
}

// GetCategoryResponse is returned by GetSubscribeMessageCategory.
type GetCategoryResponse struct {
	ErrResponse
	// Data is the list of categories the Mini Program belongs to.
	Data []CategoryItem `json:"data"`
}

// GetSubscribeMessageCategory returns the service categories of the Mini
// Program, from which subscribe message templates can be picked.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_getcategory.html
func (w *MiniProgram) GetSubscribeMessageCategory(ctx context.Context) (*GetCategoryResponse, error) {
	var result GetCategoryResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/newtmpl/getcategory", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TemplateTitle is a publishable subscribe message template title.
type TemplateTitle struct {
	// ID of the template title.
	ID int `json:"tid"`
	// Title of the template.
	Title string `json:"title"`
	// Type of the template (e.g. 0 = one-time, 2 = long-term).
	Type int `json:"type"`
	// CategoryID of the category the template belongs to.
	CategoryID int `json:"category_id"`
}

// GetTemplateTitleListResponse is returned by
// GetSubscribeMessageTemplateTitles.
type GetTemplateTitleListResponse struct {
	ErrResponse
	// Count is the total number of matching titles.
	Count int `json:"count"`
	// Data is the page of template titles.
	Data []TemplateTitle `json:"data"`
}

// GetSubscribeMessageTemplateTitles lists the template titles that can be
// subscribed under a category keyword combination.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_getpubnewtemplatetitles.html
func (w *MiniProgram) GetSubscribeMessageTemplateTitles(ctx context.Context, categoryID int, page, count int) (*GetTemplateTitleListResponse, error) {
	query := url.Values{}
	query.Set("ids", fmtInt(categoryID))
	query.Set("start", fmtInt(page))
	query.Set("limit", fmtInt(count))
	var result GetTemplateTitleListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/newtmpl/getpubtemplatetitles", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TemplateKeyword is one keyword of a subscribe message template.
type TemplateKeyword struct {
	// KeywordID of the keyword.
	KeywordID int `json:"keyword_id"`
	// Name of the keyword.
	Name string `json:"name"`
	// Example shown for the keyword.
	Example string `json:"example"`
}

// GetTemplateKeywordListResponse is returned by
// GetSubscribeMessageTemplateKeywords.
type GetTemplateKeywordListResponse struct {
	ErrResponse
	// Data is the list of keywords usable in the template.
	Data []TemplateKeyword `json:"data"`
}

// GetSubscribeMessageTemplateKeywords lists the keywords of a template title
// so the app can design its template content.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_getpubnewtemplatekeywords.html
func (w *MiniProgram) GetSubscribeMessageTemplateKeywords(ctx context.Context, templateID int) (*GetTemplateKeywordListResponse, error) {
	query := url.Values{}
	query.Set("tid", fmtInt(templateID))
	var result GetTemplateKeywordListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/newtmpl/getpubtemplatekeywords", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SubscribeTemplateItem is a template already added to the Mini Program.
type SubscribeTemplateItem struct {
	// ID (priTmplId) of the added template used when sending messages.
	ID string `json:"priTmplId"`
	// Title of the template.
	Title string `json:"title"`
	// Content of the template.
	Content string `json:"content"`
	// Example of the template.
	Example string `json:"example"`
	// Type of the template (one-time or long-term).
	Type int `json:"type"`
}

// GetSubscribeTemplateListResponse is returned by
// GetSubscribeMessageTemplateList.
type GetSubscribeTemplateListResponse struct {
	ErrResponse
	// Data is the list of added templates.
	Data []SubscribeTemplateItem `json:"data"`
}

// GetSubscribeMessageTemplateList lists the subscribe message templates already
// added to this Mini Program, returning their priTmplId for sending.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_getwxapubnewtemplate.html
func (w *MiniProgram) GetSubscribeMessageTemplateList(ctx context.Context) (*GetSubscribeTemplateListResponse, error) {
	var result GetSubscribeTemplateListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/newtmpl/gettemplate", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddSubscribeTemplateRequest selects the template to add.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_addwxanewtemplate.html
type AddSubscribeTemplateRequest struct {
	// TemplateID (tid) of the template title.
	TemplateID int `json:"tid"`
	// KeywordIDList are the keyword ids composing the template, in the order
	// the message will use them.
	KeywordIDList []int `json:"kidList"`
	// SceneDescription explains the scenario when the subscription is
	// required (up to 200 chars).
	SceneDescription string `json:"sceneDesc"`
}

// AddSubscribeTemplateResponse is returned by AddSubscribeMessageTemplate.
type AddSubscribeTemplateResponse struct {
	ErrResponse
	// TemplateID is the priTmplId assigned to the new template.
	TemplateID string `json:"priTmplId"`
}

// AddSubscribeMessageTemplate adds a new subscribe message template from an
// approved title+keyword combination. The returned TemplateID is used in
// SendSubscribeMessage and in the client-side wx.requestSubscribeMessage.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_addwxanewtemplate.html
func (w *MiniProgram) AddSubscribeMessageTemplate(ctx context.Context, req *AddSubscribeTemplateRequest) (*AddSubscribeTemplateResponse, error) {
	var result AddSubscribeTemplateResponse
	if err := w.withAccessTokenPost(ctx, "/wxaapi/newtmpl/addtemplate", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSubscribeMessageTemplateRequest identifies the template to delete.
type DeleteSubscribeMessageTemplateRequest struct {
	// TemplateID is the priTmplId to delete.
	TemplateID string `json:"priTmplId"`
}

// DeleteSubscribeMessageTemplate removes a previously added subscribe message
// template.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_delwxanewtemplate.html
func (w *MiniProgram) DeleteSubscribeMessageTemplate(ctx context.Context, templateID string) error {
	body := &DeleteSubscribeMessageTemplateRequest{TemplateID: templateID}
	return w.withAccessTokenPost(ctx, "/wxaapi/newtmpl/deltemplate", nil, defaultReqOptions(), body, nil)
}

// SubscribeMessageData maps template keyword ids to their values. The JSON
// shape is {"key1":{"value":"v1"},...} as required by the send API.
type SubscribeMessageData map[string]struct {
	Value string `json:"value"`
}

// SendSubscribeMessageRequest is the payload of SendSubscribeMessage.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_sendmessage.html
type SendSubscribeMessageRequest struct {
	// ToUser is the openid of the receiving user (required).
	ToUser string `json:"touser"`
	// TemplateID is the priTmplId of the added template (required).
	TemplateID string `json:"template_id"`
	// Page to jump to after the message is tapped (optional, Mini Program
	// internal page, may carry query).
	Page string `json:"page,omitempty"`
	// MiniprogramState selects the environment the jump page opens in:
	// developer, trial or formal (default formal).
	MiniprogramState string `json:"miniprogram_state,omitempty"`
	// Lang of the message content: zh_CN (default), en_US, zh_HK or zh_TW.
	Lang string `json:"lang,omitempty"`
	// Data maps each keyword of the template to its value.
	Data SubscribeMessageData `json:"data"`
}

// SendSubscribeMessage delivers a subscribe message to a user. The user must
// previously have granted the matching template subscription.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/mp-message-management/subscribe-message/api_sendmessage.html
func (w *MiniProgram) SendSubscribeMessage(ctx context.Context, req *SendSubscribeMessageRequest) error {
	if req.MiniprogramState == "" {
		req.MiniprogramState = w.config.Env.subscribeState()
	}
	return w.withAccessTokenPost(ctx, "/cgi-bin/message/subscribe/send", nil, defaultReqOptions(), req, nil)
}
