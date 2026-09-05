package miniprogram

import (
	"context"
)

// ============================================================
// 小程序短剧 (minidrama) — WeChat short-drama content management (微短剧/小程序
// 短剧). Routes live under /wxa/sec/vod/*. Drama ids, audit lifecycle and
// media uploads are managed here.
// ============================================================

// MiniDramaAuditCopyright is the copyright-protection payload of an audit.
type MiniDramaAuditCopyright struct {
	// ApplyCopyright asks for copyright protection when true.
	ApplyCopyright bool `json:"apply_copyright,omitempty"`
	// CopyrightMaterialID of the copyright registration material when applying.
	CopyrightMaterialID string `json:"copyright_material_id,omitempty"`
	// AuthorizedMaterialID of the rights/play authorisation statement when NOT
	// applying for protection.
	AuthorizedMaterialID string `json:"authorized_material_id,omitempty"`
}

// MiniDramaActor is one actor entry of the drama audit.
type MiniDramaActor struct {
	// Name of the actor.
	Name string `json:"name,omitempty"`
	// RoleName played by the actor.
	RoleName string `json:"role_name,omitempty"`
}

// MiniDramaReplaceMedia is one media replacement during re-audit.
type MiniDramaReplaceMedia struct {
	// MediaID of the old media being replaced.
	MediaID string `json:"media_id,omitempty"`
	// NewMediaID of the replacement media.
	NewMediaID string `json:"new_media_id,omitempty"`
}

// AuditMiniDramaRequest submits a drama (剧目) for the short-drama audit.
type AuditMiniDramaRequest struct {
	// DramaID of a re-audit (omit on first submission).
	DramaID int64 `json:"drama_id,omitempty"`
	// Name of the drama (required on first submission).
	Name string `json:"name,omitempty"`
	// MediaCount of the episodes.
	MediaCount int `json:"media_count,omitempty"`
	// MediaIDList of the episode media (count must match MediaCount).
	MediaIDList []string `json:"media_id_list,omitempty"`
	// Description of the drama (up to 200 chars).
	Description string `json:"description,omitempty"`
	// Recommendations blurb (up to 30 chars).
	Recommendations string `json:"recommendations,omitempty"`
	// CoverMaterialID of the drama poster (temporary material).
	CoverMaterialID string `json:"cover_material_id,omitempty"`
	// PromotionPosterMaterialID of the promotion poster.
	PromotionPosterMaterialID string `json:"promotion_poster_material_id,omitempty"`
	// Producer of the drama.
	Producer string `json:"producer,omitempty"`
	// QualificationType: 1 licensed, 2 unlicensed with cost < 1M yuan.
	QualificationType int `json:"qualification_type"`
	// RegistrationNumber required when QualificationType is 1.
	RegistrationNumber string `json:"registration_number,omitempty"`
	// QualificationCertificateMaterialID required when QualificationType is 1.
	QualificationCertificateMaterialID string `json:"qualification_certificate_material_id,omitempty"`
	// CostCommitmentLetterMaterialID required when QualificationType is 2.
	CostCommitmentLetterMaterialID string `json:"cost_commitment_letter_material_id,omitempty"`
	// CostOfProduction in 万元 when QualificationType is 2.
	CostOfProduction int `json:"cost_of_production,omitempty"`
	// Expedited requests audit acceleration when 1.
	Expedited int `json:"expedited,omitempty"`
	// ActorList of 2-5 actors (required for drama_type 2).
	ActorList []MiniDramaActor `json:"actor_list,omitempty"`
	// OtherMaterialMaterialID of extra materials.
	OtherMaterialMaterialID string `json:"other_material_material_id,omitempty"`
	// ReplaceMediaList of re-audit media replacements.
	ReplaceMediaList []MiniDramaReplaceMedia `json:"replace_media_list,omitempty"`
	// Copyright of the drama.
	Copyright *MiniDramaAuditCopyright `json:"copyright"`
	// DramaType: 1 漫剧, 2 真人, 3 数字真人.
	DramaType int `json:"drama_type"`
	// ContentDeclared: 1 contains AI-generated content.
	ContentDeclared int `json:"content_declared,omitempty"`
}

// AuditMiniDramaResponse is returned by AuditMiniDrama.
type AuditMiniDramaResponse struct {
	ErrResponse
	// DramaID of the submitted drama.
	DramaID int64 `json:"drama_id,omitempty"`
}

// AuditMiniDrama submits (or re-submits) a short drama for audit.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/auditdrama/api_auditdrama.html
func (w *MiniProgram) AuditMiniDrama(ctx context.Context, req *AuditMiniDramaRequest) (*AuditMiniDramaResponse, error) {
	var result AuditMiniDramaResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/auditdrama", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMiniDramaRequest selects a drama by id.
type GetMiniDramaRequest struct {
	// DramaID of the drama.
	DramaID int64 `json:"drama_id"`
}

// GetMiniDramaResponse is returned by GetMiniDrama.
type GetMiniDramaResponse struct {
	ErrResponse
	// Drama of the query.
	Drama map[string]any `json:"drama,omitempty"`
}

// GetMiniDrama returns one drama and its episodes.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/auditdrama/api_getdrama.html
func (w *MiniProgram) GetMiniDrama(ctx context.Context, dramaID int64) (*GetMiniDramaResponse, error) {
	req := &GetMiniDramaRequest{DramaID: dramaID}
	var result GetMiniDramaResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getdrama", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListMiniDramasRequest pages the dramas of the account.
type ListMiniDramasRequest struct {
	// Limit of the page (max 100).
	Limit int `json:"limit,omitempty"`
	// Offset of the page.
	Offset int `json:"offset,omitempty"`
}

// ListMiniDramasResponse is returned by ListMiniDramas.
type ListMiniDramasResponse struct {
	ErrResponse
	// DramaInfoList of the page.
	DramaInfoList []map[string]any `json:"drama_info_list,omitempty"`
}

// ListMiniDramas lists the dramas submitted by the account.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/auditdrama/api_listdramas.html
func (w *MiniProgram) ListMiniDramas(ctx context.Context, req *ListMiniDramasRequest) (*ListMiniDramasResponse, error) {
	var result ListMiniDramasResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/listdramas", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMiniDramaLatestAuditResponse is returned by GetMiniDramaLatestAudit.
type GetMiniDramaLatestAuditResponse struct {
	ErrResponse
	// Audit of the latest submission.
	Audit map[string]any `json:"audit,omitempty"`
}

// GetMiniDramaLatestAudit returns the latest audit info of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/auditdrama/api_getdramalatestauditinfo.html
func (w *MiniProgram) GetMiniDramaLatestAudit(ctx context.Context, dramaID int64) (*GetMiniDramaLatestAuditResponse, error) {
	var result GetMiniDramaLatestAuditResponse
	req := &GetMiniDramaRequest{DramaID: dramaID}
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getdramalatestauditinfo", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMiniDramaMediaLinkRequest selects a media to fetch its playback URL.
type GetMiniDramaMediaLinkRequest struct {
	// MediaID of the media.
	MediaID string `json:"media_id"`
	// DramaID of the owning drama.
	DramaID int64 `json:"drama_id,omitempty"`
	// Expire of the link in seconds.
	Expire int `json:"expire,omitempty"`
}

// GetMiniDramaMediaLinkResponse is returned by GetMiniDramaMediaLink.
type GetMiniDramaMediaLinkResponse struct {
	ErrResponse
	// MediaURL of the playable media link.
	MediaURL string `json:"media_url,omitempty"`
}

// GetMiniDramaMediaLink returns a playable URL for an episode media.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_media/api_getmedialink.html
func (w *MiniProgram) GetMiniDramaMediaLink(ctx context.Context, req *GetMiniDramaMediaLinkRequest) (*GetMiniDramaMediaLinkResponse, error) {
	var result GetMiniDramaMediaLinkResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getmedialink", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListMiniDramaMediaRequest pages the media of a drama.
type ListMiniDramaMediaRequest struct {
	// DramaID of the drama.
	DramaID int64 `json:"drama_id"`
	// Limit of the page.
	Limit int `json:"limit,omitempty"`
	// Offset of the page.
	Offset int `json:"offset,omitempty"`
}

// ListMiniDramaMediaResponse is returned by ListMiniDramaMedia.
type ListMiniDramaMediaResponse struct {
	ErrResponse
	// MediaList of the page.
	MediaList []map[string]any `json:"media_list,omitempty"`
}

// ListMiniDramaMedia lists the episode media of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_media/api_listmedia.html
func (w *MiniProgram) ListMiniDramaMedia(ctx context.Context, req *ListMiniDramaMediaRequest) (*ListMiniDramaMediaResponse, error) {
	var result ListMiniDramaMediaResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/listmedia", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMiniDramaMediaRequest removes an episode media.
type DeleteMiniDramaMediaRequest struct {
	// MediaID of the media to delete.
	MediaID string `json:"media_id"`
	// DramaID of the owning drama.
	DramaID int64 `json:"drama_id,omitempty"`
}

// DeleteMiniDramaMedia deletes an episode media of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_media/api_deletemedia.html
func (w *MiniProgram) DeleteMiniDramaMedia(ctx context.Context, req *DeleteMiniDramaMediaRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/deletemedia", nil, defaultReqOptions(), req, nil)
}

// ApplyMiniDramaUploadResponse is returned by ApplyMiniDramaUpload.
type ApplyMiniDramaUploadResponse struct {
	ErrResponse
	// RawData of the upload ticket payload.
	RawData map[string]any `json:"-"`
}

// ApplyMiniDramaUpload starts a media upload session (returns upload
// credentials).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_fileupload/api_applyupload.html
func (w *MiniProgram) ApplyMiniDramaUpload(ctx context.Context, body map[string]any) (*ApplyMiniDramaUploadResponse, error) {
	var result ApplyMiniDramaUploadResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/applyupload", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CommitMiniDramaUploadRequest commits an uploaded drama media.
type CommitMiniDramaUploadRequest struct {
	// RawData of the commit payload.
	RawData map[string]any `json:"-"`
}

// CommitMiniDramaUpload finalises an uploaded media of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_fileupload/api_commitupload.html
func (w *MiniProgram) CommitMiniDramaUpload(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/commitupload", nil, defaultReqOptions(), body, nil)
}

// GetMiniDramaUploadTaskResponse is returned by GetMiniDramaUploadTask.
type GetMiniDramaUploadTaskResponse struct {
	ErrResponse
	// RawData of the task state payload.
	RawData map[string]any `json:"-"`
}

// GetMiniDramaUploadTask returns the state of an upload task.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_fileupload/api_gettask.html
func (w *MiniProgram) GetMiniDramaUploadTask(ctx context.Context, body map[string]any) (*GetMiniDramaUploadTaskResponse, error) {
	var result GetMiniDramaUploadTaskResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/gettask", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PullMiniDramaUploadRequest pulls an upload from an external URL.
type PullMiniDramaUploadRequest struct {
	// RawData of the pull-upload payload.
	RawData map[string]any `json:"-"`
}

// PullMiniDramaUpload pulls a drama media from an external URL.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/vod_fileupload/api_pullupload.html
func (w *MiniProgram) PullMiniDramaUpload(ctx context.Context, body map[string]any) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/pullupload", nil, defaultReqOptions(), body, nil)
}

// ReplaceMiniDramaMediaRequest replaces one media during re-audit.
type ReplaceMiniDramaMediaRequest struct {
	// DramaID of the drama.
	DramaID int64 `json:"drama_id"`
	// MediaID of the old media.
	MediaID string `json:"media_id,omitempty"`
	// NewMediaID of the new media.
	NewMediaID string `json:"new_media_id,omitempty"`
	// RawData extra payload.
	RawData map[string]any `json:"-"`
}

// ReplaceMiniDramaMedia replaces an unaudited media of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/auditdrama/api_replacedramamedia.html
func (w *MiniProgram) ReplaceMiniDramaMedia(ctx context.Context, req *ReplaceMiniDramaMediaRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/replacedramamedia", nil, defaultReqOptions(), req, nil)
}

// ============================================================
// Drama authorisation (剧集授权).
// ============================================================

// MiniDramaAuthorization is one authorisation record.
type MiniDramaAuthorization struct {
	// DramaID of the authorised drama.
	DramaID int64 `json:"drama_id,omitempty"`
	// AppID of the authorised Mini Program.
	AppID string `json:"appid,omitempty"`
	// RawData of the remaining fields.
	RawData map[string]any `json:"-"`
}

// AuthorizeMiniDramaRequest authorises apps to play a drama.
type AuthorizeMiniDramaRequest struct {
	// AuthorizeList of the grants.
	AuthorizeList []MiniDramaAuthorization `json:"authorize_list,omitempty"`
}

// AuthorizeMiniDrama grants other Mini Programs the right to play a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizedrama/api_authorizedrama.html
func (w *MiniProgram) AuthorizeMiniDrama(ctx context.Context, req *AuthorizeMiniDramaRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/authorizedrama", nil, defaultReqOptions(), req, nil)
}

// DeauthorizeMiniDrama revokes the play right of a Mini Program.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizedrama/api_deauthorizedrama.html
func (w *MiniProgram) DeauthorizeMiniDrama(ctx context.Context, req *AuthorizeMiniDramaRequest) error {
	return w.withAccessTokenPost(ctx, "/wxa/sec/vod/deauthorizedrama", nil, defaultReqOptions(), req, nil)
}

// GetMiniDramaAuthorizeListResponse is returned by GetMiniDramaAuthorizeList.
type GetMiniDramaAuthorizeListResponse struct {
	ErrResponse
	// AuthorizeList of the grants.
	AuthorizeList []MiniDramaAuthorization `json:"authorize_list,omitempty"`
}

// GetMiniDramaAuthorizeList lists the app authorisations of a drama.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/minidrama/authorizedrama/api_getauthorizedobjects.html
func (w *MiniProgram) GetMiniDramaAuthorizeList(ctx context.Context, req *GetMiniDramaRequest) (*GetMiniDramaAuthorizeListResponse, error) {
	var result GetMiniDramaAuthorizeListResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/sec/vod/getauthorizedobjects", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
