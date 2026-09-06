package miniprogram

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// JsSDKConfigResponse carries the parameters for wx.config to initialise the
// WeChat JS-SDK on a web page.
type JsSDKConfigResponse struct {
	// AppID of the official account / Mini Program the page belongs to.
	AppID string `json:"appId"`
	// Timestamp of the signature generation.
	Timestamp string `json:"timestamp"`
	// NonceStr used for the signature.
	NonceStr string `json:"nonceStr"`
	// Signature of the JS-SDK configuration.
	Signature string `json:"signature"`
}

// GetJsSDKConfig builds the signed wx.config payload for a web page URL served
// under the configured account. It internally fetches a jsapi ticket. The same
// signature is also valid for JSSDK usage inside web-views opened from the Mini
// Program when the host account is bound accordingly. Any URL fragment (#...) is
// stripped by jsSignature before signing, as WeChat requires.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html
func (w *MiniProgram) GetJsSDKConfig(ctx context.Context, pageURL string) (*JsSDKConfigResponse, error) {
	ticket, err := w.GetJsTicket(ctx, false)
	if err != nil {
		return nil, err
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce, err := randomBase62(16)
	if err != nil {
		return nil, err
	}
	signature := jsSignature(ticket, nonce, timestamp, pageURL)
	return &JsSDKConfigResponse{
		AppID:     w.config.AppID,
		Timestamp: timestamp,
		NonceStr:  nonce,
		Signature: signature,
	}, nil
}

// randomBase62 returns n random characters drawn from the base62 alphabet.
// It uses crypto/rand so the produced nonce is not guessable.
func randomBase62(n int) (string, error) {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("wechat: generate random string: %w", err)
	}
	var sb strings.Builder
	sb.Grow(n)
	for _, b := range raw {
		sb.WriteByte(alphabet[int(b)%len(alphabet)])
	}
	return sb.String(), nil
}

// jsSignature computes the JS-SDK signature. WeChat's algorithm is a SHA-1 of
// the exact fixed-order string "jsapi_ticket=..&noncestr=..&timestamp=..&url=.."
// (keys are not sorted and values are not escaped). The URL fragment (#...) is
// stripped first. This matches official.jsSignature so the two platform
// packages sign identically.
func jsSignature(ticket, nonce, timestamp, pageURL string) string {
	if i := strings.IndexByte(pageURL, '#'); i >= 0 {
		pageURL = pageURL[:i]
	}
	s := "jsapi_ticket=" + ticket +
		"&noncestr=" + nonce +
		"&timestamp=" + timestamp +
		"&url=" + pageURL
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
