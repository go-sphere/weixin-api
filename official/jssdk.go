package official

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// JSSDKConfig is the signed wx.config payload for the JS-SDK.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html
type JSSDKConfig struct {
	// AppID of the official account.
	AppID string `json:"appId"`
	// Timestamp of the signature.
	Timestamp string `json:"timestamp"`
	// NonceStr of the signature.
	NonceStr string `json:"nonceStr"`
	// Signature of the JS-SDK config.
	Signature string `json:"signature"`
}

// GetJSSDKConfig builds the signed wx.config payload for a page URL served
// under the account. It fetches a jsapi ticket via the core client. Any URL
// fragment (#...) is stripped by jsSignature before signing, as WeChat requires.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html
func (oa *OfficialAccount) GetJSSDKConfig(ctx context.Context, pageURL string) (*JSSDKConfig, error) {
	ticket, err := oa.GetJsTicket(ctx, false)
	if err != nil {
		return nil, err
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce, err := randomBase62(16)
	if err != nil {
		return nil, err
	}
	return &JSSDKConfig{
		AppID:     oa.config.AppID,
		Timestamp: timestamp,
		NonceStr:  nonce,
		Signature: jsSignature(ticket, nonce, timestamp, pageURL),
	}, nil
}

// randomBase62 returns n crypto-random base62 characters.
func randomBase62(n int) (string, error) {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("official: generate nonce: %w", err)
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
// stripped first, as WeChat requires the fragment-less page URL.
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
