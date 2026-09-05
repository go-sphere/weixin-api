package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// MessageCrypto implements the WeChat Mini Program message-push signature
// verification and AES-CBC body (de)encryption. It is used by the server that
// receives the callback pushes (access token change, media check result,
// subscribe message feedback...).
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/framework/server-ability/message-push.html
type MessageCrypto struct {
	// token is the message push Token configured in the Mini Program console.
	token string
	// encodingAESKey is the message push EncodingAESKey (43 chars) configured
	// in the console; it is used both for signature verification and for the
	// AES-256-CBC payload encryption.
	encodingAESKey string
	// appID is the Mini Program appid, used to validate the decrypted body.
	appID string
	// aesKey is the derived 32 byte AES key.
	aesKey []byte
}

// NewMessageCrypto builds the crypto helper for the message push callback.
//
// Reference: https://developers.weixin.qq.com/miniprogram/dev/framework/server-ability/message-push.html
func NewMessageCrypto(token, encodingAESKey, appID string) (*MessageCrypto, error) {
	if len(encodingAESKey) != 43 {
		return nil, fmt.Errorf("wechat: encoding aes key must be 43 characters, got %d", len(encodingAESKey))
	}
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("wechat: decode encoding aes key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("wechat: invalid aes key length %d", len(key))
	}
	return &MessageCrypto{token: token, encodingAESKey: encodingAESKey, appID: appID, aesKey: key}, nil
}

// VerifySignature checks the msg_signature of a GET URL verification request.
// Params are the raw query values received.
//
// The signature is sha1 of the sorted list:
// [token, timestamp, nonce, the "echostr" (URL verify) or the encrypted body
// (message) ] joined as a single string. For GET verification the echostr
// participates; the caller passes it as echostr.
func (c *MessageCrypto) VerifySignature(signature, timestamp, nonce, echostr string) bool {
	if c.token == "" {
		return false
	}
	parts := []string{c.token, timestamp, nonce, echostr}
	sort.Strings(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return signature == hex.EncodeToString(sum[:])
}

// EncryptedPushBody is the XML envelope WeChat sends for encrypted pushes.
type EncryptedPushBody struct {
	// ToUserName echoes the receiver appid.
	ToUserName string `xml:"ToUserName"`
	// Encrypt is the base64 AES-256-CBC encrypted payload.
	Encrypt string `xml:"Encrypt"`
	// AgentID is present on some platforms and can be ignored.
	AgentID string `xml:"AgentID,omitempty"`
}

// MessagePushBody is the decrypted plaintext push payload. The concrete event
// content lives in the Event-typed sub-structures; parse further with the
// standard encoding/xml when needed.
type MessagePushBody struct {
	// ToUserName of the Mini Program receiving the push.
	ToUserName string `xml:"ToUserName"`
	// FromUserName of the source.
	FromUserName string `xml:"FromUserName"`
	// CreateTime of the push (Unix seconds).
	CreateTime int64 `xml:"CreateTime"`
	// MsgType of the message.
	MsgType string `xml:"MsgType"`
	// Event of the push (for event messages).
	Event string `xml:"Event"`
	// Content of text messages.
	Content string `xml:"Content"`
	// Raw is the full decrypted XML.
	Raw string
}

// DecryptBody decrypts an encrypted message push body and returns the
// plaintext XML.
func (c *MessageCrypto) DecryptBody(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("wechat: decode encrypted body: %w", err)
	}
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", fmt.Errorf("wechat: init aes: %w", err)
	}
	if len(data) < aes.BlockSize || len(data)%aes.BlockSize != 0 {
		return "", errors.New("wechat: invalid encrypted body length")
	}
	iv := data[:aes.BlockSize]
	payload := data[aes.BlockSize:]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(payload, payload)
	payload, err = pkcs7Unpad(payload)
	if err != nil {
		return "", fmt.Errorf("wechat: unpad decrypted body: %w", err)
	}
	// Layout: random(16) + msgLen(4, big endian) + msg + receiveid(appid)
	if len(payload) < 20 {
		return "", errors.New("wechat: decrypted body too short")
	}
	msgLen := binary.BigEndian.Uint32(payload[16:20])
	if int(msgLen) > len(payload)-20 {
		return "", errors.New("wechat: decrypted message length overflow")
	}
	msg := payload[20 : 20+msgLen]
	appID := string(payload[20+msgLen:])
	if c.appID != "" && appID != c.appID {
		return "", fmt.Errorf("wechat: appid mismatch in push body: got %q", appID)
	}
	return string(msg), nil
}

// pkcs7Unpad removes the PKCS#7 padding from a plaintext block.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > aes.BlockSize || padLen > len(data) {
		return nil, fmt.Errorf("invalid padding length %d", padLen)
	}
	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, errors.New("invalid padding bytes")
		}
	}
	return data[:len(data)-padLen], nil
}

// EncryptReplyBody encrypts a reply XML into the <Encrypt> envelope WeChat
// expects from the message push receiver.
func (c *MessageCrypto) EncryptReplyBody(plainXML string) (string, error) {
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", err
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	msgLen := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLen, uint32(len(plainXML)))
	raw := append(random, msgLen...)
	raw = append(raw, []byte(plainXML)...)
	raw = append(raw, []byte(c.appID)...)
	raw = pkcs7Pad(raw, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(raw))
	mode.CryptBlocks(out, raw)
	encrypted := append(iv, out...)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// EncryptedReplyEnvelope wraps an encrypted reply into the XML <xml><Encrypt>
// envelope required by WeChat.
func (c *MessageCrypto) EncryptedReplyEnvelope(plainXML string) (string, error) {
	encrypted, err := c.EncryptReplyBody(plainXML)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("<xml><Encrypt><![CDATA[%s]]></Encrypt></xml>", encrypted), nil
}

// SignReply computes the msg_signature for a reply message.
func (c *MessageCrypto) SignReply(timestamp, nonce, encryptedBody string) string {
	parts := []string{c.token, timestamp, nonce, encryptedBody}
	sort.Strings(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(sum[:])
}

// DecodeEncryptedPushXML parses an <xml> envelope, decrypts it and returns the
// plaintext message body.
func (c *MessageCrypto) DecodeEncryptedPushXML(raw []byte) (*MessagePushBody, error) {
	var env EncryptedPushBody
	if err := xml.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("wechat: parse push xml: %w", err)
	}
	if env.Encrypt == "" {
		return nil, errors.New("wechat: push xml has no Encrypt node")
	}
	plain, err := c.DecryptBody(env.Encrypt)
	if err != nil {
		return nil, err
	}
	var body MessagePushBody
	if err := xml.Unmarshal([]byte(plain), &body); err != nil {
		return nil, fmt.Errorf("wechat: parse decrypted push: %w", err)
	}
	body.Raw = plain
	return &body, nil
}

// pkcs7Pad pads data to a multiple of blockSize using PKCS#7.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	pad := bytesRepeat(byte(padLen), padLen)
	return append(data, pad...)
}

// bytesRepeat builds a byte slice filled with b repeated n times.
func bytesRepeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}
