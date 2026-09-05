package miniprogram

import (
	"context"
)

// QuickCheckStudentIdentity verifies whether a user is a verified college
// student (微信学生身份快速验证). The caller must first obtain the check code from
// the client side (wx.requestStudentCheck) after the user completed the
// student verification flow.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/server/API/student/api_quickcheckstudentidentity.html
func (w *MiniProgram) QuickCheckStudentIdentity(ctx context.Context, openid, checkCode string) (*StudentIdentityResponse, error) {
	var result StudentIdentityResponse
	body := map[string]string{
		"openid":               openid,
		"wx_studentcheck_code": checkCode,
	}
	if err := w.withAccessTokenPost(ctx, "/intp/quickcheckstudentidentity", nil, defaultReqOptions(), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StudentIdentityResponse is returned by QuickCheckStudentIdentity.
type StudentIdentityResponse struct {
	ErrResponse
	// BindStatus is 1 when the user is a verified student bound to this Mini
	// Program.
	BindStatus int `json:"bind_status,omitempty"`
	// IsStudent is 1 when the identity check succeeded.
	IsStudent int `json:"is_student,omitempty"`
}
