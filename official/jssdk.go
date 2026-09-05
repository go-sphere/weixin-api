package official

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
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
// under the account. It fetches a jsapi ticket via the core client.
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
	params := map[string]string{
		"jsapi_ticket": ticket,
		"noncestr":     nonce,
		"timestamp":    timestamp,
		"url":          pageURL,
	}
	return &JSSDKConfig{
		AppID:     oa.config.AppID,
		Timestamp: timestamp,
		NonceStr:  nonce,
		Signature: sha1Signature(params),
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

// sha1Signature sorts the params and computes the SHA-1 of the joined
// "k=v&k=v" string (the JS-SDK signature algorithm).
func sha1Signature(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(params[k])
	}
	sum := sha1.Sum([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}
