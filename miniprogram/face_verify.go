package miniprogram

import (
	"context"
)

// ============================================================
// 人脸核身 (face verification) via cityservice. Used by政务/服务 Mini
// Programs that need verified identity.
// ============================================================

// FaceVerifyCertInfo is the identity document of a face-verify session.
type FaceVerifyCertInfo struct {
	// CertType of the document, "IDENTITY_CARD" for ID cards.
	CertType string `json:"cert_type"`
	// CertName on the document.
	CertName string `json:"cert_name"`
	// CertNo of the document.
	CertNo string `json:"cert_no"`
}

// GetFaceVerifyIDRequest starts a face-verification session.
type GetFaceVerifyIDRequest struct {
	// OutSeqNo of the business flow (5-32 chars, unique per appid).
	OutSeqNo string `json:"out_seq_no"`
	// CertInfo of the user identity.
	CertInfo FaceVerifyCertInfo `json:"cert_info"`
	// OpenID of the user to verify.
	OpenID string `json:"openid"`
}

// GetFaceVerifyIDResponse is returned by GetFaceVerifyID.
type GetFaceVerifyIDResponse struct {
	ErrResponse
	// VerifyID of the face-verify session.
	VerifyID string `json:"verify_id,omitempty"`
	// ExpiresIn of the session (default 3600 seconds).
	ExpiresIn int `json:"expires_in,omitempty"`
}

// GetFaceVerifyID creates a face-verification session whose URL the Mini
// Program opens with wx.requestFacialVerify.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/face/api_getverifyid.html
func (w *MiniProgram) GetFaceVerifyID(ctx context.Context, req *GetFaceVerifyIDRequest) (*GetFaceVerifyIDResponse, error) {
	var result GetFaceVerifyIDResponse
	if err := w.withAccessTokenPost(ctx, "/cityservice/face/identify/getverifyid", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryFaceVerifyInfoRequest queries the result of a face-verification session.
type QueryFaceVerifyInfoRequest struct {
	// VerifyID returned by GetFaceVerifyID.
	VerifyID string `json:"verify_id"`
	// OutSeqNo matching the GetFaceVerifyID call.
	OutSeqNo string `json:"out_seq_no"`
	// CertHash generated from the identity document.
	CertHash string `json:"cert_hash"`
	// OpenID matching the GetFaceVerifyID call.
	OpenID string `json:"openid"`
}

// QueryFaceVerifyInfoResponse is returned by QueryFaceVerifyInfo.
type QueryFaceVerifyInfoResponse struct {
	ErrResponse
	// VerifyRet of the face verification result.
	VerifyRet int `json:"verify_ret,omitempty"`
}

// QueryFaceVerifyInfo returns the result of a face-verify session.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/face/api_queryverifyinfo.html
func (w *MiniProgram) QueryFaceVerifyInfo(ctx context.Context, req *QueryFaceVerifyInfoRequest) (*QueryFaceVerifyInfoResponse, error) {
	var result QueryFaceVerifyInfoResponse
	if err := w.withAccessTokenPost(ctx, "/cityservice/face/identify/queryverifyinfo", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
