package miniprogram

import (
	"context"
)

// LiveRoom is the shared shape of the live broadcast room (both in-room and
// replay results). Only the fields relevant to management are modelled;
// WeChat adds more read-only fields over time.
type LiveRoom struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomid,omitempty"`
	// Name of the live room.
	Name string `json:"name,omitempty"`
	// CoverImg is the URL of the cover image.
	CoverImg string `json:"cover_img,omitempty"`
	// ShareImg is the share image URL.
	ShareImg string `json:"share_img,omitempty"`
	// LiveStatus of the room: 101 = live, 102 = ended, 103 = offline,
	// 104 = scheduled, 106 = banned.
	LiveStatus int `json:"live_status,omitempty"`
	// StartTime of the live session.
	StartTime int64 `json:"start_time,omitempty"`
	// EndTime of the live session.
	EndTime int64 `json:"end_time,omitempty"`
	// AnchorName of the room anchor.
	AnchorName string `json:"anchor_name,omitempty"`
	// AnchorWechat of the anchor's bound WeChat id.
	AnchorWechat string `json:"anchor_wechat,omitempty"`
	// CreatorOpenID of the room creator.
	CreatorOpenID string `json:"creator_openid,omitempty"`
	// Goods of the goods added to the room.
	Goods []LiveRoomGoods `json:"goods,omitempty"`
}

// LiveRoomGoods is one goods entry attached to a live room. Only the fields
// relevant to management are modelled; WeChat may add more read-only fields.
type LiveRoomGoods struct {
	// Name of the goods.
	Name string `json:"name"`
	// CoverImg is the URL of the goods cover image.
	CoverImg string `json:"cover_img"`
	// URL of the goods detail page.
	URL string `json:"url"`
	// Price of the goods in cents.
	Price int `json:"price"`
	// PriceType of the goods: 0 = not for sale, 1 = price above the line.
	PriceType int `json:"price_type"`
	// GoodsID of the goods.
	GoodsID int64 `json:"goods_id"`
}

// LiveRoomPushItem is one live room inside a paginated list.
type LiveRoomPushItem struct {
	LiveRoom
	// ReplayMediaURL of the last replay media (if any).
	ReplayMediaURL string `json:"replay_media_url,omitempty"`
	// FeedsImg of the feeds entry image.
	FeedsImg string `json:"feeds_img,omitempty"`
}

// GetLiveRoomListResponse is returned by GetLiveRoomList.
type GetLiveRoomListResponse struct {
	ErrResponse
	// ErrMsg of the upstream call.
	ErrMsg string `json:"errmsg,omitempty"`
	// Total of the matching rooms.
	Total int `json:"total,omitempty"`
	// RoomIDs of the page.
	RoomIDs []int64 `json:"room_ids,omitempty"`
	// RoomInfos of the page.
	RoomInfos []LiveRoomPushItem `json:"room_info,omitempty"`
}

// GetLiveRoomList lists the live broadcast rooms of the account, paged by
// start index / count.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_getliveinfo.html
func (w *MiniProgram) GetLiveRoomList(ctx context.Context, start, count int) (*GetLiveRoomListResponse, error) {
	body := map[string]int{"start": start, "limit": count}
	var result GetLiveRoomListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/getliveinfo", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLiveReplayRequest selects the replays to fetch.
type GetLiveReplayRequest struct {
	// RoomID of the room.
	RoomID int64 `json:"room_id"`
	// Start index of the page.
	Start int `json:"start"`
	// Count of the page (max 100).
	Count int `json:"limit"`
}

// GetLiveReplayResponse is returned by GetLiveReplay.
type GetLiveReplayResponse struct {
	ErrResponse
	// Total of the replays.
	Total int `json:"total,omitempty"`
	// LiveReplay of the page.
	LiveReplay []LiveReplayItem `json:"live_replay,omitempty"`
}

// LiveReplayItem is one replay of a live room.
type LiveReplayItem struct {
	// MediaURL of the replay video.
	MediaURL string `json:"media_url,omitempty"`
	// ExpireTime of the replay URL (Unix seconds).
	ExpireTime int64 `json:"expire_time,omitempty"`
	// CreateTime of the replay (Unix seconds).
	CreateTime int64 `json:"create_time,omitempty"`
}

// GetLiveReplay returns the replay videos of a finished live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_updatereplay.html
func (w *MiniProgram) GetLiveReplay(ctx context.Context, req *GetLiveReplayRequest) (*GetLiveReplayResponse, error) {
	var result GetLiveReplayResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/business/getliveinfo", nil, defaultReqOptions(), map[string]any{
		"action":  "get_replay",
		"room_id": req.RoomID,
		"start":   req.Start,
		"limit":   req.Count,
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateLiveRoomRequest creates a scheduled live room.
type CreateLiveRoomRequest struct {
	// Name of the room (3-17 Chinese chars).
	Name string `json:"name"`
	// CoverImg is the cover image MediaID (valid 3 days; recommend 1080x1920).
	CoverImg string `json:"coverImg"`
	// StartTime of the live (Unix seconds; between 10 minutes and 6 months
	// from now).
	StartTime int64 `json:"startTime"`
	// EndTime of the live (30 minutes to 24 hours after start).
	EndTime int64 `json:"endTime"`
	// AnchorName of the anchor (2-15 Chinese chars).
	AnchorName string `json:"anchorName"`
	// AnchorWechat of the anchor's WeChat id (must be real-name verified).
	AnchorWechat string `json:"anchorWechat"`
	// SubAnchorWechat of a deputy anchor (optional).
	SubAnchorWechat string `json:"subAnchorWechat,omitempty"`
	// CreaterWechat of the creator; when unset the room is visible to all
	// members, when set only to the creator/admins/主播.
	CreaterWechat string `json:"createrWechat,omitempty"`
	// ShareImg is the share image MediaID (recommend 800x640).
	ShareImg string `json:"shareImg,omitempty"`
	// FeedsImg is the feeds channel cover MediaID (recommend 800x800).
	FeedsImg string `json:"feedsImg,omitempty"`
	// IsFeedsPublic enables official feeds listing (1 on, 0 off; default on).
	IsFeedsPublic int `json:"isFeedsPublic,omitempty"`
	// Type of the live: 1 = push streaming (推流), 0 = phone live.
	Type int `json:"type"`
	// CloseLike closes the like button when 1.
	CloseLike int `json:"closeLike"`
	// CloseGoods closes the goods shelf when 1.
	CloseGoods int `json:"closeGoods"`
	// CloseComment closes the comment area when 1.
	CloseComment int `json:"closeComment"`
	// CloseKf closes the customer-service button when 1.
	CloseKf int `json:"closeKf,omitempty"`
	// CloseReplay closes the replay when 1 (default closed).
	CloseReplay int `json:"closeReplay,omitempty"`
	// CloseShare closes sharing when 1 (default open).
	CloseShare int `json:"closeShare,omitempty"`
}

// CreateLiveRoomResponse is returned by CreateLiveRoom.
type CreateLiveRoomResponse struct {
	ErrResponse
	// RoomID of the created room.
	RoomID int64 `json:"roomId,omitempty"`
}

// CreateLiveRoom schedules a new live broadcast room. The anchor and deputy
// anchors must accept the invitation in the WeChat client.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_createroom.html
func (w *MiniProgram) CreateLiveRoom(ctx context.Context, req *CreateLiveRoomRequest) (*CreateLiveRoomResponse, error) {
	var result CreateLiveRoomResponse
	if err := w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/create", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddLiveGoodsRequest adds a shopping good to the room's goods shelf.
type AddLiveGoodsRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// GoodsIDs of the goods to add.
	GoodsIDs []int64 `json:"ids"`
}

// AddLiveGoods adds goods to a live room's shelf.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_pushgoods.html
func (w *MiniProgram) AddLiveGoods(ctx context.Context, roomID int64, goodsIDs []int64) error {
	body := &AddLiveGoodsRequest{RoomID: roomID, GoodsIDs: goodsIDs}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/addgoods", nil, defaultReqOptions(), body, nil)
}

// DeleteLiveGoods removes one good from a live room's shelf.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_deletedoods.html
func (w *MiniProgram) DeleteLiveGoods(ctx context.Context, roomID, goodsID int64) error {
	body := map[string]int64{"roomId": roomID, "goodsId": goodsID}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/deleteInRoom", nil, defaultReqOptions(), body, nil)
}

// RoomGoodsOrder is one good entry of the room goods sort payload.
type RoomGoodsOrder struct {
	// GoodsID of the good.
	GoodsID int64 `json:"goodsId"`
}

// SortLiveGoodsRequest sorts the goods on a room shelf.
type SortLiveGoodsRequest struct {
	// RoomID of the live room.
	RoomID int64 `json:"roomId"`
	// Goods in the desired display order.
	Goods []RoomGoodsOrder `json:"goods"`
}

// SortLiveGoods reorders the goods of a live room's shelf.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_sortgoods.html
func (w *MiniProgram) SortLiveGoods(ctx context.Context, req *SortLiveGoodsRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/sort", nil, defaultReqOptions(), req, nil)
}

// EditLiveRoomRequest updates a scheduled live room before it starts.
type EditLiveRoomRequest struct {
	// ID of the live room to edit.
	ID int64 `json:"id"`
	// Name of the room (3-17 Chinese chars).
	Name string `json:"name"`
	// CoverImg is the cover image MediaID.
	CoverImg string `json:"coverImg"`
	// StartTime of the live (Unix seconds).
	StartTime int64 `json:"startTime"`
	// EndTime of the live (Unix seconds).
	EndTime int64 `json:"endTime"`
	// AnchorName of the anchor.
	AnchorName string `json:"anchorName"`
	// AnchorWechat of the anchor's WeChat id.
	AnchorWechat string `json:"anchorWechat"`
	// ShareImg is the share image MediaID.
	ShareImg string `json:"shareImg,omitempty"`
	// FeedsImg is the feeds cover MediaID.
	FeedsImg string `json:"feedsImg,omitempty"`
	// IsFeedsPublic: 1 feeds listing on, 0 off.
	IsFeedsPublic int `json:"isFeedsPublic,omitempty"`
	// CloseLike closes likes when 1.
	CloseLike int `json:"closeLike"`
	// CloseGoods closes the goods shelf when 1.
	CloseGoods int `json:"closeGoods"`
	// CloseComment closes comments when 1.
	CloseComment int `json:"closeComment"`
	// CloseKf closes customer service when 1.
	CloseKf int `json:"closeKf,omitempty"`
	// CloseReplay closes replay when 1.
	CloseReplay int `json:"closeReplay,omitempty"`
	// CloseShare closes sharing when 1.
	CloseShare int `json:"closeShare,omitempty"`
}

// EditLiveRoom updates the configuration of a scheduled live room.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_editroom.html
func (w *MiniProgram) EditLiveRoom(ctx context.Context, req *EditLiveRoomRequest) error {
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/editroom", nil, defaultReqOptions(), req, nil)
}

// ImportLiveRoomGoods imports already-approved goods into a live room shelf.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_importgoods.html
func (w *MiniProgram) ImportLiveRoomGoods(ctx context.Context, roomID int64, goodsIDs []int64) error {
	body := map[string]any{"roomId": roomID, "ids": goodsIDs}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/room/addgoods", nil, defaultReqOptions(), body, nil)
}

// PushLiveRoomGoods adds one approved good to the live room shelf.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/livebroadcast/studio-management/api_pushgoods.html
func (w *MiniProgram) PushLiveRoomGoods(ctx context.Context, roomID, goodsID int64) error {
	body := map[string]int64{"roomId": roomID, "goodsId": goodsID}
	return w.withAccessTokenPost(ctx, "/wxaapi/broadcast/goods/push", nil, defaultReqOptions(), body, nil)
}
