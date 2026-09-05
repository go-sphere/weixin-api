package miniprogram

import (
	"context"
	"net/http"
	"net/url"
)

// ============================================================
// 视频号直播 / 小程序直播 (live broadcast) management.
// Goods, roles and room operations live under /wxaapi/broadcast/*.
// ============================================================

// BroadcastGoods describes one livestream shopping good (商品).
type BroadcastGoods struct {
	// GoodsID assigned by WeChat after creation.
	GoodsID int64 `json:"goodsId,omitempty"`
	// Name of the good (up to 14 Chinese chars).
	Name string `json:"name,omitempty"`
	// PageURL of the good detail Mini Program page.
	PageURL string `json:"url,omitempty"`
	// CoverImgURL is the cover image MediaID (valid 3 days after upload).
	CoverImgURL string `json:"coverImgUrl,omitempty"`
	// PriceType: 1 fixed price, 2 price range, 3 discounted price.
	PriceType int `json:"priceType,omitempty"`
	// Price is the left/fixed/original price in yuan.
	Price float64 `json:"price,omitempty"`
	// Price2 is the right/current price in yuan (for types 2 and 3).
	Price2 float64 `json:"price2,omitempty"`
	// ThirdPartyAppID is set when the good belongs to a third-party Mini
	// Program.
	ThirdPartyAppID string `json:"thirdPartyAppid,omitempty"`
	// ThirdPartyTag is the third-party good tag.
	ThirdPartyTag int `json:"thirdPartyTag,omitempty"`
	// AuditStatus of the good: 0 draft, 1 under review, 2 approved,
	// 3 rejected.
	AuditStatus int `json:"auditStatus,omitempty"`
}

// AddBroadcastGoodsRequest adds a good to the live goods library.
type AddBroadcastGoodsRequest struct {
	// GoodsInfo of the new good.
	GoodsInfo BroadcastGoods `json:"goodsInfo"`
}

// AddBroadcastGoodsResponse is returned by AddBroadcastGoods.
type AddBroadcastGoodsResponse struct {
	ErrResponse
	// GoodsID of the created good.
	GoodsID int64 `json:"goodsId,omitempty"`
	// AuditID of the auto-submitted audit.
	AuditID int64 `json:"auditId,omitempty"`
}

// AddBroadcastGoods adds a good and submits it for audit.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/commodity-management/api_addgoods.html
func (w *MiniProgram) AddBroadcastGoods(ctx context.Context, req *AddBroadcastGoodsRequest) (*AddBroadcastGoodsResponse, error) {
	var result AddBroadcastGoodsResponse
	if err := w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/add", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateBroadcastGoodsRequest updates an existing good before audit.
type UpdateBroadcastGoodsRequest struct {
	// GoodsID of the good to update.
	GoodsID int64 `json:"goodsId"`
	// GoodsInfo of the new good payload.
	GoodsInfo BroadcastGoods `json:"goodsInfo"`
}

// UpdateBroadcastGoods modifies a good in the library.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/commodity-management/api_updategoodsinfo.html
func (w *MiniProgram) UpdateBroadcastGoods(ctx context.Context, req *UpdateBroadcastGoodsRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/update", nil, defaultReqOptions(), req, nil)
}

// DeleteBroadcastGoods removes a good from the library.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/commodity-management/api_deletegoodsinfo.html
func (w *MiniProgram) DeleteBroadcastGoods(ctx context.Context, goodsID int64) error {
	body := map[string]int64{"goodsId": goodsID}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/delete", nil, defaultReqOptions(), body, nil)
}

// ResubmitBroadcastGoodsAuditRequest re-submits a rejected good for audit.
type ResubmitBroadcastGoodsAuditRequest struct {
	// GoodsID of the good.
	GoodsID int64 `json:"goodsId"`
}

// ResubmitBroadcastGoodsAudit resubmits a good that was rejected.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/commodity-management/api_resubmitaudit.html
func (w *MiniProgram) ResubmitBroadcastGoodsAudit(ctx context.Context, goodsID int64) error {
	body := &ResubmitBroadcastGoodsAuditRequest{GoodsID: goodsID}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/audit", nil, defaultReqOptions(), body, nil)
}

// ResetBroadcastGoodsAuditRequest resets an approved good for re-audit.
type ResetBroadcastGoodsAuditRequest struct {
	// GoodsID of the good.
	GoodsID int64 `json:"goodsId"`
	// AuditID of the previous audit.
	AuditID int64 `json:"auditId"`
}

// ResetBroadcastGoodsAudit revokes an approved good back to draft so it can be
// edited and re-audited.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/commodity-management/api_resetaudit.html
func (w *MiniProgram) ResetBroadcastGoodsAudit(ctx context.Context, goodsID, auditID int64) error {
	body := &ResetBroadcastGoodsAuditRequest{GoodsID: goodsID, AuditID: auditID}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/resetaudit", nil, defaultReqOptions(), body, nil)
}

// BroadcastWarehouseGoods is one good of the warehouse query result.
type BroadcastWarehouseGoods struct {
	// GoodsID of the good.
	GoodsID int64 `json:"goods_id"`
	// Name of the good.
	Name string `json:"name"`
	// PageURL of the good page.
	PageURL string `json:"url"`
	// CoverImgURL of the cover.
	CoverImgURL string `json:"cover_img_url"`
	// PriceType of the price model.
	PriceType int `json:"price_type"`
	// Price of the good.
	Price float64 `json:"price"`
	// Price2 of the good.
	Price2 float64 `json:"price2"`
	// AuditStatus of the good.
	AuditStatus int `json:"audit_status"`
	// ThirdPartyTag of the third-party good.
	ThirdPartyTag int `json:"third_party_tag"`
}

// GetBroadcastGoodsWarehouseResponse is returned by
// GetBroadcastGoodsWarehouse.
type GetBroadcastGoodsWarehouseResponse struct {
	ErrResponse
	// Goods of the requested ids.
	Goods []BroadcastWarehouseGoods `json:"goods,omitempty"`
}

// GetBroadcastGoodsWarehouse queries goods of the WeChat good warehouse
// (微信商品库) by their ids.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/commodity-management/api_getgoodsauditinfo.html
func (w *MiniProgram) GetBroadcastGoodsWarehouse(ctx context.Context, goodsIDs []int64) (*GetBroadcastGoodsWarehouseResponse, error) {
	body := map[string]any{"goods_ids": goodsIDs}
	var result GetBroadcastGoodsWarehouseResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/getgoodswarehouse", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetApprovedBroadcastGoodsResponse is returned by GetApprovedBroadcastGoods.
type GetApprovedBroadcastGoodsResponse struct {
	ErrResponse
	// Goods of the page (approved goods usable in rooms).
	Goods []BroadcastGoods `json:"goods,omitempty"`
	// Total number of approved goods.
	Total int64 `json:"total,omitempty"`
}

// GetApprovedBroadcastGoods lists the approved goods of the account, paged.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/commodity-management/api_getgoodsinfo.html
func (w *MiniProgram) GetApprovedBroadcastGoods(ctx context.Context, offset, limit, status int) (*GetApprovedBroadcastGoodsResponse, error) {
	query := url.Values{}
	query.Set("offset", fmtInt(offset))
	query.Set("limit", fmtInt(limit))
	if status != 0 {
		query.Set("status", fmtInt(status))
	}
	var result GetApprovedBroadcastGoodsResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/broadcast/goods/getapproved", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetBroadcastGoodsOnSaleRequest controls whether a good is on sale in a room.
type SetBroadcastGoodsOnSaleRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// GoodsID of the good.
	GoodsID int64 `json:"goodsId"`
	// OnSale enables/disables the good shelf.
	OnSale bool `json:"onSale"`
}

// SetBroadcastGoodsOnSale puts a good on sale or takes it off sale in a room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_salegoods.html
func (w *MiniProgram) SetBroadcastGoodsOnSale(ctx context.Context, req *SetBroadcastGoodsOnSaleRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/onsale", nil, defaultReqOptions(), req, nil)
}

// GetBroadcastGoodsVideoRequest fetches the promo video of a good in a room.
type GetBroadcastGoodsVideoRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// GoodsID of the good.
	GoodsID int64 `json:"goodsId"`
}

// GetBroadcastGoodsVideoResponse is returned by GetBroadcastGoodsVideo.
type GetBroadcastGoodsVideoResponse struct {
	ErrResponse
	// URL of the good video.
	URL string `json:"url,omitempty"`
}

// GetBroadcastGoodsVideo returns the promo video URL of a good.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_downloadgoodsvideo.html
func (w *MiniProgram) GetBroadcastGoodsVideo(ctx context.Context, roomID, goodsID int64) (*GetBroadcastGoodsVideoResponse, error) {
	body := &GetBroadcastGoodsVideoRequest{RoomID: roomID, GoodsID: goodsID}
	var result GetBroadcastGoodsVideoResponse
	if err := w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/getVideo", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetBroadcastDefaultGoodsKey sets the "default goods" key of the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_setdefault_goodskey.html
func (w *MiniProgram) SetBroadcastDefaultGoodsKey(ctx context.Context) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/setkey", nil, defaultReqOptions(), map[string]string{}, nil)
}

// GetBroadcastDefaultGoodsKeyResponse is returned by
// GetBroadcastDefaultGoodsKey.
type GetBroadcastDefaultGoodsKeyResponse struct {
	ErrResponse
	// RawData of the default-goods key payload.
	RawData map[string]any `json:"-"`
}

// GetBroadcastDefaultGoodsKey returns the current default-goods key.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_getdefault_goodskey.html
func (w *MiniProgram) GetBroadcastDefaultGoodsKey(ctx context.Context) (*GetBroadcastDefaultGoodsKeyResponse, error) {
	var result GetBroadcastDefaultGoodsKeyResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/broadcast/goods/getkey", nil, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// Broadcast roles (直播成员管理).
// ============================================================

// AddBroadcastRoleRequest grants a role to a WeChat user.
type AddBroadcastRoleRequest struct {
	// Role of the user: 1 主播, 2 运营者, 3 超级管理员.
	Role int `json:"role"`
	// Username of the WeChat user (微信号).
	Username string `json:"username"`
}

// AddBroadcastRole grants a broadcast role to a WeChat user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/role-management/api_getrolelistdw.html
func (w *MiniProgram) AddBroadcastRole(ctx context.Context, role int, username string) error {
	req := &AddBroadcastRoleRequest{Role: role, Username: username}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/role/addrole", nil, defaultReqOptions(), req, nil)
}

// DeleteBroadcastRole removes a broadcast role from a user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/role-management/api_deleterole.html
func (w *MiniProgram) DeleteBroadcastRole(ctx context.Context, role int, username string) error {
	req := &AddBroadcastRoleRequest{Role: role, Username: username}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/role/deleterole", nil, defaultReqOptions(), req, nil)
}

// BroadcastRoleMember is one role member of the account.
type BroadcastRoleMember struct {
	// OpenID of the member.
	OpenID string `json:"openid,omitempty"`
	// Username of the member.
	Username string `json:"username,omitempty"`
	// Nickname of the member.
	Nickname string `json:"nickname,omitempty"`
	// Headingimg of the member avatar.
	Headingimg string `json:"headingimg,omitempty"`
	// RoleList granted to the member.
	RoleList []int `json:"roleList,omitempty"`
	// UpdateTimestamp of the last role change.
	UpdateTimestamp int64 `json:"updateTimestamp,omitempty"`
}

// GetBroadcastRoleListResponse is returned by GetBroadcastRoleList.
type GetBroadcastRoleListResponse struct {
	ErrResponse
	// List of the role members.
	List []BroadcastRoleMember `json:"list,omitempty"`
	// Total number of members.
	Total int `json:"total,omitempty"`
}

// GetBroadcastRoleList lists the role members of the broadcast capability.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/role-management/api_getrolelist.html
func (w *MiniProgram) GetBroadcastRoleList(ctx context.Context, role, offset, limit int, keyword string) (*GetBroadcastRoleListResponse, error) {
	query := url.Values{}
	if role != 0 {
		query.Set("role", fmtInt(role))
	}
	query.Set("offset", fmtInt(offset))
	query.Set("limit", fmtInt(limit))
	if keyword != "" {
		query.Set("keyword", keyword)
	}
	var result GetBroadcastRoleListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/broadcast/role/getrolelist", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// Live room sub-anchor and assistant management.
// ============================================================

// AddRoomSubAnchorRequest adds a sub anchor (副主播) to a room.
type AddRoomSubAnchorRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// Username of the WeChat user to add as sub anchor.
	Username string `json:"username"`
}

// AddRoomSubAnchor adds a deputy anchor to a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_addsubanchor.html
func (w *MiniProgram) AddRoomSubAnchor(ctx context.Context, roomID int64, username string) error {
	req := &AddRoomSubAnchorRequest{RoomID: roomID, Username: username}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/addsubanchor", nil, defaultReqOptions(), req, nil)
}

// DeleteRoomSubAnchor removes the sub anchor of a room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_deletesubanchor.html
func (w *MiniProgram) DeleteRoomSubAnchor(ctx context.Context, roomID int64) error {
	req := map[string]int64{"roomId": roomID}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/deletesubanchor", nil, defaultReqOptions(), req, nil)
}

// ModifyRoomSubAnchorRequest replaces the sub anchor of a room.
type ModifyRoomSubAnchorRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// Username of the new sub anchor.
	Username string `json:"username"`
}

// ModifyRoomSubAnchor changes the deputy anchor of a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_modifysubanchor.html
func (w *MiniProgram) ModifyRoomSubAnchor(ctx context.Context, roomID int64, username string) error {
	req := &ModifyRoomSubAnchorRequest{RoomID: roomID, Username: username}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/modifysubanchor", nil, defaultReqOptions(), req, nil)
}

// GetRoomSubAnchorResponse is returned by GetRoomSubAnchor.
type GetRoomSubAnchorResponse struct {
	ErrResponse
	// Username of the current sub anchor.
	Username string `json:"username,omitempty"`
}

// GetRoomSubAnchor returns the deputy anchor of a room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_getsubanchor.html
func (w *MiniProgram) GetRoomSubAnchor(ctx context.Context, roomID int64) (*GetRoomSubAnchorResponse, error) {
	query := url.Values{}
	query.Set("roomId", fmtInt64(roomID))
	var result GetRoomSubAnchorResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/broadcast/room/getsubanchor", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BroadcastAssistant is one room assistant.
type BroadcastAssistant struct {
	// Username of the assistant's WeChat id.
	Username string `json:"username"`
	// Nickname of the assistant shown in the room.
	Nickname string `json:"nickname"`
}

// AddRoomAssistantRequest adds assistants to a live room.
type AddRoomAssistantRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// Users to add as assistants.
	Users []BroadcastAssistant `json:"users"`
}

// AddRoomAssistant adds assistant accounts (场控) to a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_addveassistant.html
func (w *MiniProgram) AddRoomAssistant(ctx context.Context, req *AddRoomAssistantRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/addassistant", nil, defaultReqOptions(), req, nil)
}

// BroadcastAssistantInfo is one assistant of the room list.
type BroadcastAssistantInfo struct {
	// Alias of the assistant.
	Alias string `json:"alias,omitempty"`
	// Nickname of the assistant.
	Nickname string `json:"nickname,omitempty"`
	// OpenID of the assistant.
	OpenID string `json:"openid,omitempty"`
	// Headimg of the assistant.
	Headimg string `json:"headimg,omitempty"`
	// Timestamp of the last update.
	Timestamp int64 `json:"timestamp,omitempty"`
}

// GetRoomAssistantListResponse is returned by GetRoomAssistantList.
type GetRoomAssistantListResponse struct {
	ErrResponse
	// List of the room assistants.
	List []BroadcastAssistantInfo `json:"list,omitempty"`
	// Count of the assistants.
	Count int `json:"count,omitempty"`
	// MaxCount of the assistants allowed in a room.
	MaxCount int `json:"maxCount,omitempty"`
}

// GetRoomAssistantList returns the assistant list of a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_getassistantlist.html
func (w *MiniProgram) GetRoomAssistantList(ctx context.Context, roomID int64) (*GetRoomAssistantListResponse, error) {
	query := url.Values{}
	query.Set("roomId", fmtInt64(roomID))
	var result GetRoomAssistantListResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/broadcast/room/getassistantlist", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ModifyRoomAssistantRequest renames an assistant of a room.
type ModifyRoomAssistantRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// Username of the assistant to modify.
	Username string `json:"username"`
	// Nickname to assign.
	Nickname string `json:"nickname"`
}

// ModifyRoomAssistant changes the nickname of a room assistant.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_modifyassistant.html
func (w *MiniProgram) ModifyRoomAssistant(ctx context.Context, req *ModifyRoomAssistantRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/modifyassistant", nil, defaultReqOptions(), req, nil)
}

// RemoveRoomAssistantRequest removes an assistant from a room.
type RemoveRoomAssistantRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// Username of the assistant to remove.
	Username string `json:"username"`
}

// RemoveRoomAssistant removes an assistant (场控) from a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_removeassistant.html
func (w *MiniProgram) RemoveRoomAssistant(ctx context.Context, roomID int64, username string) error {
	req := &RemoveRoomAssistantRequest{RoomID: roomID, Username: username}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/removeassistant", nil, defaultReqOptions(), req, nil)
}

// GetRoomPushURLResponse is returned by GetRoomPushURL.
type GetRoomPushURLResponse struct {
	ErrResponse
	// PushURL of the room for the anchor software.
	PushURL string `json:"pushAddr,omitempty"`
}

// GetRoomPushURL returns the RTMP push URL of a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_getpushurl.html
func (w *MiniProgram) GetRoomPushURL(ctx context.Context, roomID int64) (*GetRoomPushURLResponse, error) {
	query := url.Values{}
	query.Set("roomId", fmtInt64(roomID))
	var result GetRoomPushURLResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/broadcast/room/getpushurl", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRoomSharedCodeResponse is returned by GetRoomSharedCode.
type GetRoomSharedCodeResponse struct {
	ErrResponse
	// RawData of the share code payload.
	RawData map[string]any `json:"-"`
}

// GetRoomSharedCode returns the share code / QR content of a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_getsharedcode.html
func (w *MiniProgram) GetRoomSharedCode(ctx context.Context, roomID int64, params string) (*GetRoomSharedCodeResponse, error) {
	query := url.Values{}
	query.Set("roomId", fmtInt64(roomID))
	if params != "" {
		query.Set("params", params)
	}
	var result GetRoomSharedCodeResponse
	if err := w.withAccessToken(ctx, http.MethodGet, "/wxaapi/broadcast/room/getsharedcode", query, defaultReqOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RoomUpdateRequest is the shared payload of room state update methods.
type RoomUpdateRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
}

// UpdateRoomCommentRequest controls whether the room comment area is on.
type UpdateRoomCommentRequest struct {
	RoomUpdateRequest
	// Comment enabled flag.
	Comment bool `json:"comment"`
}

// UpdateRoomComment enables or disables the comment area of a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_updatecomment.html
func (w *MiniProgram) UpdateRoomComment(ctx context.Context, roomID int64, enabled bool) error {
	req := &UpdateRoomCommentRequest{RoomUpdateRequest: RoomUpdateRequest{RoomID: roomID}, Comment: enabled}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/updatecomment", nil, defaultReqOptions(), req, nil)
}

// UpdateRoomFeedPublicRequest controls whether the room is shown in the feeds.
type UpdateRoomFeedPublicRequest struct {
	RoomUpdateRequest
	// FeedPublic enables the feeds presence.
	FeedPublic int `json:"feedPublic"`
}

// UpdateRoomFeedPublic toggles the feeds (信息流) visibility of a room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_updatefeedpublic.html
func (w *MiniProgram) UpdateRoomFeedPublic(ctx context.Context, roomID int64, feedPublic int) error {
	req := &UpdateRoomFeedPublicRequest{RoomUpdateRequest: RoomUpdateRequest{RoomID: roomID}, FeedPublic: feedPublic}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/updatefeedpublic", nil, defaultReqOptions(), req, nil)
}

// UpdateRoomKfRequest controls the customer-service button of a room.
type UpdateRoomKfRequest struct {
	RoomUpdateRequest
	// CloseKf closes the customer-service entry when 1.
	CloseKf int `json:"closeKf"`
}

// UpdateRoomKf enables or disables the customer-service button of a live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_updatekf.html
func (w *MiniProgram) UpdateRoomKf(ctx context.Context, roomID int64, closeKf int) error {
	req := &UpdateRoomKfRequest{RoomUpdateRequest: RoomUpdateRequest{RoomID: roomID}, CloseKf: closeKf}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/updatekf", nil, defaultReqOptions(), req, nil)
}

// UpdateRoomReplayRequest controls the replay availability of a room.
type UpdateRoomReplayRequest struct {
	RoomUpdateRequest
	// CloseReplay closes the replay when 1.
	CloseReplay int `json:"closeReplay"`
}

// UpdateRoomReplay enables or disables the replay of a finished live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_updatereplay.html
func (w *MiniProgram) UpdateRoomReplay(ctx context.Context, roomID int64, closeReplay int) error {
	req := &UpdateRoomReplayRequest{RoomUpdateRequest: RoomUpdateRequest{RoomID: roomID}, CloseReplay: closeReplay}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/updatereplay", nil, defaultReqOptions(), req, nil)
}

// DeleteLiveRoom deletes a scheduled (not yet started) live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_deleteroom.html
func (w *MiniProgram) DeleteLiveRoom(ctx context.Context, roomID int64) error {
	body := map[string]int64{"roomId": roomID}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/deleteroom", nil, defaultReqOptions(), body, nil)
}

// ============================================================
// Live subscribe management (订阅消息).
// ============================================================

// LiveFollower is one long-term subscriber of the live reminders.
type LiveFollower struct {
	// OpenID of the subscriber.
	OpenID string `json:"openid,omitempty"`
	// SubscribeTime of the subscription.
	SubscribeTime int64 `json:"subscribe_time,omitempty"`
	// RoomID the user subscribed in.
	RoomID int64 `json:"room_id,omitempty"`
	// RoomStatus of the room at subscription time.
	RoomStatus int `json:"room_status,omitempty"`
}

// GetLiveFollowersRequest pages the live subscribers.
type GetLiveFollowersRequest struct {
	// Limit of the returned subscribers (default 200, max 2000).
	Limit int `json:"limit,omitempty"`
	// PageBreak cursor from the previous page response.
	PageBreak int64 `json:"page_break,omitempty"`
}

// GetLiveFollowersResponse is returned by GetLiveFollowers.
type GetLiveFollowersResponse struct {
	ErrResponse
	// Followers of the live capability.
	Followers []LiveFollower `json:"followers,omitempty"`
	// PageBreak cursor for the next page.
	PageBreak int64 `json:"page_break,omitempty"`
}

// GetLiveFollowers returns the followers (长期订阅用户) who can receive live
// reminders.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/subscribe-management/api_getfollowers.html
func (w *MiniProgram) GetLiveFollowers(ctx context.Context, req *GetLiveFollowersRequest) (*GetLiveFollowersResponse, error) {
	var result GetLiveFollowersResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/get_wxa_followers", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PushLiveMessageRequest pushes the live-start broadcast message to a set of
// subscribed users.
type PushLiveMessageRequest struct {
	// RoomID of the live room that is starting.
	RoomID int64 `json:"room_id"`
	// UserOpenIDs that receive the broadcast.
	UserOpenIDs []string `json:"user_openid"`
}

// PushLiveMessageResponse is returned by PushLiveMessage.
type PushLiveMessageResponse struct {
	ErrResponse
	// MessageID of the mass message, for matching the delivery result
	// callback.
	MessageID string `json:"message_id,omitempty"`
}

// PushLiveMessage sends the live-start reminder message to subscribed users.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/subscribe-management/api_pushmessage.html
func (w *MiniProgram) PushLiveMessage(ctx context.Context, req *PushLiveMessageRequest) (*PushLiveMessageResponse, error) {
	var result PushLiveMessageResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/push_message", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
