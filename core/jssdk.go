package core

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// JSSDKConfig is the signed wx.config payload used to initialise the WeChat
// JS-SDK on a web page. It is shared by every platform package (Mini Program,
// official account) and by the generated clients.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html
type JSSDKConfig struct {
	// AppID of the account the page belongs to.
	AppID string `json:"appId"`
	// Timestamp of the signature generation (Unix seconds, as a string).
	Timestamp string `json:"timestamp"`
	// NonceStr used for the signature.
	NonceStr string `json:"nonceStr"`
	// Signature of the JS-SDK configuration.
	Signature string `json:"signature"`
}

// JSSDKSignature computes the JS-SDK signature for a page URL.
//
// WeChat's algorithm is SHA-1 of the exact fixed-order string
// "jsapi_ticket=..&noncestr=..&timestamp=..&url=..": the keys are not sorted
// and values are not escaped. A URL fragment (#...) is stripped first, because
// WeChat signs the fragment-less page URL.
func JSSDKSignature(ticket, nonce, timestamp, pageURL string) string {
	if before, _, found := strings.Cut(pageURL, "#"); found {
		pageURL = before
	}
	s := "jsapi_ticket=" + ticket +
		"&noncestr=" + nonce +
		"&timestamp=" + timestamp +
		"&url=" + pageURL
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// NewJSSDKConfig builds a signed wx.config payload for pageURL from the
// account appID and its jsapi ticket. The timestamp and nonce are generated
// here, so the result is ready to hand to the page.
func NewJSSDKConfig(appID, ticket, pageURL string) (*JSSDKConfig, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce, err := RandomNonce(16)
	if err != nil {
		return nil, err
	}
	return &JSSDKConfig{
		AppID:     appID,
		Timestamp: timestamp,
		NonceStr:  nonce,
		Signature: JSSDKSignature(ticket, nonce, timestamp, pageURL),
	}, nil
}

// RandomNonce returns n crypto-random base62 characters, the alphabet WeChat
// uses for nonce strings. The bytes come from crypto/rand so a nonce is not
// guessable.
func RandomNonce(n int) (string, error) {
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
