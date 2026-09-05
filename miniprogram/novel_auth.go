package miniprogram

import (
	"context"
)

// ============================================================
// WeChat novel — authorisation (作品授权), preview settings and
// recommendation. Routes under /wxa/book and /wxa/business/novelreader.
// ============================================================

// NovelAuthBook is one book grant in an authorisation request.
type NovelAuthBook struct {
	// BookID of the book being authorised.
	BookID string `json:"book_id,omitempty"`
	// OriginalID of the provider-side book key.
	OriginalID string `json:"original_id,omitempty"`
}

// AddNovelBookAuthRequest grants book-level authorisation (按书授权).
type AddNovelBookAuthRequest struct {
	// Books to grant.
	Books []NovelAuthBook `json:"books"`
}

// AddNovelBookAuth authorises one or more books to be used by other apps.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/auth/api_addbookauth.html
func (w *MiniProgram) AddNovelBookAuth(ctx context.Context, req *AddNovelBookAuthRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/book/addbookauth", nil, defaultReqOptions(), req, nil)
}

// AddNovelBookAuthByAppIDRequest grants all books of an appid.
type AddNovelBookAuthByAppIDRequest struct {
	// Infos of the appid-level grant (appid pairs).
	Infos []map[string]any `json:"infos"`
}

// AddNovelBookAuthByAppID authorises books to another app at appid level.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/auth/api_addbookauthbyappid.html
func (w *MiniProgram) AddNovelBookAuthByAppID(ctx context.Context, req *AddNovelBookAuthByAppIDRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/book/addbookauthbyappid", nil, defaultReqOptions(), req, nil)
}

// DeleteNovelBookAuthRequest revokes a book authorisation.
type DeleteNovelBookAuthRequest struct {
	// BookID of the book.
	BookID string `json:"book_id"`
	// GranteeAppID of the authorised app.
	GranteeAppID string `json:"grantee_appid"`
}

// DeleteNovelBookAuth revokes the book authorisation of an app.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/auth/api_delbookauth.html
func (w *MiniProgram) DeleteNovelBookAuth(ctx context.Context, bookID, granteeAppID string) error {
	body := &DeleteNovelBookAuthRequest{BookID: bookID, GranteeAppID: granteeAppID}
	return w.withAccessTokenPost(ctx, "/wxa/book/delbookauth", nil, defaultReqOptions(), body, nil)
}

// DeleteNovelBookAuthByAppIDRequest revokes all grants of an appid.
type DeleteNovelBookAuthByAppIDRequest struct {
	// GranteeAppID of the authorised app.
	GranteeAppID string `json:"grantee_appid"`
}

// DeleteNovelBookAuthByAppID revokes all book authorisations of an app.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/auth/api_delbookauthbyappid.html
func (w *MiniProgram) DeleteNovelBookAuthByAppID(ctx context.Context, granteeAppID string) error {
	body := &DeleteNovelBookAuthByAppIDRequest{GranteeAppID: granteeAppID}
	return w.withAccessTokenPost(ctx, "/wxa/book/delbookauthbyappid", nil, defaultReqOptions(), body, nil)
}

// QueryNovelBookAuthRequest pages the book authorisations.
type QueryNovelBookAuthRequest struct {
	// Type of the authorisation query.
	Type int `json:"type,omitempty"`
	// Offset of the page.
	Offset int `json:"offset"`
	// Count of the page.
	Count int `json:"count"`
	// IsSum aggregates all grants when true.
	IsSum bool `json:"is_sum,omitempty"`
	// BookID of a specific book.
	BookID string `json:"book_id,omitempty"`
}

// QueryNovelBookAuthResponse is returned by QueryNovelBookAuth.
type QueryNovelBookAuthResponse struct {
	ErrResponse
	// AuthList of the query result.
	AuthList []map[string]any `json:"auth_list,omitempty"`
}

// QueryNovelBookAuth queries the granted book authorisations.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/auth/api_querybookauth.html
func (w *MiniProgram) QueryNovelBookAuth(ctx context.Context, req *QueryNovelBookAuthRequest) (*QueryNovelBookAuthResponse, error) {
	var result QueryNovelBookAuthResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/book/querybookauth", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryNovelBookAuthV2Request pages the book authorisations (v2 cursor).
type QueryNovelBookAuthV2Request struct {
	// Type of the authorisation query.
	Type int `json:"type,omitempty"`
	// Count of the page.
	Count int `json:"count"`
	// Cursor of the previous page.
	Cursor string `json:"cursor,omitempty"`
	// GrantorAppID filter.
	GrantorAppID string `json:"grantor_appid,omitempty"`
	// BookIDs filter.
	BookIDs []string `json:"book_ids,omitempty"`
}

// QueryNovelBookAuthV2Response is returned by QueryNovelBookAuthV2.
type QueryNovelBookAuthV2Response struct {
	ErrResponse
	// AuthList of the query result.
	AuthList []map[string]any `json:"auth_list,omitempty"`
	// Cursor for the next page.
	Cursor string `json:"cursor,omitempty"`
}

// QueryNovelBookAuthV2 queries the granted book authorisations with a cursor.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/auth/api_querybookauthv2.html
func (w *MiniProgram) QueryNovelBookAuthV2(ctx context.Context, req *QueryNovelBookAuthV2Request) (*QueryNovelBookAuthV2Response, error) {
	var result QueryNovelBookAuthV2Response
	if err := w.withAccessTokenPost(ctx, "/wxa/book/querybookauthv2", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetNovelPreviewSettingResponse is returned by GetNovelPreviewSetting.
type GetNovelPreviewSettingResponse struct {
	ErrResponse
	// Setting of the free-preview configuration.
	Setting map[string]any `json:"setting,omitempty"`
}

// GetNovelPreviewSetting returns the free-preview settings of a book.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/preview/api_getpreviewsetting.html
func (w *MiniProgram) GetNovelPreviewSetting(ctx context.Context, bookID string) (*GetNovelPreviewSettingResponse, error) {
	var result GetNovelPreviewSettingResponse
	body := map[string]string{"book_id": bookID}
	if err := w.withAccessTokenPost(ctx, "/wxa/business/novelreader/getpreviewsetting", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetNovelPreviewSettingRequest updates the free-preview settings of a book.
type SetNovelPreviewSettingRequest struct {
	// BookID of the book.
	BookID string `json:"book_id"`
	// DefaultWords of free preview per chapter.
	DefaultWords int `json:"default_words"`
	// ChapterIndex of a chapter-specific override.
	ChapterIndex int `json:"chapter_index,omitempty"`
	// Words for the chapter-specific override.
	Words int `json:"words,omitempty"`
}

// SetNovelPreviewSetting configures how many words of each chapter are free.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/preview/api_setpreviewsetting.html
func (w *MiniProgram) SetNovelPreviewSetting(ctx context.Context, req *SetNovelPreviewSettingRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/novelreader/setpreviewsetting", nil, defaultReqOptions(), req, nil)
}

// RecommendNovelRequest configures the reader recommendation (推荐) books.
type RecommendNovelRequest struct {
	// RecmdType of the recommendation slot.
	RecmdType int `json:"recmd_type"`
	// BookIDList in the recommended order.
	BookIDList []string `json:"book_id_list"`
}

// RecommendNovel sets the recommended novels shown to readers.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/novel/other/api_novelreadersetrecmdnovel.html
func (w *MiniProgram) RecommendNovel(ctx context.Context, req *RecommendNovelRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/business/novelreader/setrecmdnovel", nil, defaultReqOptions(), req, nil)
}
