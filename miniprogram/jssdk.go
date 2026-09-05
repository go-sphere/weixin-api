package miniprogram

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
// Program when the host account is bound accordingly.
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

// jsSignature computes the JS-SDK signature: the URL-decode sensitive string
// "jsapi_ticket=...&noncestr=...&timestamp=...&url=..." sorted lexically and
// SHA-1 hashed.
func jsSignature(ticket, nonce, timestamp, pageURL string) string {
	params := map[string]string{
		"jsapi_ticket": ticket,
		"noncestr":     nonce,
		"timestamp":    timestamp,
		"url":          pageURL,
	}
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
