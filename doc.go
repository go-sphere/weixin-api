// Package weixinapi is the root of the WeChat server API client suite. It
// contains no code; import one of the platform packages instead:
//
//   - github.com/go-sphere/weixin-api/miniprogram: WeChat Mini Program
//     (小程序) server APIs;
//   - github.com/go-sphere/weixin-api/official: Official Account
//     (公众号 / 服务号) server APIs;
//   - github.com/go-sphere/weixin-api/core: the shared transport, access-token,
//     error, message-crypto, JS-SDK and API-security runtime the platform
//     packages delegate to.
//
// The platform packages are generated from the official WeChat documentation.
// Every endpoint is a method on the platform Client that takes a context and a
// typed request struct and returns a typed response struct. Method names are
// derived from the HTTP verb and path: GET /sns/jscode2session becomes
// GetSnsJscode2Session.
//
// # Usage
//
// Build one client per account with New and the account's AppID and
// AppSecret. The client fetches the access_token from /cgi-bin/token on first
// use, caches it (in process by default), refreshes it after it expires, and
// retries a call once with a fresh token when WeChat reports an expired or
// invalid token. A Client is safe for concurrent use; share it.
//
//	import (
//		"context"
//		"errors"
//
//		"github.com/go-sphere/weixin-api/core"
//		"github.com/go-sphere/weixin-api/miniprogram"
//	)
//
//	client := miniprogram.New(miniprogram.Config{
//		AppID:     "wx-app-id",
//		AppSecret: "app-secret",
//	})
//	session, err := client.GetSnsJscode2Session(ctx, &miniprogram.GetSnsJscode2SessionRequest{
//		JsCode:    code, // from wx.login()
//		GrantType: "authorization_code",
//	})
//	if err != nil {
//		if apiErr, ok := errors.AsType[*core.APIError](err); ok {
//			// apiErr.ErrCode, apiErr.ErrMsg, apiErr.Description
//		}
//		return err
//	}
//	_ = session.OpenID
//
// The official package has the same shape: official.New(official.Config{...}).
//
// # Errors
//
// A response carrying a non-zero errcode is returned as an error, never as a
// response value. Access-token failures (errcode 40001, 40014, 42001) are the
// sentinel errors core.ErrorInvalidCredential, core.ErrorInvalidAccessToken
// and core.ErrorAccessTokenExpired, matched with errors.Is. Every other
// errcode is a *core.APIError carrying the code, message, rid, raw body and,
// when the endpoint documents the code, its description and solution. A
// non-200 HTTP status, a transport failure or a decode failure is a plain
// error.
//
// # Configuration alternatives
//
//   - Config.Cache: share access tokens across processes (for example a Redis
//     backed core.Cache) instead of the default in-process cache.
//   - Config.Token: supply tokens from elsewhere (a central token service or
//     a third-party platform); the client then neither caches nor refreshes
//     them and does not retry on token errors.
//   - Config.HTTPClient, Config.Proxy, Config.BaseURL: inject the transport,
//     an outbound proxy, or a different API host (tests, gateways).
//   - Config.Modifiers or Client.Use: install request decorators such as
//     core.APISecurity for API 二次加密和签名.
//   - Client.WithCredentials: derive a client for another account that shares
//     the cache and HTTP client.
package weixinapi
