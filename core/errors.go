package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// hmacSHA256Hex returns the lowercase hex HMAC-SHA256 of msg under key. It is
// the primitive behind the XPay request signatures.
func hmacSHA256Hex(key, msg []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return hex.EncodeToString(mac.Sum(nil))
}

// Well-known WeChat business error codes that affect access-token lifecycle
// management. Platform clients use them to decide whether a failed request
// should be retried after forcing an access-token refresh.
const (
	// ErrCodeInvalidCredential means the AppSecret is wrong or the access token
	// is invalid while obtaining it from the /cgi-bin/token endpoint.
	ErrCodeInvalidCredential = 40001
	// ErrCodeInvalidAccessToken means the access token passed to an API is not
	// valid (for example it is expired or belongs to another app).
	ErrCodeInvalidAccessToken = 40014
	// ErrCodeAccessTokenExpired means the access token has timed out.
	ErrCodeAccessTokenExpired = 42001
	// ErrCodeAPIFreqOutOfLimit means the current API daily call quota has been
	// exhausted.
	ErrCodeAPIFreqOutOfLimit = 45009
)

// Sentinel errors returned for access-token problems. Use errors.Is to match
// them when implementing custom retry/refresh logic.
var (
	ErrorInvalidCredential  = errors.New("wechat: invalid app credential (errcode 40001)")
	ErrorAccessTokenExpired = errors.New("wechat: access token expired (errcode 42001)")
	ErrorInvalidAccessToken = errors.New("wechat: invalid access token (errcode 40014)")
)

// ErrResponse mirrors the universal error envelope returned by WeChat APIs:
//
//	{"errcode": 0, "errmsg": "ok"}
//
// Almost every WeChat API embeds these two fields in its response payload even
// on success, so JSON tags are fixed to errcode/errmsg.
type ErrResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Error implements the error interface using the human readable errmsg when
// available.
func (e ErrResponse) Error() string {
	if strings.TrimSpace(e.ErrMsg) == "" {
		return fmt.Sprintf("wechat api error: errcode %d", e.ErrCode)
	}
	return fmt.Sprintf("wechat api error: %s (errcode %d)", e.ErrMsg, e.ErrCode)
}

// APIError is a WeChat business error. It carries the trace id (rid) when the
// upstream response includes one — the value to report to WeChat support — and
// the raw response bytes so callers can inspect fields the envelope does not
// cover. For endpoints whose documentation tabulates error codes, Description,
// Solution and Doc carry the documented meaning of the code.
type APIError struct {
	ErrResponse
	// RID is the WeChat request trace id (rid field), present on some error
	// responses. It is empty when the upstream response did not include one.
	RID string `json:"rid"`
	// Description is the documented description of the error code.
	Description string `json:"-"`
	// Solution is the documented remedy for the error code.
	Solution string `json:"-"`
	// Raw is the unmodified response body.
	Raw json.RawMessage `json:"-"`
	// Doc is the documented entry the code matched, when one exists.
	Doc *ErrDoc `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RID != "" {
		return fmt.Sprintf("%s (rid: %s)", e.ErrResponse.Error(), e.RID)
	}
	return e.ErrResponse.Error()
}

// ClassifyBusinessError converts a non-zero WeChat business error code into an
// error. Access-token related codes are mapped to the exported sentinel errors
// so callers can detect them with errors.Is; every other code produces an
// *APIError carrying the raw code and message.
func ClassifyBusinessError(code int, msg, rid string) error {
	switch code {
	case ErrCodeInvalidCredential:
		return ErrorInvalidCredential
	case ErrCodeAccessTokenExpired:
		return ErrorAccessTokenExpired
	case ErrCodeInvalidAccessToken:
		return ErrorInvalidAccessToken
	}
	return &APIError{ErrResponse: ErrResponse{ErrCode: code, ErrMsg: msg}, RID: rid}
}

// isNeedRetryError reports whether err is an access-token related failure that
// can be recovered by forcing an access-token refresh and retrying the request.
func isNeedRetryError(err error) bool {
	return errors.Is(err, ErrorInvalidCredential) ||
		errors.Is(err, ErrorAccessTokenExpired) ||
		errors.Is(err, ErrorInvalidAccessToken)
}

// scanErrorBody inspects a raw response body for the WeChat error envelope.
// It returns nil when the body is not JSON, is JSON without an errcode field, or
// carries errcode == 0. Otherwise it returns the classified business error.
func scanErrorBody(body []byte) error {
	var probe struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		RID     string `json:"rid"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return nil
	}
	if probe.ErrCode == 0 {
		return nil
	}
	err := ClassifyBusinessError(probe.ErrCode, probe.ErrMsg, probe.RID)
	// Token failures collapse into sentinels for errors.Is matching; every
	// other code keeps the response bytes for inspection.
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		apiErr.Raw = bytes.Clone(body)
	}
	return err
}
