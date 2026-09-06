package official

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-sphere/weixin-api/core"
)

// ============================================================
// 素材管理 (media & material management).
// Temporary media is for transient uploads (valid 3 days); permanent
// materials are news images, thumbnails and video stored indefinitely.
// ============================================================

// UploadTempMediaRequest uploads a temporary media file.
type UploadTempMediaRequest struct {
	// MediaType: image / voice / video / thumb.
	MediaType string
	// Filename of the uploaded file.
	Filename string
	// Content of the uploaded file.
	Content []byte
}

// UploadTempMediaResponse is returned by UploadTempMedia.
type UploadTempMediaResponse struct {
	ErrResponse
	// Type of the uploaded media.
	Type string `json:"type"`
	// MediaID of the uploaded media.
	MediaID string `json:"media_id"`
	// CreatedAt of the upload (unix seconds).
	CreatedAt int64 `json:"created_at"`
}

// UploadTempMedia uploads a temporary media (valid for 3 days).
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Adding_Permanent_Assets.html
func (oa *OfficialAccount) UploadTempMedia(ctx context.Context, req *UploadTempMediaRequest) (*UploadTempMediaResponse, error) {
	query := url.Values{}
	query.Set("type", req.MediaType)
	data, err := oa.WithTokenUpload(ctx, "/cgi-bin/media/upload", query, core.DefaultRequestOptions(), nil, "media", req.Filename, "", bytes.NewReader(req.Content))
	if err != nil {
		return nil, err
	}
	var result UploadTempMediaResponse
	if err := core.DecodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTempMedia returns the raw bytes of a temporary media file.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Get_temporary_materials.html
func (oa *OfficialAccount) GetTempMedia(ctx context.Context, mediaID string) ([]byte, error) {
	query := url.Values{}
	query.Set("media_id", mediaID)
	return oa.withTokenRaw(ctx, http.MethodGet, "/cgi-bin/media/get", query, core.DefaultRequestOptions(), nil)
}

// AddPermanentMaterialRequest uploads a permanent material.
type AddPermanentMaterialRequest struct {
	// MaterialType: image / voice / video / thumb.
	MaterialType string
	// Filename of the uploaded file.
	Filename string
	// Content of the uploaded file.
	Content []byte
	// Title of the video material (required for video).
	Title string
	// Introduction of the video material (required for video).
	Introduction string
}

// AddPermanentMaterialResponse is returned by AddPermanentMaterial.
type AddPermanentMaterialResponse struct {
	ErrResponse
	// MediaID of the permanent material.
	MediaID string `json:"media_id"`
	// URL of the material (images only).
	URL string `json:"url,omitempty"`
}

// AddPermanentMaterial uploads an image/thumb permanent material.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Adding_Permanent_Assets.html
func (oa *OfficialAccount) AddPermanentMaterial(ctx context.Context, req *AddPermanentMaterialRequest) (*AddPermanentMaterialResponse, error) {
	if req.MaterialType == "" {
		return nil, fmt.Errorf("wechat: add permanent material: empty material type")
	}
	query := url.Values{}
	query.Set("type", req.MaterialType)
	form := url.Values{}
	if req.Title != "" {
		form.Set("title", req.Title)
	}
	if req.Introduction != "" {
		form.Set("introduction", req.Introduction)
	}
	data, err := oa.WithTokenUpload(ctx, "/cgi-bin/material/add_material", query, core.DefaultRequestOptions(), form, "media", req.Filename, "", bytes.NewReader(req.Content))
	if err != nil {
		return nil, err
	}
	var result AddPermanentMaterialResponse
	if err := core.DecodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPermanentMaterialRequest selects a permanent material by media_id.
type GetPermanentMaterialRequest struct {
	// MediaID of the material.
	MediaID string
}

// GetPermanentMaterial returns the raw bytes (or JSON article list) of a
// permanent material.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Get_materials.html
func (oa *OfficialAccount) GetPermanentMaterial(ctx context.Context, mediaID string) ([]byte, error) {
	body := map[string]string{"media_id": mediaID}
	return oa.withTokenRaw(ctx, http.MethodPost, "/cgi-bin/material/get_material", nil, core.DefaultRequestOptions(), body)
}

// DeletePermanentMaterial removes a permanent material.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Deleting_materials.html
func (oa *OfficialAccount) DeletePermanentMaterial(ctx context.Context, mediaID string) error {
	body := map[string]string{"media_id": mediaID}
	return oa.withTokenPost(ctx, "/cgi-bin/material/del_material", nil, core.DefaultRequestOptions(), body, nil)
}

// GetMaterialCountResponse is returned by GetMaterialCount.
type GetMaterialCountResponse struct {
	ErrResponse
	// VoiceCount of permanent voice materials.
	VoiceCount int `json:"voice_count"`
	// VideoCount of permanent video materials.
	VideoCount int `json:"video_count"`
	// ImageCount of permanent image materials.
	ImageCount int `json:"image_count"`
	// NewsCount of permanent news materials.
	NewsCount int `json:"news_count"`
}

// GetMaterialCount returns the counts of the permanent materials.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Get_the_total_of_all_materials.html
func (oa *OfficialAccount) GetMaterialCount(ctx context.Context) (*GetMaterialCountResponse, error) {
	var result GetMaterialCountResponse
	if err := oa.withToken(ctx, http.MethodGet, "/cgi-bin/material/get_materialcount", nil, core.DefaultRequestOptions(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListMaterialRequest pages the permanent materials of one type.
type ListMaterialRequest struct {
	// MaterialType: image / video / voice / news.
	MaterialType string
	// Offset of the page.
	Offset int
	// Count of the page (1-20).
	Count int
}

// ListMaterialResponse is returned by ListMaterial.
type ListMaterialResponse struct {
	ErrResponse
	// TotalCount of the materials.
	TotalCount int `json:"total_count"`
	// ItemCount of the returned page.
	ItemCount int `json:"item_count"`
	// Item of the returned materials.
	Item []map[string]any `json:"item"`
}

// ListMaterial pages the permanent materials of an account.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Get_materials_list.html
func (oa *OfficialAccount) ListMaterial(ctx context.Context, req *ListMaterialRequest) (*ListMaterialResponse, error) {
	body := map[string]any{
		"type":   req.MaterialType,
		"offset": req.Offset,
		"count":  req.Count,
	}
	var result ListMaterialResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/material/batchget_material", nil, core.DefaultRequestOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ============================================================
// 二维码 (QR codes) — temporary & permanent scene QR.
// ============================================================

// CreateQRCodeRequest creates a QR ticket with an integer scene.
type CreateQRCodeRequest struct {
	// ExpireSeconds of a temporary QR (30-2592000).
	ExpireSeconds int `json:"expire_seconds,omitempty"`
	// ActionName: QR_SCENE / QR_STR_SCENE (temp), QR_LIMIT_SCENE /
	// QR_LIMIT_STR_SCENE (permanent).
	ActionName string `json:"action_name"`
	// ActionInfo of the scene.
	ActionInfo QRSceneActionInfo `json:"action_info"`
}

// QRSceneActionInfo carries the QR scene payload.
type QRSceneActionInfo struct {
	// Scene of the QR code.
	Scene QRScene `json:"scene"`
}

// QRScene carries an integer or string scene id.
type QRScene struct {
	// SceneID of an integer scene.
	SceneID int `json:"scene_id,omitempty"`
	// SceneStr of a string scene.
	SceneStr string `json:"scene_str,omitempty"`
}

// CreateQRCodeResponse is returned by CreateQRCode.
type CreateQRCodeResponse struct {
	ErrResponse
	// Ticket used to render the QR image.
	Ticket string `json:"ticket"`
	// ExpireSeconds of the QR.
	ExpireSeconds int `json:"expire_seconds"`
	// URL of the QR.
	URL string `json:"url"`
}

// CreateQRCode creates a scene QR ticket that can be rendered into an image.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Account_Management/Generating_a_Parametric_QR_Code.html
func (oa *OfficialAccount) CreateQRCode(ctx context.Context, req *CreateQRCodeRequest) (*CreateQRCodeResponse, error) {
	var result CreateQRCodeResponse
	if err := oa.withTokenPost(ctx, "/cgi-bin/qrcode/create", nil, core.DefaultRequestOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ShowQRCode renders the QR image bytes for a ticket.
//
// The render endpoint lives on mp.weixin.qq.com and needs no access_token, so
// it bypasses the token-based transport entirely; the ticket is URL-encoded
// exactly once by the query encoder.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Account_Management/Generating_a_Parametric_QR_Code.html
func (oa *OfficialAccount) ShowQRCode(ctx context.Context, ticket string) ([]byte, error) {
	query := url.Values{}
	query.Set("ticket", ticket)
	return oa.qrImage(ctx, query)
}

// qrImage performs the token-less GET that renders a QR code from its ticket.
func (oa *OfficialAccount) qrImage(ctx context.Context, query url.Values) ([]byte, error) {
	base := oa.qrBaseURL
	if base == "" {
		base = "https://mp.weixin.qq.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/cgi-bin/showqrcode", nil)
	if err != nil {
		return nil, fmt.Errorf("wechat: build showqrcode request: %w", err)
	}
	req.URL.RawQuery = query.Encode()
	httpClient := oa.qrHTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wechat: showqrcode: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wechat: showqrcode: unexpected HTTP status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 16<<20))
}

// CreateShortURL shortens a long URL with a weixin.cn short link.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/Account_Management/URL_Shortener.html
func (oa *OfficialAccount) CreateShortURL(ctx context.Context, longURL string) (string, error) {
	var result struct {
		ErrResponse
		ShortURL string `json:"short_url"`
	}
	body := map[string]string{"action": "long2short", "long_url": longURL}
	if err := oa.withTokenPost(ctx, "/cgi-bin/shorturl", nil, core.DefaultRequestOptions(), body, &result); err != nil {
		return "", err
	}
	return result.ShortURL, nil
}
