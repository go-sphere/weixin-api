package weixinapi_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/go-sphere/weixin-api/core"
	"github.com/go-sphere/weixin-api/miniprogram"
	"github.com/go-sphere/weixin-api/official"
)

// The examples point BaseURL at a local httptest server that stands in for
// api.weixin.qq.com; production code leaves BaseURL empty.

// Example exchanges a wx.login() code for the user's openid with a Mini
// Program client. /sns/jscode2session authenticates with the AppID and
// AppSecret from Config, so no access token is fetched.
func Example() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		fmt.Println("appid:", q.Get("appid"), "js_code:", q.Get("js_code"))
		fmt.Fprint(w, `{"openid":"o-user","session_key":"sk"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := miniprogram.New(miniprogram.Config{
		AppID:     "wx-app-id",
		AppSecret: "app-secret",
		BaseURL:   srv.URL,
	})
	session, err := client.GetSnsJscode2Session(context.Background(), &miniprogram.GetSnsJscode2SessionRequest{
		JsCode:    "code-from-wx-login",
		GrantType: "authorization_code",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("openid:", session.OpenID)
	// Output:
	// appid: wx-app-id js_code: code-from-wx-login
	// openid: o-user
}

// Example_accessToken shows the automatic access_token handling of an
// Official Account client: the first call fetches the token from
// /cgi-bin/token, later calls reuse the cached token.
func Example_accessToken() {
	tokenFetches := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		tokenFetches++
		fmt.Fprint(w, `{"access_token":"TOKEN","expires_in":7200}`)
	})
	mux.HandleFunc("GET /cgi-bin/getcallbackip", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("access_token:", r.URL.Query().Get("access_token"))
		fmt.Fprint(w, `{"ip_list":["127.0.0.1"]}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := official.New(official.Config{
		AppID:     "wx-app-id",
		AppSecret: "app-secret",
		BaseURL:   srv.URL,
	})
	ctx := context.Background()
	for range 2 {
		resp, err := client.GetCgiBinGetcallbackip(ctx, &official.GetCgiBinGetcallbackipRequest{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("ips:", resp.IPList)
	}
	fmt.Println("token fetches:", tokenFetches)
	// Output:
	// access_token: TOKEN
	// ips: [127.0.0.1]
	// access_token: TOKEN
	// ips: [127.0.0.1]
	// token fetches: 1
}

// Example_apiError shows how a non-zero errcode surfaces: as a *core.APIError
// enriched with the endpoint's documented description, not as a response.
func Example_apiError() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"errcode":40029,"errmsg":"invalid code","rid":"rid-1"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := miniprogram.New(miniprogram.Config{
		AppID:     "wx-app-id",
		AppSecret: "app-secret",
		BaseURL:   srv.URL,
	})
	_, err := client.GetSnsJscode2Session(context.Background(), &miniprogram.GetSnsJscode2SessionRequest{
		JsCode:    "expired-code",
		GrantType: "authorization_code",
	})
	if apiErr, ok := errors.AsType[*core.APIError](err); ok {
		fmt.Println("errcode:", apiErr.ErrCode)
		fmt.Println("rid:", apiErr.RID)
		fmt.Println("documented:", apiErr.Description)
	}
	fmt.Println("token problem:", errors.Is(err, core.ErrorInvalidCredential))
	// Output:
	// errcode: 40029
	// rid: rid-1
	// documented: code 无效
	// token problem: false
}
