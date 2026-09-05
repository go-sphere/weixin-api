package official

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sphere/weixin-api/core"
)

// newTestOfficial spins an httptest server with a token endpoint and returns
// the client plus the recorded handler.
func newTestOfficial(t *testing.T, handler http.HandlerFunc) *OfficialAccount {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			writeJSON(w, map[string]any{"access_token": "official-token", "expires_in": 7200})
			return
		}
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return newOfficialWithHTTPClient(Config{AppID: "wx-official", AppSecret: "secret"}, core.NewMemoryCache(), srv.URL, srv.Client())
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestOfficialGetUserInfo(t *testing.T) {
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/user/info" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("access_token") != "official-token" {
			t.Errorf("missing access token")
		}
		if r.URL.Query().Get("openid") != "o1" {
			t.Errorf("openid not passed")
		}
		writeJSON(w, map[string]any{
			"errcode": 0, "errmsg": "ok", "openid": "o1", "subscribe": 1,
			"nickname": "nick", "sex": 1, "city": "Shenzhen", "province": "Guangdong",
			"country": "CN", "language": "zh_CN", "subscribe_time": 123456,
			"remark": "", "groupid": 0, "tagid_list": []int{}, "subscribe_scene": "ADD_SCENE_QR_CODE",
		})
	})
	info, err := oa.GetUserInfo(t.Context(), &GetUserInfoRequest{OpenID: "o1"})
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("nil user info without error")
	}
	if info.OpenID != "o1" || info.Subscribe != 1 || info.Nickname != "nick" {
		t.Fatalf("unexpected user info %+v", info)
	}
}

func TestOfficialSendCustomerMessage(t *testing.T) {
	var body map[string]any
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/message/custom/send" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok"})
	})
	msg := &CustomerMessage{
		ToUser:  "o1",
		MsgType: MsgTypeText,
		Text:    &TextMessage{Content: "hello"},
	}
	if err := oa.SendCustomerMessage(t.Context(), msg); err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "o1" || body["msgtype"] != "text" {
		t.Fatalf("unexpected body %v", body)
	}
}

func TestOfficialJSSDKConfig(t *testing.T) {
	// The jsapi ticket endpoint must be served.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			writeJSON(w, map[string]any{"access_token": "t", "expires_in": 7200})
		case "/cgi-bin/ticket/getticket":
			writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok", "ticket": "jsapi-ticket", "expires_in": 7200})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	oa := newOfficialWithHTTPClient(Config{AppID: "wx-official", AppSecret: "s"}, core.NewMemoryCache(), srv.URL, srv.Client())
	cfg, err := oa.GetJSSDKConfig(t.Context(), "https://example.com/page")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppID != "wx-official" || cfg.Signature == "" || cfg.NonceStr == "" {
		t.Fatalf("unexpected js sdk config %+v", cfg)
	}
}

func TestOfficialMassMessage(t *testing.T) {
	var body map[string]any
	oa := newTestOfficial(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/message/mass/sendall" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		writeJSON(w, map[string]any{"errcode": 0, "errmsg": "ok", "msg_id": 100, "msg_data_id": 1})
	})
	resp, err := oa.MassSendByTag(t.Context(), 2, "text", "promo")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil mass response without error")
	}
	if resp.MsgID != 100 {
		t.Fatalf("unexpected msg id %+v", resp)
	}
	if body["msgtype"] != "text" {
		t.Fatalf("unexpected body %v", body)
	}
}
