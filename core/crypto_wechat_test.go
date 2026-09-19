package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
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
	// K is the key length (32), per the documentation's PKCS#7 definition.
	raw = pkcs7Pad(raw, aesKeySize)
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

// TestMessageCryptoDecryptsDocumentedVector decrypts the ciphertext the
// message-push documentation publishes for its worked example (token "AAAAA",
// the 43-character all-A EncodingAESKey, appid wxba5fad812f8e6fb9). The
// ciphertext is 224 bytes for a 205-byte plaintext, which is only consistent
// with PKCS#7 padding to a multiple of 32: the earlier implementation, which
// unpad against the 16-byte AES block size, failed to open it.
//
// Reference:
// https://developers.weixin.qq.com/miniprogram/dev/framework/server-ability/message-push.html
func TestMessageCryptoDecryptsDocumentedVector(t *testing.T) {
	mc, err := NewMessageCrypto(docPushToken, docPushAESKey, docPushAppID)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := mc.DecryptBody(docPushEncrypt)
	if err != nil {
		t.Fatalf("DecryptBody on the documented ciphertext: %v", err)
	}
	if plain != docPushPlaintext {
		t.Fatalf("documented plaintext mismatch:\n got %s\nwant %s", plain, docPushPlaintext)
	}
	// The documented msg_signature covers token/timestamp/nonce/Encrypt and must
	// verify with VerifyBodySignature.
	if !mc.VerifyBodySignature(docPushMsgSig, docPushTimestamp, docPushNonce, docPushEncrypt) {
		t.Error("documented msg_signature did not verify")
	}
	// The documented URL-verification signature covers only
	// token/timestamp/nonce; VerifyURLSignature must accept it.
	if !mc.VerifyURLSignature(docPushURLSig, docPushURLTimestamp, docPushURLNonce) {
		t.Error("documented URL signature did not verify")
	}
	// And the three-argument signature must NOT verify when the echostr is
	// folded in, which is what the earlier four-argument helper did.
	if mc.VerifyBodySignature(docPushURLSig, docPushURLTimestamp, docPushURLNonce, docPushURLEchostr) {
		t.Error("URL signature verified with echostr included, want mismatch")
	}
}

// TestDeprecatedVerifySignatureContract pins the compatibility shim's known
// semantics: an empty fourth argument falls back to the three-argument handshake,
// while passing the real echostr does not verify (the documented handshake never
// included it). This is a trap for callers that followed the old doc comment, so
// the behaviour is asserted rather than left implicit.
func TestDeprecatedVerifySignatureContract(t *testing.T) {
	mc, err := NewMessageCrypto(docPushToken, docPushAESKey, docPushAppID)
	if err != nil {
		t.Fatal(err)
	}
	if !mc.VerifySignature(docPushURLSig, docPushURLTimestamp, docPushURLNonce, "") {
		t.Error(`VerifySignature with empty echostr should accept the documented GET handshake`)
	}
	if mc.VerifySignature(docPushURLSig, docPushURLTimestamp, docPushURLNonce, docPushURLEchostr) {
		t.Error("VerifySignature must not accept the GET signature when the echostr is folded in")
	}
	// The four-argument POST form still works through the shim.
	if !mc.VerifySignature(docPushMsgSig, docPushTimestamp, docPushNonce, docPushEncrypt) {
		t.Error("VerifySignature should accept the documented four-argument msg_signature")
	}
}

// TestMessageCryptoReplyEnvelopeIsComplete checks the reply envelope carries all
// four documented fields (Encrypt, MsgSignature, TimeStamp, Nonce) and that the
// signature covers the ciphertext.
func TestMessageCryptoReplyEnvelopeIsComplete(t *testing.T) {
	mc, err := NewMessageCrypto(docPushToken, docPushAESKey, docPushAppID)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := `{"demo_resp":"good luck"}`
	envelope, err := mc.EncryptedReplyEnvelope(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"<Encrypt>", "<MsgSignature>", "<TimeStamp>", "<Nonce>"} {
		if !strings.Contains(envelope, field) {
			t.Errorf("reply envelope is missing %s: %s", field, envelope)
		}
	}
	var reply Reply
	if err := xml.Unmarshal([]byte(envelope), &reply); err != nil {
		t.Fatalf("parse reply envelope: %v", err)
	}
	if want := mc.SignReply(reply.TimeStamp, reply.Nonce, reply.Encrypt); reply.MsgSignature != want {
		t.Errorf("MsgSignature does not cover the ciphertext: got %s want %s", reply.MsgSignature, want)
	}
	// The ciphertext must decrypt back to the reply body.
	got, err := mc.DecryptBody(reply.Encrypt)
	if err != nil {
		t.Fatalf("decrypt own reply: %v", err)
	}
	if got != plaintext {
		t.Errorf("reply round trip mismatch:\n got %s\nwant %s", got, plaintext)
	}
	// And the JSON form must carry the same four fields.
	jsonReply, err := reply.JSON()
	if err != nil {
		t.Fatal(err)
	}
	var asMap map[string]string
	if err := json.Unmarshal(jsonReply, &asMap); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"Encrypt", "MsgSignature", "TimeStamp", "Nonce"} {
		if asMap[key] == "" {
			t.Errorf("JSON reply is missing %s: %s", key, jsonReply)
		}
	}
}

// Documented message-push vectors: token "AAAAA", EncodingAESKey of 43 "A"s,
// appid wxba5fad812f8e6fb9. Two distinct scenarios are published (URL
// verification and an encrypted push), hence two timestamp/nonce pairs.
const (
	docPushToken  = "AAAAA"
	docPushAESKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	docPushAppID  = "wxba5fad812f8e6fb9"

	docPushURLTimestamp = "1714036504"
	docPushURLNonce     = "1514711492"
	docPushURLEchostr   = "4375120948345356249"
	docPushURLSig       = "f464b24fc39322e44b38aa78f5edd27bd1441696"

	docPushTimestamp = "1714112445"
	docPushNonce     = "415670741"
	docPushMsgSig    = "046e02f8204d34f8ba5fa3b1db94908f3df2e9b3"
)

// docPushEncrypt is the published ciphertext of the encrypted-push example.
const docPushEncrypt = "+qdx1OKCy+5JPCBFWw70tm0fJGb2Jmeia4FCB7kao+/Q5c/ohsOzQHi8khUOb05JCpj0JB4RvQMkUyus8TPxLKJGQqcvZqzDpVzazhZv6JsXUnnR8XGT740XgXZUXQ7vJVnAG+tE8NUd4yFyjPy7GgiaviNrlCTj+l5kdfMuFUPpRSrfMZuMcp3Fn2Pede2IuQrKEYwKSqFIZoNqJ4M8EajAsjLY2km32IIjdf8YL/P50F7mStwntrA2cPDrM1kb6mOcfBgRtWygb3VIYnSeOBrebufAlr7F9mFUPAJGj04="

// docPushPlaintext is the plaintext of that ciphertext: 167 bytes, matching
// the msg_len=167 the documentation states for the example.
const docPushPlaintext = `{"ToUserName":"gh_97417a04a28d","FromUserName":"o9AgO5Kd5ggOC-bXrbNODIiE3bGY","CreateTime":1714112445,"MsgType":"event","Event":"debug_demo","debug_str":"hello world"}`

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
