package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			writeJSON(w, map[string]any{"access_token": "TOKEN", "expires_in": 7200})
			return
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return NewClient(CredentialsOf(Credentials{AppID: "appid", AppSecret: "secret"}), ClientOptions{BaseURL: srv.URL, HTTPClient: srv.Client()})
}

func TestGetAccessTokenCachedAndReload(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		writeJSON(w, map[string]any{"access_token": fmt.Sprintf("tok%d", calls), "expires_in": 7200})
	}))
	t.Cleanup(srv.Close)
	c := NewClient(CredentialsOf(Credentials{AppID: "a", AppSecret: "s"}), ClientOptions{BaseURL: srv.URL, HTTPClient: srv.Client()})
	ctx := t.Context()
	t1, err := c.GetAccessToken(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	t2, _ := c.GetAccessToken(ctx, false)
	if t1 != t2 || calls != 1 {
		t.Fatalf("expected cache hit, got %q %q calls=%d", t1, t2, calls)
	}
	t3, _ := c.GetAccessToken(ctx, true)
	if t3 == t1 || calls != 2 {
		t.Fatalf("expected forced reload calls=%d", calls)
	}
}

func TestErrEnvelopeSurfacesSentinel(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"errcode": 42001, "errmsg": "access_token expired"})
	}
	c := newTestClient(t, handler)
	var dst struct{ ErrResponse }
	err := c.GetJSON(t.Context(), "/cgi-bin/x", nil, &dst)
	if err != ErrorAccessTokenExpired {
		t.Fatalf("expected ErrorAccessTokenExpired, got %v", err)
	}
}

func TestAPIErrorCarriesRID(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"errcode": 40003, "errmsg": "invalid openid", "rid": "r-123"})
	}
	c := newTestClient(t, handler)
	err := c.GetJSON(t.Context(), "/x", nil, &struct{ ErrResponse }{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T %v", err, err)
	}
	if apiErr == nil || apiErr.RID != "r-123" {
		t.Fatalf("expected rid, got %v", apiErr)
	}
}

func TestMemoryCacheTTL(t *testing.T) {
	c := NewMemoryCache()
	ctx := t.Context()
	if err := c.SetWithTTL(ctx, "k", "v", 20*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if v, ok, _ := c.Get(ctx, "k"); !ok || v != "v" {
		t.Fatalf("expected hit")
	}
	time.Sleep(30 * time.Millisecond)
	if _, ok, _ := c.Get(ctx, "k"); ok {
		t.Fatal("expected expiry")
	}
}

func TestTokenRetryAfterExpiry(t *testing.T) {
	var calls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			writeJSON(w, map[string]any{"errcode": 42001, "errmsg": "expired"})
			return
		}
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok"})
	}
	c := newTestClient(t, handler)
	err := c.WithTokenPost(t.Context(), "/cgi-bin/foo", nil, DefaultRequestOptions(), map[string]string{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected a retry, got %d calls", calls)
	}
}
