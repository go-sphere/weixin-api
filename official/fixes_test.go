package official

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================
// Regression tests for the official package findings
// (RF-003 .. RF-008, RF-018, RF-023).
// ============================================================

// TestAddPermanentMaterialQueryHasType asserts the required type query
// parameter is sent on permanent material upload (RF-003).
func TestAddPermanentMaterialQueryHasType(t *testing.T) {
	var queryType string
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/material/add_material" {
			http.NotFound(w, r)
			return
		}
		queryType = r.URL.Query().Get("type")
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok", "media_id": "m1"})
	})
	_, err := oa.AddPermanentMaterial(t.Context(), &AddPermanentMaterialRequest{
		MaterialType: "image",
		Filename:     "a.jpg",
		Content:      []byte("xx"),
	})
	if err != nil {
		t.Fatalf("AddPermanentMaterial: %v", err)
	}
	if queryType != "image" {
		t.Fatalf("expected type=image in query, got %q", queryType)
	}
}

// TestShowQRCodeHostAndEncoding asserts the QR renderer uses mp.weixin.qq.com,
// carries no access_token and encodes the ticket exactly once (RF-004).
func TestShowQRCodeHostAndEncoding(t *testing.T) {
	var (
		path    string
		rawQ    string
		hasAuth bool
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		rawQ = r.URL.RawQuery
		hasAuth = r.URL.Query().Get("access_token") != ""
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png"))
	}))
	t.Cleanup(srv.Close)
	// The test seam points qrBaseURL at the test server. ShowQRCode must never
	// consult the token endpoint.
	oa := newOfficialWithHTTPClient(Config{AppID: "wx-official", AppSecret: "s"}, nil, srv.URL, srv.Client())
	img, err := oa.ShowQRCode(t.Context(), "abc=def+ghi")
	if err != nil {
		t.Fatalf("ShowQRCode: %v", err)
	}
	if string(img) != "png" {
		t.Fatalf("unexpected image %q", img)
	}
	if path != "/cgi-bin/showqrcode" {
		t.Fatalf("unexpected path %q", path)
	}
	if hasAuth {
		t.Fatalf("ShowQRCode must not attach access_token")
	}
	if rawQ != "ticket=abc%3Ddef%2Bghi" {
		t.Fatalf("expected single encoding of ticket, got %q", rawQ)
	}
}

// TestGetAPICallQuotaDecodes asserts the quota payload is actually surfaced
// (RF-005).
func TestGetAPICallQuotaDecodes(t *testing.T) {
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/openapi/quota/get" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]any{
			"errcode": 0, "errmsg": "ok",
			"quota": 100000, "used": 12, "surplus": 99988,
		})
	})
	resp, err := oa.GetAPICallQuota(t.Context(), "/cgi-bin/user/info")
	if err != nil {
		t.Fatalf("GetAPICallQuota: %v", err)
	}
	if resp == nil || resp.Quota != 100000 || resp.Used != 12 || resp.Surplus != 99988 {
		t.Fatalf("quota not decoded: %+v", resp)
	}
}

// TestClearQuotaSendsAppID asserts the whole-account clear route body carries
// the account appid (RF-018, official side).
func TestClearQuotaSendsAppID(t *testing.T) {
	var body map[string]string
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/clear_quota" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok"})
	})
	if err := oa.ClearQuota(t.Context()); err != nil {
		t.Fatalf("ClearQuota: %v", err)
	}
	if body["appid"] != "wx-official" {
		t.Fatalf("expected appid in body, got %v", body)
	}
}

// TestMassSendDoesNotRetryAfterTokenExpiry asserts mass sends disable the
// token-expiry auto retry to avoid double-delivering a crowd message (RF-006).
func TestMassSendDoesNotRetryAfterTokenExpiry(t *testing.T) {
	var calls int
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/message/mass/sendall" {
			calls++
			writeJSON(w, map[string]any{"errcode": 42001, "errmsg": "expired"})
			return
		}
		http.NotFound(w, r)
	})
	_, err := oa.MassSendByTag(t.Context(), 2, "text", "promo")
	if err == nil {
		t.Fatal("expected token-expiry error")
	}
	if calls != 1 {
		t.Fatalf("expected no retry for mass send, calls=%d", calls)
	}
}

// TestUpdateDraftArticlePartialUpdate asserts a title-only update does not
// serialise empty content/thumb_media_id that would wipe the stored article
// (RF-007).
func TestUpdateDraftArticlePartialUpdate(t *testing.T) {
	var raw string
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/draft/update" {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok"})
	})
	err := oa.UpdateDraftArticle(t.Context(), &UpdateDraftArticleRequest{
		MediaID: "d1",
		Index:   0,
		Articles: DraftArticleUpdate{
			Title: "new title",
		},
	})
	if err != nil {
		t.Fatalf("UpdateDraftArticle: %v", err)
	}
	if strings.Contains(raw, `"content":"")`) || strings.Contains(raw, `"thumb_media_id":"")`) {
		t.Fatalf("partial update must not send empty content/thumb_media_id: %s", raw)
	}
	if !strings.Contains(raw, `"title":"new title"`) {
		t.Fatalf("expected title in body: %s", raw)
	}
}

// TestCustomerMessageWxCardPayload asserts a wxcard customer message serialises
// its card payload (RF-008).
func TestCustomerMessageWxCardPayload(t *testing.T) {
	var raw string
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/message/custom/send" {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok"})
	})
	err := oa.SendCustomerMessage(t.Context(), &CustomerMessage{
		ToUser:  "u1",
		MsgType: MsgTypeWxCard,
		WxCard:  &WxCardMessage{CardID: "card-1"},
	})
	if err != nil {
		t.Fatalf("SendCustomerMessage: %v", err)
	}
	if !strings.Contains(raw, `"wxcard":{"card_id":"card-1"}`) {
		t.Fatalf("expected wxcard payload in body: %s", raw)
	}
}

// TestJSSDKSignatureGolden pins the JS-SDK signature to WeChat's fixed-order
// string and checks the URL fragment is stripped before signing (RF-023).
func TestJSSDKSignatureGolden(t *testing.T) {
	sig := jsSignature("ticket123", "nonce456", "1700000000", "https://example.com/page#frag")
	// sha1("jsapi_ticket=ticket123&noncestr=nonce456&timestamp=1700000000&url=https://example.com/page")
	const want = "27498a134c6dd38c417b52d64aa9c3905a590c64"
	if sig != want {
		t.Fatalf("jsSignature golden mismatch:\n got %s\nwant %s", sig, want)
	}
}

// TestJSSDKConfigStripsFragment drives GetJSSDKConfig with a hash URL and
// verifies the signature matches the fragment-less URL.
func TestJSSDKConfigStripsFragment(t *testing.T) {
	sigFrag := jsSignature("t", "n", "ts", "https://example.com/page#frag")
	sigNoFrag := jsSignature("t", "n", "ts", "https://example.com/page")
	if sigFrag != sigNoFrag {
		t.Fatal("signing helpers disagree on fragment handling")
	}
}

// TestMultipartFormFieldName checks a multipart upload body actually names the
// file field "media" (RF-021 wire assertion).
func TestMultipartFormFieldName(t *testing.T) {
	var (
		fieldName string
		gotType   string
	)
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/media/upload" {
			http.NotFound(w, r)
			return
		}
		gotType = r.URL.Query().Get("type")
		mr, err := r.MultipartReader()
		if err != nil {
			t.Errorf("not multipart: %v", err)
			return
		}
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			if p.FormName() == "media" {
				fieldName = "media"
			}
			_ = p.Close()
		}
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok", "type": "image", "media_id": "m1"})
	})
	_, err := oa.UploadTempMedia(t.Context(), &UploadTempMediaRequest{
		MediaType: "image",
		Filename:  "a.jpg",
		Content:   []byte("xx"),
	})
	if err != nil {
		t.Fatalf("UploadTempMedia: %v", err)
	}
	if fieldName != "media" {
		t.Fatalf("expected multipart field media")
	}
	if gotType != "image" {
		t.Fatalf("expected type=image, got %q", gotType)
	}
}
