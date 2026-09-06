package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"testing"
)

// The following helpers replicate WeChat's official message-push encryption
// (安全模式) as implemented by the WXBizMsgCrypt reference sample:
// AES-256-CBC with a fixed iv derived from the key (iv = key[:16]) and no IV
// prefix on the wire; plaintext = random(16) + uint32BE(len(msg)) + msg +
// receiveid(appid). They are deliberately independent of the production code so
// the tests below are a true cross-implementation check.

func wechatTestKey() (string, []byte) {
	// 43-char EncodingAESKey (base64 of a 32-byte key once "=" is appended).
	const keyB64 = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"
	key, err := base64.StdEncoding.DecodeString(keyB64 + "=")
	if err != nil {
		panic(err)
	}
	return keyB64, key
}

func wechatEncrypt(aesKey []byte, msg, appID string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	msgLen := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLen, uint32(len(msg)))
	raw := append(random, msgLen...)
	raw = append(raw, msg...)
	raw = append(raw, appID...)
	raw = pkcs7Pad(raw, aes.BlockSize)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(raw))
	mode.CryptBlocks(out, raw)
	return base64.StdEncoding.EncodeToString(out), nil
}

func wechatDecrypt(aesKey []byte, encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	plain := make([]byte, len(data))
	mode.CryptBlocks(plain, data)
	plain, err = pkcs7Unpad(plain)
	if err != nil {
		return "", err
	}
	if len(plain) < 20 {
		return "", errors.New("wechat: decrypted body too short")
	}
	msgLen := binary.BigEndian.Uint32(plain[16:20])
	return string(plain[20 : 20+msgLen]), nil
}

// TestMessageCryptoDecryptsOfficialSchemeCiphertext checks that DecryptBody can
// decrypt a payload produced by the official WXBizMsgCrypt scheme (regression
// guard for RF-002: the old code treated the first 16 ciphertext bytes as an IV
// and failed on every real push).
func TestMessageCryptoDecryptsOfficialSchemeCiphertext(t *testing.T) {
	keyB64, aesKey := wechatTestKey()
	mc, err := NewMessageCrypto("token", keyB64, "wx1234567890abcdef")
	if err != nil {
		t.Fatal(err)
	}
	msg := `<xml><ToUserName><![CDATA[wx1234567890abcdef]]></ToUserName><MsgType><![CDATA[event]]></MsgType><Event><![CDATA[subscribe_msg_send_result]]></Event></xml>`
	blob, err := wechatEncrypt(aesKey, msg, "wx1234567890abcdef")
	if err != nil {
		t.Fatal(err)
	}
	got, err := mc.DecryptBody(blob)
	if err != nil {
		t.Fatalf("DecryptBody failed against official-scheme ciphertext: %v", err)
	}
	if got != msg {
		t.Fatalf("decrypt mismatch:\n got %q\nwant %q", got, msg)
	}
}

// TestMessageCryptoDecryptRoundTrip encrypts with the production code and
// decrypts with the official-scheme helper (and vice versa), proving both
// directions agree with the WeChat algorithm.
func TestMessageCryptoDecryptRoundTrip(t *testing.T) {
	keyB64, aesKey := wechatTestKey()
	mc, err := NewMessageCrypto("token", keyB64, "wx1234567890abcdef")
	if err != nil {
		t.Fatal(err)
	}
	reply := `<xml><Content><![CDATA[hello]]></Content></xml>`

	// SDK-encrypted reply must be readable by the official-scheme decryptor.
	envelope, err := mc.EncryptedReplyEnvelope(reply)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Encrypt string `xml:"Encrypt"`
	}
	if err := xml.Unmarshal([]byte(envelope), &env); err != nil {
		t.Fatalf("parse reply envelope: %v", err)
	}
	got, err := wechatDecrypt(aesKey, env.Encrypt)
	if err != nil {
		t.Fatalf("official-scheme decrypt of EncryptReplyBody failed: %v", err)
	}
	if got != reply {
		t.Fatalf("reply mismatch:\n got %q\nwant %q", got, reply)
	}
}
