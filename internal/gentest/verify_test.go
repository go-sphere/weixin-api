package gentest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sphere/weixin-api/core"
	"github.com/go-sphere/weixin-api/miniprogram"
	"github.com/go-sphere/weixin-api/official"
)

// RF-001: the injected access token must survive request serialization.
func TestAccessTokenNotOverwritten(t *testing.T) {
	var q string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"REALTOKEN","expires_in":7200}`)
			return
		}
		q = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"errcode":0,"openid":"o"}`)
	}))
	t.Cleanup(srv.Close)
	c := official.New(official.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
	if _, err := c.GetCgiBinUserInfo(t.Context(), &official.GetCgiBinUserInfoRequest{OpenID: "OPENID"}); err != nil {
		t.Fatal(err)
	}
	t.Logf("query: %s", q)
	if !strings.Contains(q, "access_token=REALTOKEN") {
		t.Errorf("access_token not injected correctly: %s", q)
	}
}

// RF-006: signed XPay endpoints compute pay_sig from Config.AppKey.
func TestXPayPaySigComputed(t *testing.T) {
	var q, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"T","expires_in":7200}`)
			return
		}
		b, _ := io.ReadAll(r.Body)
		q, body = r.URL.RawQuery, string(b)
		_, _ = io.WriteString(w, `{"errcode":0}`)
	}))
	t.Cleanup(srv.Close)
	c := miniprogram.New(miniprogram.Config{AppID: "a", AppSecret: "b", AppKey: "APPKEY", BaseURL: srv.URL})
	_, err := c.PostXpayQueryOrder(t.Context(), &miniprogram.PostXpayQueryOrderRequest{})
	if err != nil {
		t.Fatalf("xpay: %v", err)
	}
	t.Logf("pay_sig query=%s body=%s", q, body)
	if !strings.Contains(q, "pay_sig=") {
		t.Errorf("pay_sig missing: %s", q)
	}
	// pay_sig = HMAC-SHA256(AppKey, path + "&" + rawBody)
	if want := hmachex("APPKEY", "/xpay/query_order&"+body); !strings.Contains(q, want) {
		t.Errorf("pay_sig mismatch\n got %s\nwant %s", q, want)
	}
	// access_token must still be present alongside the signature
	if !strings.Contains(q, "access_token=T") {
		t.Errorf("token lost when signing: %s", q)
	}
}

// RF-006: signed endpoints fail fast without AppKey instead of sending junk.
func TestXPayRequiresAppKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"T","expires_in":7200}`)
			return
		}
		_, _ = io.WriteString(w, `{"errcode":0}`)
	}))
	t.Cleanup(srv.Close)
	c := miniprogram.New(miniprogram.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
	if _, err := c.PostXpayQueryOrder(t.Context(), &miniprogram.PostXpayQueryOrderRequest{}); err == nil {
		t.Fatal("expected an error when AppKey is empty")
	}
}

// RF-006: user-level signed endpoints require a SessionKey.
func TestXPaySessionKeyRequired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"T","expires_in":7200}`)
			return
		}
		_, _ = io.WriteString(w, `{"errcode":0}`)
	}))
	t.Cleanup(srv.Close)
	c := miniprogram.New(miniprogram.Config{AppID: "a", AppSecret: "b", AppKey: "K", BaseURL: srv.URL})
	if _, err := c.PostXpayQueryUserBalance(t.Context(), &miniprogram.PostXpayQueryUserBalanceRequest{}); err == nil {
		t.Fatal("expected an error without SessionKey")
	}
}

// RF-003: numeric arrays deserialize as numbers, not strings.
func TestNumericArrayDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"T","expires_in":7200}`)
			return
		}
		_, _ = io.WriteString(w, `{"errcode":0,"tagid_list":[2,100]}`)
	}))
	t.Cleanup(srv.Close)
	c := official.New(official.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
	resp, err := c.PostCgiBinTagsGetidlist(t.Context(), &official.PostCgiBinTagsGetidlistRequest{})
	if err != nil {
		t.Fatalf("getidlist: %v", err)
	}
	if len(resp.TagidList) != 2 || resp.TagidList[0] != 2 || resp.TagidList[1] != 100 {
		t.Fatalf("tagid_list = %#v", resp.TagidList)
	}
}

// RF-006: subscribe messages default miniprogram_state from Config.Env.
func TestSubscribeMessageEnvDefault(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"T","expires_in":7200}`)
			return
		}
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		_, _ = io.WriteString(w, `{"errcode":0}`)
	}))
	t.Cleanup(srv.Close)
	c := miniprogram.New(miniprogram.Config{AppID: "a", AppSecret: "b", Env: core.MiniAppEnvTrial, BaseURL: srv.URL})
	_, err := c.PostCgiBinMessageSubscribeSend(t.Context(), &miniprogram.PostCgiBinMessageSubscribeSendRequest{
		Touser: "o", TemplateID: "t",
	})
	if err != nil {
		t.Fatalf("subscribe send: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	if got["miniprogram_state"] != "trial" {
		t.Errorf("miniprogram_state = %v, body=%s", got["miniprogram_state"], body)
	}
}

func hmachex(key, msg string) string {
	return hexhmac([]byte(key), []byte(msg))
}
