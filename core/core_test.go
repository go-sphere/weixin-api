package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
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
	c := newTestEndpointClient(t, handler)
	_, err := Call[map[string]any](c, t.Context(), http.MethodGet, "/cgi-bin/x", nil, nil, false, SignNone)
	if !errors.Is(err, ErrorAccessTokenExpired) {
		t.Fatalf("expected ErrorAccessTokenExpired, got %v", err)
	}
}

func TestAPIErrorCarriesRIDAndRaw(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"errcode": 40003, "errmsg": "invalid openid", "rid": "r-123"})
	}
	c := newTestEndpointClient(t, handler)
	_, err := Call[map[string]any](c, t.Context(), http.MethodGet, "/x", nil, nil, false, SignNone)
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T %v", err, err)
	}
	if apiErr.RID != "r-123" {
		t.Errorf("expected rid, got %q", apiErr.RID)
	}
	if len(apiErr.Raw) == 0 {
		t.Error("expected the raw response body to be retained")
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
	c := newTestEndpointClient(t, handler)
	if _, err := Call[map[string]any](c, t.Context(), http.MethodGet, "/cgi-bin/foo", nil, nil, true, SignNone); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected a retry, got %d calls", calls)
	}
}

// newTestEndpointClient builds an EndpointClient pointed at a test server,
// answering token requests with a fixed token.
func newTestEndpointClient(t *testing.T, handler http.HandlerFunc) *EndpointClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			writeJSON(w, map[string]any{"access_token": "TOKEN", "expires_in": 7200})
			return
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return NewEndpointClient(EndpointConfig{
		AppID:      "appid",
		AppSecret:  "secret",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
}

// TestPaySigSigning pins the XPay request signature to the documented formula
// pay_sig = HMAC-SHA256(AppKey, path + "&" + body).
func TestPaySigSigning(t *testing.T) {
	var query, body string
	c := newTestEndpointClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		query, body = r.URL.RawQuery, string(b)
		writeJSON(w, map[string]any{"errcode": 0})
	})
	c.appKey = "APPKEY"
	if _, err := Call[map[string]any](c, t.Context(), http.MethodPost, "/xpay/query_order",
		map[string]string{"openid": "o"}, nil, true, SignPaySig); err != nil {
		t.Fatal(err)
	}
	want := hmacSHA256Hex([]byte("APPKEY"), []byte("/xpay/query_order&"+body))
	if got := parseQueryValue(query, "pay_sig"); got != want {
		t.Errorf("pay_sig = %q, want %q", got, want)
	}
}

// TestPaySigRequiresAppKey checks a signed call fails fast without AppKey.
func TestPaySigRequiresAppKey(t *testing.T) {
	c := newTestEndpointClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"errcode": 0})
	})
	if _, err := Call[map[string]any](c, t.Context(), http.MethodPost, "/xpay/query_order",
		map[string]string{"openid": "o"}, nil, true, SignPaySig); err == nil {
		t.Fatal("expected an error when AppKey is empty")
	}
}

// TestOmitSemantics pins the tag-driven omission contract the generator relies
// on: an optional numeric field carries omitzero, so its zero value is left out
// of the body while a non-zero value is sent; a required field carries no
// option and is always sent, even at the zero value.
func TestOmitSemantics(t *testing.T) {
	type req struct {
		OptionalNum  int64  `json:"optional_num,omitzero"`
		OptionalText string `json:"optional_text,omitempty"`
		RequiredNum  int64  `json:"required_num"`
	}
	post := func(t *testing.T, body any) map[string]any {
		t.Helper()
		var got []byte
		c := newTestEndpointClient(t, func(w http.ResponseWriter, r *http.Request) {
			got, _ = io.ReadAll(r.Body)
			writeJSON(w, map[string]any{"errcode": 0})
		})
		if _, err := Call[map[string]any](c, t.Context(), http.MethodPost, "/x", body, nil, true, SignNone); err != nil {
			t.Fatal(err)
		}
		var out map[string]any
		if err := json.Unmarshal(got, &out); err != nil {
			t.Fatalf("decode body %s: %v", got, err)
		}
		return out
	}

	zero := post(t, req{})
	if _, present := zero["optional_num"]; present {
		t.Errorf("zero optional_num must be omitted, got %v", zero["optional_num"])
	}
	if _, present := zero["optional_text"]; present {
		t.Errorf("empty optional_text must be omitted, got %v", zero["optional_text"])
	}
	if _, present := zero["required_num"]; !present {
		t.Errorf("required_num must be sent even at zero, body=%v", zero)
	}

	set := post(t, req{OptionalNum: 7, OptionalText: "s", RequiredNum: 1})
	if set["optional_num"] != float64(7) {
		t.Errorf("optional_num = %v, want 7", set["optional_num"])
	}
	if set["optional_text"] != "s" {
		t.Errorf("optional_text = %v, want s", set["optional_text"])
	}
}

// TestCallerTokenAndPaySigAreSentAsWritten covers the runtime half of the
// ownership split: the generator omits access_token/pay_sig fields on endpoints
// where the runtime supplies them, so a field that IS present carries
// caller data (the /sns/ OAuth token, a B2B pay_sig) and must be sent verbatim.
func TestCallerTokenAndPaySigAreSentAsWritten(t *testing.T) {
	type req struct {
		AccessToken string `query:"access_token"`
		PaySig      string `query:"pay_sig"`
		OpenID      string `query:"openid"`
	}
	var query string
	c := newTestEndpointClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		writeJSON(w, map[string]any{"errcode": 0})
	})
	// no app token: this models the /sns/ family, which opts out of injection
	if _, err := Call[map[string]any](c, t.Context(), http.MethodGet, "/sns/userinfo",
		req{AccessToken: "USERTOKEN", PaySig: "SIG", OpenID: "o"}, nil, false, SignNone); err != nil {
		t.Fatal(err)
	}
	if got := parseQueryValue(query, "access_token"); got != "USERTOKEN" {
		t.Errorf("access_token = %q, want USERTOKEN (query=%s)", got, query)
	}
	if got := parseQueryValue(query, "pay_sig"); got != "SIG" {
		t.Errorf("pay_sig = %q, want SIG", got)
	}
}

// TestAppTokenInjectedWhenAuthRequired covers the other half: when the endpoint
// takes the application token, the runtime supplies it.
func TestAppTokenInjectedWhenAuthRequired(t *testing.T) {
	var query string
	c := newTestEndpointClient(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		writeJSON(w, map[string]any{"errcode": 0})
	})
	type req struct {
		OpenID string `query:"openid"`
	}
	if _, err := Call[map[string]any](c, t.Context(), http.MethodGet, "/x", req{OpenID: "o"}, nil, true, SignNone); err != nil {
		t.Fatal(err)
	}
	if got := parseQueryValue(query, "access_token"); got != "TOKEN" {
		t.Errorf("access_token = %q, want TOKEN (query=%s)", got, query)
	}
}

func parseQueryValue(raw, key string) string {
	v, err := url.ParseQuery(raw)
	if err != nil {
		return ""
	}
	return v.Get(key)
}
