package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
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

// VerifyURLSignature checks the "signature" of the GET URL verification
// request WeChat sends when a message-push URL is submitted.
//
// The signed set is exactly [token, timestamp, nonce] in dictionary order, as
// the documentation specifies ("将 Token、timestamp、nonce 三个参数进行字典序排序"):
//
//	https://developers.weixin.qq.com/miniprogram/dev/framework/server-ability/message-push.html
//
// The echostr does NOT participate (the doc's worked example, token "AAAAA",
// timestamp 1714036504, nonce 1514711492, yields sha1 of
// "15147114921714036504AAAAA"). On success the caller echoes echostr back
// verbatim; use VerifyBodySignature for the four-argument POST check.
func (c *MessageCrypto) VerifyURLSignature(signature, timestamp, nonce string) bool {
	if c.token == "" {
		return false
	}
	return signature == c.sign(timestamp, nonce)
}

// VerifyBodySignature checks the "msg_signature" of an encrypted push request.
//
// The signed set is [token, timestamp, nonce, encrypt] in dictionary order,
// where encrypt is the base64 Encrypt field of the body. Note that the plain
// "signature" query parameter must NOT be used for encrypted pushes ("注意：不要
// 使用 signature 验证！").
func (c *MessageCrypto) VerifyBodySignature(signature, timestamp, nonce, encrypt string) bool {
	if c.token == "" {
		return false
	}
	return signature == c.sign(timestamp, nonce, encrypt)
}

// VerifySignature checks the four-argument callback signature
// [token, timestamp, nonce, value] in dictionary order. It is retained for
// callers that predate the split below; it also accepts the documented
// three-argument GET handshake when value is empty, so a caller that passes ""
// for the echostr still verifies.
//
// Migration: an earlier version of this documentation claimed the echostr
// participated in the GET handshake, which is wrong — the handshake signs only
// token/timestamp/nonce. A caller that followed that comment and passes the
// real echostr here will therefore NOT verify; switch it to
// VerifyURLSignature(signature, timestamp, nonce).
//
// Deprecated: prefer VerifyURLSignature or VerifyBodySignature.
func (c *MessageCrypto) VerifySignature(signature, timestamp, nonce, value string) bool {
	if c.token == "" {
		return false
	}
	if signature == c.sign(timestamp, nonce, value) {
		return true
	}
	return value == "" && signature == c.sign(timestamp, nonce)
}

// sign returns the lowercase hex sha1 of the token plus the given values,
// sorted in dictionary order and concatenated.
func (c *MessageCrypto) sign(values ...string) string {
	parts := append([]string{c.token}, values...)
	slices.Sort(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(sum[:])
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
//
// WeChat's message-push encryption (安全模式) uses AES-256-CBC with a fixed IV
// derived from the AES key (iv = key[:16]) and does NOT prepend an IV to the
// ciphertext. The plaintext layout is:
//
//	random(16) + msg_len(4, big endian) + msg + receiveid(appid)
func (c *MessageCrypto) DecryptBody(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("wechat: decode encrypted body: %w", err)
	}
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", fmt.Errorf("wechat: init aes: %w", err)
	}
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return "", errors.New("wechat: invalid encrypted body length")
	}
	iv := c.aesKey[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	payload := make([]byte, len(data))
	mode.CryptBlocks(payload, data)
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
//
// WeChat specifies the block size as K = 32, the AES key length, not the AES
// block size ("PKCS#7：K 为秘钥字节数（采用 32）"), so the pad byte ranges
// 1..32 and roughly half of all real messages carry a value above 16. Its own
// worked example pads a 205-byte plaintext by 19 bytes of 0x13 to reach 224.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > aesKeySize || padLen > len(data) {
		return nil, fmt.Errorf("invalid padding length %d", padLen)
	}
	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, errors.New("invalid padding bytes")
		}
	}
	return data[:len(data)-padLen], nil
}

// aesKeySize is the AES key length of the message-push scheme. WeChat uses it
// as the PKCS#7 block size K as well, so padding runs to a multiple of 32
// rather than the AES block size of 16.
const aesKeySize = 32

// EncryptReplyBody encrypts a reply body into the ciphertext WeChat expects from
// the message push receiver. It mirrors the DecryptBody scheme: AES-256-CBC
// with iv = key[:16], plaintext random(16)+len(4)+msg+appid, and no IV prefix on
// the wire.
func (c *MessageCrypto) EncryptReplyBody(plainText string) (string, error) {
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", err
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	msgLen := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLen, uint32(len(plainText)))
	raw := append(random, msgLen...)
	raw = append(raw, []byte(plainText)...)
	raw = append(raw, []byte(c.appID)...)
	raw = pkcs7Pad(raw, aesKeySize)
	iv := c.aesKey[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(raw))
	mode.CryptBlocks(out, raw)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Reply is the four-field envelope WeChat requires for an encrypted reply: the
// platform verifies MsgSignature, so the ciphertext alone is not a valid reply.
//
// Reference:
// https://developers.weixin.qq.com/miniprogram/dev/framework/server-ability/message-push.html
type Reply struct {
	// Encrypt is the base64 AES-256-CBC ciphertext.
	Encrypt string `json:"Encrypt" xml:"Encrypt"`
	// MsgSignature authenticates the reply: sha1 over the sorted
	// token/timestamp/nonce/Encrypt values.
	MsgSignature string `json:"MsgSignature" xml:"MsgSignature"`
	// TimeStamp is the reply timestamp (Unix seconds).
	TimeStamp string `json:"TimeStamp" xml:"TimeStamp"`
	// Nonce is a random string; the doc's example echoes the request nonce.
	Nonce string `json:"Nonce" xml:"Nonce"`
}

// BuildReply encrypts plainText and returns the complete signed reply envelope,
// generating the timestamp and nonce. The data format of the reply (XML or JSON)
// follows the format configured for the account's message push; use XML or JSON
// to render it.
//
// Callers that must echo the request's nonce can set Reply.Nonce and recompute
// MsgSignature with SignReply.
func (c *MessageCrypto) BuildReply(plainText string) (*Reply, error) {
	encrypted, err := c.EncryptReplyBody(plainText)
	if err != nil {
		return nil, err
	}
	nonce, err := RandomNonce(16)
	if err != nil {
		return nil, err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	return &Reply{
		Encrypt:      encrypted,
		MsgSignature: c.SignReply(ts, nonce, encrypted),
		TimeStamp:    ts,
		Nonce:        nonce,
	}, nil
}

// XML renders the reply in the XML envelope:
//
//	<xml><Encrypt>..</Encrypt><MsgSignature>..</MsgSignature>
//	<TimeStamp>..</TimeStamp><Nonce>..</Nonce></xml>
func (r *Reply) XML() string {
	var sb strings.Builder
	sb.WriteString("<xml>")
	for _, f := range []struct{ tag, value string }{
		{"Encrypt", r.Encrypt},
		{"MsgSignature", r.MsgSignature},
		{"TimeStamp", r.TimeStamp},
		{"Nonce", r.Nonce},
	} {
		sb.WriteString("<")
		sb.WriteString(f.tag)
		sb.WriteString("><![CDATA[")
		sb.WriteString(f.value)
		sb.WriteString("]]></")
		sb.WriteString(f.tag)
		sb.WriteString(">")
	}
	sb.WriteString("</xml>")
	return sb.String()
}

// JSON renders the reply in the JSON envelope, for accounts configured with the
// JSON data format.
func (r *Reply) JSON() ([]byte, error) {
	buf, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("wechat: marshal reply: %w", err)
	}
	return buf, nil
}

// EncryptedReplyEnvelope encrypts plainXML and returns the complete signed XML
// reply envelope (see Reply for the four fields). Use BuildReply when the reply
// must reuse the request's nonce or be rendered as JSON.
func (c *MessageCrypto) EncryptedReplyEnvelope(plainXML string) (string, error) {
	reply, err := c.BuildReply(plainXML)
	if err != nil {
		return "", err
	}
	return reply.XML(), nil
}

// SignReply computes the msg_signature for a reply message: the lowercase hex
// sha1 of token, timestamp, nonce and the encrypted body in dictionary order.
func (c *MessageCrypto) SignReply(timestamp, nonce, encryptedBody string) string {
	parts := []string{c.token, timestamp, nonce, encryptedBody}
	slices.Sort(parts)
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
	return append(data, bytes.Repeat([]byte{byte(padLen)}, padLen)...)
}
