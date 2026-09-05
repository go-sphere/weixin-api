package miniprogram

import (
	"context"
)

// ============================================================
// 微信小说 (WeChat novel) — content management for Mini Programs running
// novel content. Routes live under /wxa/book/*.
// ============================================================

// CreateNovelBookRequest creates a novel (作品) on the WeChat novel platform.
type CreateNovelBookRequest struct {
	// Title of the novel (1-30 chars).
	Title string `json:"title"`
	// Intro of the novel (1-500 chars).
	Intro string `json:"intro"`
	// CoverMediaID uploaded via the temporary-media API.
	CoverMediaID string `json:"cover_media_id"`
	// Author of the novel (1-100 chars).
	Author string `json:"author"`
	// FirstCategoryID of the novel category.
	FirstCategoryID int64 `json:"first_category_id"`
	// SecondCategoryID of the novel category.
	SecondCategoryID int64 `json:"second_category_id"`
	// ThirdCategoryID of the novel category.
	ThirdCategoryID int64 `json:"third_category_id"`
	// CompleteStatus: 1 serialising, 2 completed.
	CompleteStatus int `json:"complete_status"`
	// OriginalID of the provider-side novel key (deduplication).
	OriginalID string `json:"original_id,omitempty"`
	// ChapterOrderMethod: 0 append, 1 ascending by seq (default 0).
	ChapterOrderMethod int `json:"chapter_order_method,omitempty"`
	// CustomInfo of the provider (up to 128 bytes).
	CustomInfo string `json:"custom_info,omitempty"`
	// KeywordList of up to 3 topic keywords (1-4 chars each).
	KeywordList []string `json:"keyword_list,omitempty"`
	// AwesomeParagraph excerpt of the novel (400-1000 chars).
	AwesomeParagraph string `json:"awesome_paragraph,omitempty"`
}

// CreateNovelBookResponse is returned by CreateNovelBook.
type CreateNovelBookResponse struct {
	ErrResponse
	// BookID of the created novel.
	BookID string `json:"book_id,omitempty"`
}

// CreateNovelBook registers a novel with the WeChat novel platform.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_createbook.html
func (w *MiniProgram) CreateNovelBook(ctx context.Context, req *CreateNovelBookRequest) (*CreateNovelBookResponse, error) {
	var result CreateNovelBookResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/createbook", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateNovelBookRequest updates an existing novel.
type UpdateNovelBookRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// Title of the novel (1-30 chars).
	Title string `json:"title,omitempty"`
	// Intro of the novel.
	Intro string `json:"intro,omitempty"`
	// CoverMediaID of the cover.
	CoverMediaID string `json:"cover_media_id,omitempty"`
	// Author of the novel.
	Author string `json:"author,omitempty"`
	// FirstCategoryID of the category.
	FirstCategoryID int64 `json:"first_category_id,omitempty"`
	// SecondCategoryID of the category.
	SecondCategoryID int64 `json:"second_category_id,omitempty"`
	// ThirdCategoryID of the category.
	ThirdCategoryID int64 `json:"third_category_id,omitempty"`
	// CompleteStatus: 1 serialising, 2 completed.
	CompleteStatus int `json:"complete_status,omitempty"`
	// ChapterIDList in the desired reading order.
	ChapterIDList []string `json:"chapter_id_list,omitempty"`
	// NeedVolume enables volumes.
	NeedVolume bool `json:"need_volume,omitempty"`
	// VolumeList of the volumes.
	VolumeList []map[string]any `json:"volume_list,omitempty"`
	// ChapterOrderMethod: 0 append, 1 ascending by seq.
	ChapterOrderMethod int `json:"chapter_order_method,omitempty"`
	// CustomInfo of the provider.
	CustomInfo string `json:"custom_info,omitempty"`
	// UpdateKeyword re-applies the keywords.
	UpdateKeyword bool `json:"update_keyword,omitempty"`
	// KeywordList of up to 3 topic keywords.
	KeywordList []string `json:"keyword_list,omitempty"`
	// AwesomeParagraph excerpt.
	AwesomeParagraph string `json:"awesome_paragraph,omitempty"`
}

// UpdateNovelBook updates a novel.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_updatebook.html
func (w *MiniProgram) UpdateNovelBook(ctx context.Context, req *UpdateNovelBookRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/book/updatebook", nil, defaultReqOptions(), req, nil)
}

// DeleteNovelBook deletes a novel.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_deletebook.html
func (w *MiniProgram) DeleteNovelBook(ctx context.Context, bookID string) error {
	body := map[string]string{"book_id": bookID}
	return w.withAccessTokenPost(ctx, "/wxa/book/deletebook", nil, defaultReqOptions(), body, nil)
}

// GetNovelBookRequest selects a novel to fetch.
type GetNovelBookRequest struct {
	// BookID of the novel (alternative to OriginalID).
	BookID string `json:"book_id,omitempty"`
	// NeedEditedData fetches the edited (not published) version when true.
	NeedEditedData bool `json:"need_edited_data,omitempty"`
	// OriginalID of the provider-side key (alternative to BookID).
	OriginalID string `json:"original_id,omitempty"`
}

// GetNovelBookResponse is returned by GetNovelBook.
type GetNovelBookResponse struct {
	ErrResponse
	// Book of the query.
	Book map[string]any `json:"book,omitempty"`
}

// GetNovelBook returns a novel by id or original id.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_getbook.html
func (w *MiniProgram) GetNovelBook(ctx context.Context, req *GetNovelBookRequest) (*GetNovelBookResponse, error) {
	var result GetNovelBookResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/getbook", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListNovelBookRequest pages the provider's novels.
type ListNovelBookRequest struct {
	// Limit of the page (1-100, default 100).
	Limit int `json:"limit,omitempty"`
	// Offset of the page (alternative to LastID).
	Offset int64 `json:"offset,omitempty"`
	// LastID cursor; first call 0, subsequent calls use the returned last_id.
	LastID int64 `json:"last_id,omitempty"`
	// NeedEditedData fetches edited (not published) data when true.
	NeedEditedData bool `json:"need_edited_data,omitempty"`
}

// ListNovelBookResponse is returned by ListNovelBook.
type ListNovelBookResponse struct {
	ErrResponse
	// BookList of the page.
	BookList []map[string]any `json:"book_list,omitempty"`
	// LastID cursor for the next page.
	LastID int64 `json:"last_id,omitempty"`
}

// ListNovelBook lists the novels of the provider, paged.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_listbook.html
func (w *MiniProgram) ListNovelBook(ctx context.Context, req *ListNovelBookRequest) (*ListNovelBookResponse, error) {
	var result ListNovelBookResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/listbook", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AuditNovelBookRequest submits a novel for audit.
type AuditNovelBookRequest struct {
	// BookID of the novel to audit.
	BookID string `json:"book_id"`
}

// AuditNovelBook submits a novel for the WeChat audit flow.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_auditbook.html
func (w *MiniProgram) AuditNovelBook(ctx context.Context, bookID string) error {
	body := &AuditNovelBookRequest{BookID: bookID}
	return w.withAccessTokenPost(ctx, "/wxa/book/auditbook", nil, defaultReqOptions(), body, nil)
}

// NovelChapter is one chapter being created.
type NovelChapter struct {
	// Title of the chapter (1-80 chars).
	Title string `json:"title,omitempty"`
	// Content of the chapter (1-20000 chars).
	Content string `json:"content,omitempty"`
	// Seq of the chapter for ordering.
	Seq int64 `json:"seq,omitempty"`
	// VolumeIndex of the containing volume.
	VolumeIndex int `json:"volume_index,omitempty"`
}

// CreateNovelChapterRequest adds a chapter to a novel.
type CreateNovelChapterRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// Chapter of the new chapter.
	Chapter NovelChapter `json:"chapter"`
}

// CreateNovelChapterResponse is returned by CreateNovelChapter.
type CreateNovelChapterResponse struct {
	ErrResponse
	// ChapterID of the created chapter.
	ChapterID string `json:"chapter_id,omitempty"`
}

// CreateNovelChapter appends a chapter to a novel.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_createchapter.html
func (w *MiniProgram) CreateNovelChapter(ctx context.Context, req *CreateNovelChapterRequest) (*CreateNovelChapterResponse, error) {
	var result CreateNovelChapterResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/createchapter", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchCreateNovelChapterRequest adds up to 10 chapters at once.
type BatchCreateNovelChapterRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// ChapterList of the chapters (max 10).
	ChapterList []NovelChapter `json:"chapter_list"`
}

// BatchCreateNovelChapterResponse is returned by BatchCreateNovelChapter.
type BatchCreateNovelChapterResponse struct {
	ErrResponse
	// ChapterIDList of the created chapters.
	ChapterIDList []string `json:"chapter_id_list,omitempty"`
}

// BatchCreateNovelChapter appends multiple chapters in one call.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_batchcreatechapter.html
func (w *MiniProgram) BatchCreateNovelChapter(ctx context.Context, req *BatchCreateNovelChapterRequest) (*BatchCreateNovelChapterResponse, error) {
	var result BatchCreateNovelChapterResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/batchcreatechapter", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteNovelChapterRequest deletes one chapter.
type DeleteNovelChapterRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// ChapterID of the chapter to delete.
	ChapterID string `json:"chapter_id"`
}

// DeleteNovelChapter removes a chapter from a novel.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_deletechapter.html
func (w *MiniProgram) DeleteNovelChapter(ctx context.Context, bookID, chapterID string) error {
	body := &DeleteNovelChapterRequest{BookID: bookID, ChapterID: chapterID}
	return w.withAccessTokenPost(ctx, "/wxa/book/deletechapter", nil, defaultReqOptions(), body, nil)
}

// GetNovelChapterRequest selects a chapter.
type GetNovelChapterRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id,omitempty"`
	// ChapterID of the chapter.
	ChapterID string `json:"chapter_id,omitempty"`
	// NeedEditedData fetches edited data when true.
	NeedEditedData bool `json:"need_edited_data,omitempty"`
}

// GetNovelChapterResponse is returned by GetNovelChapter.
type GetNovelChapterResponse struct {
	ErrResponse
	// Chapter of the query.
	Chapter map[string]any `json:"chapter,omitempty"`
}

// GetNovelChapter returns one chapter.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_getchapter.html
func (w *MiniProgram) GetNovelChapter(ctx context.Context, req *GetNovelChapterRequest) (*GetNovelChapterResponse, error) {
	var result GetNovelChapterResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/getchapter", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListNovelChapterRequest pages the chapters of a novel.
type ListNovelChapterRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// NeedEditedData fetches edited data when true.
	NeedEditedData bool `json:"need_edited_data,omitempty"`
	// Limit of the page (default 10, max 100).
	Limit int `json:"limit,omitempty"`
	// Offset of the page.
	Offset int64 `json:"offset,omitempty"`
	// VolumeIndex filters to one volume when set.
	VolumeIndex int `json:"volume_index,omitempty"`
}

// ListNovelChapterResponse is returned by ListNovelChapter.
type ListNovelChapterResponse struct {
	ErrResponse
	// ChapterList of the page.
	ChapterList []map[string]any `json:"chapter_list,omitempty"`
}

// ListNovelChapter lists the chapters of a novel.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_listchapter.html
func (w *MiniProgram) ListNovelChapter(ctx context.Context, req *ListNovelChapterRequest) (*ListNovelChapterResponse, error) {
	var result ListNovelChapterResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/listchapter", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReplaceNovelChapterRequest replaces the content of a published chapter.
type ReplaceNovelChapterRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// ChapterID of the chapter.
	ChapterID string `json:"chapter_id"`
	// NewChapterTitle of the chapter (1-80 chars).
	NewChapterTitle string `json:"new_chapter_title"`
	// NewContent of the chapter (1-20000 chars).
	NewContent string `json:"new_content"`
}

// ReplaceNovelChapter replaces a chapter's title and content.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_replacechapter.html
func (w *MiniProgram) ReplaceNovelChapter(ctx context.Context, req *ReplaceNovelChapterRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/book/replacechapter", nil, defaultReqOptions(), req, nil)
}

// ReorderNovelChapterRequest moves a chapter relative to another.
type ReorderNovelChapterRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// ChapterID of the chapter to move.
	ChapterID string `json:"chapter_id"`
	// TargetChapterID anchors the move.
	TargetChapterID string `json:"target_chapter_id"`
	// Operation: 1 swap, 2 insert before the target, 3 insert after.
	Operation int `json:"operation"`
}

// ReorderNovelChapter reorders a chapter in a novel.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_reorderchapter.html
func (w *MiniProgram) ReorderNovelChapter(ctx context.Context, req *ReorderNovelChapterRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/book/reorderchapter", nil, defaultReqOptions(), req, nil)
}

// NovelChapterSeq is one seq assignment of a chapter.
type NovelChapterSeq struct {
	// ChapterID of the chapter.
	ChapterID string `json:"chapter_id,omitempty"`
	// Seq assigned to the chapter.
	Seq int64 `json:"seq,omitempty"`
}

// UpdateNovelChapterSeqRequest reassigns the ordering seq of chapters.
type UpdateNovelChapterSeqRequest struct {
	// BookID of the novel.
	BookID string `json:"book_id"`
	// ChapterSeqList of the new seq assignments.
	ChapterSeqList []NovelChapterSeq `json:"chapter_seq_list"`
}

// UpdateNovelChapterSeq updates the seq ordering of chapters.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/business/api_updatechapterseq.html
func (w *MiniProgram) UpdateNovelChapterSeq(ctx context.Context, req *UpdateNovelChapterSeqRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/book/updatechapterseq", nil, defaultReqOptions(), req, nil)
}
