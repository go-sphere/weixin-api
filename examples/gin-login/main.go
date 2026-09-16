// Command gin-login is a runnable Gin example that shows the two most common
// WeChat login flows implemented with the github.com/go-sphere/weixin-api
// library:
//
//   - Mini Program login (小程序登录):
//     wx.login() code -> JsCode2Session -> openid/session_key
//     Endpoint: POST /api/mp/login
//   - Official Account / Service Account web login (公众号/服务号网页授权):
//     OAuth2 authorize (snsapi_base / snsapi_userinfo) -> code -> openid (+ profile)
//     Endpoints: GET /api/oa/login, GET /api/oa/callback
//
// The module depends on the library through a local replace directive in
// go.mod:
//
//	replace github.com/go-sphere/weixin-api => ../..
//
// Credentials are read from environment variables (see README.md). Both
// clients are created with a nil cache, which means the library uses its
// built-in in-process memory cache for access tokens.
package main

import (
	"cmp"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Config bundles the settings of both WeChat accounts plus the local server
// options. In a real deployment these would come from a config file or a
// config center; environment variables keep the example dependency-free.
type Config struct {
	// Port the Gin server listens on.
	Port string

	// Mini Program (小程序) credentials.
	MPAppID     string
	MPAppSecret string

	// Official Account / Service Account (公众号/服务号) credentials.
	OAAppID     string
	OAAppSecret string

	// OACallbackURL is the full URL WeChat redirects to after the user
	// authorizes, e.g. "https://wx.example.com/api/oa/callback". The domain
	// must be configured as the 网页授权域名 in the official account console.
	OACallbackURL string
}

func loadConfig() Config {
	cfg := Config{
		Port:          cmp.Or(os.Getenv("PORT"), "8080"),
		MPAppID:       os.Getenv("MP_APP_ID"),
		MPAppSecret:   os.Getenv("MP_APP_SECRET"),
		OAAppID:       os.Getenv("OA_APP_ID"),
		OAAppSecret:   os.Getenv("OA_APP_SECRET"),
		OACallbackURL: os.Getenv("OA_REDIRECT_URI"),
	}
	if cfg.MPAppID == "" || cfg.MPAppSecret == "" {
		log.Println("warning: MP_APP_ID / MP_APP_SECRET not set — /api/mp/login will fail with a credential error")
	}
	if cfg.OAAppID == "" || cfg.OAAppSecret == "" {
		log.Println("warning: OA_APP_ID / OA_APP_SECRET not set — the official account flow will fail")
	}
	if cfg.OACallbackURL == "" {
		log.Println("warning: OA_REDIRECT_URI not set — /api/oa/login has no redirect target (see README.md)")
	}
	return cfg
}

func main() {
	cfg := loadConfig()

	store := NewStore()
	mpAPI := newMiniProgramAPI(cfg, store)
	oaAPI := newOfficialAccountAPI(cfg, store)

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, strings.Join([]string{
			"weixin-api x gin login demo",
			"",
			"Mini Program (小程序):",
			"  POST /api/mp/login   body: {\"code\": \"<wx.login() code>\"}",
			"  GET  /api/mp/me      header: Authorization: Bearer <token>",
			"",
			"Official Account (公众号/服务号):",
			"  GET  /api/oa/login?scope=snsapi_userinfo   (302 to WeChat, or &json=1 for the URL)",
			"  GET  /api/oa/callback?code=..&state=..      (WeChat redirect target)",
			"  GET  /api/oa/me      session cookie wx_session",
			"",
		}, "\n"))
	})

	mp := router.Group("/api/mp")
	{
		mp.POST("/login", mpAPI.login)
		mp.GET("/me", bearerOrCookieAuth(store), mpAPI.me)
	}

	oa := router.Group("/api/oa")
	{
		oa.GET("/login", oaAPI.login)
		oa.GET("/callback", oaAPI.callback)
		oa.GET("/me", bearerOrCookieAuth(store), oaAPI.me)
	}

	log.Printf("listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

// accountJSON is the JSON shape returned by the /me endpoints.
type accountJSON struct {
	Channel   Channel   `json:"channel"`
	OpenID    string    `json:"openid"`
	UnionID   string    `json:"unionid,omitempty"`
	Nickname  string    `json:"nickname,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func toAccountJSON(acc *Account) accountJSON {
	return accountJSON{
		Channel:   acc.Channel,
		OpenID:    acc.OpenID,
		UnionID:   acc.UnionID,
		Nickname:  acc.Nickname,
		Avatar:    acc.Avatar,
		CreatedAt: acc.CreatedAt,
	}
}
