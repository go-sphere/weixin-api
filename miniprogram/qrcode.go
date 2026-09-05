package miniprogram

import (
	"context"
	"net/http"
)

// RGBColor is a decimal RGB triple used by the QR code line_color option.
type RGBColor struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// QrCodeBase carries the options shared by the wxacode QR code endpoints.
type QrCodeBase struct {
	// Width of the code in pixels, default 430, min 280, max 1280.
	Width int `json:"width,omitempty"`
	// AutoColor lets WeChat pick the line colour automatically (default
	// false). It wins over LineColor when true.
	AutoColor bool `json:"auto_color,omitempty"`
	// LineColor is the RGB line colour used when AutoColor is false. Defaults
	// to black {0,0,0}.
	LineColor *RGBColor `json:"line_color,omitempty"`
	// IsHyaline makes the background transparent when true (default false).
	IsHyaline bool `json:"is_hyaline,omitempty"`
}

// GetMiniProgramCodeRequest are the options of GetMiniProgramCode (wxacode).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_getqrcode.html
type GetMiniProgramCodeRequest struct {
	// Path is the page the code points to, e.g. "pages/index/index". It must
	// exist in the published Mini Program. Query parameters belong in the path,
	// e.g. "pages/index/index?foo=bar".
	Path string `json:"path"`
	// EnvVersion selects the Mini Program version to open: develop, trial or
	// release (default). Codes pointing at non-release versions only open for
	// authorised developers.
	EnvVersion string `json:"env_version,omitempty"`
	// CheckPath checks that Path exists in the published Mini Program (default
	// true). Set false only when the page is not yet published.
	CheckPath *bool `json:"check_path,omitempty"`
	QrCodeBase
}

// GetUnlimitedMiniProgramCodeRequest are the options of
// GetUnlimitedMiniProgramCode (wxacodeunlimit), the recommended endpoint for
// dynamic scenes because the encoded scene can be passed at run time.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_getunlimitedqrcode.html
type GetUnlimitedMiniProgramCodeRequest struct {
	// Scene is the scene value carried by the code (required). Up to 32
	// visible characters: digits, upper/lower case letters and
	// !#$&'()*+,/:;=?@-._~ are allowed; other characters must be encoded.
	// Reads back in the Mini Program as the "scene" query parameter.
	Scene string `json:"scene"`
	// Page is the page the code opens (default homepage). Do not start it with
	// "/" and do not append query parameters (they belong in Scene).
	Page string `json:"page,omitempty"`
	// EnvVersion selects the Mini Program version to open (default release).
	EnvVersion string `json:"env_version,omitempty"`
	// CheckPath checks that Page exists (default true). When false the page
	// may not be published yet; the number of such pages is capped at 60000.
	CheckPath *bool `json:"check_path,omitempty"`
	QrCodeBase
}

// CreateMiniProgramQRCodeRequest is the body of CreateMiniProgramQRCode
// (createwxaqrcode). This endpoint generates a QR code (not a Mini Program
// code) and is the right choice when the code must be scanned outside WeChat.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_createqrcode.html
type CreateMiniProgramQRCodeRequest struct {
	// Path is the encoded page path, e.g. "pages/index?query=1".
	Path string `json:"path"`
	// Width of the code in pixels, default 430, min 280, max 1280.
	Width int `json:"width,omitempty"`
}

// GetMiniProgramCode generates a Mini Program code for a fixed path
// (GET /wxa/getwxacode). Suitable when a small, fixed number of codes is
// needed. Returns the code image bytes (JPEG/PNG); media-type errors return a
// JSON APIError.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_getqrcode.html
func (w *MiniProgram) GetMiniProgramCode(ctx context.Context, req *GetMiniProgramCodeRequest) ([]byte, error) {
	return w.withAccessTokenRaw(ctx, http.MethodPost, "/wxa/getwxacode", nil, defaultReqOptions(), req)
}

// GetUnlimitedMiniProgramCode generates a Mini Program code whose scene is
// embedded in the image (POST /wxa/getwxacodeunlimit), allowing an unlimited
// number of distinct codes. Returns the code image bytes; errors return JSON.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_getunlimitedqrcode.html
func (w *MiniProgram) GetUnlimitedMiniProgramCode(ctx context.Context, req *GetUnlimitedMiniProgramCodeRequest) ([]byte, error) {
	return w.withAccessTokenRaw(ctx, http.MethodPost, "/wxa/getwxacodeunlimit", nil, defaultReqOptions(), req)
}

// CreateMiniProgramQRCode creates a scannable QR code linking into the Mini
// Program (POST /cgi-bin/wxaapp/createwxaqrcode). The returned image is a
// conventional QR code readable outside WeChat.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_createqrcode.html
func (w *MiniProgram) CreateMiniProgramQRCode(ctx context.Context, req *CreateMiniProgramQRCodeRequest) ([]byte, error) {
	return w.withAccessTokenRaw(ctx, http.MethodPost, "/cgi-bin/wxaapp/createwxaqrcode", nil, defaultReqOptions(), req)
}
