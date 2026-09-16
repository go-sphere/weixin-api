package parity

import "github.com/go-sphere/weixin-api/core"

// core_JSSDKSignature is a thin alias so the parity tests read clearly and the
// shared-algorithm dependency is explicit.
func core_JSSDKSignature(ticket, nonce, timestamp, pageURL string) string {
	return core.JSSDKSignature(ticket, nonce, timestamp, pageURL)
}
