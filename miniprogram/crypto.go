package miniprogram

import (
	"github.com/go-sphere/weixin-api/core"
)

// MessageCrypto is the WeChat message-push signature verification and AES-CBC
// body (de)encryption helper, shared with the official package through core.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/framework/server-ability/message-push.html
type MessageCrypto = core.MessageCrypto

// NewMessageCrypto builds a MessageCrypto for the message push callback.
func NewMessageCrypto(token, encodingAESKey, appID string) (*MessageCrypto, error) {
	return core.NewMessageCrypto(token, encodingAESKey, appID)
}
