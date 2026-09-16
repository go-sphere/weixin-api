package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := Config{
		MPAppID:       "wx-mp-test",
		MPAppSecret:   "mp-secret",
		OAAppID:       "wx-oa-test",
		OAAppSecret:   "oa-secret",
		OACallbackURL: "https://example.com/api/oa/callback",
	}
	store := NewStore()
	mpAPI := newMiniProgramAPI(cfg, store)
	oaAPI := newOfficialAccountAPI(cfg, store)

	r := gin.New()
	r.POST("/api/mp/login", mpAPI.login)
	r.GET("/api/oa/login", oaAPI.login)
	r.GET("/api/oa/callback", oaAPI.callback)
	return r
}

func TestMPLoginRejectsMissingCode(t *testing.T) {
	r := testRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/mp/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body=%s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestOALoginBuildsAuthorizeURL(t *testing.T) {
	r := testRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/oa/login?json=1&scope=snsapi_userinfo", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	var out struct {
		AuthorizeURL string `json:"authorize_url"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	u, err := url.Parse(out.AuthorizeURL)
	if err != nil {
		t.Fatalf("parse authorize_url: %v", err)
	}
	if u.Scheme != "https" || u.Host != "open.weixin.qq.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
	q := u.Query()
	if q.Get("appid") != "wx-oa-test" {
		t.Errorf("appid = %q", q.Get("appid"))
	}
	if q.Get("redirect_uri") != "https://example.com/api/oa/callback" {
		t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
	}
	if q.Get("scope") != "snsapi_userinfo" {
		t.Errorf("scope = %q", q.Get("scope"))
	}
	// gin's test recorder keeps Set-Cookie headers so the CSRF state is stored.
	if got := w.Header().Get("Set-Cookie"); !strings.Contains(got, "wx_oauth_state=") {
		t.Errorf("missing wx_oauth_state cookie, got %q", got)
	}
}

func TestOACallbackRejectsStateMismatch(t *testing.T) {
	r := testRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/oa/callback?code=abc&state=wrong", nil)
	req.Header.Set("Cookie", "wx_oauth_state=expected")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", w.Code, w.Body.String())
	}
}

func TestStoreUpsertAndToken(t *testing.T) {
	s := NewStore()

	// First sighting creates the account.
	acc := s.Upsert(ChannelMiniProgram, "openid-1", "union-1", "", "")
	if acc.Channel != ChannelMiniProgram || acc.OpenID != "openid-1" {
		t.Fatalf("unexpected account: %+v", acc)
	}

	// Same openid resolves to the same *Account.
	if again := s.Upsert(ChannelMiniProgram, "openid-1", "", "", ""); again != acc {
		t.Fatal("Upsert should return the stored account for the same key")
	}

	// A different openid sharing the unionid links to the existing account.
	linked := s.Upsert(ChannelOfficialAccount, "openid-2", "union-1", "", "")
	if linked != acc {
		t.Fatalf("openids of one unionid should share an account, got %+v and %+v", linked, acc)
	}

	token := s.IssueToken(acc)
	if got, ok := s.LookupToken(token); !ok || got != acc {
		t.Fatalf("LookupToken(%q) = %+v, %v; want the issued account", token, got, ok)
	}

	// Tampering must not authenticate.
	if _, ok := s.LookupToken(token[:len(token)-2] + "00"); ok {
		t.Fatal("tampered token authenticated")
	}
	if _, ok := s.LookupToken("garbage"); ok {
		t.Fatal("garbage token authenticated")
	}
}
