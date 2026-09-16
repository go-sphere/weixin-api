package gentest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func hexhmac(key, msg []byte) string {
	m := hmac.New(sha256.New, key)
	m.Write(msg)
	return hex.EncodeToString(m.Sum(nil))
}
