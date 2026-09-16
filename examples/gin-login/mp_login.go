package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-sphere/weixin-api/miniprogram"
)

// miniProgramAPI groups the Gin handlers that implement the Mini Program login
// flow (小程序登录).
type miniProgramAPI struct {
	client *miniprogram.Client
	store  *Store
}

func newMiniProgramAPI(cfg Config, store *Store) *miniProgramAPI {
	return &miniProgramAPI{
		// Leaving Cache nil makes the client use its built-in in-process
		// memory cache for the access token; set Cache to a shared core.Cache
		// when the client runs in several processes.
		client: miniprogram.New(miniprogram.Config{
			AppID:     cfg.MPAppID,
			AppSecret: cfg.MPAppSecret,
		}),
		store: store,
	}
}

// loginRequest is the payload of POST /api/mp/login. The code is produced
// client-side by wx.login() and is single-use with a short lifetime.
type loginRequest struct {
	Code string `json:"code" binding:"required"`
}

// loginResponse is returned on a successful Mini Program login.
type loginResponse struct {
	Token string      `json:"token"`
	User  accountJSON `json:"user"`
}

// Mini program login flow (小程序登录流程):
//
//  1. The Mini Program calls wx.login() and posts the resulting code here.
//  2. JsCode2Session exchanges the code for the user's openid and session_key.
//  3. We mint our own session token and return it to the Mini Program; it is
//     then sent as "Authorization: Bearer <token>" on later requests.
//
// The session_key is the credential that decrypts data such as the user's phone
// number (getPhoneNumber). In a real backend you would keep it next to the
// user id and use it for the decrypt calls of this library.
func (a *miniProgramAPI) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid 'code'"})
		return
	}

	session, err := a.client.GetSnsJscode2Session(c.Request.Context(), &miniprogram.GetSnsJscode2SessionRequest{
		JsCode:    req.Code,
		GrantType: "authorization_code",
	})
	if err != nil {
		// err is a *miniprogram.APIError (or core.ErrorInvalidCredential when
		// the appid/secret pair is wrong); for a demo a plain 502 is enough.
		c.JSON(http.StatusBadGateway, gin.H{"error": "code2session failed", "detail": err.Error()})
		return
	}

	acc := a.store.Upsert(ChannelMiniProgram, session.OpenID, session.UnionID, "", "")
	token := a.store.IssueToken(acc)

	c.JSON(http.StatusOK, loginResponse{Token: token, User: toAccountJSON(acc)})
}

// me returns the account behind the bearer token, which demonstrates how the
// Mini Program authenticates subsequent API calls.
func (a *miniProgramAPI) me(c *gin.Context) {
	acc, _ := currentAccount(c)
	c.JSON(http.StatusOK, toAccountJSON(acc))
}
