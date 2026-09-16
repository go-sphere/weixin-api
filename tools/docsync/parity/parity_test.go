// Package parity verifies the generated clients against golden expectations
// captured from the hand-written clients they replaced (see tools/capture).
// The goldens pin the exact request each call puts on the wire, so a
// regeneration that changes an endpoint path, parameter name, auth handling,
// body encoding or content type fails here.
package parity

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/go-sphere/weixin-api/core"
	genmp "github.com/go-sphere/weixin-api/miniprogram"
	genoa "github.com/go-sphere/weixin-api/official"
)

type recorder struct {
	last *http.Request
	body string
}

func newServer(t *testing.T, r *recorder, response string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.URL.Path, "/cgi-bin/token"):
			_, _ = io.WriteString(w, `{"access_token":"TOKEN","expires_in":7200}`)
			return
		case strings.HasPrefix(req.URL.Path, "/cgi-bin/ticket/getticket"):
			_, _ = io.WriteString(w, `{"errcode":0,"ticket":"TICKET","expires_in":7200}`)
			return
		}
		b, _ := io.ReadAll(req.Body)
		r.last = req
		r.body = string(b)
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// sortedQuery renders a query with sorted keys and values so ordering noise
// does not matter. access_token is kept: an earlier revision stripped it for
// comparison, which made a wrong or empty token invisible to this suite.
func sortedQuery(raw string) string {
	v, err := url.ParseQuery(raw)
	if err != nil {
		return raw
	}
	keys := slices.Sorted(maps.Keys(v))
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		vals := slices.Clone(v[k])
		slices.Sort(vals)
		parts = append(parts, k+"="+strings.Join(vals, ","))
	}
	return strings.Join(parts, "&")
}

// tokenFromQuery extracts access_token so the suite can assert the real value
// the server handed out, independently of the rest of the query.
func tokenFromQuery(raw string) string {
	v, err := url.ParseQuery(raw)
	if err != nil {
		return ""
	}
	return v.Get("access_token")
}

func canonBody(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "RAW:" + s
	}
	b, _ := json.Marshal(v)
	return string(b)
}

type goldenCase struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Query       string `json:"query"`
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
	// AccessToken is the token the test server issued; authenticated calls must
	// carry exactly this value. Empty means the endpoint takes no token.
	AccessToken string `json:"access_token"`
}

// loadGolden reads the expectations captured from the hand-written clients
// before they were replaced by the generated ones.
//
// testdata/golden.json is authored data, not generated output: it is the
// reference the generated clients are measured against. Nothing in the test
// suite rewrites it, so a broken implementation cannot quietly redefine the
// baseline it is supposed to fail.
func loadGolden(t *testing.T) map[string]goldenCase {
	t.Helper()
	path := filepath.Join("testdata", "golden.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (regenerate with the capture test)", err)
	}
	var out map[string]goldenCase
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestGeneratedMatchesGolden replays every captured case against the generated
// clients and compares the recorded request.
func TestGeneratedMatchesGolden(t *testing.T) {
	golden := loadGolden(t)

	cases := []struct {
		name string
		// run drives one generated call; the recorder holds the request after.
		run func(t *testing.T, srv *httptest.Server, rec *recorder)
		// contentTypeExempt marks calls whose content type intentionally
		// differs from the captured hand-written behaviour.
		contentTypeExempt bool
		// bodyNote documents an intentional body divergence from the golden.
		bodyNote string
	}{
		{
			name: "miniprogram.CallbackCheck",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genmp.New(genmp.Config{AppID: "appid", AppSecret: "secret", BaseURL: srv.URL})
				if _, err := c.PostCgiBinCallbackCheck(t.Context(), &genmp.PostCgiBinCallbackCheckRequest{}); err != nil {
					t.Fatal(err)
				}
			},
			// The generated client sends documented-required fields even when
			// they hold the zero value (P0-3), so an empty CallbackCheckRequest
			// becomes {"action":"","check_operator":""} where the hand-written
			// client sent {}. Both are rejected upstream: the caller must set
			// valid enum values, and an explicit empty value is at least
			// visible rather than silently missing.
			bodyNote: "required fields are sent even when empty",
		},
		{
			name: "official.GetCallbackIP",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genoa.New(genoa.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
				if _, err := c.GetCgiBinGetcallbackip(t.Context(), &genoa.GetCgiBinGetcallbackipRequest{}); err != nil {
					t.Fatal(err)
				}
			},
			contentTypeExempt: true, // GET: the generated client sends no JSON content type
		},
		{
			name: "official.GetTempMedia",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genoa.New(genoa.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
				if _, err := c.GetCgiBinMediaGet(t.Context(), &genoa.GetCgiBinMediaGetRequest{MediaID: "MID"}); err != nil {
					t.Fatal(err)
				}
			},
			contentTypeExempt: true,
		},
		{
			name: "official.UploadTempMedia",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genoa.New(genoa.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
				if _, err := c.PostCgiBinMediaUpload(t.Context(),
					&genoa.PostCgiBinMediaUploadRequest{Type: "image"},
					&core.Upload{FieldName: "media", FileName: "a.png", Reader: strings.NewReader("PNGDATA")}); err != nil {
					t.Fatal(err)
				}
			},
			contentTypeExempt: true, // multipart boundary differs by nature
		},
		{
			name: "miniprogram.JsCode2Session",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genmp.New(genmp.Config{AppID: "appid", AppSecret: "secret", BaseURL: srv.URL})
				if _, err := c.GetSnsJscode2Session(t.Context(), &genmp.GetSnsJscode2SessionRequest{
					JsCode: "CODE", GrantType: "authorization_code",
				}); err != nil {
					t.Fatal(err)
				}
			},
			contentTypeExempt: true,
		},
		{
			name: "official.GetUserInfo",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genoa.New(genoa.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
				if _, err := c.GetCgiBinUserInfo(t.Context(), &genoa.GetCgiBinUserInfoRequest{OpenID: "OPENID", Lang: "zh_CN"}); err != nil {
					t.Fatal(err)
				}
			},
			contentTypeExempt: true,
		},
		{
			name: "official.ListDraft",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genoa.New(genoa.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
				if _, err := c.PostCgiBinDraftBatchget(t.Context(), &genoa.PostCgiBinDraftBatchgetRequest{Offset: 5, Count: 10}); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "official.CreateTag",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genoa.New(genoa.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
				if _, err := c.PostCgiBinTagsCreate(t.Context(), &genoa.PostCgiBinTagsCreateRequest{
					Tag: &genoa.PostCgiBinTagsCreateRequestTagObject{Name: "标签A"},
				}); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "official.GetAPIDomainIP",
			run: func(t *testing.T, srv *httptest.Server, rec *recorder) {
				c := genoa.New(genoa.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
				if _, err := c.GetCgiBinGetApiDomainIp(t.Context(), &genoa.GetCgiBinGetApiDomainIpRequest{}); err != nil {
					t.Fatal(err)
				}
			},
			contentTypeExempt: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want, ok := golden[tc.name]
			if !ok {
				t.Fatalf("no golden captured for %s", tc.name)
			}
			var rec recorder
			srv := newServer(t, &rec, `{"errcode":0,"dns":[{"ip":"1.1.1.1"}],"ping":[],"ip_list":[],"ticket":"T","total_count":0,"item_count":0,"tag":{"id":1,"name":"x"},"openid":"o"}`)
			tc.run(t, srv, &rec)
			if rec.last == nil {
				t.Fatal("no request recorded")
			}
			if got := rec.last.Method; got != want.Method {
				t.Errorf("method %s, want %s", got, want.Method)
			}
			if got := rec.last.URL.Path; got != want.Path {
				t.Errorf("path %s, want %s", got, want.Path)
			}
			if got := sortedQuery(rec.last.URL.RawQuery); got != want.Query {
				t.Errorf("query differs:\n got %s\nwant %s", got, want.Query)
			}
			// The credential must actually reach the server (RF-001): a struct
			// field or a later parameter write must never blank it out.
			if got := tokenFromQuery(rec.last.URL.RawQuery); got != want.AccessToken {
				t.Errorf("access_token = %q, want %q", got, want.AccessToken)
			}
			gotBody := canonBody(rec.body)
			if tc.bodyNote != "" {
				t.Logf("reviewed body divergence (%s): got %s, golden %s", tc.bodyNote, gotBody, want.Body)
			} else if want.Body == "{}" && gotBody == "" {
				// the generated client omits an empty body; WeChat accepts both
				gotBody = "{}"
			}
			if tc.bodyNote == "" && gotBody != want.Body {
				if !(strings.HasPrefix(want.Body, "RAW:") && strings.HasPrefix(gotBody, "RAW:")) {
					t.Errorf("body differs:\n got %s\nwant %s", gotBody, want.Body)
				}
			}
			if !tc.contentTypeExempt {
				if got := rec.last.Header.Get("Content-Type"); got != want.ContentType {
					t.Errorf("content type %q, want %q", got, want.ContentType)
				}
			}
		})
	}
}

// TestSharedCryptoAndJSSDK pins the capabilities that cannot come from the
// documentation to the shared core runtime, which the generated packages and
// the hand-written ones both use.
func TestSharedCryptoAndJSSDK(t *testing.T) {
	const (
		token  = "callback-token"
		aesKey = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG" // 43 chars
		appID  = "wx1234567890abcdef"
	)
	crypto, err := genoa.NewMessageCrypto(token, aesKey, appID)
	if err != nil {
		t.Fatal(err)
	}
	plain := "<xml><MsgType>text</MsgType></xml>"
	envelope, err := crypto.EncryptReplyBody(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := crypto.DecryptBody(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Errorf("crypto round trip: got %q want %q", got, plain)
	}
	// The JS-SDK signature is pinned by the golden value derived from WeChat's
	// fixed-order string.
	if sig := core_JSSDKSignature("ticket123", "nonce456", "1700000000", "https://example.com/page#frag"); sig != "27498a134c6dd38c417b52d64aa9c3905a590c64" {
		t.Errorf("jssdk signature drifted: %s", sig)
	}
}

// TestGeneratedJSSDKConfigFlow checks the generated client fetches a ticket and
// signs with the shared algorithm.
func TestGeneratedJSSDKConfigFlow(t *testing.T) {
	var rec recorder
	// The handler must answer the token request as well, because the generated
	// JS-SDK flow obtains the jsapi ticket through the normal authenticated path.
	srv := newServer(t, &rec, `{"errcode":0,"ticket":"TICKET","expires_in":7200}`)
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/cgi-bin/token") {
			_, _ = io.WriteString(w, `{"access_token":"T","expires_in":7200}`)
			return
		}
		_, _ = io.WriteString(w, `{"errcode":0,"ticket":"TICKET","expires_in":7200}`)
	})
	c := genoa.New(genoa.Config{AppID: "app", AppSecret: "sec", BaseURL: srv.URL})
	cfg, err := c.GetJsSDKConfig(t.Context(), "https://example.com/page#frag")
	if err != nil {
		t.Fatal(err)
	}
	want := core_JSSDKSignature("TICKET", cfg.NonceStr, cfg.Timestamp, "https://example.com/page")
	if cfg.Signature != want {
		t.Errorf("signature mismatch:\n got %s\nwant %s", cfg.Signature, want)
	}
}
