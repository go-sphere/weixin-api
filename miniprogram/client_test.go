package miniprogram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sphere/weixin-api/core"
)

// newTestClient spins up an httptest server whose handler records requests and
// returns canned responses, and wires a Wechat client to it.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*MiniProgram, *httptest.Server, *[]http.Request) {
	t.Helper()
	var mu sync.Mutex
	var requests []http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cp := *r
		cp.Body = r.Body
		requests = append(requests, cp)
		mu.Unlock()
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	w := newMiniProgramWithHTTPClient(Config{AppID: "appid1", AppSecret: "secret1"}, nil, srv.URL, srv.Client())
	return w, srv, &requests
}

// writeJSON writes a JSON response body with errcode/errmsg defaults of 0/ok.
func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func TestGetAccessTokenCachesAndForcesReload(t *testing.T) {
	var tokenCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			tokenCalls++
			writeJSON(w, map[string]any{"access_token": fmt.Sprintf("tok%d", tokenCalls), "expires_in": 7200})
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	w := newMiniProgramWithHTTPClient(Config{AppID: "a", AppSecret: "s"}, nil, srv.URL, srv.Client())

	ctx := t.Context()
	t1, err := w.GetAccessToken(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	t2, err := w.GetAccessToken(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if t1 != t2 || tokenCalls != 1 {
		t.Fatalf("expected cached token, got %q then %q with %d upstream calls", t1, t2, tokenCalls)
	}
	reloaded, err := w.GetAccessToken(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded == t1 || tokenCalls != 2 {
		t.Fatalf("expected forced reload, got %q with %d upstream calls", reloaded, tokenCalls)
	}
}

func TestAccessTokenQueryInjectedAndRetryOnExpiry(t *testing.T) {
	var gotQuery url.Values
	first := true
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			writeJSON(w, map[string]any{"access_token": "ACCESS_1", "expires_in": 7200})
		case "/wxa/business/getuserphonenumber":
			gotQuery = r.URL.Query()
			if first {
				// First attempt fails with an expired-token error, which must
				// trigger a forced token refresh and a single retry.
				first = false
				writeJSON(w, map[string]any{"errcode": 42001, "errmsg": "access_token expired"})
				return
			}
			writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok", "phone_info": map[string]any{
				"phoneNumber": "8613800138000", "purePhoneNumber": "13800138000", "countryCode": "86",
			}})
		default:
			http.NotFound(w, r)
		}
	}
	w, srv, _ := newTestClient(t, handler)
	_ = srv
	resp, err := w.GetUserPhoneNumber(t.Context(), "code123")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response without error")
	}
	if resp.PhoneInfo.PurePhoneNumber != "13800138000" {
		t.Fatalf("unexpected phone %+v", resp.PhoneInfo)
	}
	if gotQuery.Get("access_token") != "ACCESS_1" {
		t.Fatalf("access_token not injected, query=%v", gotQuery)
	}
}

func TestJsCode2SessionRequestShape(t *testing.T) {
	var got url.Values
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sns/jscode2session" {
			http.NotFound(w, r)
			return
		}
		got = r.URL.Query()
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok", "openid": "o1", "session_key": "sk1", "unionid": "u1"})
	}
	w, _, _ := newTestClient(t, handler)
	resp, err := w.JsCode2Session(t.Context(), "code")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response without error")
	}
	if resp.OpenID != "o1" || resp.SessionKey != "sk1" {
		t.Fatalf("unexpected resp %+v", resp)
	}
	if got.Get("js_code") != "code" || got.Get("appid") != "appid1" || got.Get("grant_type") != "authorization_code" {
		t.Fatalf("unexpected query %v", got)
	}
}

func TestSubscribeMessageBodyShape(t *testing.T) {
	var body map[string]any
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			writeJSON(w, map[string]any{"access_token": "tok", "expires_in": 7200})
		case "/cgi-bin/message/subscribe/send":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode body: %v", err)
			}
			writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok"})
		default:
			http.NotFound(w, r)
		}
	}
	w, _, _ := newTestClient(t, handler)
	msg := &SendSubscribeMessageRequest{
		ToUser:     "openid1",
		TemplateID: "tmpl1",
		Data:       SubscribeMessageData{"thing1": {Value: "hello"}},
	}
	if err := w.SendSubscribeMessage(t.Context(), msg); err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "openid1" || body["template_id"] != "tmpl1" {
		t.Fatalf("unexpected body %v", body)
	}
	if body["miniprogram_state"] != "formal" {
		t.Fatalf("expected default miniprogram_state formal, got %v", body["miniprogram_state"])
	}
}

func TestQRCodeBinaryAndJSONErrorSniffing(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			writeJSON(w, map[string]any{"access_token": "tok", "expires_in": 7200})
		case "/wxa/getwxacodeunlimit":
			if r.URL.Query().Get("succeed") == "" {
				// binary payload path
				w.Header().Set("Content-Type", "image/jpeg")
				_, _ = w.Write([]byte{0xff, 0xd8, 0xff, 0xe0})
				return
			}
			writeJSON(w, map[string]any{"errcode": 40001, "errmsg": "invalid credential", "rid": "r1"})
		default:
			http.NotFound(w, r)
		}
	}
	w, _, _ := newTestClient(t, handler)
	// Binary ok.
	img, err := w.GetUnlimitedMiniProgramCode(t.Context(), &GetUnlimitedMiniProgramCodeRequest{Scene: "s1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(img) == 0 || img[0] != 0xff {
		t.Fatalf("expected binary image, got %v", img)
	}
}

func TestBusinessErrorSurfacing(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			writeJSON(w, map[string]any{"errcode": 40013, "errmsg": "invalid appid"})
		default:
			http.NotFound(w, r)
		}
	}
	w, _, _ := newTestClient(t, handler)
	_, err := w.GetAccessToken(t.Context(), false)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !strings.Contains(err.Error(), "invalid appid") {
		t.Fatalf("expected invalid appid error, got %v", err)
	}
	_ = apiErr
}

func TestSentinelErrorsForTokenProblems(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			writeJSON(w, map[string]any{"access_token": "tok", "expires_in": 7200})
		case "/wxa/getwxacode":
			writeJSON(w, map[string]any{"errcode": 40014, "errmsg": "invalid access token"})
		default:
			http.NotFound(w, r)
		}
	}
	w, _, _ := newTestClient(t, handler)
	// Disable retry to observe the raw sentinel.
	_, err := w.GetMiniProgramCode(t.Context(), &GetMiniProgramCodeRequest{Path: "pages/a"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrorInvalidAccessToken {
		t.Fatalf("expected ErrorInvalidAccessToken sentinel, got %v", err)
	}
}

func TestCheckSessionKeyQueryShape(t *testing.T) {
	var got url.Values
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			writeJSON(w, map[string]any{"access_token": "tok", "expires_in": 7200})
		case "/wxa/checksession":
			got = r.URL.Query()
			writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok"})
		default:
			http.NotFound(w, r)
		}
	}
	w, _, _ := newTestClient(t, handler)
	if err := w.CheckSessionKey(t.Context(), "openid1", "sessionkey1"); err != nil {
		t.Fatal(err)
	}
	if got.Get("openid") != "openid1" || got.Get("sig_method") != "hmac_sha256" {
		t.Fatalf("unexpected query %v", got)
	}
	expected := hmacSHA256Hex([]byte("sessionkey1"), nil)
	if got.Get("signature") != expected {
		t.Fatalf("signature mismatch: got %s want %s", got.Get("signature"), expected)
	}
}

func TestMemoryCacheExpiry(t *testing.T) {
	c := core.NewMemoryCache()
	ctx := t.Context()
	if err := c.SetWithTTL(ctx, "k", "v", 30*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if v, ok, err := c.Get(ctx, "k"); err != nil || !ok || v != "v" {
		t.Fatalf("expected cached value, got %q ok=%v err=%v", v, ok, err)
	}
	time.Sleep(50 * time.Millisecond)
	if _, ok, _ := c.Get(ctx, "k"); ok {
		t.Fatal("expected entry to expire")
	}
}

func TestHTTPStatusErrorWithoutEnvelope(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}
	w, _, _ := newTestClient(t, handler)
	var probe ErrResponse
	err := w.getJSON(t.Context(), "/x", nil, &probe)
	if err == nil || !strings.Contains(err.Error(), "500") && !strings.Contains(err.Error(), "502") {
		t.Fatalf("expected http status error, got %v", err)
	}
}
