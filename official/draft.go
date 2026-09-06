package official

import (
	"context"

	"github.com/go-sphere/weixin-api/core"
)

// ============================================================
// 草稿箱 / 发布 (draft box and publishing).
// Drafts are the modern way to create 图文 (articles) which are then
// published through the 发布接口 to go live.
// ============================================================

// DraftArticle is one article inside a draft.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Add_draft.html
type DraftArticle struct {
	// Title of the article.
	Title string `json:"title"`
	// Author of the article.
	Author string `json:"author,omitempty"`
	// Digest shown in the list.
	Digest string `json:"digest,omitempty"`
	// Content HTML of the article.
	Content string `json:"content"`
	// ContentSourceURL of the original article.
	ContentSourceURL string `json:"content_source_url,omitempty"`
	// ThumbMediaID of the cover (permanent material).
	ThumbMediaID string `json:"thumb_media_id"`
	// NeedOpenComment enables comments.
	NeedOpenComment int `json:"need_open_comment,omitempty"`
	// OnlyFansCanComment restricts comments to followers.
	OnlyFansCanComment int `json:"only_fans_can_comment,omitempty"`
	// PicCrop2351 uses 2.35:1 cropping when true.
	PicCrop2351 int `json:"pic_crop_235_1,omitempty"`
	// PicCrop11 uses 1:1 cropping when true.
	PicCrop11 int `json:"pic_crop_1_1,omitempty"`
}

// AddDraftRequest adds a draft with 1-8 articles.
type AddDraftRequest struct {
	// Articles of the draft.
	Articles []DraftArticle `json:"articles"`
}

// AddDraftResponse is returned by AddDraft.
type AddDraftResponse struct {
	ErrResponse
	// MediaID of the created draft.
	MediaID string `json:"media_id"`
}

// AddDraft creates a draft box entry.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Add_draft.html
func (oa *OfficialAccount) AddDraft(ctx context.Context, req *AddDraftRequest) (*AddDraftResponse, error) {
	var result AddDraftResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/draft/add", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDraftRequest selects a draft.
type GetDraftRequest struct {
	// MediaID of the draft.
	MediaID string `json:"media_id"`
}

// GetDraftResponse is returned by GetDraft.
type GetDraftResponse struct {
	ErrResponse
	// NewsItem of the draft articles.
	NewsItem []DraftArticle `json:"news_item"`
}

// GetDraft returns the articles of a draft.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Get_draft.html
func (oa *OfficialAccount) GetDraft(ctx context.Context, mediaID string) (*GetDraftResponse, error) {
	var result GetDraftResponse
	req := &GetDraftRequest{MediaID: mediaID}
	if err := oa.withTokenPost(ctx, "/cgi-bin/draft/get", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDraft removes a draft.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Delete_draft.html
func (oa *OfficialAccount) DeleteDraft(ctx context.Context, mediaID string) error {
	req := &GetDraftRequest{MediaID: mediaID}
	return oa.withTokenPost(ctx, "/cgi-bin/draft/delete", nil, core.DefaultRequestOptions(), req, nil)
}

// ListDraftRequest pages the drafts.
type ListDraftRequest struct {
	// Offset of the page.
	Offset int `json:"offset"`
	// Count of the page (1-20).
	Count int `json:"count"`
	// NoContent skips the article content when true.
	NoContent int `json:"no_content,omitempty"`
}

// ListDraftResponse is returned by ListDraft.
type ListDraftResponse struct {
	ErrResponse
	// TotalCount of the drafts.
	TotalCount int `json:"total_count"`
	// ItemCount of the returned page.
	ItemCount int `json:"item_count"`
	// Item of the returned drafts.
	Item []struct {
		// MediaID of the draft.
		MediaID string `json:"media_id"`
		// Content of the draft (list of news items).
		Content struct {
			// NewsItem of the draft.
			NewsItem []DraftArticle `json:"news_item"`
		} `json:"content"`
		// UpdateTime of the draft (unix seconds).
		UpdateTime int64 `json:"update_time"`
	} `json:"item"`
}

// ListDraft pages the drafts of the account.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Get_draft_list.html
func (oa *OfficialAccount) ListDraft(ctx context.Context, req *ListDraftRequest) (*ListDraftResponse, error) {
	var result ListDraftResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/draft/batchget", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PublishDraftRequest publishes a draft.
type PublishDraftRequest struct {
	// MediaID of the draft to publish.
	MediaID string `json:"media_id"`
}

// PublishDraftResponse is returned by PublishDraft.
type PublishDraftResponse struct {
	ErrResponse
	// PublishID of the publish task.
	PublishID int64 `json:"publish_id"`
}

// PublishDraft publishes a draft (发布). The result is async; poll with
// GetPublishStatus.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Publish/Publish.html
func (oa *OfficialAccount) PublishDraft(ctx context.Context, mediaID string) (*PublishDraftResponse, error) {
	var result PublishDraftResponse
	req := &PublishDraftRequest{MediaID: mediaID}
	if err := oa.withTokenPost(ctx, "/cgi-bin/freepublish/submit", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PublishStatusResponse is returned by GetPublishStatus.
type PublishStatusResponse struct {
	ErrResponse
	// PublishID of the publish task.
	PublishID int64 `json:"publish_id"`
	// PublishStatus of the task: 0 success, 1 publishing, 2 original failed,
	// 3 general failed, 4 no auth, 5 params error.
	PublishStatus int `json:"publish_status"`
	// ArticleDetail of the published article.
	ArticleDetail struct {
		// Count of the published articles.
		Count int `json:"count"`
		// Item of the published articles.
		Item []struct {
			// URL of the published article.
			URL string `json:"url"`
		} `json:"item"`
	} `json:"article_detail"`
	// FailIdx of the failed article.
	FailIdx []int `json:"fail_idx"`
}

// GetPublishStatus polls the state of a publish task.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Publish/Get_the_status_of_publish.html
func (oa *OfficialAccount) GetPublishStatus(ctx context.Context, publishID int64) (*PublishStatusResponse, error) {
	var result PublishStatusResponse
	body := map[string]int64{"publish_id": publishID}
	if err := oa.withTokenPost(ctx, "/cgi-bin/freepublish/get", nil, core.DefaultRequestOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeletePublishRequest deletes a published article.
type DeletePublishRequest struct {
	// ArticleID of the published article.
	ArticleID string `json:"article_id"`
	// Index of the article in a multi-article publish.
	Index int `json:"index,omitempty"`
}

// DeletePublish deletes a published article.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Publish/Delete_publish.html
func (oa *OfficialAccount) DeletePublish(ctx context.Context, req *DeletePublishRequest) error {
	return oa.withTokenPost(ctx, "/cgi-bin/freepublish/delete", nil, core.DefaultRequestOptions(), req, nil)
}

// DraftArticleUpdate carries the fields to change on one article of an
// existing draft. Every field is optional (omitempty): the draft/update API
// applies a partial update, so omitting a field leaves the stored value
// untouched. Use DraftArticle (full, required fields) when adding a draft.
type DraftArticleUpdate struct {
	// Title of the article.
	Title string `json:"title,omitempty"`
	// Author of the article.
	Author string `json:"author,omitempty"`
	// Digest shown in the list.
	Digest string `json:"digest,omitempty"`
	// Content HTML of the article.
	Content string `json:"content,omitempty"`
	// ContentSourceURL of the original article.
	ContentSourceURL string `json:"content_source_url,omitempty"`
	// ThumbMediaID of the cover (permanent material).
	ThumbMediaID string `json:"thumb_media_id,omitempty"`
	// NeedOpenComment enables comments.
	NeedOpenComment int `json:"need_open_comment,omitempty"`
	// OnlyFansCanComment restricts comments to followers.
	OnlyFansCanComment int `json:"only_fans_can_comment,omitempty"`
	// PicCrop2351 uses 2.35:1 cropping when true.
	PicCrop2351 int `json:"pic_crop_235_1,omitempty"`
	// PicCrop11 uses 1:1 cropping when true.
	PicCrop11 int `json:"pic_crop_1_1,omitempty"`
}

// UpdateDraftArticleRequest updates one article inside a draft.
type UpdateDraftArticleRequest struct {
	// MediaID of the draft.
	MediaID string `json:"media_id"`
	// Index of the article inside the draft.
	Index int `json:"index"`
	// Articles carries only the fields to change.
	Articles DraftArticleUpdate `json:"articles"`
}

// UpdateDraftArticle modifies an article of an existing draft.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Update_draft.html
func (oa *OfficialAccount) UpdateDraftArticle(ctx context.Context, req *UpdateDraftArticleRequest) error {
	return oa.withTokenPost(ctx, "/cgi-bin/draft/update", nil, core.DefaultRequestOptions(), req, nil)
}
