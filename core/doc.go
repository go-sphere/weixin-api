// Package core is the shared runtime and foundation of the WeChat API clients
// in this module. The platform packages (miniprogram for Mini Programs,
// official for 公众号 / 服务号 official accounts) are generated code that
// delegates every request to this package, so all cross-platform behaviour
// lives here exactly once.
//
// It provides:
//
//   - the plain net/http transport (JSON bodies, form fields, raw bytes and
//     multipart uploads) with proxy and timeout configuration;
//   - access-token management: cache with singleflight de-duplication and a
//     transparent single retry after a token-expiry error;
//   - the universal {"errcode":..,"errmsg":..} envelope and its
//     classification — token problems collapse into sentinel errors that
//     errors.Is matches, every other code becomes an *APIError carrying the
//     code, message, rid and raw body;
//   - EndpointClient, the runtime behind the generated endpoint methods:
//     tag-driven request assembly, credential and token injection, request
//     signing (XPay pay_sig and user-level signatures), and documented
//     error-code enrichment;
//   - message-callback crypto (signature verification and AES-CBC
//     (de)encryption) and JS-SDK config signing, neither of which can be
//     derived from the HTTP documentation;
//   - the request/response decorator hook (RequestModifier, installed with
//     Client.Use) and the optional API 二次加密和签名 layer built on it
//     (APISecurity): AES-256-GCM body encryption with RSA-PSS request signing
//     and platform-certificate response verification, applied to the endpoint
//     paths the caller names;
//   - a pluggable Cache interface with an in-process memory implementation.
//
// Callers normally use a platform package rather than this one directly; core
// is imported for the shared types those packages expose (APIError, Cache,
// Credentials, MessageCrypto, MiniAppEnv, ...).
//
// # Usage
//
// Build a platform client, then use core to match errors and to configure
// optional layers:
//
//	import (
//		"errors"
//		"log"
//
//		"github.com/go-sphere/weixin-api/core"
//		"github.com/go-sphere/weixin-api/miniprogram"
//	)
//
//	client := miniprogram.New(miniprogram.Config{
//		AppID:     "wx-app-id",
//		AppSecret: "app-secret",
//		Cache:     myCache, // optional core.Cache shared across processes
//	})
//	resp, err := client.GetSnsJscode2Session(ctx, req)
//	switch {
//	case errors.Is(err, core.ErrorInvalidCredential):
//		// wrong AppID/AppSecret
//	case err != nil:
//		if apiErr, ok := errors.AsType[*core.APIError](err); ok {
//			log.Println(apiErr.ErrCode, apiErr.ErrMsg, apiErr.RID)
//		}
//	}
//
// Handle message-push callbacks with [NewMessageCrypto]:
//
//	mc, err := core.NewMessageCrypto(token, encodingAESKey, appID)
//	if err != nil {
//		return err
//	}
//	if !mc.VerifyBodySignature(msgSignature, timestamp, nonce, encrypt) {
//		return errors.New("bad signature")
//	}
//	body, err := mc.DecodeEncryptedPushXML(rawXML)
//
// Install API 二次加密和签名 with [NewAPISecurity] and the client's Use method
// (or the platform Config.Modifiers field) before the client serves requests.
//
// # Concurrency
//
// [Client], [EndpointClient], [MessageCrypto], [APISecurity] and the cache
// returned by [NewMemoryCache] are safe for concurrent use once constructed.
// [EndpointClient.Use] is the exception: call it before the client is shared.
// Custom [Cache] and [RequestModifier] implementations must be safe for
// concurrent use too.
package core
