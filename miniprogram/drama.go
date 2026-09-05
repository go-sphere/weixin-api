package miniprogram

import (
	"context"
)

// ============================================================
// 短剧 (wxadrama) — WeChat short-drama developer capabilities. All payloads
// identify dramas by {src_appid, drama_id, drama_name}.
// ============================================================

// DramaRef identifies one drama in a request list.
type DramaRef struct {
	// SrcAppID of the Mini Program that submitted the drama.
	SrcAppID string `json:"src_appid"`
	// DramaID of the drama.
	DramaID string `json:"drama_id"`
	// DramaName of the drama.
	DramaName string `json:"drama_name"`
}

// SetDramaFlushRequest selects the dramas whose player preloads (预拉取) are
// enabled.
type SetDramaFlushRequest struct {
	// List of the dramas to enable (max 1000; empty clears the config).
	List []DramaRef `json:"list"`
}

// SetDramaFlush enables the pre-fetch (刷剧) capability for a set of dramas;
// each call overwrites the previous configuration.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_developersetflushdrama.html
func (w *MiniProgram) SetDramaFlush(ctx context.Context, list []DramaRef) error {
	req := &SetDramaFlushRequest{List: list}
	return w.withAccessTokenPost(ctx, "/wxadrama/developersetflushdrama", nil, defaultReqOptions(), req, nil)
}

// PublishDramaRequest publishes dramas to the production store.
type PublishDramaRequest struct {
	// List of the dramas to publish.
	List []DramaRef `json:"list"`
}

// PublishDrama publishes developer dramas (发布上线).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_developerpublishdrama.html
func (w *MiniProgram) PublishDrama(ctx context.Context, list []DramaRef) error {
	req := &PublishDramaRequest{List: list}
	return w.withAccessTokenPost(ctx, "/wxadrama/developerpublishdrama", nil, defaultReqOptions(), req, nil)
}

// GetPublishedDramaResponse is returned by GetPublishedDrama.
type GetPublishedDramaResponse struct {
	ErrResponse
	// RawData of the published dramas payload.
	RawData map[string]any `json:"-"`
}

// GetPublishedDrama returns the published dramas of the developer.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_developergetpublisheddrama.html
func (w *MiniProgram) GetPublishedDrama(ctx context.Context) (*GetPublishedDramaResponse, error) {
	var result GetPublishedDramaResponse
	if err := w.withAccessTokenPost(ctx, "/wxadrama/developergetpublisheddrama", nil, defaultReqOptions(), map[string]string{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetIaaDramaRequest selects the IAA (激励广告) dramas.
type SetIaaDramaRequest struct {
	// List of the IAA dramas.
	List []DramaRef `json:"list"`
}

// SetIaaDrama sets the dramas using the IAA monetisation model.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_developersetiaadrama.html
func (w *MiniProgram) SetIaaDrama(ctx context.Context, list []DramaRef) error {
	req := &SetIaaDramaRequest{List: list}
	return w.withAccessTokenPost(ctx, "/wxadrama/developersetiaadrama", nil, defaultReqOptions(), req, nil)
}

// GetIaaDramaRequest queries the IAA status of dramas.
type GetIaaDramaRequest struct {
	// List of the dramas to query.
	List []DramaRef `json:"list"`
}

// GetIaaDramaResponse is returned by GetIaaDrama.
type GetIaaDramaResponse struct {
	ErrResponse
	// RawData of the query payload.
	RawData map[string]any `json:"-"`
}

// GetIaaDrama queries the IAA monetisation status of dramas.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_developergetiaadrama.html
func (w *MiniProgram) GetIaaDrama(ctx context.Context, list []DramaRef) (*GetIaaDramaResponse, error) {
	req := &GetIaaDramaRequest{List: list}
	var result GetIaaDramaResponse
	if err := w.withAccessTokenPost(ctx, "/wxadrama/developergetiaadrama", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchProcessDramaPromotionRequest joins/queries/leaves the drama promotion
// plans in batch.
type BatchProcessDramaPromotionRequest struct {
	// ActionType: 1 join a plan, 2 query plans, 3 leave a plan.
	ActionType int `json:"action_type"`
	// List of the dramas to process.
	List []DramaRef `json:"list"`
}

// BatchProcessDramaPromotion manages the drama promotion (推广) plan
// membership.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_batchprocessdramapromotion.html
func (w *MiniProgram) BatchProcessDramaPromotion(ctx context.Context, req *BatchProcessDramaPromotionRequest) error {
	return w.withAccessTokenPost(ctx, "/wxadrama/batchprocessdramapromotion", nil, defaultReqOptions(), req, nil)
}

// GetFinderEventRequest queries the finder cooperation events of dramas.
type GetFinderEventRequest struct {
	// EventIDList filters the cooperation events (empty = all).
	EventIDList []string `json:"event_id_list,omitempty"`
}

// GetFinderEventResponse is returned by GetFinderEvent.
type GetFinderEventResponse struct {
	ErrResponse
	// RawData of the event list payload.
	RawData map[string]any `json:"-"`
}

// GetFinderEvent returns the short-drama cooperation promotion events.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_getfinderevent.html
func (w *MiniProgram) GetFinderEvent(ctx context.Context, eventIDs []string) (*GetFinderEventResponse, error) {
	var result GetFinderEventResponse
	req := &GetFinderEventRequest{EventIDList: eventIDs}
	if err := w.withAccessTokenPost(ctx, "/wxadrama/getfinderevent", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetPlayerDramaRecommendSwitchRequest toggles the player drama
// recommendation.
type SetPlayerDramaRecommendSwitchRequest struct {
	// EntryType of the recommendation entry (e.g. 2002).
	EntryType int `json:"entry_type"`
	// SwitchStatus of the recommendation.
	SwitchStatus bool `json:"switch_status"`
}

// SetPlayerDramaRecommendSwitch enables or disables the in-player drama
// recommendations.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_setplayerdramarecmdswitch.html
func (w *MiniProgram) SetPlayerDramaRecommendSwitch(ctx context.Context, entryType int, enabled bool) error {
	req := &SetPlayerDramaRecommendSwitchRequest{EntryType: entryType, SwitchStatus: enabled}
	return w.withAccessTokenPost(ctx, "/wxadrama/setplayerdramarecmdswitch", nil, defaultReqOptions(), req, nil)
}

// DramaRecommendItem is one recommended drama entry.
type DramaRecommendItem struct {
	// SrcAppID of the drama's submitting Mini Program.
	SrcAppID string `json:"src_appid"`
	// DramaID of the recommended drama.
	DramaID string `json:"drama_id"`
	// DramaName of the recommended drama.
	DramaName string `json:"drama_name"`
}

// SetRecommendedDramaRequest configures the recommended dramas of an entry.
type SetRecommendedDramaRequest struct {
	// EntryType of the recommendation slot: 1 剧结束, 2 选集最右侧推荐,
	// 3 剧集profile页相关推荐.
	EntryType int `json:"entry_type"`
	// SrcAppID of the source Mini Program.
	SrcAppID string `json:"src_appid,omitempty"`
	// DramaID of the source drama.
	DramaID string `json:"drama_id,omitempty"`
	// List of the recommended dramas.
	List []DramaRecommendItem `json:"list,omitempty"`
}

// SetRecommendedDrama configures which dramas are recommended after a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/dramaOthersAPI/api_developersetrecmddrama.html
func (w *MiniProgram) SetRecommendedDrama(ctx context.Context, req *SetRecommendedDramaRequest) error {
	return w.withAccessTokenPost(ctx, "/wxadrama/developersetrecmddrama", nil, defaultReqOptions(), req, nil)
}
