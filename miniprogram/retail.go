package miniprogram

import (
	"context"
)

// ============================================================
// B2B 零售 (B2b retail / 微信小店供应商) — store assistant 门店助手 and
// notification APIs.
// ============================================================

// RetailInfo is one store (门店) of a batch import.
type RetailInfo struct {
	// Name of the store.
	Name string `json:"name,omitempty"`
	// Address of the store.
	Address string `json:"address,omitempty"`
	// Province of the store.
	Province string `json:"province,omitempty"`
	// City of the store.
	City string `json:"city,omitempty"`
	// District of the store.
	District string `json:"district,omitempty"`
	// Lng of the store.
	Lng float64 `json:"lng,omitempty"`
	// Lat of the store.
	Lat float64 `json:"lat,omitempty"`
	// AdminName of the store admin.
	AdminName string `json:"admin_name,omitempty"`
	// AdminMobile of the store admin.
	AdminMobile string `json:"admin_mobile,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// BatchCreateRetailRequest imports stores in batch.
type BatchCreateRetailRequest struct {
	// RetailInfoList of up to 100 stores.
	RetailInfoList []map[string]any `json:"retail_info_list"`
}

// BatchCreateRetail imports up to 100 stores (门店) per call.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/store_assistant/api_batchcreateretail.html
func (w *MiniProgram) BatchCreateRetail(ctx context.Context, stores []map[string]any) error {
	req := &BatchCreateRetailRequest{RetailInfoList: stores}
	return w.withAccessTokenPost(ctx, "/wxa/business/batchcreateretail", nil, defaultReqOptions(), req, nil)
}

// GetRetailInfoRequest selects a store by openid or mobile.
type GetRetailInfoRequest struct {
	// OpenID of the store admin/employee (alternative to MobilePhone).
	OpenID string `json:"openid,omitempty"`
	// MobilePhone of the store admin (alternative to OpenID).
	MobilePhone string `json:"mobile_phone,omitempty"`
}

// GetRetailInfoResponse is returned by GetRetailInfo.
type GetRetailInfoResponse struct {
	ErrResponse
	// RetailInfo of the queried store.
	RetailInfo map[string]any `json:"retail_info,omitempty"`
}

// GetRetailInfo returns the store bound to an admin openid or phone.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/store_assistant/api_getretailinfo.html
func (w *MiniProgram) GetRetailInfo(ctx context.Context, req *GetRetailInfoRequest) (*GetRetailInfoResponse, error) {
	var result GetRetailInfoResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/getretailinfo", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRetailOpenIDListRequest pages the store-admin openids.
type GetRetailOpenIDListRequest struct {
	// Limit of the page (1-100).
	Limit int `json:"limit"`
	// PageContext cursor of the previous page (empty on first call).
	PageContext string `json:"page_context"`
}

// GetRetailOpenIDListResponse is returned by GetRetailOpenIDList.
type GetRetailOpenIDListResponse struct {
	ErrResponse
	// OpenIDList of the page.
	OpenIDList []string `json:"openid_list,omitempty"`
	// PageContext cursor for the next page.
	PageContext string `json:"page_context,omitempty"`
}

// GetRetailOpenIDList pages through all store-admin openids of the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/store_assistant/api_getretailopenidlist.html
func (w *MiniProgram) GetRetailOpenIDList(ctx context.Context, req *GetRetailOpenIDListRequest) (*GetRetailOpenIDListResponse, error) {
	var result GetRetailOpenIDListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/getretailopenidlist", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RetailBusinessApplyRequest applies to become a B2b retail service provider.
type RetailBusinessApplyRequest struct {
	// GoodsTypeList of the main product categories.
	GoodsTypeList []string `json:"goods_type_list"`
	// GoodsSaleList of the main offline sales channels.
	GoodsSaleList []string `json:"goods_sale_list"`
	// CoverNum of covered stores.
	CoverNum string `json:"cover_num"`
	// ServiceList of the requested services.
	ServiceList []string `json:"service_list"`
	// Description of the Mini Program solution (21-100 chars).
	Description string `json:"description"`
	// ContactName of the contact (1-7 chars).
	ContactName string `json:"contact_name"`
	// ContactPhone of the contact.
	ContactPhone string `json:"contact_phone"`
	// ContactEmail of the contact.
	ContactEmail string `json:"contact_email"`
}

// RetailBusinessApply submits a service-provider application (商家申请).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/store_assistant/api_retailbusinessapply.html
func (w *MiniProgram) RetailBusinessApply(ctx context.Context, req *RetailBusinessApplyRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/retailbusinessapply", nil, defaultReqOptions(), req, nil)
}

// GetRetailMessageListRequest pages the retail template messages.
type GetRetailMessageListRequest struct {
	// Start of the page (from 0).
	Start int `json:"start"`
	// Offset of the page size (max 1000, default 20).
	Offset int `json:"offset"`
	// BeginDate of the window ("xxxx-xx-xx").
	BeginDate string `json:"begin_date"`
	// EndDate of the window.
	EndDate string `json:"end_date"`
}

// GetRetailMessageListResponse is returned by GetRetailMessageList.
type GetRetailMessageListResponse struct {
	ErrResponse
	// MessageList of the sent template messages.
	MessageList []map[string]any `json:"message_list,omitempty"`
}

// GetRetailMessageList lists the retail-assistant template messages sent.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/notify/api_getretailmessagelist.html
func (w *MiniProgram) GetRetailMessageList(ctx context.Context, req *GetRetailMessageListRequest) (*GetRetailMessageListResponse, error) {
	var result GetRetailMessageListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/getretailmessagelist", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// NotifyRetailBusinessRequest sends a template message to store owners.
type NotifyRetailBusinessRequest struct {
	// Type of the message template (see the B2b 门店助手 template list).
	Type int `json:"type"`
	// ToUserList of the store owner openids (max 200).
	ToUserList []string `json:"to_user_list"`
	// Content is a JSON string of the message data per template.
	Content string `json:"content,omitempty"`
}

// NotifyRetailBusiness sends a retail-assistant template message to store
// owners.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/notify/api_retailnotifybusiness.html
func (w *MiniProgram) NotifyRetailBusiness(ctx context.Context, req *NotifyRetailBusinessRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/retailnotifybusiness", nil, defaultReqOptions(), req, nil)
}
