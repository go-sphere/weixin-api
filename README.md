# weixin-api

A typed, idiomatic Go client suite for the WeChat **server APIs**. Each WeChat
platform gets its own package built on a shared HTTP/token foundation:

| Package | Platform | Reference |
| --- | --- | --- |
| [`miniprogram`](./miniprogram) | 微信小程序 Mini Program | https://developers.weixin.qq.com/miniprogram/dev/server/API/ |
| [`official`](./official) | 公众号 / 服务号 Official Account | https://developers.weixin.qq.com/doc/subscription/api/ · https://developers.weixin.qq.com/doc/service/guide/ |
| [`core`](./core) | shared foundation (transport, tokens, errors, crypto) | — |

All packages are built with a plain `net/http` implementation — **no third
party HTTP library** — and every endpoint method carries typed request/response
structs plus a `Reference:` link straight to its official WeChat documentation
page (English Go doc throughout, AI/agent friendly).

## Packages

### core

Shared infrastructure extracted from the Mini Program client so every platform
package stays thin:

- `core.Client`: HTTP transport (JSON / form / raw / multipart upload), proxy
  and timeout configuration, per-account `access_token` & JS-SDK ticket caching
  with singleflight deduplication and transparent retry after token-expiry
  errors.
- `core.ErrResponse` / `core.APIError`: the universal `{"errcode":..,"errmsg":..}`
  envelope and classified errors; token failures map to sentinel errors
  (`core.ErrorInvalidCredential`, ...) for `errors.Is`.
- `core.Cache` + `core.NewMemoryCache()`: pluggable credential cache.
- `core.MessageCrypto`: message-push callback signature verification and
  AES-256-CBC payload (de)encryption.

### miniprogram

The Mini Program server API client (moved from the former single `wechat`
package, now one method per upstream endpoint across the whole catalogue:
login/session, subscribe messages, QR codes & URL links, content security,
customer service, datacube, express/intracity logistics, order shipping &
transaction guarantee, live broadcast, plugin & nearby POI, hardware/IoT,
XPay virtual payment, CloudBase, novel/short-drama, city service and more).

```go
client := miniprogram.NewMiniProgram(miniprogram.Config{
	AppID:     "wx1234567890abcdef",
	AppSecret: "0123...",
}, nil) // nil cache -> in-process memory cache

session, err := client.JsCode2Session(ctx, "wx-login-code")
```

### official

Official Account (公众号/服务号) APIs: user & tag management, blacklist,
customer-service messages, template messages, mass messages, media & permanent
materials, QR codes & short URLs, custom menus, draft box & publishing,
article/user/interface analysis, callback IP/quota and JS-SDK signing.

```go
oa := official.NewOfficialAccount(official.Config{
	AppID:     "wxb0000000000000",
	AppSecret: "abcd...",
}, nil)

info, err := oa.GetUserInfo(ctx, &official.GetUserInfoRequest{OpenID: "o1"})
```

## Conventions

- Methods that require an access token acquire (and cache) it automatically and
  transparently retry once after a token-expiry style error.
- JSON responses are checked for the universal error envelope; a non-zero code
  is surfaced as an error instead of silently returning partial data.
- Binary endpoints (QR codes, media files) return raw bytes and automatically
  surface an embedded JSON business error.
- Every exported method documents its parameters, response shape and a
  `Reference:` URL to the official page.

## Development

```sh
make fmt      # gofmt + goimports
make lint     # go vet + golangci-lint + nilaway
make test     # unit tests (httptest based, no network access)
make check    # everything CI runs
```

## License

MIT
