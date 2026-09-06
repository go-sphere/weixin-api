package miniprogram

import (
	"context"
)

// ============================================================
// Remaining endpoints: quota clearing, express local real mock &
// provider order update, delivery message push (provider side),
// cloudbase extension upload, boot performance data, retail B2b
// merchant registration, and the minidrama authorisation/usage
// endpoints. Each method keeps the exact upstream route.
// ============================================================

// ClearQuotaLegacy clears the whole account's daily API quota via the legacy
// route /cgi-bin/clear_quota (the request body carries the account appid, not a
// cgi_path). Prefer ClearAPICallQuota when available.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_clearquota.html
func (w *MiniProgram) ClearQuotaLegacy(ctx context.Context) error {
	body := map[string]string{"appid": w.config.AppID}
	return w.withAccessTokenPost(ctx, "/cgi-bin/clear_quota", nil, defaultReqOptions(), body, nil)
}

// ClearQuotaByAppSecretRequest clears the quota of a Mini Program identified
// by appid/appsecret.
type ClearQuotaByAppSecretRequest struct {
	// AppID of the account to clear.
	AppID string `json:"appid"`
	// AppSecret of the account to clear.
	AppSecret string `json:"appsecret"`
}

// ClearQuotaByAppSecret clears an account's API quota using its appsecret.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/openApi-mgnt/api_clearquotabyappsecret.html
func (w *MiniProgram) ClearQuotaByAppSecret(ctx context.Context, req *ClearQuotaByAppSecretRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/clear_quota/v2", nil, defaultReqOptions(), req, nil)
}

// QueryDeliveryUserBinding checks whether a phone number is bound to a WeChat
// delivery user (provider side).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/msgpush/api_deliveryuserquery.html
func (w *MiniProgram) QueryDeliveryUserBinding(ctx context.Context, phone string) (bool, error) {
	var result struct {
		ErrResponse
		Exist int `json:"exist"`
	}
	body := map[string]string{"phone": phone}
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/userquery", nil, defaultReqOptions(), body, &result); err != nil {
		return false, err
	}
	return result.Exist == 1, nil
}

// DeliveryPathNotifyRequest is the payload of the delivery path push notify
// (provider pushes a new tracking point to the user).
type DeliveryPathNotifyRequest struct {
	// Sender of the waybill.
	Sender map[string]any `json:"sender"`
	// Receiver of the waybill.
	Receiver map[string]any `json:"receiver"`
	// WaybillID of the waybill.
	WaybillID string `json:"waybill_id"`
	// Path is the current tracking point to push.
	Path map[string]any `json:"path"`
	// CreateTime of the waybill (unix seconds).
	CreateTime int64 `json:"create_time"`
}

// DeliveryPathNotify pushes a delivery path update message to the recipient.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/weixin-express/msgpush/api_deliverypathnotify.html
func (w *MiniProgram) DeliveryPathNotify(ctx context.Context, req *DeliveryPathNotifyRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/delivery/pathnotify", nil, defaultReqOptions(), req, nil)
}

// RealMockUpdateOrderRequest simulates a real order update during testing.
type RealMockUpdateOrderRequest struct {
	// ShopID assigned by the delivery provider.
	ShopID string `json:"shopid"`
	// ShopOrderID of the merchant order.
	ShopOrderID string `json:"shop_order_id"`
	// OrderStatus of the delivery.
	OrderStatus int `json:"order_status"`
	// ActionTime of the state change (unix seconds).
	ActionTime int64 `json:"action_time"`
	// ActionMsg extra info.
	ActionMsg string `json:"action_msg,omitempty"`
	// DeliverySign checksum signed with the provider appsecret.
	DeliverySign string `json:"delivery_sign"`
}

// RealMockUpdateOrder simulates a real provider order update (sandbox).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-business/api_realmockupdateorder.html
func (w *MiniProgram) RealMockUpdateOrder(ctx context.Context, req *RealMockUpdateOrderRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/local/business/realmock_update_order", nil, defaultReqOptions(), req, nil)
}

// UpdateLocalOrderByProviderRequest updates an order by the local delivery
// provider.
type UpdateLocalOrderByProviderRequest struct {
	// WxToken pushed by the order event.
	WxToken string `json:"wx_token"`
	// OrderStatus of the delivery.
	OrderStatus int `json:"order_status"`
	// WaybillID of the delivery order.
	WaybillID string `json:"waybill_id"`
	// ActionMsg extra info.
	ActionMsg string `json:"action_msg,omitempty"`
	// ActionTime of the change (unix seconds).
	ActionTime int64 `json:"action_time"`
	// Agent courier info (required when a courier accepts).
	Agent map[string]any `json:"agent,omitempty"`
	// ShopID assigned by the provider.
	ShopID string `json:"shopid"`
	// ShopOrderID of the merchant order.
	ShopOrderID string `json:"shop_order_id"`
	// ShopNo of the store registered at the provider.
	ShopNo string `json:"shop_no,omitempty"`
	// WxaPath of the provider Mini Program page for the user.
	WxaPath string `json:"wxa_path"`
	// ExpectedDeliveryTime (unix seconds) when a courier accepts.
	ExpectedDeliveryTime int64 `json:"expected_delivery_time,omitempty"`
}

// UpdateLocalOrderByProvider reports an order-state change from the local
// delivery provider side.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/immediate-delivery/deliver-by-provider/api_updateorder.html
func (w *MiniProgram) UpdateLocalOrderByProvider(ctx context.Context, req *UpdateLocalOrderByProviderRequest) error {
	return w.withAccessTokenPost(ctx, "/cgi-bin/express/local/delivery/update_order", nil, defaultReqOptions(), req, nil)
}

// RegisterOnlyWqf activates 微信支付分 only for an existing B2b onboarding
// order.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_registeronlywqf.html
func (w *MiniProgram) RegisterOnlyWqf(ctx context.Context, outRegistrationID string) error {
	body := map[string]string{"out_registration_id": outRegistrationID}
	return w.b2bPost(ctx, "/retail/B2b/registeronlywqf", body, nil)
}

// ListRetailMchOrderRequest pages the B2b merchant onboarding orders.
type ListRetailMchOrderRequest struct {
	// OutRegistrationID of a specific onboarding order.
	OutRegistrationID string `json:"out_registration_id,omitempty"`
	// PageIndex of the page.
	PageIndex int `json:"page_index,omitempty"`
	// PageSize of the page.
	PageSize int `json:"page_size,omitempty"`
}

// ListRetailMchOrderResponse is returned by ListRetailMchOrder.
type ListRetailMchOrderResponse struct {
	ErrResponse
	// RawData of the onboarding order list.
	RawData map[string]any `json:"-"`
}

// ListRetailMchOrder lists the B2b merchant onboarding orders.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_retailgetmchorder.html
func (w *MiniProgram) ListRetailMchOrder(ctx context.Context, req *ListRetailMchOrderRequest) (*ListRetailMchOrderResponse, error) {
	var result ListRetailMchOrderResponse
	if err := w.b2bPost(ctx, "/retail/B2b/retailgetmchorder", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RetailRegisterMchRequest opens a B2b retail merchant (进件). The payload is
// deep and versioned, so it is accepted as a raw JSON object; see the official
// doc for the exact schema.
type RetailRegisterMchRequest struct {
	// RawData of the onboarding request.
	RawData map[string]any `json:"-"`
}

// RetailRegisterMch opens (registers) a retail sub merchant for B2b payment.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_retailregistermch.html
func (w *MiniProgram) RetailRegisterMch(ctx context.Context, body map[string]any) error {
	return w.b2bPost(ctx, "/retail/B2b/retailregistermch", body, nil)
}

// RetailUploadMchFileResponse is returned by RetailUploadMchFile.
type RetailUploadMchFileResponse struct {
	ErrResponse
	// MediaID of the uploaded merchant file.
	MediaID string `json:"media_id,omitempty"`
	// RawData of the response.
	RawData map[string]any `json:"-"`
}

// RetailUploadMchFile uploads a merchant material image used by the B2b
// onboarding APIs.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/B2b/bill/api_retailuploadmchfile.html
func (w *MiniProgram) RetailUploadMchFile(ctx context.Context, fileName string, file []byte) (*RetailUploadMchFileResponse, error) {
	body := map[string]any{"file_name": fileName, "file": file}
	var result RetailUploadMchFileResponse
	if err := w.b2bPost(ctx, "/retail/B2b/retailuploadmchfile", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TCBExtensionFile is one file of an extension-upload info request.
type TCBExtensionFile struct {
	// FileName of the file.
	FileName string `json:"file_name,omitempty"`
	// RawData of the file entry.
	RawData map[string]any `json:"-"`
}

// TCBDescribeExtensionUploadInfo returns the upload info (signed URLs) for the
// cloud extension files.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/cloudbase/others/api_describeextensionuploadinfo.html
func (w *MiniProgram) TCBDescribeExtensionUploadInfo(ctx context.Context, files []map[string]any) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	body := map[string]any{"ExtensionFiles": files}
	if err := w.withAccessTokenPost(ctx, "/tcb/describeextensionuploadinfo", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// BootPerformanceDataRequest queries the boot (启动) performance data.
type BootPerformanceDataRequest struct {
	// Module of the query data type.
	Module int `json:"module"`
	// Time is the {begin,end} window (max 30 days).
	Time map[string]int64 `json:"time"`
	// Params of the query conditions (device, network, etc).
	Params []map[string]any `json:"params"`
}

// BootPerformanceData returns Mini Program boot performance statistics.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/data-analysis/others/api_getperformancedata.html
func (w *MiniProgram) BootPerformanceData(ctx context.Context, req *BootPerformanceDataRequest) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := w.withAccessTokenPost(ctx, "/wxa/business/performance/boot", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// GetMiniDramaMedia returns one episode media of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_media/api_getmedia.html
func (w *MiniProgram) GetMiniDramaMedia(ctx context.Context, mediaID int64) (map[string]any, error) {
	var result struct {
		ErrResponse
		Media map[string]any `json:"media,omitempty"`
	}
	body := map[string]int64{"media_id": mediaID}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getmedia", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return result.Media, nil
}

// ListMiniDramaPackagesResponse is returned by ListMiniDramaPackages.
type ListMiniDramaPackagesResponse struct {
	ErrResponse
	// PackageList of the CDN packages.
	PackageList []map[string]any `json:"package_list,omitempty"`
}

// ListMiniDramaPackages lists the CDN usage packages of the drama account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/usagedata/api_listpackages.html
func (w *MiniProgram) ListMiniDramaPackages(ctx context.Context) (*ListMiniDramaPackagesResponse, error) {
	var result ListMiniDramaPackagesResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/listpackages", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MiniDramaCDNUsageRequest selects the CDN usage statistics window.
type MiniDramaCDNUsageRequest struct {
	// StartTime of the window (unix seconds).
	StartTime int64 `json:"start_time"`
	// EndTime of the window (unix seconds).
	EndTime int64 `json:"end_time"`
	// DataInterval in minutes: 5, 60 or 1440.
	DataInterval string `json:"data_interval"`
	// QueryType: 0 all, 1 short-drama directed, 2 generic play traffic.
	QueryType int `json:"query_type,omitempty"`
}

// GetMiniDramaCDNUsageData returns the CDN usage statistics of the drama
// media.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/usagedata/api_getcdnusagedata.html
func (w *MiniProgram) GetMiniDramaCDNUsageData(ctx context.Context, req *MiniDramaCDNUsageRequest) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getcdnusagedata", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// GetMiniDramaCDNLogsRequest selects the CDN log period.
type GetMiniDramaCDNLogsRequest struct {
	// StartTime of the window (unix seconds).
	StartTime int64 `json:"start_time,omitempty"`
	// EndTime of the window (unix seconds).
	EndTime int64 `json:"end_time,omitempty"`
	// RawData of the remaining selectors.
	RawData map[string]any `json:"-"`
}

// GetMiniDramaCDNLogs returns the CDN access logs of the drama media.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/usagedata/api_getcdnlogs.html
func (w *MiniProgram) GetMiniDramaCDNLogs(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getcdnlogs", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// ModifyMiniDramaBasicInfo updates the basic info of a drama. The payload is
// deep and versioned; pass the JSON object described by the reference doc.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/auditdrama/api_submitmodifydramabasicinforeq.html
func (w *MiniProgram) ModifyMiniDramaBasicInfo(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/modifydramabasicinfo", nil, defaultReqOptions(), body, nil)
}

// SingleFileMiniDramaUploadRequest uploads a single drama media file.
type SingleFileMiniDramaUploadRequest struct {
	// MediaName of the media (e.g. "我的演艺 - 第1集").
	MediaName string `json:"media_name"`
	// MediaType e.g. "MP4".
	MediaType string `json:"media_type"`
	// MediaData binary content (base64 string).
	MediaData string `json:"media_data"`
	// CoverType e.g. "JPG".
	CoverType string `json:"cover_type,omitempty"`
	// CoverData binary content (base64 string).
	CoverData string `json:"cover_data,omitempty"`
	// SourceContext of the upload.
	SourceContext string `json:"source_context,omitempty"`
}

// SingleFileMiniDramaUpload uploads a complete episode file.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_fileupload/api_singlefileupload.html
func (w *MiniProgram) SingleFileMiniDramaUpload(ctx context.Context, req *SingleFileMiniDramaUploadRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/singlefileupload", nil, defaultReqOptions(), req, nil)
}

// UploadMiniDramaPartRequest uploads one part of a chunked media upload.
type UploadMiniDramaPartRequest struct {
	// UploadID returned by ApplyMiniDramaUpload.
	UploadID string `json:"upload_id"`
	// PartNumber of the part (1-100).
	PartNumber int `json:"part_number"`
	// ResourceType: 1 video, 2 image.
	ResourceType int `json:"resource_type"`
	// Data binary part content.
	Data string `json:"data"`
}

// UploadMiniDramaPart uploads one chunk of a media file.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_fileupload/api_uploadpart.html
func (w *MiniProgram) UploadMiniDramaPart(ctx context.Context, req *UploadMiniDramaPartRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/uploadpart", nil, defaultReqOptions(), req, nil)
}

// SubmitMiniDramaReplaceMedias submits a set of media replacements for
// re-audit.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/auditdrama/api_submitreplacemedias.html
func (w *MiniProgram) SubmitMiniDramaReplaceMedias(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/submitreplacedramamedias", nil, defaultReqOptions(), body, nil)
}

// AuthorizeMiniDramaApp grants another app the use of a drama (app-level
// authorisation).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizeapp/api_authorizeapp.html
func (w *MiniProgram) AuthorizeMiniDramaApp(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/authorizeapp", nil, defaultReqOptions(), body, nil)
}

// DeauthorizeMiniDramaApp revokes the app-level drama authorisation.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizeapp/api_deauthorizeapp.html
func (w *MiniProgram) DeauthorizeMiniDramaApp(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/deauthorizeapp", nil, defaultReqOptions(), body, nil)
}

// GetMiniDramaAuthorizedApps returns the apps authorised for a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizeapp/api_getauthorizeapps.html
func (w *MiniProgram) GetMiniDramaAuthorizedApps(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getauthorizeapps", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// GetMiniDramaAuthorizedObjects returns the authorised objects of an app.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizedrama/api_getauthorizeobjects.html
func (w *MiniProgram) GetMiniDramaAuthorizedObjects(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getauthorizeobjects", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// AuthorizeMiniDramaCopyright applies for copyright protection of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizecopyright/api_authorizecopyright.html
func (w *MiniProgram) AuthorizeMiniDramaCopyright(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/authorizecopyright", nil, defaultReqOptions(), body, nil)
}

// DeauthorizeMiniDramaCopyright revokes a drama copyright application.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizecopyright/api_deauthorizecopyright.html
func (w *MiniProgram) DeauthorizeMiniDramaCopyright(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/deauthorizecopyright", nil, defaultReqOptions(), body, nil)
}

// ListMiniDramaCopyrightAuthorization returns the copyright applications the
// account received.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizecopyright/api_getcopyrightauthorizationlist.html
func (w *MiniProgram) ListMiniDramaCopyrightAuthorization(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getcopyrightauthorizationlist", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// ListMiniDramaCopyrightAuthorized returns the dramas the account is
// authorised for.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizecopyright/api_getcopyrightauthorizedlist.html
func (w *MiniProgram) ListMiniDramaCopyrightAuthorized(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getcopyrightauthorizedlist", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}

// GetXPayPunishmentReasons lists the punishment (处罚) reason options of the
// virtual-payment account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/VirtualPayment/api_query_punishment_reasons.html
func (w *MiniProgram) GetXPayPunishmentReasons(ctx context.Context) (map[string]any, error) {
	data, err := w.xpayRequest(ctx, "/xpay/query_punishment_reasons", map[string]string{}, "", false)
	if err != nil {
		return nil, err
	}
	var result struct {
		ErrResponse
		Raw map[string]any `json:"-"`
	}
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return result.Raw, nil
}
