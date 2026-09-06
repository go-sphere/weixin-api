package miniprogram

import (
	"bytes"
	"context"
	"net/url"
)

// UserProfile holds optional user context that WeChat may use to improve the
// risk scoring of a content check.
type UserProfile struct {
	// OpenID of the user who created the content.
	OpenID string `json:"openid,omitempty"`
	// Nickname of the user.
	Nickname string `json:"nickname,omitempty"`
	// Country of the user (ISO 3166-1 alpha-2).
	Country string `json:"country,omitempty"`
	// Province of the user.
	Province string `json:"province,omitempty"`
	// City of the user.
	City string `json:"city,omitempty"`
	// AvatarURL of the user.
	AvatarURL string `json:"avatar_url,omitempty"`
	// PhoneNumber of the user (country code + number).
	PhoneNumber string `json:"phonenumber,omitempty"`
}

// MsgSecCheckResult is the risk verdict of one checked fragment.
type MsgSecCheckResult struct {
	// Label is the risk label of the content: 100 = normal, 10001 =
	// advertising, 20001 = abuse, 20002 = sexual, 20003 = political, 20004 =
	// violence, 20006 = illegal drugs, 20008 = gambling, 20013 = piracy,
	// 21000 = other.
	Label int `json:"label"`
	// Suggestion: "pass", "review" or "risky".
	Suggestion string `json:"suggest"`
	// RiskLevel for risk-level-2 accounts: "normal", "risky" or "review".
	RiskLevel string `json:"risk_level,omitempty"`
}

// MsgSecCheckRequest carries user generated text content to check.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_msgseccheck.html
type MsgSecCheckRequest struct {
	// Content is the text to check (required, up to 2500 bytes of UTF-8).
	Content string `json:"content"`
	// Version selects the risk-level model: 1 (default) or 2. Version 2
	// requires the profile fields below and returns risk_level per hit.
	Version int `json:"version,omitempty"`
	// Scene is the usage scene, required for version 2.
	Scene int `json:"scene,omitempty"`
	// OpenID of the user who created the content, required for version 2.
	OpenID string `json:"openid,omitempty"`
	// Title of the content, optional context for version 2.
	Title string `json:"title,omitempty"`
	// Nickname of the user, optional context for version 2.
	Nickname string `json:"nickname,omitempty"`
	// Signature of the user, optional context for version 2.
	Signature string `json:"signature,omitempty"`
}

// MsgSecCheckResponse is returned by MsgSecCheck.
type MsgSecCheckResponse struct {
	ErrResponse
	// Result is the overall check result. errcode 0 and
	// Result.Suggestion == "pass" mean the content is safe.
	Result MsgSecCheckResult `json:"result"`
	// TraceID can be reported back to WeChat when disputing a verdict.
	TraceID string `json:"trace_id"`
	// Detail is filled for version 2 with one entry per checked fragment.
	Detail []MsgSecCheckResult `json:"detail,omitempty"`
}

// MsgSecCheck synchronously checks user generated text (nicknames, signatures,
// messages) for risky content.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_msgseccheck.html
func (w *MiniProgram) MsgSecCheck(ctx context.Context, req *MsgSecCheckRequest) (*MsgSecCheckResponse, error) {
	var result MsgSecCheckResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/msg_sec_check", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MediaCheckAsyncRequest triggers an asynchronous media content check.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_mediacheckasync.html
type MediaCheckAsyncRequest struct {
	// MediaURL is the publicly reachable URL of the media to check
	// (recommended size 5 KB to 10 MB).
	MediaURL string `json:"media_url"`
	// MediaType of the media: 1 = image, 2 = video, 3 = audio.
	MediaType int `json:"media_type"`
	// Version selects the risk-level model (1 or 2; default 1).
	Version int `json:"version,omitempty"`
	// OpenID of the user who uploaded the media.
	OpenID string `json:"openid,omitempty"`
	// Scene for version 2.
	Scene int `json:"scene,omitempty"`
	UserProfile
}

// MediaCheckAsyncResponse is returned by MediaCheckAsync. The actual verdict
// arrives later via the message push configured for the Mini Program
// (wxa_media_check event).
type MediaCheckAsyncResponse struct {
	ErrResponse
	// TraceID of the check; the pushed result references it.
	TraceID string `json:"trace_id"`
}

// MediaCheckAsync queues an asynchronous content check for an image, video or
// audio file. The result is pushed to the message callback server configured
// for the Mini Program (event wxa_media_check).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_mediacheckasync.html
func (w *MiniProgram) MediaCheckAsync(ctx context.Context, req *MediaCheckAsyncRequest) (*MediaCheckAsyncResponse, error) {
	var result MediaCheckAsyncResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/media_check_async", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImgSecCheckRequest carries an image to check synchronously. The image is
// uploaded as multipart media, so it does not need a public URL.
type ImgSecCheckRequest struct {
	// Filename of the uploaded image.
	Filename string
	// Content is the raw image bytes.
	Content []byte
}

// ImgSecCheckResponse is returned by ImgSecCheck.
type ImgSecCheckResponse struct {
	ErrResponse
	// Result of the image check.
	Result MsgSecCheckResult `json:"result"`
	// TraceID of the check.
	TraceID string `json:"trace_id"`
}

// ImgSecCheck synchronously checks one uploaded image for risky content
// (multipart upload, the counterpart of MediaCheckAsync for images).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_imgseccheck.html
func (w *MiniProgram) ImgSecCheck(ctx context.Context, req *ImgSecCheckRequest) (*ImgSecCheckResponse, error) {
	if req.Filename == "" {
		req.Filename = "image.jpg"
	}
	data, err := w.withAccessTokenUpload(ctx, "/wxa/img_sec_check", nil, defaultReqOptions(), url.Values{}, "media", req.Filename, "", bytes.NewReader(req.Content))
	if err != nil {
		return nil, err
	}
	var result ImgSecCheckResponse
	if err := decodeJSON(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UserRiskRankRequest asks WeChat to score the risk level of a user.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/safety-control-capability/api_getuserriskrank.html
type UserRiskRankRequest struct {
	// AppID must be the current Mini Program appid.
	AppID string `json:"appid"`
	// OpenID of the user to score.
	OpenID string `json:"openid"`
	// Scene of the operation: 0 = login (default), 1 = register, 2 =
	// add friend, 3 = pay, 4 = create group/room, 5 = other. Scenes 0 and 1
	// require MobileNo.
	Scene int `json:"scene,omitempty"`
	// MobileNo is the user's phone number (required for scenes 0 and 1).
	MobileNo string `json:"mobile_no,omitempty"`
	// Email of the user (optional context).
	Email string `json:"email,omitempty"`
	// UserIP of the user (optional context).
	UserIP string `json:"user_ip,omitempty"`
	// NickName of the user (optional context).
	NickName string `json:"nick_name,omitempty"`
	// CertificateNumber is the ID card number (optional context).
	CertificateNumber string `json:"certificate_number,omitempty"`
	// UnionID of the user (optional context).
	UnionID string `json:"union_id,omitempty"`
}

// UserRiskRankResponse is returned by GetUserRiskRank.
type UserRiskRankResponse struct {
	ErrResponse
	// RiskRank of the user for the scene: 0 = unknown/no risk, 1 = low risk,
	// 2 = medium risk, 3 = high risk. Non-zero suggests the user is risky for
	// the queried scene.
	RiskRank int `json:"risk_rank"`
	// UnionID is echoed when a unionid context was used.
	UnionID int `json:"union_id,omitempty"`
}

// GetUserRiskRank scores how risky a user operation is (fraud / abuse control
// capability). The returned risk_rank guides decisions such as captcha or
// manual review.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/safety-control-capability/api_getuserriskrank.html
func (w *MiniProgram) GetUserRiskRank(ctx context.Context, req *UserRiskRankRequest) (*UserRiskRankResponse, error) {
	if req.AppID == "" {
		req.AppID = w.config.AppID
	}
	var result UserRiskRankResponse
	if err := w.withAccessTokenPost(ctx, "/wxa/getuserriskrank", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
