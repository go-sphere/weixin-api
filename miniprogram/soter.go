package miniprogram

import (
	"context"
)

// VerifySignatureRequest is the payload of VerifySignature. The SOTER (生物认证
// 生物认证) signature verification confirms that a biometric authentication
// (fingerprint / face) actually happened on the user's device.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/soter/api_verifysignature.html
type VerifySignatureRequest struct {
	// OpenID of the user who authenticated.
	OpenID string `json:"openid"`
	// JsonString is the raw json produced by the client SOTER flow.
	JsonString string `json:"json_string"`
	// JsonSignature is the signature over JsonString produced by the client
	// SOTER flow.
	JsonSignature string `json:"json_signature"`
}

// VerifySignatureResponse is returned by VerifySignature.
type VerifySignatureResponse struct {
	ErrResponse
	// IsOk is true when the signature is valid.
	IsOk bool `json:"is_ok"`
}

// VerifySignature verifies a SOTER (biometric) signature produced by the
// wx.checkIsSupportSoterAuthentication / wx.startSoterAuthentication flows.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/soter/api_verifysignature.html
func (w *MiniProgram) VerifySignature(ctx context.Context, req *VerifySignatureRequest) (*VerifySignatureResponse, error) {
	var result VerifySignatureResponse
	if err := w.withAccessTokenPost(ctx, "/cgi-bin/soter/verify_signature", nil, defaultReqOptions(), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
