package miniprogram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

// ============================================================
// Regression tests for the review findings (RF-009 .. RF-022).
// ============================================================

// recordClient serves the access-token bootstrap (not recorded), delegates the
// real request to h, and records non-token requests for wire assertions.
func recordClient(t *testing.T, h http.HandlerFunc) (*MiniProgram, *[]http.Request) {
	t.Helper()
	var mu sync.Mutex
	var requests []http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			writeJSON(w, map[string]any{"access_token": "tok", "expires_in": 7200})
			return
		}
		mu.Lock()
		cp := *r
		cp.Body = r.Body
		requests = append(requests, cp)
		mu.Unlock()
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	w := newMiniProgramWithHTTPClient(Config{AppID: "appid1", AppSecret: "secret1"}, nil, srv.URL, srv.Client())
	return w, &requests
}

func simpleClient(t *testing.T, h http.HandlerFunc) *MiniProgram {
	w, _ := recordClient(t, h)
	return w
}

// TestGetUserEncryptKeyIsSignedGET asserts the contract of
// /wxa/business/getuserencryptkey: a GET whose query carries the HMAC signature
// and never the raw session_key (RF-011).
func TestGetUserEncryptKeyIsSignedGET(t *testing.T) {
	w, requests := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		writeJSON(rw, map[string]any{
			"errcode": 0, "errmsg": "ok",
			"key_info_list": []any{map[string]any{
				"encrypt_key": "abc", "version": 1, "expire_in": 3600,
				"iv": "abcdef", "create_time": 1000,
			}},
		})
	})
	req := &GetUserEncryptKeyRequest{OpenID: "o1", SessionKey: "sk-secret"}
	resp, err := w.GetUserEncryptKey(t.Context(), req)
	if err != nil {
		t.Fatalf("GetUserEncryptKey: %v", err)
	}
	got := (*requests)[0]
	if got.Method != http.MethodGet {
		t.Fatalf("expected GET, got %s", got.Method)
	}
	raw := got.URL.RawQuery
	if strings.Contains(raw, "sk-secret") {
		t.Fatalf("session_key leaked into query: %s", raw)
	}
	q, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	if q.Get("sig_method") != "hmac_sha256" || q.Get("signature") == "" || q.Get("openid") != "o1" {
		t.Fatalf("unexpected query: %v", q)
	}
	if len(resp.EncryptKeyInfos) != 1 || resp.EncryptKeyInfos[0].ExpireIn != 3600 {
		t.Fatalf("key info not decoded: %+v", resp.EncryptKeyInfos)
	}
}

// TestResetUserSessionKeyReturnsNewKey asserts the response's replacement
// session key is surfaced to the caller (RF-019).
func TestResetUserSessionKeyReturnsNewKey(t *testing.T) {
	w, _ := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/resetusersessionkey" {
			http.NotFound(rw, r)
			return
		}
		writeJSON(rw, map[string]any{
			"errcode": 0, "errmsg": "ok", "openid": "o1", "session_key": "new-key",
		})
	})
	resp, err := w.ResetUserSessionKey(t.Context(), "o1", "old-key")
	if err != nil {
		t.Fatalf("ResetUserSessionKey: %v", err)
	}
	if resp.SessionKey != "new-key" || resp.OpenID != "o1" {
		t.Fatalf("expected new session key, got %+v", resp)
	}
}

// TestGetPluginOpenPIDUsesCode asserts the correct helper sends the pluginLogin
// code as the request body (RF-012).
func TestGetPluginOpenPIDUsesCode(t *testing.T) {
	var body map[string]string
	w, _ := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/getpluginopenpid" {
			http.NotFound(rw, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		writeJSON(rw, map[string]any{"errcode": 0, "errmsg": "ok", "openpid": "openpid-1"})
	})
	resp, err := w.GetPluginOpenPID(t.Context(), "plugin-code-123")
	if err != nil {
		t.Fatalf("GetPluginOpenPID: %v", err)
	}
	if body["code"] != "plugin-code-123" {
		t.Fatalf("expected body code, got %v", body)
	}
	if resp.OpenPID != "openpid-1" {
		t.Fatalf("unexpected response %+v", resp)
	}
}

// TestUploadKfMediaTypeInQuery asserts type travels as a query parameter, not a
// form field (RF-014).
func TestUploadKfMediaTypeInQuery(t *testing.T) {
	w, requests := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		writeJSON(rw, map[string]any{"errcode": 0, "errmsg": "ok", "type": "image", "media_id": "m1"})
	})
	_, err := w.UploadKfMedia(t.Context(), "image", "a.jpg", []byte("xx"))
	if err != nil {
		t.Fatalf("UploadKfMedia: %v", err)
	}
	req := (*requests)[0]
	if got := req.URL.Query().Get("type"); got != "image" {
		t.Fatalf("expected type=image in query, got %q (query %s)", got, req.URL.RawQuery)
	}
	if ct := req.Header.Get("Content-Type"); !strings.Contains(ct, "multipart/form-data") {
		t.Fatalf("expected multipart upload, got %q", ct)
	}
}

// TestSecCheckSuggestionTag asserts the risk verdict decodes from the upstream
// "suggest" field (RF-016).
func TestSecCheckSuggestionTag(t *testing.T) {
	w, _ := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/msg_sec_check" {
			http.NotFound(rw, r)
			return
		}
		writeJSON(rw, map[string]any{
			"errcode": 0, "errmsg": "ok",
			"result": map[string]any{"suggest": "pass", "label": 100},
			"detail": []any{map[string]any{"suggest": "risky", "label": 20001}},
		})
	})
	resp, err := w.MsgSecCheck(t.Context(), &MsgSecCheckRequest{Content: "hi"})
	if err != nil {
		t.Fatalf("MsgSecCheck: %v", err)
	}
	if resp.Result.Suggestion != "pass" {
		t.Fatalf("expected Suggestion=pass, got %q", resp.Result.Suggestion)
	}
	if len(resp.Detail) == 0 || resp.Detail[0].Suggestion != "risky" {
		t.Fatalf("expected detail suggest decoded, got %+v", resp.Detail)
	}
}

// TestUserPortraitResponseDecode asserts the portrait object shape decodes
// (RF-009).
func TestUserPortraitResponseDecode(t *testing.T) {
	body := []byte(`{"errcode":0,"errmsg":"ok","ref_date":"2026-09-01","visit_uv_new":{"province":[{"id":1,"name":"x","value":3}],"city":[],"genders":[],"platforms":[],"devices":[],"ages":[]},"visit_uv":{"province":[],"city":[],"genders":[],"platforms":[],"devices":[],"ages":[]}}`)
	var resp GetUserPortraitResponse
	if err := decodeJSON(body, &resp); err != nil {
		t.Fatalf("decode portrait: %v", err)
	}
	if len(resp.VisitUVNew.Province) != 1 || resp.VisitUVNew.Province[0].Value != 3 {
		t.Fatalf("expected province bucket decoded, got %+v", resp.VisitUVNew.Province)
	}
}

// TestTemplateTitleCategoryIDString asserts the camelCase string categoryId
// decodes without failing the whole list (RF-017).
func TestTemplateTitleCategoryIDString(t *testing.T) {
	body := []byte(`{"errcode":0,"errmsg":"ok","count":1,"data":[{"tid":99,"title":"t","type":0,"categoryId":"616"}]}`)
	var resp GetTemplateTitleListResponse
	if err := decodeJSON(body, &resp); err != nil {
		t.Fatalf("decode template titles: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].CategoryID != "616" {
		t.Fatalf("expected categoryId decoded, got %+v", resp.Data)
	}
}

// TestLiveRoomGoodsObjects asserts room goods decode as objects (RF-010).
func TestLiveRoomGoodsObjects(t *testing.T) {
	body := []byte(`{"errcode":0,"errmsg":"ok","total":1,"room_info":[{"roomid":1,"name":"r","goods":[{"name":"g","cover_img":"c","url":"u","price":100,"price_type":1,"goods_id":77}]}]}`)
	var resp GetLiveRoomListResponse
	if err := decodeJSON(body, &resp); err != nil {
		t.Fatalf("decode live rooms: %v", err)
	}
	if len(resp.RoomInfos) != 1 || len(resp.RoomInfos[0].Goods) != 1 || resp.RoomInfos[0].Goods[0].GoodsID != 77 {
		t.Fatalf("expected goods objects, got %+v", resp.RoomInfos)
	}
}

// TestUniformMessageWeappDataShape asserts weapp_template_msg.data serialises
// as per-key {value} objects (RF-015).
func TestUniformMessageWeappDataShape(t *testing.T) {
	req := &SendUniformMessageRequest{
		ToUser: "u",
		WeappTemplateMsg: &MiniProgramTemplateMessage{
			TemplateID: "tpl",
			Data:       SubscribeMessageData{"keyword1": {Value: "v1"}},
		},
	}
	raw, err := jsonMarshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var probe map[string]any
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatal(err)
	}
	weapp, ok := probe["weapp_template_msg"].(map[string]any)
	if !ok {
		t.Fatalf("no weapp_template_msg: %s", raw)
	}
	data, ok := weapp["data"].(map[string]any)
	if !ok {
		t.Fatalf("no data: %s", raw)
	}
	kw, ok := data["keyword1"].(map[string]any)
	if !ok || kw["value"] != "v1" {
		t.Fatalf("expected keyword1:{value:...} object, got %s", raw)
	}
}

// TestClearQuotaLegacySendsAppID asserts the legacy clear route body carries the
// account appid (RF-018).
func TestClearQuotaLegacySendsAppID(t *testing.T) {
	var body map[string]string
	w, _ := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/clear_quota" {
			http.NotFound(rw, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		writeJSON(rw, map[string]any{"errcode": 0, "errmsg": "ok"})
	})
	if err := w.ClearQuotaLegacy(t.Context()); err != nil {
		t.Fatalf("ClearQuotaLegacy: %v", err)
	}
	if body["appid"] != "appid1" {
		t.Fatalf("expected appid in body, got %v", body)
	}
	if _, ok := body["cgi_path"]; ok {
		t.Fatalf("cgi_path must not be sent: %v", body)
	}
}

// TestB2bRequiresPaySig asserts the money-moving endpoints fail fast without a
// caller-supplied signature (RF-022).
func TestB2bRequiresPaySig(t *testing.T) {
	w, _ := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		writeJSON(rw, map[string]any{"errcode": 0, "errmsg": "ok"})
	})
	err := w.CreateProfitSharingOrder(t.Context(), &CreateProfitSharingOrderRequest{})
	if err == nil || !strings.Contains(err.Error(), "pay_sig") {
		t.Fatalf("expected missing-pay_sig error, got %v", err)
	}
	err = w.ManualWithdraw(t.Context(), &ManualWithdrawRequest{})
	if err == nil || !strings.Contains(err.Error(), "pay_sig") {
		t.Fatalf("expected missing-pay_sig error, got %v", err)
	}
}

// TestFmtInt64LargeValue asserts 64-bit values survive formatting on any word
// size (RF-020).
func TestFmtInt64LargeValue(t *testing.T) {
	if got := fmtInt64(1 << 40); got != "1099511627776" {
		t.Fatalf("expected full int64 string, got %q", got)
	}
}
