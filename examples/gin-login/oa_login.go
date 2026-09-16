package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-sphere/weixin-api/official"
)

// oauthEndpoint is the WeChat web authorization consent page. The user visits
// it (or is redirected to it) and, after authorizing, WeChat redirects back to
// the redirect_uri with a single-use code.
const oauthEndpoint = "https://open.weixin.qq.com/connect/oauth2/authorize"

// officialAccountAPI groups the Gin handlers that implement the Official
// Account / Service Account (公众号/服务号) web login flow.
type officialAccountAPI struct {
	client *official.Client
	store  *Store
	// callbackURL is the OACallbackURL from the config.
	callbackURL string
	// appID is kept because the sns/oauth2 endpoints take appid as a plain
	// query parameter rather than an access token.
	appID string
}

func newOfficialAccountAPI(cfg Config, store *Store) *officialAccountAPI {
	return &officialAccountAPI{
		client: official.New(official.Config{
			AppID:     cfg.OAAppID,
			AppSecret: cfg.OAAppSecret,
		}),
		store:       store,
		appID:       cfg.OAAppID,
		callbackURL: cfg.OACallbackURL,
	}
}

// login handles GET /api/oa/login. It redirects the browser to the WeChat
// consent page so the user can authorize the official account.
//
// Query parameters:
//   - scope: snsapi_base (default, silent, openid only) or snsapi_userinfo
//     (requires the user to consent, grants profile fields).
//   - json:  when set to 1 the authorize URL is returned as JSON instead of a
//     redirect (handy for non-browser clients or SPA pages).
//   - state: optional caller state, echoed back by WeChat on the callback.
//
// Reference: https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
func (a *officialAccountAPI) login(c *gin.Context) {
	if a.appID == "" || a.callbackURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "OA_APP_ID / OA_REDIRECT_URI not configured",
			"details": "set OA_APP_ID, OA_APP_SECRET and OA_REDIRECT_URI (see README.md)",
		})
		return
	}

	scope := c.DefaultQuery("scope", "snsapi_base")
	if scope != "snsapi_base" && scope != "snsapi_userinfo" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope must be snsapi_base or snsapi_userinfo"})
		return
	}

	// The state parameter round-trips through WeChat back to the callback; the
	// server stores it in a cookie and verifies the match to defeat CSRF. WeChat
	// itself does not validate it.
	state := c.Query("state")
	if state == "" {
		state = randomHex(8)
	}
	c.SetCookie("wx_oauth_state", state, 10*60, "/", "", false, true) // 10 minutes, HttpOnly

	u, err := url.Parse(oauthEndpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	q := url.Values{}
	q.Set("appid", a.appID)
	q.Set("redirect_uri", a.callbackURL)
	q.Set("response_type", "code")
	q.Set("scope", scope)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	authorizeURL := u.String() + "#wechat_redirect"

	if c.Query("json") == "1" {
		c.JSON(http.StatusOK, gin.H{"authorize_url": authorizeURL, "scope": scope})
		return
	}
	c.Redirect(http.StatusFound, authorizeURL)
}

// callback handles GET /api/oa/callback, the URL configured as the 网页授权回调
// domain. WeChat redirects here with the code after the user authorizes.
//
// The flow mirrors the official docs:
//
//	code -> sns/oauth2/access_token -> openid (+ access_token)
//	scope == snsapi_userinfo -> sns/userinfo -> nickname, avatar, ...
//	then the user is logged into the local session (wx_session cookie).
func (a *officialAccountAPI) callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'code' query parameter"})
		return
	}
	// Optional CSRF check: the state echoed by WeChat must match the cookie we
	// set in /api/oa/login.
	if expected, err := c.Cookie("wx_oauth_state"); err == nil {
		if actual := c.Query("state"); actual == "" || actual != expected {
			c.JSON(http.StatusBadRequest, gin.H{"error": "state mismatch"})
			return
		}
	}

	// Step 1: exchange the code for the user's openid.
	//
	// The generated method is typed and injects appid/secret from the client
	// configuration; a non-zero errcode (e.g. 40029 "invalid code") arrives as
	// an error just like every other endpoint.
	token, err := a.client.GetSnsOauth2AccessToken(c.Request.Context(), &official.GetSnsOauth2AccessTokenRequest{
		Code:      code,
		GrantType: "authorization_code",
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "oauth token exchange failed", "detail": err.Error()})
		return
	}

	// Step 2: when the user granted snsapi_userinfo, fetch the profile so the
	// demo account has a nickname/avatar to show.
	// The profile is only available when the user granted snsapi_userinfo; the
	// docs do not tabulate the returned scope, so the call is attempted and a
	// refusal is tolerated rather than predicted.
	nickname, avatar := "", ""
	if token.AccessToken != "" {
		// sns/userinfo carries the *user's* OAuth token, so it is passed on the
		// request rather than injected from the account configuration.
		info, err := a.client.GetSnsUserinfo(c.Request.Context(), &official.GetSnsUserinfoRequest{
			AccessToken: token.AccessToken,
			OpenID:      token.OpenID,
		})
		if err == nil {
			nickname, avatar = info.Nickname, info.Headimgurl
		}
	}

	// Step 3: create/refresh the local session and hand the browser its cookie.
	acc := a.store.Upsert(ChannelOfficialAccount, token.OpenID, token.UnionID, nickname, avatar)
	session := a.store.IssueToken(acc)
	c.SetCookie("wx_session", session, 7*24*3600, "/", "", false, true) // 7 days, HttpOnly

	// Browsers get redirected home; a page fetching this callback with json=1
	// (see /api/oa/login) receives the session details as JSON instead.
	if c.Query("json") == "1" {
		c.JSON(http.StatusOK, gin.H{"token": session, "user": toAccountJSON(acc)})
		return
	}
	target := c.Query("state")
	if target == "" || !strings.HasPrefix(target, "/") {
		target = "/"
	}
	c.Redirect(http.StatusFound, target)
}

// me returns the account of the current browser session.
func (a *officialAccountAPI) me(c *gin.Context) {
	acc, _ := currentAccount(c)
	c.JSON(http.StatusOK, toAccountJSON(acc))
}

// randomHex returns n random bytes hex-encoded.
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
