package core_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/go-sphere/weixin-api/core"
	"github.com/go-sphere/weixin-api/miniprogram"
)

// ExampleEndpointClient_AccessToken shows the managed access-token lifecycle:
// the first request fetches from /cgi-bin/token, later ones hit the cache, and
// reload forces a refresh.
func ExampleEndpointClient_AccessToken() {
	fetches := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		fetches++
		fmt.Fprintf(w, `{"access_token":"TOKEN-%d","expires_in":7200}`, fetches)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := core.NewEndpointClient(core.EndpointConfig{
		AppID:     "wx-app-id",
		AppSecret: "app-secret",
		BaseURL:   srv.URL, // tests only; empty means core.DefaultBaseURL
	})
	ctx := context.Background()
	for _, reload := range []bool{false, false, true} {
		token, err := client.AccessToken(ctx, reload)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(token)
	}
	// Output:
	// TOKEN-1
	// TOKEN-1
	// TOKEN-2
}

// ExampleNewMessageCrypto decrypts an encrypted message push after checking
// its msg_signature. The push is produced locally with BuildReply so the
// example is self-contained; a real handler reads it from the request.
func ExampleNewMessageCrypto() {
	const (
		token          = "push-token"
		encodingAESKey = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG" // 43 chars
		appID          = "wx-app-id"
	)
	mc, err := core.NewMessageCrypto(token, encodingAESKey, appID)
	if err != nil {
		log.Fatal(err)
	}

	// What WeChat would send: an encrypted XML body plus msg_signature,
	// timestamp and nonce query parameters.
	push, err := mc.BuildReply(`<xml><ToUserName>wx-app-id</ToUserName><MsgType>event</MsgType><Event>wxa_media_check</Event></xml>`)
	if err != nil {
		log.Fatal(err)
	}
	rawXML := []byte(push.XML())

	if !mc.VerifyBodySignature(push.MsgSignature, push.TimeStamp, push.Nonce, push.Encrypt) {
		log.Fatal("bad msg_signature")
	}
	body, err := mc.DecodeEncryptedPushXML(rawXML)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(body.MsgType, body.Event)
	// Output: event wxa_media_check
}

// ExampleJSSDKSignature signs a page URL for wx.config. The fragment is
// ignored, so both URLs produce the same signature.
func ExampleJSSDKSignature() {
	a := core.JSSDKSignature("ticket", "nonce", "1700000000", "https://example.com/page?x=1")
	b := core.JSSDKSignature("ticket", "nonce", "1700000000", "https://example.com/page?x=1#top")
	fmt.Println(len(a), a == b)
	// Output: 40 true
}

// ExampleNewAPISecurity installs the API 二次加密和签名 decorator on a Mini
// Program client for the endpoints listed in Paths. Real keys come from the
// MP backend; this example generates a throwaway RSA key.
func ExampleNewAPISecurity() {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}
	symKey := make([]byte, 32)
	if _, err := rand.Read(symKey); err != nil {
		log.Fatal(err)
	}

	sec, err := core.NewAPISecurity(core.APISecurityConfig{
		SymKey:     base64.StdEncoding.EncodeToString(symKey),
		SymSN:      "sym-sn",
		PrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
		// 平台证书编号 -> PEM, downloaded from the MP backend.
		PlatformCerts: map[string]string{},
		Paths:         []string{"/wxa/getuserriskrank"},
	})
	if err != nil {
		log.Fatal(err)
	}
	client := miniprogram.New(miniprogram.Config{
		AppID:     "wx-app-id",
		AppSecret: "app-secret",
		Modifiers: []core.RequestModifier{sec},
	})
	_ = client
	fmt.Println("api security installed")
	// Output: api security installed
}
